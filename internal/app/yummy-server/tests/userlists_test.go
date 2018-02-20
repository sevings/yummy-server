package tests

import (
	"testing"

	usersImpl "github.com/sevings/yummy-server/internal/app/yummy-server/users"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"
	"github.com/sevings/yummy-server/restapi/operations/me"
	"github.com/stretchr/testify/require"
)

func TestUserLists(t *testing.T) {
	checkFollow(t, userIDs[0], userIDs[1], models.RelationshipRelationFollowed)
	checkFollow(t, userIDs[0], userIDs[2], models.RelationshipRelationFollowed)
	checkFollow(t, userIDs[1], userIDs[2], models.RelationshipRelationFollowed)

	api := operations.YummyAPI{}
	usersImpl.ConfigureAPI(db, &api)

	var skip int64
	var limit int64 = 100
	myfers := api.MeGetUsersMeFollowersHandler.Handle
	params := me.GetUsersMeFollowersParams{
		Skip:  &skip,
		Limit: &limit,
	}
	resp := myfers(params, userIDs[0])
	body, ok := resp.(*me.GetUsersMeFollowersOK)

	req := require.New(t)
	req.True(ok)

	list := body.Payload.Users
	req.Empty(list)

	resp = myfers(params, userIDs[1])
	body, ok = resp.(*me.GetUsersMeFollowersOK)
	req.True(ok)

	list = body.Payload.Users
	req.Equal(1, len(list))

	req.Equal(profiles[0].ID, list[0].ID)
}
