package test

import (
	"testing"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/relations"
	"github.com/stretchr/testify/require"
)

func checkFromRelation(t *testing.T, user, to *models.UserID, relation string) {
	load := api.RelationsGetRelationsFromNameHandler.Handle
	params := relations.GetRelationsFromNameParams{
		Name: to.Name,
	}
	resp := load(params, user)
	body, ok := resp.(*relations.GetRelationsFromNameOK)
	req := require.New(t)
	req.True(ok)

	status := body.Payload
	req.Equal(params.Name, status.To)
	req.Equal(user.Name, status.From)
	req.Equal(relation, status.Relation)
}

func checkToRelation(t *testing.T, user, from *models.UserID, relation string) {
	load := api.RelationsGetRelationsToNameHandler.Handle
	params := relations.GetRelationsToNameParams{
		Name: from.Name,
	}
	resp := load(params, user)
	body, ok := resp.(*relations.GetRelationsToNameOK)
	req := require.New(t)
	req.True(ok)

	status := body.Payload
	req.Equal(params.Name, status.From)
	req.Equal(user.Name, status.To)
	req.Equal(relation, status.Relation)
}

func checkRelation(t *testing.T, from, to *models.UserID, relation string) {
	checkFromRelation(t, from, to, relation)
	checkToRelation(t, to, from, relation)
}

func checkFollow(t *testing.T, user *models.UserID, to *models.AuthProfile, relation string) {
	put := api.RelationsPutRelationsToNameHandler.Handle
	params := relations.PutRelationsToNameParams{
		Name: to.Name,
		R:    relation,
	}
	resp := put(params, user)
	body, ok := resp.(*relations.PutRelationsToNameOK)
	req := require.New(t)
	req.True(ok)

	status := body.Payload
	req.Equal(params.Name, status.To)
	req.Equal(user.Name, status.From)
	req.Equal(relation, status.Relation)

	if relation == models.RelationshipRelationFollowed && to.Account.Verified {
		esm.CheckEmail(t, to.Account.Email)
	} else {
		req.Empty(esm.Emails)
	}
}

func checkPermitFollow(t *testing.T, user, from *models.UserID, success bool) {
	put := api.RelationsPutRelationsFromNameHandler.Handle
	params := relations.PutRelationsFromNameParams{
		Name: from.Name,
	}
	resp := put(params, user)
	body, ok := resp.(*relations.PutRelationsFromNameOK)
	req := require.New(t)
	req.Equal(success, ok)
	if !success {
		return
	}

	status := body.Payload
	req.Equal(user.ID, status.To)
	req.Equal(params.Name, status.From)
	req.Equal(models.RelationshipRelationFollowed, status.Relation)
}

func checkUnfollow(t *testing.T, user, to *models.UserID) {
	del := api.RelationsDeleteRelationsToNameHandler.Handle
	params := relations.DeleteRelationsToNameParams{
		Name: to.Name,
	}
	resp := del(params, user)
	body, ok := resp.(*relations.DeleteRelationsToNameOK)
	req := require.New(t)
	req.True(ok)

	status := body.Payload
	req.Equal(params.Name, status.To)
	req.Equal(user.Name, status.From)
	req.Equal(models.RelationshipRelationNone, status.Relation)
}

func checkCancelFollow(t *testing.T, user, from *models.UserID, success bool) {
	del := api.RelationsDeleteRelationsFromNameHandler.Handle
	params := relations.DeleteRelationsFromNameParams{
		Name: from.Name,
	}
	resp := del(params, user)
	body, ok := resp.(*relations.DeleteRelationsToNameOK)
	req := require.New(t)
	req.True(ok)

	status := body.Payload
	req.Equal(user.ID, status.To)
	req.Equal(params.Name, status.From)
	req.Equal(models.RelationshipRelationNone, status.Relation)
}

func TestRelationship(t *testing.T) {
	checkRelation(t, userIDs[0], userIDs[1], models.RelationshipRelationNone)
	checkRelation(t, userIDs[1], userIDs[0], models.RelationshipRelationNone)

	checkFollow(t, userIDs[0], profiles[1], models.RelationshipRelationFollowed)
	checkRelation(t, userIDs[0], userIDs[1], models.RelationshipRelationFollowed)
	checkRelation(t, userIDs[1], userIDs[0], models.RelationshipRelationNone)

	checkFollow(t, userIDs[0], profiles[1], models.RelationshipRelationIgnored)
	checkRelation(t, userIDs[0], userIDs[1], models.RelationshipRelationIgnored)
	checkRelation(t, userIDs[1], userIDs[0], models.RelationshipRelationNone)

	checkUnfollow(t, userIDs[0], userIDs[1])
	checkRelation(t, userIDs[0], userIDs[1], models.RelationshipRelationNone)
	checkRelation(t, userIDs[1], userIDs[0], models.RelationshipRelationNone)
}
