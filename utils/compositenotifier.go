package utils

import (
	"database/sql"
	"github.com/sevings/mindwell-server/models"
	"gitlab.com/golang-commonmark/markdown"
)

type CompositeNotifier struct {
	srv  *MindwellServer
	md   *markdown.Markdown
	Mail MailSender
	Ntf  *Notifier
	Tg   *TelegramBot
}

func NewCompositeNotifier(srv *MindwellServer) *CompositeNotifier {
	ntfURL := srv.ConfigString("centrifugo.api_url")
	ntfKey := srv.ConfigString("centrifugo.api_key")

	return &CompositeNotifier{
		srv: srv,
		md:  markdown.New(markdown.Typographer(false), markdown.Breaks(true), markdown.Tables(false)),
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
	var sendEmail, sendTg bool
	var tg sql.NullInt64

	const q = `
		SELECT email, name, show_name, verified AND email_invites, telegram, telegram_invites
		FROM users WHERE id = $1
	`

	tx.Query(q, userID)
	tx.Scan(&email, &name, &showName, &sendEmail, &tg, &sendTg)

	ntf.Ntf.Notify(tx, 0, typeInvite, name)

	if tg.Valid && sendTg {
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
			ntf.srv.LogSystem().Error(tx.Error().Error())
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
			ntf.srv.LogSystem().Error(tx.Error().Error())
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

func (ntf *CompositeNotifier) entryTitle(tx *AutoTx, entryID int64) string {
	var title, content string
	tx.Query("SELECT title, edit_content FROM entries WHERE id = $1", entryID)
	tx.Scan(&title, &content)

	if title != "" {
		return title
	}

	content = ntf.md.RenderToString([]byte(content))
	content = RemoveHTML(content)
	content, _ = CutText(content, 100)

	return content
}

func (ntf *CompositeNotifier) SendNewComment(tx *AutoTx, cmt *models.Comment) {
	title := ntf.entryTitle(tx, cmt.EntryID)

	const fromQ = `
		SELECT gender.type 
		FROM users, gender 
		WHERE users.id = $1 AND users.gender = gender.id
	`

	var fromGender string
	tx.Query(fromQ, cmt.Author.ID).Scan(&fromGender)

	const toQ = `
		SELECT users.name, show_name, email, verified AND email_comments, telegram, telegram_comments
		FROM users 
		INNER JOIN watching ON watching.user_id = users.id 
		WHERE watching.entry_id = $1 AND users.id <> $2 
			AND can_view_entry(users.id, $1)`

	tx.Query(toQ, cmt.EntryID, cmt.Author.ID)

	var toNames []string
	var toName string
	var sendEmail, sendTg bool
	var toShowName, email string
	var tg sql.NullInt64
	for tx.Scan(&toName, &toShowName, &email, &sendEmail, &tg, &sendTg) {
		if sendEmail {
			ntf.Mail.SendNewComment(email, fromGender, toShowName, title, cmt)
		}

		if tg.Valid && sendTg {
			ntf.Tg.SendNewComment(tg.Int64, title, cmt)
		}

		toNames = append(toNames, toName)
	}

	for _, name := range toNames {
		ntf.Ntf.Notify(tx, cmt.ID, typeComment, name)
	}
}

func (ntf *CompositeNotifier) SendUpdateComment(tx *AutoTx, cmt *models.Comment) {
	title := ntf.entryTitle(tx, cmt.EntryID)

	ntf.Tg.SendUpdateComment(title, cmt)
	ntf.Ntf.NotifyUpdate(tx, cmt.ID, typeComment)
}

func (ntf *CompositeNotifier) SendRemoveComment(tx *AutoTx, commentID int64) {
	ntf.Tg.SendRemoveComment(commentID)
	ntf.Ntf.NotifyRemove(tx, commentID, typeComment)
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

	ntf.Ntf.Notify(tx, fromID, typeInvited, to)
}

func (ntf *CompositeNotifier) SendNewFollower(tx *AutoTx, toPrivate bool, from, to string) {
	const toQ = `
		SELECT show_name, email, verified AND email_followers, telegram, telegram_followers
		FROM users 
		WHERE lower(name) = lower($1)
	`

	var sendEmail, sendTg bool
	var toShowName, email string
	var tg sql.NullInt64
	tx.Query(toQ, to).Scan(&toShowName, &email, &sendEmail, &tg, &sendTg)

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

	if tg.Valid && sendTg {
		ntf.Tg.SendNewFollower(tg.Int64, from, fromShowName, fromGender, toPrivate)
	}

	if toPrivate {
		ntf.Ntf.Notify(tx, fromID, typeRequest, to)
	} else {
		ntf.Ntf.Notify(tx, fromID, typeFollower, to)
	}
}

func (ntf *CompositeNotifier) SendNewAccept(tx *AutoTx, from, to string) {
	const toQ = `
		SELECT show_name, email, verified AND email_followers, telegram, telegram_followers
		FROM users 
		WHERE lower(name) = lower($1)
	`

	var sendEmail, sendTg bool
	var toShowName, email string
	var tg sql.NullInt64
	tx.Query(toQ, to).Scan(&toShowName, &email, &sendEmail, &tg, &sendTg)

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

	if tg.Valid && sendTg {
		ntf.Tg.SendNewAccept(tg.Int64, from, fromShowName, fromGender)
	}

	ntf.Ntf.Notify(tx, fromID, typeAccept, to)
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

const retryQuery = `
SELECT EXISTS(SELECT 1 
	FROM notifications 
	JOIN notification_type ON notifications.type = notification_type.id
	JOIN users on user_id = users.id
	WHERE users.name = $1 AND notification_type.type = $2 AND age(notifications.created_at) < interval '6 month')
`

func (ntf *CompositeNotifier) SendAdmSent(tx *AutoTx, grandson string) {
	if tx.QueryBool(retryQuery, grandson, "adm_sent") {
		return
	}

	const toQ = `
		SELECT show_name, email, verified, telegram
		FROM users 
		WHERE lower(name) = lower($1)
	`

	var sendEmail bool
	var toShowName, email string
	var tg sql.NullInt64
	tx.Query(toQ, grandson).Scan(&toShowName, &email, &sendEmail, &tg)

	if sendEmail {
		ntf.Mail.SendAdmSent(email, toShowName)
	}

	if tg.Valid {
		ntf.Tg.SendAdmSent(tg.Int64)
	}

	ntf.Ntf.Notify(tx, 0, typeAdmSent, grandson)
}

func (ntf *CompositeNotifier) SendAdmReceived(tx *AutoTx, grandfather string) {
	if tx.QueryBool(retryQuery, grandfather, "adm_received") {
		return
	}

	const toQ = `
		SELECT show_name, email, verified, telegram
		FROM users 
		WHERE lower(name) = lower($1)
	`

	var sendEmail bool
	var toShowName, email string
	var tg sql.NullInt64
	tx.Query(toQ, grandfather).Scan(&toShowName, &email, &sendEmail, &tg)

	if sendEmail {
		ntf.Mail.SendAdmReceived(email, toShowName)
	}

	if tg.Valid {
		ntf.Tg.SendAdmReceived(tg.Int64)
	}

	ntf.Ntf.Notify(tx, 0, typeAdmReceived, grandfather)
}

func (ntf *CompositeNotifier) NotifyMessage(tx *AutoTx, msg *models.Message, user string) {
	const q = "SELECT telegram, telegram_messages FROM users WHERE lower(name) = lower($1)"

	var tg sql.NullInt64
	var sendTg bool
	tx.Query(q, user).Scan(&tg, &sendTg)

	if tg.Valid && sendTg {
		ntf.Tg.SendNewMessage(tg.Int64, msg)
	}

	ntf.Ntf.NotifyMessage(msg.ChatID, msg.ID, user)
}

func (ntf *CompositeNotifier) NotifyMessageUpdate(msg *models.Message, user string) {
	ntf.Tg.SendUpdateMessage(msg)
	ntf.Ntf.NotifyMessageUpdate(msg.ChatID, msg.ID, user)
}

func (ntf *CompositeNotifier) NotifyMessageRemove(chatID, msgID int64, user string) {
	ntf.Tg.SendRemoveMessage(msgID)
	ntf.Ntf.NotifyMessageRemove(chatID, msgID, user)
}

func (ntf *CompositeNotifier) NotifyMessageRead(chatID, msgID int64, user string) {
	ntf.Ntf.NotifyMessageRead(chatID, msgID, user)
}
