package users

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/restapi/operations/users"
)

const profileQueryByID = profileQuery + "WHERE long_users.id = $1"

func newUserLoader(db *sql.DB) func(users.GetUsersIDParams) middleware.Responder {
	return func(params users.GetUsersIDParams) middleware.Responder {
		return loadProfile(db, profileQueryByID, params.XUserKey, params.ID)
	}
}

const privacyQueryID = privacyQueryStart + "users.id = $1"
const usersQueryToID = usersQueryStart + "relations.to_id = $1 AND relations.from_id = short_users.id" + usersQueryEnd
const usersQueryFromID = usersQueryStart + "relations.from_id = $1 AND relations.to_id = short_users.id" + usersQueryEnd

func loadUsersRelatedToID(db *sql.DB, usersQuery string,
	apiKey *string,
	arg interface{}, relation string, limit, offset int64) middleware.Responder {
	return loadUsers(db, usersQuery, privacyQueryID, relationToIDQuery,
		apiKey, arg, relation, limit, offset)
}

func newFollowersLoader(db *sql.DB) func(users.GetUsersIDFollowersParams) middleware.Responder {
	return func(params users.GetUsersIDFollowersParams) middleware.Responder {
		return loadUsersRelatedToID(db, usersQueryToID,
			params.XUserKey,
			params.ID, "followed", *params.Limit, *params.Skip)
	}
}

func newFollowingsLoader(db *sql.DB) func(users.GetUsersIDFollowingsParams) middleware.Responder {
	return func(params users.GetUsersIDFollowingsParams) middleware.Responder {
		return loadUsersRelatedToID(db, usersQueryFromID,
			params.XUserKey,
			params.ID, "followed", *params.Limit, *params.Skip)
	}
}
