package notifications

import (
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/internal/app/mindwell-server/comments"
	"github.com/sevings/mindwell-server/internal/app/mindwell-server/entries"
	"github.com/sevings/mindwell-server/internal/app/mindwell-server/users"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/notifications"
	"github.com/sevings/mindwell-server/utils"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	srv.API.NotificationsPutNotificationsReadHandler = notifications.PutNotificationsReadHandlerFunc(newNotificationsReader(srv))
	srv.API.NotificationsGetNotificationsHandler = notifications.GetNotificationsHandlerFunc(newNotificationsLoader(srv))
}

func unreadCount(tx *utils.AutoTx, userID int64) int64 {
	var unread int64
	tx.Query("SELECT count(*) FROM notifications WHERE user_id = $1 AND NOT read", userID).Scan(&unread)

	return unread
}

func newNotificationsReader(srv *utils.MindwellServer) func(notifications.PutNotificationsReadParams, *models.UserID) middleware.Responder {
	return func(params notifications.PutNotificationsReadParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {

			q := `UPDATE notifications
				SET read = true
				WHERE user_id = $1`

			if *params.Time > 0 {
				q = q + " AND extract(epoch from created_at) <= $2"
				tx.Exec(q, uID.ID, *params.Time)
			} else {
				tx.Exec(q, uID.ID)
			}

			unread := unreadCount(tx, uID.ID)
			feed := notifications.PutNotificationsReadOKBody{Unread: unread}
			return notifications.NewPutNotificationsReadOK().WithPayload(&feed)
		})
	}
}

type notice struct {
	id   int64
	at   float64
	tpe  string
	subj int64
	read bool
}

func loadFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID int64, reverse bool) *models.NotificationList {
	var notices []notice
	for {
		var not notice
		ok := tx.Scan(&not.id, &not.at, &not.tpe, &not.subj, &not.read)
		if !ok {
			break
		}

		notices = append(notices, not)
	}

	feed := models.NotificationList{}
	feed.UnreadCount = unreadCount(tx, userID)

	for _, not := range notices {
		notif := models.Notification{
			ID:        not.id,
			CreatedAt: not.at,
			Read:      not.read,
			Type:      not.tpe,
		}

		switch not.tpe {
		case "comment":
			notif.Comment = comments.LoadComment(srv, tx, userID, not.subj)
			notif.Entry = entries.LoadEntry(srv, tx, notif.Comment.EntryID, userID)
			break
		case "follower":
			fallthrough
		case "request":
			fallthrough
		case "accept":
			notif.User = users.LoadUserByID(srv, tx, not.subj)
			break
		default:
			log.Println("Unknown notification type:", not.tpe)
		}

		feed.Notifications = append(feed.Notifications, &notif)
	}

	if reverse {
		list := feed.Notifications
		for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
			list[i], list[j] = list[j], list[i]
		}
	}

	return &feed
}

func newNotificationsLoader(srv *utils.MindwellServer) func(notifications.GetNotificationsParams, *models.UserID) middleware.Responder {
	return func(params notifications.GetNotificationsParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			var q = `SELECT notifications.id, extract(epoch from created_at), notification_type.type, subject_id, read
				FROM notifications, notification_type
				WHERE user_id = $1 AND notifications.type = notification_type.id
				`

			if *params.Unread {
				q = q + "AND NOT read "
			}

			before := utils.ParseFloat(*params.Before)
			after := utils.ParseFloat(*params.After)

			if after > 0 {
				q := q + " AND created_at > to_timestamp($2) ORDER BY created_at ASC LIMIT $3"
				tx.Query(q, userID.ID, after, *params.Limit)
			} else if before > 0 {
				q := q + " AND created_at < to_timestamp($2) ORDER BY created_at DESC LIMIT $3"
				tx.Query(q, userID.ID, before, *params.Limit)
			} else {
				q := q + " ORDER BY created_at DESC LIMIT $2"
				tx.Query(q, userID.ID, *params.Limit)
			}

			feed := loadFeed(srv, tx, userID.ID, after > 0)

			if len(feed.Notifications) == 0 {
				return notifications.NewGetNotificationsOK().WithPayload(feed)
			}

			nextBefore := feed.Notifications[len(feed.Notifications)-1].CreatedAt
			feed.NextBefore = utils.FormatFloat(nextBefore)

			const beforeQuery = `SELECT EXISTS(
				SELECT 1 
				FROM notifications
				WHERE user_id = $1 AND created_at < to_timestamp($2))`

			tx.Query(beforeQuery, userID.ID, nextBefore)
			tx.Scan(&feed.HasBefore)

			const afterQuery = `SELECT EXISTS(
				SELECT 1 
				FROM notifications
				WHERE user_id = $1 AND created_at > to_timestamp($2))`

			nextAfter := feed.Notifications[0].CreatedAt
			feed.NextAfter = utils.FormatFloat(nextAfter)
			tx.Query(afterQuery, userID.ID, nextAfter)
			tx.Scan(&feed.HasAfter)

			return notifications.NewGetNotificationsOK().WithPayload(feed)
		})
	}
}