package tests

import (
	"testing"

	relationsImpl "github.com/sevings/yummy-server/internal/app/yummy-server/relations"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"
	"github.com/sevings/yummy-server/restapi/operations/relations"
	"github.com/stretchr/testify/require"
)

func checkFromRelation(t *testing.T, user, to *models.UserID, relation string) {
	api := operations.YummyAPI{}
	relationsImpl.ConfigureAPI(db, &api)

	load := api.RelationsGetRelationsFromIDHandler.Handle
	params := relations.GetRelationsFromIDParams{
		ID: int64(*to),
	}
	resp := load(params, user)
	body, ok := resp.(*relations.GetRelationsFromIDOK)
	req := require.New(t)
	req.True(ok)

	status := body.Payload
	req.Equal(params.ID, status.To)
	req.Equal(int64(*user), status.From)
	req.Equal(relation, status.Relation)
}

func checkToRelation(t *testing.T, user, from *models.UserID, relation string) {
	api := operations.YummyAPI{}
	relationsImpl.ConfigureAPI(db, &api)

	load := api.RelationsGetRelationsToIDHandler.Handle
	params := relations.GetRelationsToIDParams{
		ID: int64(*from),
	}
	resp := load(params, user)
	body, ok := resp.(*relations.GetRelationsToIDOK)
	req := require.New(t)
	req.True(ok)

	status := body.Payload
	req.Equal(params.ID, status.From)
	req.Equal(int64(*user), status.To)
	req.Equal(relation, status.Relation)
}

func checkRelation(t *testing.T, from, to *models.UserID, relation string) {
	checkFromRelation(t, from, to, relation)
	checkToRelation(t, to, from, relation)
}

func checkFollow(t *testing.T, user, to *models.UserID, relation string) {
	api := operations.YummyAPI{}
	relationsImpl.ConfigureAPI(db, &api)

	put := api.RelationsPutRelationsToIDHandler.Handle
	params := relations.PutRelationsToIDParams{
		ID: int64(*to),
		R:  relation,
	}
	resp := put(params, user)
	body, ok := resp.(*relations.PutRelationsToIDOK)
	req := require.New(t)
	req.True(ok)

	status := body.Payload
	req.Equal(params.ID, status.To)
	req.Equal(int64(*user), status.From)
	req.Equal(relation, status.Relation)
}

func checkPermitFollow(t *testing.T, user, from *models.UserID, success bool) {
	api := operations.YummyAPI{}
	relationsImpl.ConfigureAPI(db, &api)

	put := api.RelationsPutRelationsFromIDHandler.Handle
	params := relations.PutRelationsFromIDParams{
		ID: int64(*from),
	}
	resp := put(params, user)
	body, ok := resp.(*relations.PutRelationsFromIDOK)
	req := require.New(t)
	req.Equal(success, ok)
	if !success {
		return
	}

	status := body.Payload
	req.Equal(int64(*user), status.To)
	req.Equal(params.ID, status.From)
	req.Equal(models.RelationshipRelationFollowed, status.Relation)
}

func checkUnfollow(t *testing.T, user, to *models.UserID) {
	api := operations.YummyAPI{}
	relationsImpl.ConfigureAPI(db, &api)

	del := api.RelationsDeleteRelationsToIDHandler.Handle
	params := relations.DeleteRelationsToIDParams{
		ID: int64(*to),
	}
	resp := del(params, user)
	body, ok := resp.(*relations.DeleteRelationsToIDOK)
	req := require.New(t)
	req.True(ok)

	status := body.Payload
	req.Equal(params.ID, status.To)
	req.Equal(int64(*user), status.From)
	req.Equal(models.RelationshipRelationNone, status.Relation)
}

func checkCancelFollow(t *testing.T, user, from *models.UserID, success bool) {
	api := operations.YummyAPI{}
	relationsImpl.ConfigureAPI(db, &api)

	del := api.RelationsDeleteRelationsFromIDHandler.Handle
	params := relations.DeleteRelationsFromIDParams{
		ID: int64(*from),
	}
	resp := del(params, user)
	body, ok := resp.(*relations.DeleteRelationsToIDOK)
	req := require.New(t)
	req.True(ok)

	status := body.Payload
	req.Equal(int64(*user), status.To)
	req.Equal(params.ID, status.From)
	req.Equal(models.RelationshipRelationNone, status.Relation)
}

func TestRelationship(t *testing.T) {
	checkRelation(t, userIDs[0], userIDs[1], models.RelationshipRelationNone)
	checkRelation(t, userIDs[1], userIDs[0], models.RelationshipRelationNone)

	checkFollow(t, userIDs[0], userIDs[1], models.RelationshipRelationFollowed)
	checkRelation(t, userIDs[0], userIDs[1], models.RelationshipRelationFollowed)
	checkRelation(t, userIDs[1], userIDs[0], models.RelationshipRelationNone)

	checkFollow(t, userIDs[0], userIDs[1], models.RelationshipRelationIgnored)
	checkRelation(t, userIDs[0], userIDs[1], models.RelationshipRelationIgnored)
	checkRelation(t, userIDs[1], userIDs[0], models.RelationshipRelationNone)

	checkUnfollow(t, userIDs[0], userIDs[1])
	checkRelation(t, userIDs[0], userIDs[1], models.RelationshipRelationNone)
	checkRelation(t, userIDs[1], userIDs[0], models.RelationshipRelationNone)
}
