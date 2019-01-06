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

func checkLoadSingleNotification(t *testing.T, userID *models.UserID, ntf *models.Notification, success bool) {
	params := notifications.GetNotificationsIDParams{
		ID: ntf.ID,
	}

	get := api.NotificationsGetNotificationsIDHandler.Handle
	resp := get(params, userID)
	body, ok := resp.(*notifications.GetNotificationsIDOK)

	req := require.New(t)
	req.Equal(success, ok)
	if !ok {
		return
	}

	loaded := body.Payload
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
	checkFollow(t, userIDs[1], userIDs[0], profiles[0], "followed")

	postComment(userIDs[0], e.ID)
	checkFollow(t, userIDs[1], userIDs[2], profiles[2], "followed")

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
}
