package users

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations/users"
)

const profileQueryByName = profileQuery + "WHERE long_users.name = $1"

func newUserLoaderByName(db *sql.DB) func(users.GetUsersByNameNameParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersByNameNameParams, userID *models.UserID) middleware.Responder {
		return loadProfile(db, profileQueryByName, userID, params.Name)
	}
}

const privacyQueryName = privacyQueryStart + "users.name = $1"
const usersQueryToName = usersQueryStart + "name = $1 AND relations.to_id = short_users.id" + usersQueryEnd
const usersQueryFromName = usersQueryStart + "name = $1 AND relations.from_id = short_users.id" + usersQueryEnd

func loadUsersRelatedToName(db *sql.DB, usersQuery string,
	userID *models.UserID,
	arg interface{}, relation string, limit, offset int64) middleware.Responder {
	return loadUsers(db, usersQuery, privacyQueryName, relationToNameQuery,
		userID, arg, relation, limit, offset)
}

func newFollowersLoaderByName(db *sql.DB) func(users.GetUsersByNameNameFollowersParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersByNameNameFollowersParams, userID *models.UserID) middleware.Responder {
		return loadUsersRelatedToName(db, usersQueryToName,
			userID,
			params.Name, "followed", *params.Limit, *params.Skip)
	}
}
