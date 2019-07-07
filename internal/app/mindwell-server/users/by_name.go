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
const usersQueryToName = usersQueryStart + "relations.to_id = " + idFromName + " AND relations.from_id = users.id" + usersQueryEnd
const usersQueryFromName = usersQueryStart + "relations.from_id = " + idFromName + " AND relations.to_id = users.id" + usersQueryEnd

func loadUsersRelatedToName(srv *utils.MindwellServer, usersQuery, relation string,
	userID *models.UserID, args ...interface{}) middleware.Responder {

	return loadUsers(srv, usersQuery, loadUserQueryName, relation,
		userID, args...)
}

func newFollowersLoader(srv *utils.MindwellServer) func(users.GetUsersNameFollowersParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersNameFollowersParams, userID *models.UserID) middleware.Responder {
		return loadUsersRelatedToName(srv, usersQueryToName, models.FriendListRelationFollowers,
			userID,
			params.Name, models.RelationshipRelationFollowed, *params.Limit, *params.Skip)
	}
}

func newFollowingsLoader(srv *utils.MindwellServer) func(users.GetUsersNameFollowingsParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersNameFollowingsParams, userID *models.UserID) middleware.Responder {
		return loadUsersRelatedToName(srv, usersQueryFromName, models.FriendListRelationFollowings, userID,
			params.Name, models.RelationshipRelationFollowed, *params.Limit, *params.Skip)
	}
}

func newInvitedLoader(srv *utils.MindwellServer) func(users.GetUsersNameInvitedParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersNameInvitedParams, userID *models.UserID) middleware.Responder {
		return loadUsersRelatedToName(srv, invitedUsersQuery, models.FriendListRelationInvited,
			userID, params.Name, *params.Limit, *params.Skip)
	}
}
