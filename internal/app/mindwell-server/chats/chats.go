package chats

import (
	"database/sql"
	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/internal/app/mindwell-server/users"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/chats"
	"github.com/sevings/mindwell-server/utils"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	srv.API.ChatsGetChatsHandler = chats.GetChatsHandlerFunc(newChatsListLoader(srv))
	srv.API.ChatsGetChatsNameHandler = chats.GetChatsNameHandlerFunc(newChatLoader(srv))
	srv.API.ChatsPutChatsNameReadHandler = chats.PutChatsNameReadHandlerFunc(newChatReader(srv))

	srv.API.ChatsGetChatsNameMessagesHandler = chats.GetChatsNameMessagesHandlerFunc(newMessageListLoader(srv))
	srv.API.ChatsPostChatsNameMessagesHandler = chats.PostChatsNameMessagesHandlerFunc(newMessageCreator(srv))
	srv.API.ChatsGetMessagesIDHandler = chats.GetMessagesIDHandlerFunc(newMessageLoader(srv))
	srv.API.ChatsPutMessagesIDHandler = chats.PutMessagesIDHandlerFunc(newMessageEditor(srv))
	srv.API.ChatsDeleteMessagesIDHandler = chats.DeleteMessagesIDHandlerFunc(newMessageDeleter(srv))
}

const loadChatsQuery = `
    SELECT chats.id, talkers.unread_count, talkers.can_send,
        chats.creator_id, chats.partner_id,
        messages.id, extract(epoch from messages.created_at), messages.author_id,
        messages.content, messages.edit_content
    FROM talkers
        JOIN chats ON talkers.chat_id = chats.id 
        JOIN users AS partners ON chats.partner_id = partners.id
        JOIN messages ON chats.last_message = messages.id
    WHERE talkers.user_id = $1
`

const unreadChatsQuery = `
    SELECT count(*)
    FROM talkers
    WHERE user_id = $1 AND unread_count > 0
`

func loadChatList(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, reverse bool) *models.ChatList {
	var result models.ChatList

	for {
		chat := models.Chat{
			LastMessage: &models.Message{
				Author: &models.User{},
			},
			Rights: &models.ChatRights{},
		}
		var creatorID, partnerID int64
		ok := tx.Scan(&chat.ID, &chat.UnreadCount, &chat.Rights.Send,
			&creatorID, &partnerID,
			&chat.LastMessage.ID, &chat.LastMessage.CreatedAt, &chat.LastMessage.Author.ID,
			&chat.LastMessage.Content, &chat.LastMessage.EditContent)

		if !ok {
			break
		}

		chat.LastMessage.ChatID = chat.ID

		if creatorID == userID.ID {
			chat.Partner = &models.User{ID: partnerID}
		} else {
			chat.Partner = &models.User{ID: creatorID}
		}

		if chat.LastMessage.Author.ID == userID.ID {
			chat.LastMessage.Rights = &models.MessageRights{
				Delete: true,
				Edit:   true,
			}
		} else {
			chat.LastMessage.EditContent = ""
		}

		result.Data = append(result.Data, &chat)
	}

	var me *models.User

	for _, chat := range result.Data {
		chat.Partner = users.LoadUserByID(srv, tx, chat.Partner.ID)

		if chat.LastMessage.Author.ID == userID.ID {
			if me == nil {
				me = users.LoadUserByID(srv, tx, userID.ID)
			}
			chat.LastMessage.Author = me
		} else {
			chat.LastMessage.Author = chat.Partner
		}

		setMessageRead(tx, chat.LastMessage, userID.ID)
	}

	if reverse {
		list := result.Data
		for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
			list[i], list[j] = list[j], list[i]
		}
	}

	result.UnreadCount = tx.QueryInt64(unreadChatsQuery, userID.ID)

	return &result
}

func newChatsListLoader(srv *utils.MindwellServer) func(chats.GetChatsParams, *models.UserID) middleware.Responder {
	return func(params chats.GetChatsParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			var q = loadChatsQuery

			before := utils.ParseInt64(*params.Before)
			after := utils.ParseInt64(*params.After)

			if after > 0 {
				q := q + " AND last_message > $2 ORDER BY last_message ASC LIMIT $3"
				tx.Query(q, userID.ID, after, *params.Limit)
			} else if before > 0 {
				q := q + " AND last_message < $2 ORDER BY last_message DESC LIMIT $3"
				tx.Query(q, userID.ID, before, *params.Limit)
			} else {
				q := q + " ORDER BY last_message DESC LIMIT $2"
				tx.Query(q, userID.ID, *params.Limit)
			}

			list := loadChatList(srv, tx, userID, after > 0)

			if len(list.Data) == 0 {
				return chats.NewGetChatsOK().WithPayload(list)
			}

			nextBefore := list.Data[len(list.Data)-1].LastMessage.ID
			list.NextBefore = utils.FormatInt64(nextBefore)

			const beforeQuery = `SELECT EXISTS(
				SELECT 1 
				FROM talkers
                    JOIN chats ON talkers.chat_id = chats.id
				WHERE user_id = $1 AND last_message < $2)`

			tx.Query(beforeQuery, userID.ID, nextBefore)
			tx.Scan(&list.HasBefore)

			const afterQuery = `SELECT EXISTS(
				SELECT 1 
				FROM talkers
                    JOIN chats ON talkers.chat_id = chats.id
				WHERE user_id = $1 AND last_message > $2)`

			nextAfter := list.Data[0].LastMessage.ID
			list.NextAfter = utils.FormatInt64(nextAfter)
			tx.Query(afterQuery, userID.ID, nextAfter)
			tx.Scan(&list.HasAfter)

			return chats.NewGetChatsOK().WithPayload(list)
		})
	}
}

