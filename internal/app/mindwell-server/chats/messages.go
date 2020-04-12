package chats

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/internal/app/mindwell-server/comments"
	"github.com/sevings/mindwell-server/internal/app/mindwell-server/users"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/chats"
	"github.com/sevings/mindwell-server/utils"
)

const loadMessagesQuery = `
    SELECT messages.id, extract(epoch from created_at), author_id,
        content, edit_content
    FROM messages
    WHERE chat_id = $1
`

func loadMessageList(srv *utils.MindwellServer, tx *utils.AutoTx, userID, chatID int64, reverse bool) *models.MessageList {
	var result models.MessageList

	for {
		msg := models.Message{
			Author: &models.User{},
			ChatID: chatID,
		}
		ok := tx.Scan(&msg.ID, &msg.CreatedAt, &msg.Author.ID,
			&msg.Content, &msg.EditContent)

		if !ok {
			break
		}

		if msg.Author.ID == userID {
			msg.Rights = &models.MessageRights{
				Delete: true,
				Edit:   true,
			}
		} else {
			msg.EditContent = ""
		}

		result.Data = append(result.Data, &msg)
	}

	talkers := make(map[int64]*models.User, 2)

	for _, msg := range result.Data {
		author, found := talkers[msg.Author.ID]
		if !found {
			author = users.LoadUserByID(srv, tx, msg.Author.ID)
			talkers[msg.Author.ID] = author
		}
		msg.Author = author
	}

	if reverse {
		list := result.Data
		for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
			list[i], list[j] = list[j], list[i]
		}
	}

	return &result
}

func newMessageListLoader(srv *utils.MindwellServer) func(chats.GetChatsNameMessagesParams, *models.UserID) middleware.Responder {
	return func(params chats.GetChatsNameMessagesParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			chatID, partnerID := findDialog(tx, userID.ID, params.Name)
			if partnerID == 0 {
				return chats.NewGetChatsNameMessagesNotFound()
			}
			if chatID == 0 {
				return chats.NewGetChatsNameMessagesOK()
			}

			var q = loadMessagesQuery

			before := utils.ParseInt64(*params.Before)
			after := utils.ParseInt64(*params.After)

			if after > 0 {
				q := q + " AND messages.id > $2 ORDER BY messages.id ASC LIMIT $3"
				tx.Query(q, chatID, after, *params.Limit)
			} else if before > 0 {
				q := q + " AND messages.id < $2 ORDER BY messages.id DESC LIMIT $3"
				tx.Query(q, chatID, before, *params.Limit)
			} else {
				q := q + " ORDER BY messages.id DESC LIMIT $2"
				tx.Query(q, chatID, *params.Limit)
			}

			list := loadMessageList(srv, tx, userID.ID, chatID, after <= 0)

			if len(list.Data) == 0 {
				return chats.NewGetChatsNameMessagesOK().WithPayload(list)
			}

			const beforeQuery = `SELECT EXISTS(
				SELECT 1 
				FROM messages
                WHERE chat_id = $1 AND id < $2)`

			nextBefore := list.Data[0].ID
			list.NextBefore = utils.FormatInt64(nextBefore)
			tx.Query(beforeQuery, chatID, nextBefore)
			tx.Scan(&list.HasBefore)

			const afterQuery = `SELECT EXISTS(
				SELECT 1 
				FROM messages
                WHERE chat_id = $1 AND id > $2)`

			nextAfter := list.Data[len(list.Data)-1].ID
			list.NextAfter = utils.FormatInt64(nextAfter)
			tx.Query(afterQuery, chatID, nextAfter)
			tx.Scan(&list.HasAfter)

			return chats.NewGetChatsNameMessagesOK().WithPayload(list)
		})
	}
}

func canSendMessage(tx *utils.AutoTx, userID, chatID int64) bool {
	const q = "SELECT can_send FROM talkers WHERE user_id = $1 AND chat_id = $2"
	return tx.QueryBool(q, userID, chatID)
}

const createMessageQuery = `
    INSERT INTO messages(chat_id, author_id, content, edit_content)
    VALUES($1, $2, $3, $4)
    RETURNING id, extract(epoch from created_at)
`

func createMessage(srv *utils.MindwellServer, tx *utils.AutoTx, userID, chatID int64, content string) *models.Message {
	msg := &models.Message{
		ChatID:      chatID,
		Author:      users.LoadUserByID(srv, tx, userID),
		Content:     comments.HtmlContent(content),
		EditContent: content,
		Rights: &models.MessageRights{
			Delete: true,
			Edit:   true,
		},
	}

	tx.Query(createMessageQuery, chatID, userID, msg.Content, msg.EditContent)
	tx.Scan(&msg.ID, &msg.CreatedAt)

	return msg
}

