package users

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations/users"
)

func newUserLoader(db *sql.DB) func(users.GetUsersIDParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersIDParams, userID *models.UserID) middleware.Responder {
		const q = profileQuery + "WHERE long_users.id = $1"
		return loadProfile(db, q, userID, params.ID)
	}
}

const privacyQueryID = privacyQueryStart + "users.id = $1"
const usersQueryToID = usersQueryStart + "relations.to_id = $1 AND relations.from_id = short_users.id" + usersQueryEnd
const usersQueryFromID = usersQueryStart + "relations.from_id = $1 AND relations.to_id = short_users.id" + usersQueryEnd

func loadUsersRelatedToID(db *sql.DB, usersQuery string,
	userID *models.UserID,
	arg interface{}, relation string, limit, offset int64) middleware.Responder {
	return loadUsers(db, usersQuery, privacyQueryID, relationToIDQuery,
		userID, arg, relation, limit, offset)
}

func newFollowersLoader(db *sql.DB) func(users.GetUsersIDFollowersParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersIDFollowersParams, userID *models.UserID) middleware.Responder {
		return loadUsersRelatedToID(db, usersQueryToID,
			userID,
			params.ID, "followed", *params.Limit, *params.Skip)
	}
}

func newFollowingsLoader(db *sql.DB) func(users.GetUsersIDFollowingsParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersIDFollowingsParams, userID *models.UserID) middleware.Responder {
		return loadUsersRelatedToID(db, usersQueryFromID,
			userID,
			params.ID, "followed", *params.Limit, *params.Skip)
	}
}
