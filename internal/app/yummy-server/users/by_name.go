package users

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations/users"
)

func newUserLoaderByName(db *sql.DB) func(users.GetUsersByNameNameParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersByNameNameParams, userID *models.UserID) middleware.Responder {
		const q = profileQuery + "WHERE lower(long_users.name) = lower($1)"
		return loadProfile(db, q, userID, params.Name)
	}
}

const privacyQueryName = privacyQueryStart + "lower(users.name) = lower($1)"

const idFromName = "(SELECT id from users WHERE lower(name) = lower($1))"
const usersQueryToName = usersQueryStart + "relations.to_id = " + idFromName + " AND relations.from_id = short_users.id" + usersQueryEnd
const usersQueryFromName = usersQueryStart + "relations.from_id = " + idFromName + " AND relations.to_id = short_users.id" + usersQueryEnd

func loadUsersRelatedToName(db *sql.DB, usersQuery string,
	userID *models.UserID, args ...interface{}) middleware.Responder {
	return loadUsers(db, usersQuery, privacyQueryName, relationToNameQuery,
		userID, args...)
}

func newFollowersLoaderByName(db *sql.DB) func(users.GetUsersByNameNameFollowersParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersByNameNameFollowersParams, userID *models.UserID) middleware.Responder {
		return loadUsersRelatedToName(db, usersQueryToName,
			userID,
			params.Name, models.RelationshipRelationFollowed, *params.Limit, *params.Skip)
	}
}

func newFollowingsLoaderByName(db *sql.DB) func(users.GetUsersByNameNameFollowingsParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersByNameNameFollowingsParams, userID *models.UserID) middleware.Responder {
		return loadUsersRelatedToName(db, usersQueryFromName, userID,
			params.Name, models.RelationshipRelationFollowed, *params.Limit, *params.Skip)
	}
}

func newInvitedLoaderByName(db *sql.DB) func(users.GetUsersByNameNameInvitedParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersByNameNameInvitedParams, userID *models.UserID) middleware.Responder {
		return loadUsersRelatedToName(db, invitedUsersByNameQuery,
			userID, params.Name, *params.Limit, *params.Skip)
	}
}
