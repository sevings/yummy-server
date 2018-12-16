package utils

type Notifier struct {
}

func (ntf *Notifier) Notify(tx *AutoTx, userID, subjectID int64, tpe string) {
	const q = `
		INSERT INTO notifications(user_id, subject_id, type)
		VALUES($1, $2, (SELECT id FROM notification_type WHERE type = $3))
	`

	tx.Exec(q, userID, subjectID, tpe)
}
