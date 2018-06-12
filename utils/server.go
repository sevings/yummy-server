package utils

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations"
	goconf "github.com/zpatrick/go-config"
)

type MailSender interface {
	SendGreeting(address, name, code string)
	SendNewComment(address, name, gender, entryTitle string, cmt *models.Comment)
}

type MindwellServer struct {
	DB   *sql.DB
	API  *operations.MindwellAPI
	Mail MailSender
	cfg  *goconf.Config
}

func NewMindwellServer(api *operations.MindwellAPI, configPath string) *MindwellServer {
	config := LoadConfig(configPath)
	db := OpenDatabase(config)

	return &MindwellServer{
		DB:  db,
		API: api,
		cfg: config,
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
		X100: base + "100/" + avatar,
		X400: base + "400/" + avatar,
		X800: base + "800/" + avatar,
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
	sha := base64.URLEncoding.EncodeToString(sum[:])
	return sha
}
