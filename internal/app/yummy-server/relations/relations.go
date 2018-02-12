package relations

import (
	"database/sql"

	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	"github.com/sevings/yummy-server/restapi/operations/relations"
	"github.com/sevings/yummy-server/models"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.RelationsGetRelationsToIDHandler = relations.GetRelationsToIDHandlerFunc(newToRelationLoader(db))
	api.RelationsGetRelationsFromIDHandler = relations.GetRelationsFromIDHandlerFunc(newFromRelationLoader(db))
	
	api.RelationsPutRelationsToIDHandler = relations.PutRelationsToIDHandlerFunc(newToRelationSetter(db))
	api.RelationsPutRelationsFromIDHandler = relations.PutRelationsFromIDHandlerFunc(newFromRelationSetter(db))
	
	api.RelationsDeleteRelationsToIDHandler = relations.DeleteRelationsToIDHandlerFunc(newToRelationDeleter(db))
	api.RelationsDeleteRelationsFromIDHandler = relations.DeleteRelationsFromIDHandlerFunc(newFromRelationDeleter(db))
}

func newToRelationLoader(db *sql.DB) func(relations.GetRelationsToIDParams, *models.UserID) middleware.Responder {
	return func(params relations.GetRelationsToIDParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			relation := relationship(tx, userID, params.ID)
			return relations.NewGetRelationsToIDOK().WithPayload(relation)
		}
	}
}

func newFromRelationLoader(db *sql.DB) func(relations.GetRelationsFromIDParams, *models.UserID) middleware.Responder {
	return func(params relations.GetRelationsFromIDParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			relation := relationship(tx, params.ID, userID)
			return relations.NewGetRelationsFromIDOK().WithPayload(relation)
		}
	}
}

func newToRelationSetter(db *sql.DB) func(relations.PutRelationsToIDParams, *models.UserID) middleware.Responder {
	return func(params relations.PutRelationsToIDParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			var relation *models.Relationship
			var ok bool
			if params.R == models.RelationshipRelationIgnored || !isPrivateTlog(tx, params.ID) {
				relation, ok := setRelationship(tx, userID, params.ID, params.R)
			} else {
				relation, ok := setRelationship(tx, userID, params.ID, models.RelationshipRelationRequested)
			}

			if !ok {
				return relations.NewPutRelationsToIDNotFound()
			}
			
			return relations.NewPutRelationsToIDOK().WithPayload(relation)
		}
	}
}

func newFromRelationSetter(db *sql.DB) func(relations.PutRelationsFromIDParams, *models.UserID) middleware.Responder {
	return func(params relations.PutRelationsFromIDParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
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
		}
	}
}

func newToRelationDeleter(db *sql.DB) func(relations.DeleteRelationsToIDParams, *models.UserID) middleware.Responder {
	return func(params relations.DeleteRelationsToIDParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			relation := removeRelationship(tx, userID, params.ID)
			return relations.NewDeleteRelationsToIDOK().WithPayload(relation)
		}
	}
}

func newFromRelationDeleter(db *sql.DB) func(relations.DeleteRelationsFromIDParams, *models.UserID) middleware.Responder {
	return func(params relations.DeleteRelationsFromIDParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			relation := relationship(tx, params.ID, userID)
			if relation.Relation != models.RelationshipRelationRequested {
				if tx.Error() == sql.ErrNoRows {
					return relations.NewDeleteRelationsToIDNotFound()
				}
				return relations.NewDeleteRelationsFromIDForbidden()
			}

			relation, _ = removeRelationship(tx, userID, params.ID)
			return relations.NewDeleteRelationsFromIDOK().WithPayload(relation)
		}
	}
}
