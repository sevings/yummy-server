package adm

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/adm"
	"github.com/sevings/mindwell-server/utils"
)

var finishedErr *models.Error
var notRegErr *models.Error
var admBanErr *models.Error

var regFinished bool
var admFinished bool

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	finishedErr = srv.NewError(&i18n.Message{ID: "adm_reg_finished", Other: "ADM registration finished."})
	notRegErr = srv.NewError(&i18n.Message{ID: "not_in_adm", Other: "You are not registered in ADM."})
	admBanErr = srv.NewError(&i18n.Message{ID: "cant_be_adm", Other: "You are not allowed to participate in ADM."})

	regFinished = srv.ConfigBool("adm.reg_finished")
	admFinished = srv.ConfigBool("adm.adm_finished")

	srv.API.AdmGetAdmGrandsonHandler = adm.GetAdmGrandsonHandlerFunc(newGrandsonLoader(srv))
	srv.API.AdmPostAdmGrandsonHandler = adm.PostAdmGrandsonHandlerFunc(newGrandsonUpdater(srv))

	srv.API.AdmGetAdmGrandsonStatusHandler = adm.GetAdmGrandsonStatusHandlerFunc(newGrandsonStatusLoader(srv))
	srv.API.AdmPostAdmGrandsonStatusHandler = adm.PostAdmGrandsonStatusHandlerFunc(newGrandsonStatusUpdater(srv))

	srv.API.AdmGetAdmGrandfatherHandler = adm.GetAdmGrandfatherHandlerFunc(newGrandfatherLoader(srv))

	srv.API.AdmGetAdmGrandfatherStatusHandler = adm.GetAdmGrandfatherStatusHandlerFunc(newGrandfatherStatusLoader(srv))
	srv.API.AdmPostAdmGrandfatherStatusHandler = adm.PostAdmGrandfatherStatusHandlerFunc(newGrandfatherStatusUpdater(srv))

	srv.API.AdmGetAdmStatHandler = adm.GetAdmStatHandlerFunc(newAdmStatLoader(srv))
}

