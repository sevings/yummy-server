package users

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/users"
	"github.com/sevings/mindwell-server/utils"
)

func newUserLoader(srv *utils.MindwellServer) func(users.GetUsersNameParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersNameParams, userID *models.UserID) middleware.Responder {
		const query = profileQuery + "WHERE lower(users.name) = lower($1)"

		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			profile := loadUserProfile(srv, tx, query, userID, params.Name)
			if profile.ID == 0 {
				return users.NewGetUsersNameNotFound()
			}

			return users.NewGetUsersNameOK().WithPayload(profile)
		})
	}
}

const idFromName = "(SELECT id FROM users WHERE lower(name) = lower($1))"
const usersToNameQueryWhere = "relations.to_id = " + idFromName + " AND relations.from_id = users.id" + usersQueryEnd
const usersFromNameQueryWhere = "relations.from_id = " + idFromName + " AND relations.to_id = users.id" + usersQueryEnd
const usersQueryToName = usersQueryStart + usersToNameQueryWhere
const usersQueryFromName = usersQueryStart + usersFromNameQueryWhere
const invitedByQueryWhere = "invited_by = " + idFromName + usersQueryJoins
const invitedUsersQuery = usersQuerySelect + `, users.id FROM users, gender, user_privacy WHERE ` + invitedByQueryWhere

func newFollowersLoader(srv *utils.MindwellServer) func(users.GetUsersNameFollowersParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersNameFollowersParams, userID *models.UserID) middleware.Responder {
		return loadRelatedUsers(srv, userID, usersQueryToName, usersToNameQueryWhere,
			models.RelationshipRelationFollowed, params.Name, models.FriendListRelationFollowers,
			*params.After, *params.Before, *params.Limit)
	}
}

func newFollowingsLoader(srv *utils.MindwellServer) func(users.GetUsersNameFollowingsParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersNameFollowingsParams, userID *models.UserID) middleware.Responder {
		return loadRelatedUsers(srv, userID, usersQueryFromName, usersFromNameQueryWhere,
			models.RelationshipRelationFollowed, params.Name, models.FriendListRelationFollowings,
			*params.After, *params.Before, *params.Limit)
	}
}

func newInvitedLoader(srv *utils.MindwellServer) func(users.GetUsersNameInvitedParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersNameInvitedParams, userID *models.UserID) middleware.Responder {
		return loadInvitedUsers(srv, userID, invitedUsersQuery, invitedByQueryWhere,
			params.Name, *params.After, *params.Before, *params.Limit)
	}
}
