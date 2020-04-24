package test

import (
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/chats"
	"github.com/sevings/mindwell-server/utils"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func compareMessages(t *testing.T, exp, act *models.Message, user *models.UserID) {
	req := require.New(t)

	req.Equal(exp.ID, act.ID)
	req.Equal(exp.ChatID, act.ChatID)
	req.Equal(exp.Content, act.Content)
	req.Equal(exp.CreatedAt, act.CreatedAt)
	req.Equal(exp.Read, act.Read)

	if exp.Author.ID == user.ID {
		req.NotEmpty(act.EditContent)
	} else {
		req.Empty(act.EditContent)
	}

	rights := act.Rights
	req.Equal(act.Author.ID == user.ID, rights != nil && rights.Edit)
	req.Equal(act.Author.ID == user.ID, rights != nil && rights.Delete)
}

func checkLoadMessage(t *testing.T, userID *models.UserID, msg *models.Message, success bool) {
	load := api.ChatsGetMessagesIDHandler.Handle
	params := chats.GetMessagesIDParams{ID: msg.ID}
	resp := load(params, userID)
	body, ok := resp.(*chats.GetMessagesIDOK)
	require.Equal(t, success, ok)
	if !ok {
		return
	}

	compareMessages(t, msg, body.Payload, userID)
}

func checkSendMessage(t *testing.T, userID *models.UserID, otherName string, uid int64, success bool) *models.Message {
	send := api.ChatsPostChatsNameMessagesHandler.Handle
	params := chats.PostChatsNameMessagesParams{
		Content: "test message",
		Name:    otherName,
		UID:     uid,
	}
	resp := send(params, userID)
	body, ok := resp.(*chats.PostChatsNameMessagesCreated)
	require.Equal(t, success, ok)
	if !ok {
		return nil
	}

	msg := body.Payload
	require.NotZero(t, msg.ID)
	require.NotZero(t, msg.ChatID)
	require.Equal(t, msg.EditContent, params.Content)
	require.Equal(t, msg.Content, "<p>"+params.Content+"</p>")
	require.Equal(t, msg.Author.ID, userID.ID)
	require.Equal(t, userID.Name == otherName, msg.Read)
	require.True(t, msg.Rights.Edit)
	require.True(t, msg.Rights.Delete)

	checkLoadMessage(t, userID, msg, true)

	return msg
}

func checkEditMessage(t *testing.T, userID *models.UserID, msg *models.Message, success bool) {
	edit := api.ChatsPutMessagesIDHandler.Handle
	params := chats.PutMessagesIDParams{
		ID:      msg.ID,
		Content: "edit msg",
	}
	resp := edit(params, userID)
	body, ok := resp.(*chats.PutMessagesIDOK)
	require.Equal(t, success, ok)
	if !ok {
		return
	}

	msg.Content = "<p>" + params.Content + "</p>"
	msg.EditContent = params.Content

	require.Equal(t, *msg, *body.Payload)

	checkLoadMessage(t, userID, msg, true)
}

func checkDeleteMessage(t *testing.T, userID *models.UserID, msg *models.Message, success bool) {
	del := api.ChatsDeleteMessagesIDHandler.Handle
	params := chats.DeleteMessagesIDParams{
		ID: msg.ID,
	}
	resp := del(params, userID)
	_, ok := resp.(*chats.DeleteMessagesIDOK)
	require.Equal(t, success, ok)
	if !ok {
		return
	}

	checkLoadMessage(t, userID, msg, false)
}

func TestSendMessage(t *testing.T) {
	msg1 := checkSendMessage(t, userIDs[0], userIDs[1].Name, rand.Int63(), true)
	msg2 := checkSendMessage(t, userIDs[0], userIDs[1].Name, rand.Int63(), true)
	msg3 := checkSendMessage(t, userIDs[1], userIDs[0].Name, rand.Int63(), true)
	msg4 := checkSendMessage(t, userIDs[0], userIDs[0].Name, rand.Int63(), true)

	checkLoadMessage(t, userIDs[2], msg1, false)

	checkDeleteMessage(t, userIDs[0], msg1, true)
	checkDeleteMessage(t, userIDs[0], msg2, true)
	checkDeleteMessage(t, userIDs[1], msg3, true)
	checkDeleteMessage(t, userIDs[0], msg4, true)

	checkSendMessage(t, userIDs[0], "unknown name", rand.Int63(), false)
}

func TestDoubleMessage(t *testing.T) {
	msg1 := checkSendMessage(t, userIDs[0], userIDs[1].Name, 999, true)
	msg2 := checkSendMessage(t, userIDs[0], userIDs[1].Name, 999, true)
	require.Equal(t, msg1, msg2)

	msg3 := checkSendMessage(t, userIDs[1], userIDs[0].Name, 999, true)
	require.NotEqual(t, msg1, msg3)

	msg4 := checkSendMessage(t, userIDs[0], userIDs[2].Name, 999, true)
	require.NotEqual(t, msg1, msg4)

	checkDeleteMessage(t, userIDs[0], msg2, true)
	checkDeleteMessage(t, userIDs[1], msg3, true)
	checkDeleteMessage(t, userIDs[0], msg4, true)
}

func TestMessageRights(t *testing.T) {
	msg := checkSendMessage(t, userIDs[0], userIDs[1].Name, rand.Int63(), true)
	checkEditMessage(t, userIDs[0], msg, true)
	checkEditMessage(t, userIDs[1], msg, false)
	checkDeleteMessage(t, userIDs[1], msg, false)
	checkDeleteMessage(t, userIDs[0], msg, true)
}

func checkLoadMessages(t *testing.T, id *models.UserID, limit int64, name, before, after string, size int) *models.MessageList {
	params := chats.GetChatsNameMessagesParams{
		Limit:  &limit,
		Before: &before,
		After:  &after,
		Name:   name,
	}

	load := api.ChatsGetChatsNameMessagesHandler.Handle
	resp := load(params, id)
	body, ok := resp.(*chats.GetChatsNameMessagesOK)
	if !ok {
		t.Fatal("error load messages")
	}

	msgs := body.Payload

	if msgs == nil {
		require.Zero(t, size)
	} else {
		require.Equal(t, size, len(msgs.Data))
	}

	return msgs
}

func TestLoadMessages(t *testing.T) {
	m0 := checkSendMessage(t, userIDs[1], userIDs[0].Name, rand.Int63(), true)
	m1 := checkSendMessage(t, userIDs[1], userIDs[0].Name, rand.Int63(), true)
	m2 := checkSendMessage(t, userIDs[0], userIDs[1].Name, rand.Int63(), true)
	m3 := checkSendMessage(t, userIDs[0], userIDs[1].Name, rand.Int63(), true)

	msgs := checkLoadMessages(t, userIDs[0], 10, userIDs[1].Name, "", "", 4)
	compareMessages(t, m0, msgs.Data[0], userIDs[0])
	compareMessages(t, m1, msgs.Data[1], userIDs[0])
	compareMessages(t, m2, msgs.Data[2], userIDs[0])
	compareMessages(t, m3, msgs.Data[3], userIDs[0])

	req := require.New(t)
	req.False(msgs.HasAfter)
	req.False(msgs.HasBefore)

	msgs = checkLoadMessages(t, userIDs[0], 1, userIDs[1].Name, "", "", 1)
	compareMessages(t, m3, msgs.Data[0], userIDs[0])
	req.False(msgs.HasAfter)
	req.True(msgs.HasBefore)

	msgs = checkLoadMessages(t, userIDs[0], 2, userIDs[1].Name, msgs.NextBefore, "", 2)
	compareMessages(t, m1, msgs.Data[0], userIDs[0])
	compareMessages(t, m2, msgs.Data[1], userIDs[0])
	req.True(msgs.HasAfter)
	req.True(msgs.HasBefore)

	msgs = checkLoadMessages(t, userIDs[0], 1, userIDs[1].Name, msgs.NextBefore, "", 1)
	compareMessages(t, m0, msgs.Data[0], userIDs[0])
	req.True(msgs.HasAfter)
	req.False(msgs.HasBefore)

	msgs = checkLoadMessages(t, userIDs[0], 2, userIDs[1].Name, "", msgs.NextAfter, 2)
	compareMessages(t, m1, msgs.Data[0], userIDs[0])
	compareMessages(t, m2, msgs.Data[1], userIDs[0])
	req.True(msgs.HasAfter)
	req.True(msgs.HasBefore)

	msgs = checkLoadMessages(t, userIDs[0], 1, userIDs[1].Name, "", msgs.NextAfter, 1)
	compareMessages(t, m3, msgs.Data[0], userIDs[0])
	req.False(msgs.HasAfter)
	req.True(msgs.HasBefore)

	checkLoadMessages(t, userIDs[0], 1, userIDs[1].Name, "", msgs.NextAfter, 0)
	checkLoadMessages(t, userIDs[1], 10, userIDs[0].Name, "", "", 4)
	checkLoadMessages(t, userIDs[0], 10, userIDs[2].Name, "", "", 0)

	checkDeleteMessage(t, userIDs[1], m0, true)
	checkDeleteMessage(t, userIDs[1], m1, true)
	checkDeleteMessage(t, userIDs[0], m2, true)
	checkDeleteMessage(t, userIDs[0], m3, true)
}

func loadChat(t *testing.T, userID *models.UserID, otherName string) *models.Chat {
	load := api.ChatsGetChatsNameHandler.Handle
	params := chats.GetChatsNameParams{Name: otherName}
	resp := load(params, userID)
	body, ok := resp.(*chats.GetChatsNameOK)
	require.True(t, ok)

	return body.Payload
}

func checkLoadChats(t *testing.T, id *models.UserID, limit int64, before, after string, size int) *models.ChatList {
	params := chats.GetChatsParams{
		Limit:  &limit,
		Before: &before,
		After:  &after,
	}

	load := api.ChatsGetChatsHandler.Handle
	resp := load(params, id)
	body, ok := resp.(*chats.GetChatsOK)
	require.True(t, ok)
	if !ok {
		t.Fatal("error load chats")
	}

	list := body.Payload

	if list == nil {
		require.Zero(t, size)
	} else {
		require.Equal(t, size, len(list.Data))
	}

	return list
}

func compareChats(t *testing.T, exp, act *models.Chat, user *models.UserID, canSend bool) {
	req := require.New(t)

	req.Equal(exp.ID, act.ID)

	if exp.LastMessage == nil {
		req.Nil(act.LastMessage)
	} else {
		compareMessages(t, exp.LastMessage, act.LastMessage, user)
	}

	req.NotNil(act.Partner)
	if user.ID != act.Partner.ID {
		req.Equal(*exp.Partner, *act.Partner)
		req.Equal(exp.UnreadCount, act.UnreadCount)
	}

	rights := act.Rights
	req.Equal(canSend, rights != nil && rights.Send)
}

func TestLoadChats(t *testing.T) {
	req := require.New(t)

	list := checkLoadChats(t, userIDs[0], 10, "", "", 0)
	req.False(list.HasAfter)
	req.False(list.HasBefore)

	m3 := checkSendMessage(t, userIDs[2], userIDs[1].Name, rand.Int63(), true)
	m2 := checkSendMessage(t, userIDs[0], userIDs[2].Name, rand.Int63(), true)
	m1 := checkSendMessage(t, userIDs[1], userIDs[0].Name, rand.Int63(), true)
	m0 := checkSendMessage(t, userIDs[0], userIDs[0].Name, rand.Int63(), true)

	loadChat(t, userIDs[0], "Mindwell")
	c0 := loadChat(t, userIDs[0], userIDs[0].Name)
	c1 := loadChat(t, userIDs[0], userIDs[1].Name)
	c2 := loadChat(t, userIDs[0], userIDs[2].Name)

	list = checkLoadChats(t, userIDs[0], 10, "", "", 3)
	compareChats(t, c0, list.Data[0], userIDs[0], true)
	compareChats(t, c1, list.Data[1], userIDs[0], true)
	compareChats(t, c2, list.Data[2], userIDs[0], true)
	req.False(list.HasAfter)
	req.False(list.HasBefore)

	list = checkLoadChats(t, userIDs[0], 2, "", "", 2)
	compareChats(t, c0, list.Data[0], userIDs[0], true)
	compareChats(t, c1, list.Data[1], userIDs[0], true)
	req.False(list.HasAfter)
	req.True(list.HasBefore)

	list = checkLoadChats(t, userIDs[0], 10, list.NextBefore, "", 1)
	compareChats(t, c2, list.Data[0], userIDs[0], true)
	req.True(list.HasAfter)
	req.False(list.HasBefore)

	list = checkLoadChats(t, userIDs[0], 10, "", list.NextAfter, 2)
	compareChats(t, c0, list.Data[0], userIDs[0], true)
	compareChats(t, c1, list.Data[1], userIDs[0], true)
	req.False(list.HasAfter)
	req.True(list.HasBefore)

	checkLoadChats(t, userIDs[0], 10, "", list.NextAfter, 0)

	checkDeleteMessage(t, userIDs[0], m0, true)
	checkDeleteMessage(t, userIDs[1], m1, true)
	checkDeleteMessage(t, userIDs[0], m2, true)
	checkDeleteMessage(t, userIDs[2], m3, true)
}

func checkReadMessage(t *testing.T, user *models.UserID, otherName string, msg *models.Message, unread int64, success bool) {
	read := api.ChatsPutChatsNameReadHandler.Handle
	params := chats.PutChatsNameReadParams{
		Message: msg.ID,
		Name:    otherName,
	}
	resp := read(params, user)
	body, ok := resp.(*chats.PutChatsNameReadOK)
	require.Equal(t, success, ok)
	if !ok {
		return
	}

	require.Equal(t, unread, body.Payload.Unread)

	msg.Read = user.ID != msg.Author.ID
	checkLoadMessage(t, user, msg, true)
}

func TestReadMessages(t *testing.T) {
	req := require.New(t)

	m0 := checkSendMessage(t, userIDs[0], userIDs[1].Name, rand.Int63(), true)
	m1 := checkSendMessage(t, userIDs[0], userIDs[1].Name, rand.Int63(), true)

	c := loadChat(t, userIDs[0], userIDs[1].Name)
	req.Zero(c.UnreadCount)

	c = loadChat(t, userIDs[1], userIDs[0].Name)
	req.Equal(int64(2), c.UnreadCount)

	checkReadMessage(t, userIDs[0], userIDs[1].Name, m0, 0, true)
	checkReadMessage(t, userIDs[1], userIDs[0].Name, m0, 1, true)

	c = loadChat(t, userIDs[1], userIDs[0].Name)
	req.Equal(int64(1), c.UnreadCount)

	checkLoadMessage(t, userIDs[0], m1, true)
	checkReadMessage(t, userIDs[1], userIDs[0].Name, m1, 0, true)

	c = loadChat(t, userIDs[1], userIDs[0].Name)
	req.Zero(c.UnreadCount)

	checkDeleteMessage(t, userIDs[0], m0, true)
	checkDeleteMessage(t, userIDs[0], m1, true)
}

func TestCanSendMessage(t *testing.T) {
	checkFollow(t, userIDs[1], userIDs[0], profiles[0], models.RelationshipRelationIgnored, true)
	checkSendMessage(t, userIDs[0], userIDs[1].Name, rand.Int63(), false)
	checkSendMessage(t, userIDs[1], userIDs[0].Name, rand.Int63(), true)
	checkSendMessage(t, userIDs[0], userIDs[1].Name, rand.Int63(), false)
	checkUnfollow(t, userIDs[1], userIDs[0])
	checkSendMessage(t, userIDs[0], userIDs[1].Name, rand.Int63(), true)

	checkSendMessage(t, userIDs[3], userIDs[0].Name, rand.Int63(), false)
	_, err := db.Exec("UPDATE users SET invited_by = 1 WHERE id = $1", userIDs[3].ID)
	require.Nil(t, err)
	checkSendMessage(t, userIDs[3], userIDs[0].Name, rand.Int63(), true)

	user, _ := register("test_msg")
	checkSendMessage(t, userIDs[0], user.Name, rand.Int63(), true)
	checkSendMessage(t, user, userIDs[0].Name, rand.Int63(), true)
	checkSendMessage(t, user, userIDs[1].Name, rand.Int63(), false)
	checkSendMessage(t, userIDs[1], user.Name, rand.Int63(), true)
	checkSendMessage(t, user, userIDs[1].Name, rand.Int63(), true)

	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)
	esm.Clear()
}