func newGrandsonLoader(srv *utils.MindwellServer) func(adm.GetAdmGrandsonParams, *models.UserID) middleware.Responder {
	return func(params adm.GetAdmGrandsonParams, userID *models.UserID) middleware.Responder {
		if regFinished {
			return adm.NewGetAdmGrandsonGone().WithPayload(finishedErr)
		}

		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			banned := tx.QueryBool("SELECT adm_ban FROM users WHERE id = $1", userID.ID)
			if banned {
				return adm.NewGetAdmGrandsonStatusForbidden().WithPayload(admBanErr)
			}

			address := adm.GetAdmGrandsonOKBody{}

			const q = `
				SELECT anonymous, fullname, postcode, country, address, comment
				FROM adm
				WHERE lower(name) = lower($1)
			`

			tx.Query(q, userID.Name).
				Scan(&address.Anonymous, &address.Name, &address.Postcode,
					&address.Country, &address.Address, &address.Comment)

			if len(address.Country) == 0 {
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
		if regFinished {
			return adm.NewPostAdmGrandsonGone().WithPayload(finishedErr)
		}

		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			banned := tx.QueryBool("SELECT adm_ban FROM users WHERE id = $1", userID.ID)
			if banned {
				return adm.NewGetAdmGrandsonStatusForbidden().WithPayload(admBanErr)
			}

			const q = `
				INSERT INTO adm(name, anonymous, fullname, postcode, country, address, comment)
				VALUES($1, $2, $3, $4, $5, $6, $7)
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

func newGrandsonStatusLoader(srv *utils.MindwellServer) func(adm.GetAdmGrandsonStatusParams, *models.UserID) middleware.Responder {
	return func(params adm.GetAdmGrandsonStatusParams, userID *models.UserID) middleware.Responder {
		if admFinished {
			return adm.NewGetAdmGrandsonStatusGone().WithPayload(finishedErr)
		}

		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			status := adm.GetAdmGrandsonStatusOKBody{}

			tx.Query("SELECT received, sent FROM adm WHERE lower(name) = lower($1)", userID.Name).
				Scan(&status.Received, &status.Sent)

			if tx.Error() == sql.ErrNoRows {
				return adm.NewGetAdmGrandsonStatusForbidden().WithPayload(notRegErr)
			}

			return adm.NewGetAdmGrandsonStatusOK().WithPayload(&status)
		})
	}
}

func newGrandsonStatusUpdater(srv *utils.MindwellServer) func(adm.PostAdmGrandsonStatusParams, *models.UserID) middleware.Responder {
	return func(params adm.PostAdmGrandsonStatusParams, userID *models.UserID) middleware.Responder {
		if admFinished {
			return adm.NewPostAdmGrandsonStatusGone().WithPayload(finishedErr)
		}

		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			tx.Exec("UPDATE adm SET received = $2 WHERE lower(name) = lower($1)", userID.Name, params.Received)

			if tx.RowsAffected() == 0 {
				return adm.NewPostAdmGrandsonStatusForbidden().WithPayload(notRegErr)
			}

			return adm.NewPostAdmGrandsonStatusOK()
		})
	}
}

func newGrandfatherLoader(srv *utils.MindwellServer) func(adm.GetAdmGrandfatherParams, *models.UserID) middleware.Responder {
	return func(params adm.GetAdmGrandfatherParams, userID *models.UserID) middleware.Responder {
		if admFinished {
			return adm.NewGetAdmGrandsonStatusGone().WithPayload(finishedErr)
		}

		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			address := adm.GetAdmGrandfatherOKBody{}

			const q = `
				SELECT anonymous, fullname, postcode, country, address, comment, name
				FROM adm
				WHERE lower(grandfather) = lower($1)
			`

			var anon bool
			tx.Query(q, userID.Name).
				Scan(&anon, &address.Fullname, &address.Postcode,
					&address.Country, &address.Address, &address.Comment, &address.Name)

			if tx.Error() == sql.ErrNoRows {
				return adm.NewGetAdmGrandfatherForbidden().WithPayload(notRegErr)
			}

			if anon {
				address.Name = ""
			}

			return adm.NewGetAdmGrandfatherOK().WithPayload(&address)
		})
	}
}

func newGrandfatherStatusLoader(srv *utils.MindwellServer) func(adm.GetAdmGrandfatherStatusParams, *models.UserID) middleware.Responder {
	return func(params adm.GetAdmGrandfatherStatusParams, userID *models.UserID) middleware.Responder {
		if admFinished {
			return adm.NewGetAdmGrandfatherStatusGone().WithPayload(finishedErr)
		}

		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			status := adm.GetAdmGrandfatherStatusOKBody{}

			tx.Query("SELECT received, sent FROM adm WHERE lower(grandfather) = lower($1)", userID.Name).
				Scan(&status.Received, &status.Sent)

			if tx.Error() == sql.ErrNoRows {
				return adm.NewGetAdmGrandfatherStatusForbidden().WithPayload(notRegErr)
			}

			return adm.NewGetAdmGrandfatherStatusOK().WithPayload(&status)
		})
	}
}

func newGrandfatherStatusUpdater(srv *utils.MindwellServer) func(adm.PostAdmGrandfatherStatusParams, *models.UserID) middleware.Responder {
	return func(params adm.PostAdmGrandfatherStatusParams, userID *models.UserID) middleware.Responder {
		if admFinished {
			return adm.NewPostAdmGrandfatherStatusGone().WithPayload(finishedErr)
		}

		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			tx.Exec("UPDATE adm SET sent = $2 WHERE lower(grandfather) = lower($1)", userID.Name, params.Sent)

			if tx.RowsAffected() == 0 {
				return adm.NewPostAdmGrandfatherStatusForbidden().WithPayload(notRegErr)
			}

			return adm.NewPostAdmGrandfatherStatusOK()
		})
	}
}

func newAdmStatLoader(srv *utils.MindwellServer) func(adm.GetAdmStatParams, *models.UserID) middleware.Responder {
	return func(params adm.GetAdmStatParams, userID *models.UserID) middleware.Responder {
		if admFinished {
			return adm.NewGetAdmStatGone().WithPayload(finishedErr)
		}

		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			stat := adm.GetAdmStatOKBody{}

			tx.Query("SELECT count(*) FROM adm").
				Scan(&stat.Grandsons)

			tx.Query("SELECT count(*) FROM adm WHERE sent").
				Scan(&stat.Sent)

			tx.Query("SELECT count(*) FROM adm WHERE received").
				Scan(&stat.Received)

			return adm.NewGetAdmStatOK().WithPayload(&stat)
		})
	}
}
