package utils

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"
	"strconv"
	"time"

	"github.com/roylee0704/gron/xtime"

	"github.com/BurntSushi/toml"
	"github.com/go-openapi/runtime/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/roylee0704/gron"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations"
	goconf "github.com/zpatrick/go-config"
	"golang.org/x/text/language"
)

type MailSender interface {
	SendGreeting(address, name, code string)
	SendPasswordChanged(address, name string)
	SendEmailChanged(address, name string)
	SendResetPassword(address, name, gender, code string, date int64)
	SendNewComment(address, fromGender, toShowName, entryTitle string, cmt *models.Comment)
	SendNewFollower(address, fromName, fromShowName, fromGender string, toPrivate bool, toShowName string)
	SendNewAccept(address, fromName, fromShowName, fromGender, toShowName string)
	SendNewInvite(address, name string)
	SendInvited(address, fromShowName, fromGender, toShowName string)
}

type MindwellServer struct {
	DB    *sql.DB
	API   *operations.MindwellAPI
	Mail  MailSender
	Ntf   *Notifier
	Tg    *TelegramBot
	cfg   *goconf.Config
	local *i18n.Localizer
	cron  *gron.Cron
	errs  map[string]*i18n.Message
}

func NewMindwellServer(api *operations.MindwellAPI, configPath string) *MindwellServer {
	config := LoadConfig(configPath)
	db := OpenDatabase(config)

	trFile, err := config.String("server.tr_file")
	if err != nil {
		log.Print(err)
	}

	bundle := &i18n.Bundle{DefaultLanguage: language.Russian}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile(trFile)

	srv := &MindwellServer{
		DB:    db,
		API:   api,
		cfg:   config,
		local: i18n.NewLocalizer(bundle),
		cron:  gron.New(),
		errs: map[string]*i18n.Message{
			"no_entry":       &i18n.Message{ID: "no_entry", Other: "Entry not found or you have no access rights."},
			"no_comment":     &i18n.Message{ID: "no_comment", Other: "Comment not found or you have no access rights."},
			"no_tlog":        &i18n.Message{ID: "no_tlog", Other: "Tlog not found or you have no access rights."},
			"no_request":     &i18n.Message{ID: "no_friend_request", Other: "You have no friend request from this user."},
			"invalid_invite": &i18n.Message{ID: "invalid_invite", Other: "Invite is invalid."},
		},
	}

	ntfURL := srv.ConfigString("centrifugo.api_url")
	ntfKey := srv.ConfigString("centrifugo.api_key")
	srv.Ntf = NewNotifier(ntfURL, ntfKey)
	srv.Tg = NewTelegramBot(srv)

	srv.cron.AddFunc(gron.Every(xtime.Day).At("00:10"), func() { srv.recalcKarma() })
	srv.cron.AddFunc(gron.Every(xtime.Day).At("00:15"), func() { srv.giveInvites() })
	srv.cron.Start()

	return srv
}

func (srv *MindwellServer) ConfigString(field string) string {
	value, err := srv.cfg.String(field)
	if err != nil {
		log.Println(err)
	}

	return value
}

func (srv *MindwellServer) ConfigInt(field string) int {
	value, err := srv.cfg.Int(field)
	if err != nil {
		log.Println(err)
	}

	return value
}

func (srv *MindwellServer) NewAvatar(avatar string) *models.Avatar {
	base := srv.ConfigString("images.base_url") + "avatars/"

	return &models.Avatar{
		X42:  base + "42/" + avatar,
		X92:  base + "92/" + avatar,
		X124: base + "124/" + avatar,
	}
}

func (srv *MindwellServer) NewCover(id int64, cover string) *models.Cover {
	if len(cover) == 0 {
		cnt := int64(srv.ConfigInt("images.covers"))
		n := int(id % cnt)
		cover = "default/" + strconv.Itoa(n) + ".jpg"
	}

	base := srv.ConfigString("images.base_url")

	return &models.Cover{
		ID:    id,
		X1920: base + "covers/1920/" + cover,
		X318:  base + "covers/318/" + cover,
	}
}

func (srv *MindwellServer) ImagesFolder() string {
	return srv.ConfigString("images.folder")
}

func (srv *MindwellServer) Transact(txFunc func(*AutoTx) middleware.Responder) middleware.Responder {
	return Transact(srv.DB, txFunc)
}

func (srv *MindwellServer) PasswordHash(password string) []byte {
	salt := srv.ConfigString("server.pass_salt")
	sum := sha256.Sum256([]byte(password + salt))
	return sum[:]
}

func (srv *MindwellServer) VerificationCode(email string) string {
	salt := srv.ConfigString("server.mail_salt")
	sum := sha256.Sum256([]byte(email + salt))
	sha := hex.EncodeToString(sum[:])
	return sha
}

func (srv *MindwellServer) resetCode(email string, date int64) string {
	salt := srv.ConfigString("server.mail_salt")
	str := email + salt + strconv.FormatInt(date, 16)
	sum := sha256.Sum256([]byte(str))
	sha := hex.EncodeToString(sum[:])
	return sha
}

func (srv *MindwellServer) ResetPasswordCode(email string) (string, int64) {
	date := time.Now().Unix()
	code := srv.resetCode(email, date)
	return code, date
}

func (srv *MindwellServer) CheckResetPasswordCode(email, code string, date int64) bool {
	now := time.Now().Unix()
	if (now - date) >= 60*60 {
		return false
	}

	if date > now {
		return false
	}

	return srv.resetCode(email, date) == code
}

// NewError returns error object with some message
func (srv *MindwellServer) NewError(msg *i18n.Message) *models.Error {
	if msg == nil {
		msg = &i18n.Message{ID: "internal_error", Other: "Internal server error."}
	}

	message, err := srv.local.Localize(&i18n.LocalizeConfig{DefaultMessage: msg})
	if err != nil {
		log.Print(err)
	}

	return &models.Error{Message: message}
}

func (srv *MindwellServer) StandardError(name string) *models.Error {
	msg := srv.errs[name]
	if msg == nil {
		log.Printf("Standard error not found: %s\n", name)
	}

	return srv.NewError(msg)
}

func (srv *MindwellServer) recalcKarma() {
	log.Println("Updating karma...")

	_, err := srv.DB.Exec("SELECT recalc_karma()")
	if err != nil {
		log.Println(err)
	}
}

func (srv *MindwellServer) giveInvites() {
	srv.Transact(func(tx *AutoTx) middleware.Responder {
		tx.Query("SELECT user_id FROM give_invites()")

		var ids []int
		for {
			var id int
			if !tx.Scan(&id) {
				break
			}

			ids = append(ids, id)
		}

		for _, id := range ids {
			var email, name, showName string
			var sendEmail bool
			var tg sql.NullInt64

			tx.Query("SELECT email, name, show_name, verified AND email_invites, telegram FROM users WHERE id = $1", id)
			tx.Scan(&email, &name, &showName, &sendEmail, &tg)

			srv.Ntf.Notify(tx, 0, "invite", name)

			if tg.Valid {
				srv.Tg.SendNewInvite(tg.Int64)
			}

			if sendEmail {
				srv.Mail.SendNewInvite(email, showName)
			}
		}

		log.Printf("%d new invites given.\n", len(ids))

		return nil
	})
}
