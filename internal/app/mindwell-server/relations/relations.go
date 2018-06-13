package relations

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/relations"
	"github.com/sevings/mindwell-server/utils"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	srv.API.RelationsGetRelationsToIDHandler = relations.GetRelationsToIDHandlerFunc(newToRelationLoader(srv))
	srv.API.RelationsGetRelationsFromIDHandler = relations.GetRelationsFromIDHandlerFunc(newFromRelationLoader(srv))

	srv.API.RelationsPutRelationsToIDHandler = relations.PutRelationsToIDHandlerFunc(newToRelationSetter(srv))
	srv.API.RelationsPutRelationsFromIDHandler = relations.PutRelationsFromIDHandlerFunc(newFromRelationSetter(srv))

	srv.API.RelationsDeleteRelationsToIDHandler = relations.DeleteRelationsToIDHandlerFunc(newToRelationDeleter(srv))
	srv.API.RelationsDeleteRelationsFromIDHandler = relations.DeleteRelationsFromIDHandlerFunc(newFromRelationDeleter(srv))
}

func newToRelationLoader(srv *utils.MindwellServer) func(relations.GetRelationsToIDParams, *models.UserID) middleware.Responder {
	return func(params relations.GetRelationsToIDParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			relation := relationship(tx, params.ID, userID)
			return relations.NewGetRelationsToIDOK().WithPayload(relation)
		})
	}
}

func newFromRelationLoader(srv *utils.MindwellServer) func(relations.GetRelationsFromIDParams, *models.UserID) middleware.Responder {
	return func(params relations.GetRelationsFromIDParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			relation := relationship(tx, userID, params.ID)
			return relations.NewGetRelationsFromIDOK().WithPayload(relation)
		})
	}
}

func sendNewFollower(srv *utils.MindwellServer, tx *utils.AutoTx, isPrivate bool, from, to int64) {
	const toQ = `
		SELECT show_name, name, gender.type, verified
		FROM users, gender 
		WHERE users.id = $1 AND users.gender = gender.id
	`

	var hisName, hisShowName, hisGender string
	var verified bool
	tx.Query(toQ, to).Scan(&hisShowName, &hisName, &hisGender, &verified)
	if !verified {
		return
	}

	const fromQ = "SELECT email, show_name FROM users WHERE id = $1"

	var email, name string
	tx.Query(fromQ, from).Scan(&email, &name)

	srv.Mail.SendNewFollower(email, name, isPrivate, hisShowName, hisName, hisGender)
}

func newToRelationSetter(srv *utils.MindwellServer) func(relations.PutRelationsToIDParams, *models.UserID) middleware.Responder {
	return func(params relations.PutRelationsToIDParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)

			if userID == params.ID {
				return relations.NewPutRelationsToIDForbidden()
			}

			isPrivate := isPrivateTlog(tx, params.ID)
			var relation *models.Relationship
			var ok bool
			if params.R == models.RelationshipRelationIgnored || !isPrivate {
				relation, ok = setRelationship(tx, userID, params.ID, params.R)
			} else {
				relation, ok = setRelationship(tx, userID, params.ID, models.RelationshipRelationRequested)
			}

			if !ok {
				return relations.NewPutRelationsToIDNotFound()
			}

			if params.R == models.RelationshipRelationFollowed {
				sendNewFollower(srv, tx, isPrivate, userID, params.ID)
			}

			return relations.NewPutRelationsToIDOK().WithPayload(relation)
		})
	}
}

func newFromRelationSetter(srv *utils.MindwellServer) func(relations.PutRelationsFromIDParams, *models.UserID) middleware.Responder {
	return func(params relations.PutRelationsFromIDParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			relation := relationship(tx, params.ID, userID)
			if relation.Relation != models.RelationshipRelationRequested {
				if tx.Error() == sql.ErrNoRows {
					return relations.NewPutRelationsFromIDNotFound()
				}
				return relations.NewPutRelationsFromIDForbidden()
			}

			relation, _ = setRelationship(tx, userID, params.ID, models.RelationshipRelationFollowed)

			return relations.NewPutRelationsFromIDOK().WithPayload(relation)
		})
	}
}

func newToRelationDeleter(srv *utils.MindwellServer) func(relations.DeleteRelationsToIDParams, *models.UserID) middleware.Responder {
	return func(params relations.DeleteRelationsToIDParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			relation := removeRelationship(tx, userID, params.ID)
			return relations.NewDeleteRelationsToIDOK().WithPayload(relation)
		})
	}
}

func newFromRelationDeleter(srv *utils.MindwellServer) func(relations.DeleteRelationsFromIDParams, *models.UserID) middleware.Responder {
	return func(params relations.DeleteRelationsFromIDParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			relation := relationship(tx, params.ID, userID)
			if relation.Relation != models.RelationshipRelationRequested {
				if tx.Error() == sql.ErrNoRows {
					return relations.NewDeleteRelationsToIDNotFound()
				}
				return relations.NewDeleteRelationsFromIDForbidden()
			}

			relation = removeRelationship(tx, params.ID, userID)
			return relations.NewDeleteRelationsFromIDOK().WithPayload(relation)
		})
	}
}
