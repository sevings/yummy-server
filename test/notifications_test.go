package test

import (
	"testing"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/notifications"
	"github.com/sevings/mindwell-server/utils"
	"github.com/stretchr/testify/require"
)

func checkLoadNotifications(t *testing.T, id *models.UserID, limit int64, before, after string, unread bool, size int) *models.NotificationList {
	params := notifications.GetNotificationsParams{
		Limit:  &limit,
		Before: &before,
		After:  &after,
		Unread: &unread,
	}

	get := api.NotificationsGetNotificationsHandler.Handle
	resp := get(params, id)
	body, ok := resp.(*notifications.GetNotificationsOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Notifications))

	return list
}

func loadSingleNotification(userID *models.UserID, ntfID int64) *models.Notification {
	params := notifications.GetNotificationsIDParams{
		ID: ntfID,
	}

	get := api.NotificationsGetNotificationsIDHandler.Handle
	resp := get(params, userID)
	body, ok := resp.(*notifications.GetNotificationsIDOK)
	if !ok {
		return nil
	}

	return body.Payload
}

func checkLoadSingleNotification(t *testing.T, userID *models.UserID, ntf *models.Notification, success bool) {
	loaded := loadSingleNotification(userID, ntf.ID)

	req := require.New(t)
	req.Equal(success, loaded != nil)
	if !success {
		return
	}

	req.Equal(ntf.ID, loaded.ID)
	req.Equal(ntf.Type, loaded.Type)
	req.Equal(ntf.Read, loaded.Read)
}

func checkReadNotifications(t *testing.T, id *models.UserID, time float64, unread int64) {
	params := notifications.PutNotificationsReadParams{
		Time: &time,
	}

	put := api.NotificationsPutNotificationsReadHandler.Handle
	resp := put(params, id)
	body, ok := resp.(*notifications.PutNotificationsReadOK)

	require.True(t, ok)

	data := body.Payload
	require.Equal(t, unread, data.Unread)
}

func TestNotification(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)
	esm.Clear()

	e := postEntry(userIDs[0], "all", true)

	var ids = make([]int64, 3)
	ids[0] = postComment(userIDs[1], e.ID)
	ids[1] = postComment(userIDs[2], e.ID)

	ids[2] = userIDs[1].ID
	checkFollow(t, userIDs[1], userIDs[0], profiles[0], models.RelationshipRelationFollowed, true)

	postComment(userIDs[0], e.ID)
	checkFollow(t, userIDs[1], userIDs[2], profiles[2], models.RelationshipRelationFollowed, true)

	req := require.New(t)

	nots := checkLoadNotifications(t, userIDs[0], 2, "", "", false, 2)
	req.Equal(ids[2], nots.Notifications[0].User.ID)
	req.Equal(ids[1], nots.Notifications[1].Comment.ID)
	req.Equal(e.ID, nots.Notifications[1].Entry.ID)
	nots = checkLoadNotifications(t, userIDs[0], 20, nots.NextBefore, "", false, 1)
	req.Equal(ids[0], nots.Notifications[0].Comment.ID)
	req.Equal(e.ID, nots.Notifications[0].Entry.ID)

	checkLoadSingleNotification(t, userIDs[0], nots.Notifications[0], true)
	checkLoadSingleNotification(t, userIDs[1], nots.Notifications[0], false)

	req.True(nots.HasAfter)
	req.False(nots.HasBefore)

	checkLoadNotifications(t, userIDs[0], 20, "", nots.NextAfter, false, 2)

	checkReadNotifications(t, userIDs[0], nots.Notifications[0].CreatedAt, 2)
	checkLoadNotifications(t, userIDs[0], 20, "", "", true, 2)

	checkReadNotifications(t, userIDs[0], 0, 0)
	checkLoadNotifications(t, userIDs[0], 20, "", "", true, 0)

	nots = checkLoadNotifications(t, userIDs[2], 20, "", "", false, 2)
	checkReadNotifications(t, userIDs[2], nots.Notifications[0].CreatedAt, 0)
	checkFollow(t, userIDs[0], userIDs[2], profiles[2], models.RelationshipRelationIgnored, true)

	cID := postComment(userIDs[1], e.ID)
	nots = checkLoadNotifications(t, userIDs[0], 20, "", "", true, 1)
	checkLoadSingleNotification(t, userIDs[0], nots.Notifications[0], true)
	checkLoadNotifications(t, userIDs[2], 20, "", "", true, 0)

	checkDeleteComment(t, cID, userIDs[1], true)
	checkLoadNotifications(t, userIDs[0], 20, "", "", true, 0)
	checkLoadSingleNotification(t, userIDs[0], nots.Notifications[0], false)
}

func TestNotificationInfo(t *testing.T) {
	tx := utils.NewAutoTx(db)
	infoID := tx.QueryInt64(`
		INSERT INTO info(content, link) 
		VALUES('test content', 'test link')
		RETURNING id
	`)

	typeID := tx.QueryInt64("SELECT id FROM notification_type WHERE type = 'info'")
	ntfID := tx.QueryInt64(`
		INSERT INTO notifications(user_id, type, subject_id)
		VALUES($1, $2, $3)
		RETURNING id
	`, userIDs[0].ID, typeID, infoID)

	tx.Finish()

	ntf := loadSingleNotification(userIDs[0], ntfID)

	req := require.New(t)
	req.NotNil(ntf)
	req.Equal(ntfID, ntf.ID)
	req.Equal("info", ntf.Type)
	req.Equal("test content", ntf.Info.Content)
	req.Equal("test link", ntf.Info.Link)
}
