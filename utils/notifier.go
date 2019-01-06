package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/centrifugal/gocent"
)

type message struct {
	ID   int64  `json:"id"`
	Subj int64  `json:"subject,omitempty"`
	Type string `json:"type,omitempty"`
	Read bool   `json:"read,omitempty"`
	ch   string
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
			ID:   id,
			Subj: subjectID,
			Type: tpe,
			ch:   notificationsChannel(userID),
		}
	}
}

func (ntf *Notifier) NotifyRead(userID, ntfID int64) {
	if ntf.ch == nil {
		return
	}

	ntf.ch <- &message{
		ID:   ntfID,
		Read: true,
		ch:   notificationsChannel(userID),
	}
}