const loadChatQuery = `
    SELECT chats.id, talkers.unread_count, talkers.can_send,
        messages.id, extract(epoch from messages.created_at), messages.author_id,
        messages.content, messages.edit_content
    FROM chats
        JOIN talkers ON chats.id = talkers.chat_id AND talkers.user_id = $1
        LEFT JOIN messages ON chats.last_message = messages.id
    WHERE chats.id = $2
`

func loadChat(srv *utils.MindwellServer, tx *utils.AutoTx, userID, partnerID, chatID int64) *models.Chat {
	tx.Query(loadChatQuery, userID, chatID)

	chat := models.Chat{
		Rights: &models.ChatRights{},
	}

	var msgID, authorID sql.NullInt64
	var msgCreatedAt sql.NullFloat64
	var msgContent, msgEditContent sql.NullString

	ok := tx.Scan(&chat.ID, &chat.UnreadCount, &chat.Rights.Send,
		&msgID, &msgCreatedAt, &authorID,
		&msgContent, &msgEditContent)

	if !ok {
		return nil
	}

	chat.Partner = users.LoadUserByID(srv, tx, partnerID)

	if msgID.Valid {
		chat.LastMessage = &models.Message{
			Content:   msgContent.String,
			CreatedAt: msgCreatedAt.Float64,
			ChatID:    chat.ID,
			ID:        msgID.Int64,
		}

		if authorID.Int64 == userID {
			chat.LastMessage.EditContent = msgEditContent.String
			chat.LastMessage.Rights = &models.MessageRights{
				Delete: true,
				Edit:   true,
			}
		}

		if authorID.Int64 == chat.Partner.ID {
			chat.LastMessage.Author = chat.Partner
		} else {
			chat.LastMessage.Author = users.LoadUserByID(srv, tx, authorID.Int64)
		}

		setMessageRead(tx, chat.LastMessage, userID)
	}

	return &chat
}

const createChatQuery = `
    INSERT INTO chats(creator_id, partner_id)
    VALUES($1, $2)
    RETURNING id
`

const createTalkerQuery = `
    INSERT INTO talkers(chat_id, user_id)
    VALUES($1, $2)
`

func createChat(srv *utils.MindwellServer, tx *utils.AutoTx, userID, otherID int64) *models.Chat {
	chat := models.Chat{
		Rights: &models.ChatRights{
			Send: true,
		},
		Partner: users.LoadUserByID(srv, tx, otherID),
	}
	if chat.Partner.ID == 0 {
		return nil
	}

	var creatorID, partnerID int64
	if userID < otherID {
		creatorID = userID
		partnerID = otherID
	} else {
		creatorID = otherID
		partnerID = userID
	}
	chat.ID = tx.QueryInt64(createChatQuery, creatorID, partnerID)
	if chat.ID == 0 {
		return nil
	}

	tx.Exec(createTalkerQuery, chat.ID, creatorID)

	if creatorID != partnerID {
		tx.Exec(createTalkerQuery, chat.ID, partnerID)
	}

	return &chat
}

func newChatLoader(srv *utils.MindwellServer) func(chats.GetChatsNameParams, *models.UserID) middleware.Responder {
	return func(params chats.GetChatsNameParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			var chat *models.Chat

			chatID, partnerID := findDialog(tx, userID.ID, params.Name)
			if partnerID == 0 {
				err := srv.StandardError("no_chat")
				return chats.NewGetChatsNameNotFound().WithPayload(err)
			}

			if chatID == 0 {
				chat = createChat(srv, tx, userID.ID, partnerID)
			} else {
				chat = loadChat(srv, tx, userID.ID, partnerID, chatID)
			}

			if chat == nil {
				err := srv.StandardError("no_chat")
				return chats.NewGetChatsNameNotFound().WithPayload(err)
			}

			return chats.NewGetChatsNameOK().WithPayload(chat)
		})
	}
}

func loadReadMessages(tx *utils.AutoTx, chatID, userID, lastRead int64) []int64 {
	const q = `
		SELECT id
		FROM messages 
		WHERE chat_id = $2 AND author_id <> $1 AND id <= $3
			AND id > (SELECT last_read FROM talkers WHERE user_id = $1 AND chat_id = $2)
	`
	return tx.QueryInt64s(q, userID, chatID, lastRead)
}

const readChatQuery = `
    WITH cnt AS (
        SELECT count(*) AS unread
        FROM messages 
        WHERE chat_id = $2 AND id > $3 AND author_id <> $1
    )
    UPDATE talkers
    SET last_read = $3, unread_count = cnt.unread
    FROM cnt
    WHERE user_id = $1 AND chat_id = $2
    RETURNING cnt.unread
`

func newChatReader(srv *utils.MindwellServer) func(chats.PutChatsNameReadParams, *models.UserID) middleware.Responder {
	return func(params chats.PutChatsNameReadParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			chatID, partnerID := findDialog(tx, userID.ID, params.Name)
			if partnerID == 0 {
				err := srv.StandardError("no_chat")
				return chats.NewPutChatsNameReadNotFound().WithPayload(err)
			}
			if chatID == 0 {
				return chats.NewPutChatsNameReadOK()
			}

			name := findPartner(tx, chatID, userID.ID)
			msgIDs := loadReadMessages(tx, chatID, userID.ID, params.Message)
			for _, msgID := range msgIDs {
				srv.Ntf.Ntf.NotifyMessageRead(chatID, msgID, name)
			}

			var result chats.PutChatsNameReadOKBody
			result.Unread = tx.QueryInt64(readChatQuery, userID.ID, chatID, params.Message)

			return chats.NewPutChatsNameReadOK().WithPayload(&result)
		})
	}
}
