package test

import (
	"testing"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/relations"
	"github.com/sevings/mindwell-server/utils"
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

func checkFollow(t *testing.T, user *models.UserID, toID *models.UserID, to *models.AuthProfile, relation string) {
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

	if relation == models.RelationshipRelationFollowed && to.Account.Verified && getEmailSettings(t, toID).Followers {
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
	req.Equal(user.Name, status.To)
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
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)
	esm.Clear()

	checkRelation(t, userIDs[0], userIDs[1], models.RelationshipRelationNone)
	checkRelation(t, userIDs[1], userIDs[0], models.RelationshipRelationNone)

	checkFollow(t, userIDs[0], userIDs[1], profiles[1], models.RelationshipRelationFollowed)
	checkRelation(t, userIDs[0], userIDs[1], models.RelationshipRelationFollowed)
	checkRelation(t, userIDs[1], userIDs[0], models.RelationshipRelationNone)

	checkUpdateEmailSettings(t, userIDs[2], true, false, false)
	checkFollow(t, userIDs[0], userIDs[2], profiles[2], models.RelationshipRelationFollowed)
	checkRelation(t, userIDs[0], userIDs[2], models.RelationshipRelationFollowed)
	checkRelation(t, userIDs[2], userIDs[0], models.RelationshipRelationNone)

	checkUpdateEmailSettings(t, userIDs[2], true, true, false)

	checkFollow(t, userIDs[0], userIDs[1], profiles[1], models.RelationshipRelationIgnored)
	checkRelation(t, userIDs[0], userIDs[1], models.RelationshipRelationIgnored)
	checkRelation(t, userIDs[1], userIDs[0], models.RelationshipRelationNone)

	checkUnfollow(t, userIDs[0], userIDs[1])
	checkRelation(t, userIDs[0], userIDs[1], models.RelationshipRelationNone)
	checkRelation(t, userIDs[1], userIDs[0], models.RelationshipRelationNone)

	checkUnfollow(t, userIDs[0], userIDs[2])
	checkRelation(t, userIDs[0], userIDs[2], models.RelationshipRelationNone)
	checkRelation(t, userIDs[2], userIDs[0], models.RelationshipRelationNone)
}

func TestInvite(t *testing.T) {
	req := require.New(t)

	invite := func(name, inv string, from *models.UserID, success bool) {
		post := api.RelationsPostRelationsInvitedNameHandler.Handle
		params := relations.PostRelationsInvitedNameParams{
			Invite: inv,
			Name:   name,
		}
		resp := post(params, from)
		_, ok := resp.(*relations.PostRelationsInvitedNameNoContent)
		req.Equal(success, ok)
	} 

	from := &models.UserID{
		ID:   1,
		Name: "Mindwell",
	}

	invite("test3", "acknown acknown acknown", from, false)

	e1 := postEntry(userIDs[3], models.EntryPrivacyAll, false)
	e2 := postEntry(userIDs[3], models.EntryPrivacyAll, false)
	e3 := postEntry(userIDs[3], models.EntryPrivacyAll, false)

	voteForEntry(userIDs[0], e1.ID, true)
	voteForEntry(userIDs[0], e2.ID, true)
	voteForEntry(userIDs[1], e1.ID, true)
	voteForEntry(userIDs[1], e2.ID, true)
	voteForEntry(userIDs[1], e3.ID, true)
	voteForEntry(userIDs[2], e2.ID, false)

	invite("test3", "acknown acknown acknown", from, false)

	voteForEntry(userIDs[2], e3.ID, true)

	invite("test0", "acknown acknown acknown", from, false)
	invite("fsdf", "acknown acknown acknown", from, false)
	invite("test0", "", from, false)
	invite("test0", "acknown acknown sd", from, false)
	invite("test0", "acknown acknown acknown", userIDs[2], false)
	invite("test3", "acknown acknown acknown", from, true)

	if profiles[3].Account.Verified {
		esm.CheckEmail(t, profiles[3].Account.Email)
	} else {
		req.Empty(esm.Emails)
	}

	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)
	esm.Clear()
}
