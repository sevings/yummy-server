package utils

import (
	"context"
	"encoding/json"
	"log"

	"github.com/centrifugal/gocent"
)

type message struct {
	ID    int64  `json:"id"`
	Subj  int64  `json:"subject,omitempty"`
	Type  string `json:"type,omitempty"`
	State string `json:"state,omitempty"`
	ch    string
}

type Notifier struct {
	cent *gocent.Client
	ch   chan *message
	stop chan interface{}
}

func NewNotifier(apiURL, apiKey string) *Notifier {
	if len(apiKey) == 0 {
		return &Notifier{}
	}

	cfg := gocent.Config{
		Addr: apiURL,
		Key:  apiKey,
	}

	ntf := &Notifier{
		cent: gocent.New(cfg),
		ch:   make(chan *message, 200),
		stop: make(chan interface{}),
	}

	go func() {
		ctx := context.Background()

		for msg := range ntf.ch {
			data, err := json.Marshal(msg)
			if err != nil {
				log.Println(err)
				continue
			}

			err = ntf.cent.Publish(ctx, msg.ch, data)
			if err != nil {
				log.Println(err)
			}
		}

		close(ntf.stop)
	}()

	return ntf
}

func (ntf *Notifier) Stop() {
	close(ntf.ch)
	<-ntf.stop
}

func notificationsChannel(userName string) string {
	return "notifications#" + userName
}

func (ntf *Notifier) Notify(tx *AutoTx, subjectID int64, tpe, user string) {
	const q = `
		INSERT INTO notifications(user_id, subject_id, type)
		VALUES((SELECT id from users WHERE lower(name) = lower($1)), 
			$2, (SELECT id FROM notification_type WHERE type = $3))
		RETURNING id
	`

	var id int64
	tx.Query(q, user, subjectID, tpe).Scan(&id)

	if ntf.ch != nil {
		ntf.ch <- &message{
			ID:    id,
			Subj:  subjectID,
			State: "new",
			Type:  tpe,
			ch:    notificationsChannel(user),
		}
	}
}

func (ntf *Notifier) NotifyUpdate(tx *AutoTx, subjectID int64, tpe string) {
	if ntf.ch == nil {
		return
	}

	const q = `
		SELECT notifications.id, name 
		FROM notifications, users
		WHERE subject_id = $1 AND type = (SELECT id FROM notification_type WHERE type = $2)
			AND user_id = users.id
	`

	tx.Query(q, subjectID, tpe)

	for {
		var id int64
		var user string
		ok := tx.Scan(&id, &user)
		if !ok {
			break
		}

		ntf.ch <- &message{
			ID:    id,
			State: "updated",
			ch:    notificationsChannel(user),
		}
	}
}

func (ntf *Notifier) NotifyRemove(tx *AutoTx, subjectID int64, tpe string) {
	const q = `
		DELETE FROM notifications 
		WHERE subject_id = $1 AND type = (SELECT id FROM notification_type WHERE type = $2)
		RETURNING id, (SELECT name FROM users WHERE id = user_id)
	`

	tx.Query(q, subjectID, tpe)

	for {
		var id int64
		var user string
		ok := tx.Scan(&id, &user)
		if !ok {
			break
		}

		if ntf.ch == nil {
			continue
		}

		ntf.ch <- &message{
			ID:    id,
			State: "removed",
			ch:    notificationsChannel(user),
		}
	}
}

func (ntf *Notifier) NotifyRead(user string, ntfID int64) {
	if ntf.ch == nil {
		return
	}

	ntf.ch <- &message{
		ID:    ntfID,
		State: "read",
		ch:    notificationsChannel(user),
	}
}

func messagesChannel(userName string) string {
	return "messages#" + userName
}

func (ntf *Notifier) NotifyMessage(chatID, msgID int64, user string) {
	if ntf.ch != nil {
		ntf.ch <- &message{
			ID:    chatID,
			Subj:  msgID,
			State: "new",
			Type:  "message",
			ch:    messagesChannel(user),
		}
	}
}

func (ntf *Notifier) NotifyMessageUpdate(chatID, msgID int64, user string) {
	if ntf.ch != nil {
		ntf.ch <- &message{
			ID:    chatID,
			Subj:  msgID,
			State: "updated",
			Type:  "message",
			ch:    messagesChannel(user),
		}
	}
}

func (ntf *Notifier) NotifyMessageRemove(chatID, msgID int64, user string) {
	if ntf.ch != nil {
		ntf.ch <- &message{
			ID:    chatID,
			Subj:  msgID,
			State: "removed",
			Type:  "message",
			ch:    messagesChannel(user),
		}
	}
}

func (ntf *Notifier) NotifyMessageRead(chatID, msgID int64, user string) {
	if ntf.ch != nil {
		ntf.ch <- &message{
			ID:    chatID,
			Subj:  msgID,
			State: "read",
			Type:  "message",
			ch:    messagesChannel(user),
		}
	}
}