func newMessageCreator(srv *utils.MindwellServer) func(chats.PostChatsNameMessagesParams, *models.UserID) middleware.Responder {
	return func(params chats.PostChatsNameMessagesParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			msg := getCachedMessage(userID.ID, params.UID, params.Name)
			if msg != nil {
				return chats.NewPostChatsNameMessagesCreated().WithPayload(msg)
			}

			chatID, partnerID := findDialog(tx, userID.ID, params.Name)
			if partnerID == 0 {
				return chats.NewPostChatsNameMessagesNotFound()
			}

			if chatID == 0 {
				chatID = createChat(srv, tx, userID.ID, partnerID).ID
			}

			if !canSendMessage(tx, userID.ID, chatID) {
				return chats.NewPostChatsNameMessagesForbidden()
			}

			msg = createMessage(srv, tx, userID.ID, chatID, params.Content)
			setCachedMessage(userID.ID, params.UID, params.Name, msg)
			return chats.NewPostChatsNameMessagesCreated().WithPayload(msg)
		})
	}
}

const loadMessageQuery = `
    SELECT chat_id, extract(epoch from messages.created_at), 
        content, edit_content,
        users.id, users.name, users.show_name,
        is_online(users.last_seen_at), users.avatar
    FROM messages
    JOIN users ON users.id = messages.author_id
    WHERE messages.id = $1
`

func loadMessage(srv *utils.MindwellServer, tx *utils.AutoTx, userID, msgID int64) *models.Message {
	var avatar string
	msg := &models.Message{
		ID:     msgID,
		Author: &models.User{},
	}

	tx.Query(loadMessageQuery, msgID)
	tx.Scan(&msg.ChatID, &msg.CreatedAt,
		&msg.Content, &msg.EditContent,
		&msg.Author.ID, &msg.Author.Name, &msg.Author.ShowName,
		&msg.Author.IsOnline, &avatar)

	msg.Author.Avatar = srv.NewAvatar(avatar)

	if msg.Author.ID == userID {
		msg.Rights = &models.MessageRights{
			Delete: true,
			Edit:   true,
		}
	} else {
		msg.EditContent = ""
	}

	return msg
}

func canViewChat(tx *utils.AutoTx, userID, chatID int64) bool {
	const q = "SELECT true FROM talkers WHERE user_id = $1 AND chat_id = $2"
	return tx.QueryBool(q, userID, chatID)
}

func newMessageLoader(srv *utils.MindwellServer) func(chats.GetMessagesIDParams, *models.UserID) middleware.Responder {
	return func(params chats.GetMessagesIDParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			msg := loadMessage(srv, tx, userID.ID, params.ID)
			if msg.CreatedAt == 0 {
				return chats.NewGetMessagesIDNotFound()
			}

			if !canViewChat(tx, userID.ID, msg.ChatID) {
				return chats.NewGetMessagesIDForbidden()
			}

			return chats.NewGetMessagesIDOK().WithPayload(msg)
		})
	}
}

func canEditMessage(tx *utils.AutoTx, userID, msgID int64) bool {
	const q = "SELECT author_id = $1 FROM messages WHERE id = $2"
	return tx.QueryBool(q, userID, msgID)
}

const editMessageQuery = `
    UPDATE messages
    SET content = $2, edit_content = $3
    WHERE id = $1
    RETURNING chat_id, extract(epoch from created_at)
`

func editMessage(srv *utils.MindwellServer, tx *utils.AutoTx, userID, msgID int64, content string) *models.Message {
	msg := &models.Message{
		ID:          msgID,
		Author:      users.LoadUserByID(srv, tx, userID),
		Content:     comments.HtmlContent(content),
		EditContent: content,
		Rights: &models.MessageRights{
			Delete: true,
			Edit:   true,
		},
	}

	tx.Query(editMessageQuery, msgID, msg.Content, msg.EditContent)
	tx.Scan(&msg.ChatID, &msg.CreatedAt)

	return msg
}

func newMessageEditor(srv *utils.MindwellServer) func(chats.PutMessagesIDParams, *models.UserID) middleware.Responder {
	return func(params chats.PutMessagesIDParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			if !canEditMessage(tx, userID.ID, params.ID) {
				return chats.NewPutMessagesIDForbidden()
			}

			msg := editMessage(srv, tx, userID.ID, params.ID, params.Content)
			if msg.CreatedAt == 0 {
				return chats.NewPutMessagesIDNotFound()
			}

			return chats.NewPutMessagesIDOK().WithPayload(msg)
		})
	}
}

func deleteMessage(tx *utils.AutoTx, msgID int64) {
	const q = "DELETE FROM messages WHERE id = $1"
	tx.Exec(q, msgID)
}

func newMessageDeleter(srv *utils.MindwellServer) func(chats.DeleteMessagesIDParams, *models.UserID) middleware.Responder {
	return func(params chats.DeleteMessagesIDParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			if !canEditMessage(tx, userID.ID, params.ID) {
				return chats.NewDeleteMessagesIDForbidden()
			}

			deleteMessage(tx, params.ID)
			return chats.NewDeleteMessagesIDOK()
		})
	}
}
