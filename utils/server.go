package utils

import (
	"database/sql"
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations"
	goconf "github.com/zpatrick/go-config"
)

type MindwellServer struct {
	DB  *sql.DB
	API *operations.MindwellAPI
	Cfg *goconf.Config
}

func NewMindwellServer(api *operations.MindwellAPI, configPath string) *MindwellServer {
	config := LoadConfig(configPath)
	db := OpenDatabase(config)

	return &MindwellServer{
		DB:  db,
		API: api,
		Cfg: config,
	}
}

func (srv *MindwellServer) ConfigString(field string) string {
	value, err := srv.Cfg.String(field)
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
