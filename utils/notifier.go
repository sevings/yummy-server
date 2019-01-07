package utils

import (
	"context"
	"encoding/json"
	"fmt"
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
	}()

	return ntf
}

func notificationsChannel(userID int64) string {
	return fmt.Sprintf("notifications#%d", userID)
}

func (ntf *Notifier) Notify(tx *AutoTx, userID, subjectID int64, tpe string) {
	const q = `
		INSERT INTO notifications(user_id, subject_id, type)
		VALUES($1, $2, (SELECT id FROM notification_type WHERE type = $3))
		RETURNING id
	`

	var id int64
	tx.Query(q, userID, subjectID, tpe).Scan(&id)

	if ntf.ch != nil {
		ntf.ch <- &message{
			ID:    id,
			Subj:  subjectID,
			State: "new",
			Type:  tpe,
			ch:    notificationsChannel(userID),
		}
	}
}

func (ntf *Notifier) NotifyUpdate(tx *AutoTx, subjectID int64, tpe string) {
	if ntf.ch == nil {
		return
	}

	const q = `
		SELECT id, user_id 
		FROM notifications 
		WHERE subject_id = $1 AND type = (SELECT id FROM notification_type WHERE type = $2)
	`

	tx.Query(q, subjectID, tpe)

	for {
		var id, userID int64
		ok := tx.Scan(&id, &userID)
		if !ok {
			break
		}

		ntf.ch <- &message{
			ID:    id,
			State: "updated",
			ch:    notificationsChannel(userID),
		}
	}
}

func (ntf *Notifier) NotifyRemove(tx *AutoTx, subjectID int64, tpe string) {
	const q = `
		DELETE FROM notifications 
		WHERE subject_id = $1 AND type = (SELECT id FROM notification_type WHERE type = $2)
		RETURNING id, user_id
	`

	tx.Query(q, subjectID, tpe)

	for {
		var id, userID int64
		ok := tx.Scan(&id, &userID)
		if !ok {
			break
		}

		if ntf.ch == nil {
			continue
		}

		ntf.ch <- &message{
			ID:    id,
			State: "removed",
			ch:    notificationsChannel(userID),
		}
	}
}

func (ntf *Notifier) NotifyRead(userID, ntfID int64) {
	if ntf.ch == nil {
		return
	}

	ntf.ch <- &message{
		ID:    ntfID,
		State: "read",
		ch:    notificationsChannel(userID),
	}
}
