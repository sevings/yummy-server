package utils

import (
	"database/sql"
	"log"

	"github.com/sevings/mindwell-server/models"
)

type CompositeNotifier struct {
	srv  *MindwellServer
	Mail MailSender
	Ntf  *Notifier
	Tg   *TelegramBot
}

func NewCompositeNotifier(srv *MindwellServer) *CompositeNotifier {
	ntfURL := srv.ConfigString("centrifugo.api_url")
	ntfKey := srv.ConfigString("centrifugo.api_key")

	return &CompositeNotifier{
		srv: srv,
		Ntf: NewNotifier(ntfURL, ntfKey),
		Tg:  NewTelegramBot(srv),
	}
}

func (ntf *CompositeNotifier) Stop() {
	ntf.Mail.Stop()
	ntf.Tg.Stop()
	ntf.Ntf.Stop()
}

func (ntf *CompositeNotifier) SendNewInvite(tx *AutoTx, userID int) {
	var email, name, showName string
	var sendEmail bool
	var tg sql.NullInt64

	tx.Query("SELECT email, name, show_name, verified AND email_invites, telegram FROM users WHERE id = $1", userID)
	tx.Scan(&email, &name, &showName, &sendEmail, &tg)

	ntf.Ntf.Notify(tx, 0, models.NotificationTypeInvite, name)

	if tg.Valid {
		ntf.Tg.SendNewInvite(tg.Int64)
	}

	if sendEmail {
		ntf.Mail.SendNewInvite(email, showName)
	}
}

func (ntf *CompositeNotifier) SendEmailChanged(tx *AutoTx, userID *models.UserID, oldEmail, newEmail string) {
	const q = `
		SELECT show_name, telegram
		FROM users
		WHERE id = $1
	`

	var name string
	var tg sql.NullInt64
	tx.Query(q, userID.ID).Scan(&name, &tg)

	if tx.Error() != nil {
		if tx.Error() != sql.ErrNoRows {
			log.Print(tx.Error())
		}
		return
	}

	if len(oldEmail) > 0 {
		ntf.Mail.SendEmailChanged(oldEmail, name)
	}

	code := ntf.srv.VerificationCode(newEmail)
	ntf.Mail.SendGreeting(newEmail, name, code)

	if tg.Valid {
		ntf.Tg.SendEmailChanged(tg.Int64)
	}
}

func (ntf *CompositeNotifier) SendPasswordChanged(tx *AutoTx, userID *models.UserID) {
	const q = `
		SELECT email, verified, show_name, telegram
		FROM users
		WHERE id = $1
	`

	var email, name string
	var verified bool
	var tg sql.NullInt64
	tx.Query(q, userID.ID).Scan(&email, &verified, &name, &tg)

	if tx.Error() != nil {
		if tx.Error() != sql.ErrNoRows {
			log.Print(tx.Error())
		}
		return
	}

	if verified {
		ntf.Mail.SendPasswordChanged(email, name)
	}

	if tg.Valid {
		ntf.Tg.SendPasswordChanged(tg.Int64)
	}
}

func (ntf *CompositeNotifier) SendGreeting(address, showName string) {
	code := ntf.srv.VerificationCode(address)
	ntf.Mail.SendGreeting(address, showName, code)
}

func (ntf *CompositeNotifier) SendResetPassword(email, showName, gender string) {
	code, date := ntf.srv.ResetPasswordCode(email)
	ntf.Mail.SendResetPassword(email, showName, gender, code, date)
}

func (ntf *CompositeNotifier) SendNewComment(tx *AutoTx, cmt *models.Comment) {
	const titleQ = "SELECT title FROM entries WHERE id = $1"
	var title string
	tx.Query(titleQ, cmt.EntryID).Scan(&title)

	title, _ = CutText(title, 80)

	const fromQ = `
		SELECT gender.type 
		FROM users, gender 
		WHERE users.id = $1 AND users.gender = gender.id
	`

	var fromGender string
	tx.Query(fromQ, cmt.Author.ID).Scan(&fromGender)

	const toQ = `
		SELECT users.name, show_name, email, verified AND email_comments, telegram
		FROM users 
		INNER JOIN watching ON watching.user_id = users.id 
		WHERE watching.entry_id = $1 AND users.id <> $2 
			AND can_view_entry(users.id, $1)`

	tx.Query(toQ, cmt.EntryID, cmt.Author.ID)

	var toNames []string
	var toName string
	var sendEmail bool
	var toShowName, email string
	var tg sql.NullInt64
	for tx.Scan(&toName, &toShowName, &email, &sendEmail, &tg) {
		if sendEmail {
			ntf.Mail.SendNewComment(email, fromGender, toShowName, title, cmt)
		}

		if tg.Valid {
			ntf.Tg.SendNewComment(tg.Int64, title, cmt)
		}

		toNames = append(toNames, toName)
	}

	for _, name := range toNames {
		ntf.Ntf.Notify(tx, cmt.ID, models.NotificationTypeComment, name)
	}
}

