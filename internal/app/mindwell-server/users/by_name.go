package users

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/users"
	"github.com/sevings/mindwell-server/utils"
)

func newUserLoaderByName(srv *utils.MindwellServer) func(users.GetUsersByNameNameParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersByNameNameParams, userID *models.UserID) middleware.Responder {
		const q = profileQuery + "WHERE lower(long_users.name) = lower($1)"
		return loadProfile(srv, q, userID, params.Name)
	}
}

const privacyQueryName = privacyQueryStart + "lower(users.name) = lower($1)"

const idFromName = "(SELECT id from users WHERE lower(name) = lower($1))"
const usersQueryToName = usersQueryStart + "relations.to_id = " + idFromName + " AND relations.from_id = short_users.id" + usersQueryEnd
const usersQueryFromName = usersQueryStart + "relations.from_id = " + idFromName + " AND relations.to_id = short_users.id" + usersQueryEnd

func loadUsersRelatedToName(srv *utils.MindwellServer, usersQuery, relation string,
	userID *models.UserID, args ...interface{}) middleware.Responder {
	return loadUsers(srv, usersQuery, privacyQueryName, relationToNameQuery, loadUserQueryName, relation,
		userID, args...)
}

func newFollowersLoaderByName(srv *utils.MindwellServer) func(users.GetUsersByNameNameFollowersParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersByNameNameFollowersParams, userID *models.UserID) middleware.Responder {
		return loadUsersRelatedToName(srv, usersQueryToName, models.UserListRelationFollowers,
			userID,
			params.Name, models.RelationshipRelationFollowed, *params.Limit, *params.Skip)
	}
}

func newFollowingsLoaderByName(srv *utils.MindwellServer) func(users.GetUsersByNameNameFollowingsParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersByNameNameFollowingsParams, userID *models.UserID) middleware.Responder {
		return loadUsersRelatedToName(srv, usersQueryFromName, models.UserListRelationFollowings, userID,
			params.Name, models.RelationshipRelationFollowed, *params.Limit, *params.Skip)
	}
}

func newInvitedLoaderByName(srv *utils.MindwellServer) func(users.GetUsersByNameNameInvitedParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersByNameNameInvitedParams, userID *models.UserID) middleware.Responder {
		return loadUsersRelatedToName(srv, invitedUsersByNameQuery, models.UserListRelationInvited,
			userID, params.Name, *params.Limit, *params.Skip)
	}
}
