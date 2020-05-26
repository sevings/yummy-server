package relations

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/relations"
	"github.com/sevings/mindwell-server/utils"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	srv.API.RelationsGetRelationsToNameHandler = relations.GetRelationsToNameHandlerFunc(newToRelationLoader(srv))
	srv.API.RelationsGetRelationsFromNameHandler = relations.GetRelationsFromNameHandlerFunc(newFromRelationLoader(srv))

	srv.API.RelationsPutRelationsToNameHandler = relations.PutRelationsToNameHandlerFunc(newToRelationSetter(srv))
	srv.API.RelationsPutRelationsFromNameHandler = relations.PutRelationsFromNameHandlerFunc(newFromRelationSetter(srv))

	srv.API.RelationsDeleteRelationsToNameHandler = relations.DeleteRelationsToNameHandlerFunc(newToRelationDeleter(srv))
	srv.API.RelationsDeleteRelationsFromNameHandler = relations.DeleteRelationsFromNameHandlerFunc(newFromRelationDeleter(srv))

	srv.API.RelationsPostRelationsInvitedNameHandler = relations.PostRelationsInvitedNameHandlerFunc(newInviter(srv))
}

func newToRelationLoader(srv *utils.MindwellServer) func(relations.GetRelationsToNameParams, *models.UserID) middleware.Responder {
	return func(params relations.GetRelationsToNameParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			relation := relationship(tx, params.Name, uID.Name)
			return relations.NewGetRelationsToNameOK().WithPayload(relation)
		})
	}
}

func newFromRelationLoader(srv *utils.MindwellServer) func(relations.GetRelationsFromNameParams, *models.UserID) middleware.Responder {
	return func(params relations.GetRelationsFromNameParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			relation := relationship(tx, uID.Name, params.Name)
			return relations.NewGetRelationsFromNameOK().WithPayload(relation)
		})
	}
}

func newToRelationSetter(srv *utils.MindwellServer) func(relations.PutRelationsToNameParams, *models.UserID) middleware.Responder {
	return func(params relations.PutRelationsToNameParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			if uID.Name == params.Name {
				err := srv.NewError(&i18n.Message{ID: "self_relation", Other: "You can't have relationship with youself."})
				return relations.NewPutRelationsToNameForbidden().WithPayload(err)
			}

			if params.R == models.RelationshipRelationFollowed {
				toRelation := relationship(tx, params.Name, uID.Name)
				if toRelation.Relation == models.RelationshipRelationIgnored {
					err := srv.NewError(&i18n.Message{ID: "relation_from_ignored", Other: "You can't follow this user."})
					return relations.NewPutRelationsToNameForbidden().WithPayload(err)
				}
			}

			isPrivate := isPrivateTlog(tx, params.Name)
			var relation *models.Relationship
			var ok bool
			if params.R == models.RelationshipRelationIgnored || !isPrivate {
				relation, ok = setRelationship(tx, uID.Name, params.Name, params.R)
			} else {
				relation, ok = setRelationship(tx, uID.Name, params.Name, models.RelationshipRelationRequested)
			}

			if !ok {
				err := srv.StandardError("no_tlog")
				return relations.NewPutRelationsToNameNotFound().WithPayload(err)
			}

			if params.R == models.RelationshipRelationFollowed {
				srv.Ntf.SendNewFollower(tx, isPrivate, uID.Name, params.Name)
			}

			return relations.NewPutRelationsToNameOK().WithPayload(relation)
		})
	}
}

func newFromRelationSetter(srv *utils.MindwellServer) func(relations.PutRelationsFromNameParams, *models.UserID) middleware.Responder {
	return func(params relations.PutRelationsFromNameParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			relation := relationship(tx, params.Name, uID.Name)
			if relation.Relation != models.RelationshipRelationRequested {
				err := srv.StandardError("no_request")
				return relations.NewPutRelationsFromNameForbidden().WithPayload(err)
			}

			relation, _ = setRelationship(tx, params.Name, uID.Name, models.RelationshipRelationFollowed)
			srv.Ntf.SendNewAccept(tx, uID.Name, params.Name)

			return relations.NewPutRelationsFromNameOK().WithPayload(relation)
		})
	}
}

func newToRelationDeleter(srv *utils.MindwellServer) func(relations.DeleteRelationsToNameParams, *models.UserID) middleware.Responder {
	return func(params relations.DeleteRelationsToNameParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			relation := removeRelationship(tx, uID.Name, params.Name)
			return relations.NewDeleteRelationsToNameOK().WithPayload(relation)
		})
	}
}

func newFromRelationDeleter(srv *utils.MindwellServer) func(relations.DeleteRelationsFromNameParams, *models.UserID) middleware.Responder {
	return func(params relations.DeleteRelationsFromNameParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			relation := relationship(tx, params.Name, uID.Name)
			if relation.Relation != models.RelationshipRelationRequested && relation.Relation != models.RelationshipRelationFollowed {
				err := srv.StandardError("no_request")
				return relations.NewDeleteRelationsFromNameForbidden().WithPayload(err)
			}

			relation = removeRelationship(tx, params.Name, uID.Name)
			return relations.NewDeleteRelationsFromNameOK().WithPayload(relation)
		})
	}
}

func newInviter(srv *utils.MindwellServer) func(relations.PostRelationsInvitedNameParams, *models.UserID) middleware.Responder {
	return func(params relations.PostRelationsInvitedNameParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			if userID.Ban.Invite {
				err := srv.NewError(&i18n.Message{ID: "cant_invite", Other: "You are not allowed to invite users."})
				return relations.NewPostRelationsInvitedNameForbidden().WithPayload(err)
			}

			exists, invited := isTlogExistsAndInvited(tx, params.Name)

			if !exists {
				err := srv.StandardError("no_tlog")
				return relations.NewPostRelationsInvitedNameNotFound().WithPayload(err)
			}

			if invited {
				err := srv.NewError(&i18n.Message{ID: "already_invited", Other: "The user already has an invite."})
				return relations.NewPostRelationsInvitedNameForbidden().WithPayload(err)
			}

			if !canInvite(tx, params.Name) {
				err := srv.NewError(&i18n.Message{ID: "cant_be_invited", Other: "The user can't be invited."})
				return relations.NewPostRelationsInvitedNameForbidden().WithPayload(err)
			}

			if ok := removeInvite(tx, params.Invite, userID.ID); !ok {
				err := srv.StandardError("invalid_invite")
				return relations.NewPostRelationsInvitedNameForbidden().WithPayload(err)
			}

			setInvited(tx, userID.ID, params.Name)
			srv.Ntf.SendInvited(tx, userID.Name, params.Name)

			return relations.NewPostRelationsInvitedNameNoContent()
		})
	}
}
