package users

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/restapi/operations/users"
)

const profileQueryByName = profileQuery + "WHERE long_users.name = $1"

func newUserLoaderByName(db *sql.DB) func(users.GetUsersByNameNameParams) middleware.Responder {
	return func(params users.GetUsersByNameNameParams) middleware.Responder {
		return loadProfile(db, profileQueryByName, params.XUserKey, params.Name)
	}
}

const privacyQueryName = privacyQueryStart + "users.name = $1"
const usersQueryToName = usersQueryStart + "name = $1 AND relations.to_id = short_users.id" + usersQueryEnd
const usersQueryFromName = usersQueryStart + "name = $1 AND relations.from_id = short_users.id" + usersQueryEnd

func loadUsersRelatedToName(db *sql.DB, usersQuery string,
	apiKey *string,
	arg interface{}, relation string, limit, offset int64) middleware.Responder {
	return loadUsers(db, usersQuery, privacyQueryName, relationToNameQuery,
		apiKey, arg, relation, limit, offset)
}

func newFollowersLoaderByName(db *sql.DB) func(users.GetUsersByNameNameFollowersParams) middleware.Responder {
	return func(params users.GetUsersByNameNameFollowersParams) middleware.Responder {
		return loadUsersRelatedToName(db, usersQueryToName,
			params.XUserKey,
			params.Name, "followed", *params.Limit, *params.Skip)
	}
}
