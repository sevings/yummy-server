package images

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/me"
	"github.com/sevings/mindwell-server/utils"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	srv.API.MeGetMeImagesHandler = me.GetMeImagesHandlerFunc(newMyImagesLoader(srv))
}

func newMyImagesLoader(srv *utils.MindwellServer) func(me.GetMeImagesParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeImagesParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			const query = `
				SELECT id, path 
				FROM images
				WHERE user_id = $1
				ORDER BY created_at desc
				LIMIT $2
			`

			return me.NewGetMeImagesNotFound()
		})
	}
}
