package adm

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/adm"
	"github.com/sevings/mindwell-server/utils"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	srv.API.AdmGetAdmGrandsonHandler = adm.GetAdmGrandsonHandlerFunc(newGrandsonLoader(srv))
	srv.API.AdmPostAdmGrandsonHandler = adm.PostAdmGrandsonHandlerFunc(newGrandsonUpdater(srv))

	srv.API.AdmGetAdmStatHandler = adm.GetAdmStatHandlerFunc(newAdmStatLoader(srv))
}

func newGrandsonLoader(srv *utils.MindwellServer) func(adm.GetAdmGrandsonParams, *models.UserID) middleware.Responder {
	return func(params adm.GetAdmGrandsonParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			address := adm.GetAdmGrandsonOKBody{}

			const q = `
				SELECT anonymous, fullname, postcode, country, address, comment
				FROM adm
				WHERE lower(name) = lower($1)
			`

			tx.Query(q, userID.Name).
				Scan(&address.Anonymous, &address.Name, &address.Postcode,
					&address.Country, &address.Address, &address.Comment)

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
				INSERT INTO adm(name, anonymous, fullname, postcode, country, address, comment)
				VALUES($1, $2, $3, $4, $5, $6)
				ON CONFLICT (lower(name)) DO UPDATE 
				SET anonymous = EXCLUDED.anonymous, fullname = EXCLUDED.fullname,
					postcode = EXCLUDED.postcode, country = EXCLUDED.country,
					address = EXCLUDED.address, comment = EXCLUDED.comment
			`

			tx.Exec(q, userID.Name, params.Anonymous, params.Name, params.Postcode,
				params.Country, params.Address, params.Comment)

			return adm.NewPostAdmGrandsonOK()
		})
	}
}

func newAdmStatLoader(srv *utils.MindwellServer) func(adm.GetAdmStatParams, *models.UserID) middleware.Responder {
	return func(params adm.GetAdmStatParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			stat := adm.GetAdmStatOKBody{}

			tx.Query("SELECT count(*) FROM adm").
				Scan(&stat.Grandsons)

			return adm.NewGetAdmStatOK().WithPayload(&stat)
		})
	}
}
