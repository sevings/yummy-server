package utils

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/go-openapi/runtime/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations"
	goconf "github.com/zpatrick/go-config"
	"golang.org/x/text/language"
)

type MailSender interface {
	SendGreeting(address, name, code string)
	SendNewComment(address, name, gender, entryTitle string, cmt *models.Comment)
	SendNewFollower(address, name string, isPrivate bool, hisShowName, hisName, gender string)
}

type MindwellServer struct {
	DB    *sql.DB
	API   *operations.MindwellAPI
	Mail  MailSender
	cfg   *goconf.Config
	local *i18n.Localizer
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

	return &MindwellServer{
		DB:    db,
		API:   api,
		cfg:   config,
		local: i18n.NewLocalizer(bundle),
		errs: map[string]*i18n.Message{
			"no_entry":   &i18n.Message{ID: "no_entry", Other: "Entry not found or you have no access rights."},
			"no_comment": &i18n.Message{ID: "no_comment", Other: "Comment not found or you have no access rights."},
			"no_tlog":    &i18n.Message{ID: "no_tlog", Other: "Tlog not found or you have no access rights."},
			"no_request": &i18n.Message{ID: "no_friend_request", Other: "You have no friend request from this user."},
		},
	}
}

func (srv *MindwellServer) ConfigString(field string) string {
	value, err := srv.cfg.String(field)
	if err != nil {
		log.Println(err)
	}

	return value
}

func (srv *MindwellServer) NewAvatar(avatar string) *models.Avatar {
	base := srv.ConfigString("images.base_url")

	return &models.Avatar{
		X42:  base + "42/" + avatar,
		X92:  base + "92/" + avatar,
		X124: base + "124/" + avatar,
	}
}

func (srv *MindwellServer) NewCover(id int64, cover string) *models.Cover {
	if len(cover) == 0 {
		cover = srv.ConfigString("images.cover")
	}

	base := srv.ConfigString("images.base_url")

	return &models.Cover{
		ID:    id,
		X1920: base + "cover/1920/" + cover,
		X318:  base + "cover/318/" + cover,
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