func (ntf *CompositeNotifier) SendUpdateComment(tx *AutoTx, commentID int64) {
	ntf.Ntf.NotifyUpdate(tx, commentID, models.NotificationTypeComment)
}

func (ntf *CompositeNotifier) SendRemoveComment(tx *AutoTx, commentID int64) {
	ntf.Ntf.NotifyRemove(tx, commentID, models.NotificationTypeComment)
}

func (ntf *CompositeNotifier) SendRead(name string, id int64) {
	ntf.Ntf.NotifyRead(name, id)
}

func (ntf *CompositeNotifier) SendInvited(tx *AutoTx, from, to string) {
	const toQ = `
		SELECT show_name, email, verified, telegram
		FROM users 
		WHERE lower(name) = lower($1)
	`

	var sendEmail bool
	var toShowName, email string
	var tg sql.NullInt64
	tx.Query(toQ, to).Scan(&toShowName, &email, &sendEmail, &tg)

	const fromQ = `
		SELECT users.id, show_name, gender.type 
		FROM users, gender 
		WHERE lower(users.name) = lower($1) AND users.gender = gender.id`

	var fromID int64
	var fromShowName, fromGender string
	tx.Query(fromQ, from).Scan(&fromID, &fromShowName, &fromGender)

	if sendEmail {
		ntf.Mail.SendInvited(email, fromShowName, fromGender, toShowName)
	}

	if tg.Valid {
		ntf.Tg.SendInvited(tg.Int64, from, fromShowName, fromGender)
	}

	ntf.Ntf.Notify(tx, fromID, models.NotificationTypeInvited, to)
}

func (ntf *CompositeNotifier) SendNewFollower(tx *AutoTx, toPrivate bool, from, to string) {
	const toQ = `
		SELECT show_name, email, verified AND email_followers, telegram
		FROM users 
		WHERE lower(name) = lower($1)
	`

	var sendEmail bool
	var toShowName, email string
	var tg sql.NullInt64
	tx.Query(toQ, to).Scan(&toShowName, &email, &sendEmail, &tg)

	const fromQ = `
		SELECT users.id, show_name, gender.type 
		FROM users, gender 
		WHERE lower(users.name) = lower($1) AND users.gender = gender.id`

	var fromID int64
	var fromShowName, fromGender string
	tx.Query(fromQ, from).Scan(&fromID, &fromShowName, &fromGender)

	if sendEmail {
		ntf.Mail.SendNewFollower(email, from, fromShowName, fromGender, toPrivate, toShowName)
	}

	if tg.Valid {
		ntf.Tg.SendNewFollower(tg.Int64, from, fromShowName, fromGender, toPrivate)
	}

	if toPrivate {
		ntf.Ntf.Notify(tx, fromID, models.NotificationTypeRequest, to)
	} else {
		ntf.Ntf.Notify(tx, fromID, models.NotificationTypeFollower, to)
	}
}

func (ntf *CompositeNotifier) SendNewAccept(tx *AutoTx, from, to string) {
	const toQ = `
		SELECT show_name, email, verified AND email_followers, telegram
		FROM users 
		WHERE lower(name) = lower($1)
	`

	var sendEmail bool
	var toShowName, email string
	var tg sql.NullInt64
	tx.Query(toQ, to).Scan(&toShowName, &email, &sendEmail, &tg)

	const fromQ = `
		SELECT users.id, show_name, gender.type 
		FROM users, gender 
		WHERE lower(users.name) = lower($1) AND users.gender = gender.id`

	var fromID int64
	var fromShowName, fromGender string
	tx.Query(fromQ, from).Scan(&fromID, &fromShowName, &fromGender)

	if sendEmail {
		ntf.Mail.SendNewAccept(email, from, fromShowName, fromGender, toShowName)
	}

	if tg.Valid {
		ntf.Tg.SendNewAccept(tg.Int64, from, fromShowName, fromGender)
	}

	ntf.Ntf.Notify(tx, fromID, "accept", to)
}

func (ntf *CompositeNotifier) SendNewCommentComplain(tx *AutoTx, commentID int64, from, content string) {
	const q = `
		SELECT entry_id, edit_content, name 
		FROM comments, users 
		WHERE comments.id = $1 AND users.id = comments.author_id`

	var entryID int64
	var comment, against string
	tx.Query(q, commentID).Scan(&entryID, &comment, &against)

	ntf.Mail.SendCommentComplain(from, against, content, comment, commentID, entryID)
}

func (ntf *CompositeNotifier) SendNewEntryComplain(tx *AutoTx, entryID int64, from, content string) {
	const q = `
		SELECT edit_content, name
		FROM entries, users 
		WHERE entries.id = $1 AND users.id = entries.author_id`

	var entry, against string
	tx.Query(q, entryID).Scan(&entry, &against)

	ntf.Mail.SendEntryComplain(from, against, content, entry, entryID)
}
