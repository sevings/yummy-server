package adm

import (
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/adm"
	"github.com/sevings/mindwell-server/utils"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	srv.API.AdmGetAdmGrandsonHandler = adm.GetAdmGrandsonHandlerFunc(newGrandsonLoader(srv))
	srv.API.AdmPostAdmGrandsonHandler = adm.PostAdmGrandsonHandlerFunc(newGrandsonUpdater(srv))
}

func newGrandsonLoader(srv *utils.MindwellServer) func(adm.GetAdmGrandsonParams, *models.UserID) middleware.Responder {
	return func(params adm.GetAdmGrandsonParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			address := adm.GetAdmGrandsonOKBody{}

			const q = `
				SELECT anonymous, fullname, postcode, country, address
				FROM adm
				WHERE lower(name) = lower($1)
			`

			log.Println(q)
			tx.Query(q, userID.Name).
				Scan(&address.Anonymous, &address.Name, &address.Postcode, &address.Country, &address.Address)

			if address.Country == "" {
				const usersQ = `
					SELECT country, city
					FROM users
					WHERE id = $1
				`

				tx.Query(q, userID.ID).Scan(&address.Country, &address.Address)
			}

			return adm.NewGetAdmGrandsonOK().WithPayload(&address)
		})
	}
}

func newGrandsonUpdater(srv *utils.MindwellServer) func(adm.PostAdmGrandsonParams, *models.UserID) middleware.Responder {
	return func(params adm.PostAdmGrandsonParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			const q = `
				INSERT INTO adm(name, anonymous, fullname, postcode, country, address)
				VALUES($1, $2, $3, $4, $5, $6)
				ON CONFLICT (lower(name)) DO UPDATE 
				SET anonymous = EXCLUDED.anonymous, fullname = EXCLUDED.fullname,
					postcode = EXCLUDED.postcode, country = EXCLUDED.country, address = EXCLUDED.address
			`

			tx.Exec(q, userID.Name, params.Anonymous, params.Name, params.Postcode, params.Country, params.Address)

			return adm.NewPostAdmGrandsonOK()
		})
	}
}
