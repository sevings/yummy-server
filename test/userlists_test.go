package test

import (
	"strings"
	"testing"
	"time"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/me"
	"github.com/sevings/mindwell-server/restapi/operations/users"
	"github.com/stretchr/testify/require"
)

func checkMyFollowers(t *testing.T, user *models.UserID, skip, limit int64, size int) []*models.Friend {
	get := api.MeGetMeFollowersHandler.Handle
	params := me.GetMeFollowersParams{
		Skip:  &skip,
		Limit: &limit,
	}
	resp := get(params, user)
	body, ok := resp.(*me.GetMeFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, user.ID, list.Subject.ID)
	require.Equal(t, models.FriendListRelationFollowers, list.Relation)

	return list.Users
}

func checkMyFollowings(t *testing.T, user *models.UserID, skip, limit int64, size int) []*models.Friend {
	get := api.MeGetMeFollowingsHandler.Handle
	params := me.GetMeFollowingsParams{
		Skip:  &skip,
		Limit: &limit,
	}
	resp := get(params, user)
	body, ok := resp.(*me.GetMeFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, user.ID, list.Subject.ID)
	require.Equal(t, models.FriendListRelationFollowings, list.Relation)

	return list.Users
}

func checkMyIgnored(t *testing.T, user *models.UserID, skip, limit int64, size int) []*models.Friend {
	get := api.MeGetMeIgnoredHandler.Handle
	params := me.GetMeIgnoredParams{
		Skip:  &skip,
		Limit: &limit,
	}
	resp := get(params, user)
	body, ok := resp.(*me.GetMeFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, user.ID, list.Subject.ID)
	require.Equal(t, models.FriendListRelationIgnored, list.Relation)

	return list.Users
}

func checkMyInvited(t *testing.T, user *models.UserID, skip, limit int64, size int) []*models.Friend {
	get := api.MeGetMeInvitedHandler.Handle
	params := me.GetMeInvitedParams{
		Skip:  &skip,
		Limit: &limit,
	}
	resp := get(params, user)
	body, ok := resp.(*me.GetMeFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, user.ID, list.Subject.ID)
	require.Equal(t, models.FriendListRelationInvited, list.Relation)

	return list.Users
}

func checkNameFollowers(t *testing.T, user *models.UserID, name string, skip, limit int64, size int) []*models.Friend {
	get := api.UsersGetUsersNameFollowersHandler.Handle
	params := users.GetUsersNameFollowersParams{
		Skip:  &skip,
		Limit: &limit,
		Name:  name,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersNameFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, strings.ToLower(name), strings.ToLower(list.Subject.Name))
	require.Equal(t, models.FriendListRelationFollowers, list.Relation)

	return list.Users
}

func checkNameFollowings(t *testing.T, user *models.UserID, name string, skip, limit int64, size int) []*models.Friend {
	get := api.UsersGetUsersNameFollowingsHandler.Handle
	params := users.GetUsersNameFollowingsParams{
		Skip:  &skip,
		Limit: &limit,
		Name:  name,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersNameFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, strings.ToLower(name), strings.ToLower(list.Subject.Name))
	require.Equal(t, models.FriendListRelationFollowings, list.Relation)

	return list.Users
}

func checkNameInvited(t *testing.T, user *models.UserID, name string, skip, limit int64, size int) []*models.Friend {
	get := api.UsersGetUsersNameInvitedHandler.Handle
	params := users.GetUsersNameInvitedParams{
		Skip:  &skip,
		Limit: &limit,
		Name:  name,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersNameFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, strings.ToLower(name), strings.ToLower(list.Subject.Name))
	require.Equal(t, models.FriendListRelationInvited, list.Relation)

	return list.Users
}

func TestOpenFriendLists(t *testing.T) {
	checkFollow(t, userIDs[0], userIDs[1], profiles[1], models.RelationshipRelationFollowed)
	time.Sleep(10 * time.Millisecond)
	checkFollow(t, userIDs[0], userIDs[2], profiles[2], models.RelationshipRelationFollowed)
	time.Sleep(10 * time.Millisecond)
	checkFollow(t, userIDs[1], userIDs[2], profiles[2], models.RelationshipRelationFollowed)

	req := require.New(t)
	var list []*models.Friend

	checkMyFollowers(t, userIDs[0], 0, 100, 0)

	list = checkMyFollowers(t, userIDs[1], 0, 100, 1)
	req.Equal(profiles[0].ID, list[0].ID)

	checkMyFollowings(t, userIDs[2], 0, 100, 0)

	list = checkMyFollowings(t, userIDs[0], 0, 100, 2)
	req.Equal(profiles[2].ID, list[0].ID)
	req.Equal(profiles[1].ID, list[1].ID)

	list = checkMyFollowings(t, userIDs[0], 0, 1, 1)
	req.Equal(profiles[2].ID, list[0].ID)
	list = checkMyFollowings(t, userIDs[0], 1, 10, 1)
	req.Equal(profiles[1].ID, list[0].ID)

	checkNameFollowers(t, userIDs[0], profiles[0].Name, 0, 100, 0)

	list = checkNameFollowers(t, userIDs[0], profiles[1].Name, 0, 100, 1)
	req.Equal(profiles[0].ID, list[0].ID)

	checkNameFollowings(t, userIDs[0], profiles[2].Name, 0, 100, 0)

	list = checkNameFollowings(t, userIDs[0], strings.ToLower(profiles[0].Name), 0, 1, 1)
	req.Equal(profiles[2].ID, list[0].ID)
	list = checkNameFollowings(t, userIDs[0], strings.ToUpper(profiles[0].Name), 1, 1, 1)
	req.Equal(profiles[1].ID, list[0].ID)

	checkUnfollow(t, userIDs[0], userIDs[1])
	checkUnfollow(t, userIDs[0], userIDs[2])
	checkUnfollow(t, userIDs[1], userIDs[2])

	checkFollow(t, userIDs[0], userIDs[1], profiles[1], models.RelationshipRelationIgnored)
	time.Sleep(10 * time.Millisecond)
	checkFollow(t, userIDs[0], userIDs[2], profiles[2], models.RelationshipRelationIgnored)

	checkMyIgnored(t, userIDs[2], 0, 100, 0)

	list = checkMyIgnored(t, userIDs[0], 0, 100, 2)
	req.Equal(profiles[2].ID, list[0].ID)
	req.Equal(profiles[1].ID, list[1].ID)

	list = checkMyIgnored(t, userIDs[0], 0, 1, 1)
	req.Equal(profiles[2].ID, list[0].ID)
	list = checkMyIgnored(t, userIDs[0], 1, 10, 1)
	req.Equal(profiles[1].ID, list[0].ID)

	checkUnfollow(t, userIDs[0], userIDs[1])
	checkUnfollow(t, userIDs[0], userIDs[2])

	checkMyInvited(t, userIDs[2], 0, 100, 0)

	inviter := models.UserID{
		ID:   1,
		Name: "mindwell",
	}
	list = checkMyInvited(t, &inviter, 0, 100, 4)
	req.Equal(profiles[2].ID, list[0].ID)
	req.Equal(profiles[1].ID, list[1].ID)
	req.Equal(profiles[0].ID, list[2].ID)
	req.Equal(int64(1), list[3].ID)

	list = checkMyInvited(t, &inviter, 0, 2, 2)
	req.Equal(profiles[2].ID, list[0].ID)
	req.Equal(profiles[1].ID, list[1].ID)

	list = checkMyInvited(t, &inviter, 2, 2, 2)
	req.Equal(profiles[0].ID, list[0].ID)
	req.Equal(int64(1), list[1].ID)

	list = checkMyInvited(t, &inviter, 4, 2, 0)

	list = checkNameInvited(t, userIDs[0], "minDWEll", 0, 100, 4)
	req.Equal(profiles[2].ID, list[0].ID)
	req.Equal(profiles[1].ID, list[1].ID)
	req.Equal(profiles[0].ID, list[2].ID)
	req.Equal(int64(1), list[3].ID)
}

func TestPrivateFriendLists(t *testing.T) {
	checkFollow(t, userIDs[0], userIDs[2], profiles[2], models.RelationshipRelationFollowed)
	time.Sleep(10 * time.Millisecond)
	checkFollow(t, userIDs[1], userIDs[2], profiles[2], models.RelationshipRelationFollowed)

	profiles[2].Privacy = "followers"

	params := me.PutMeParams{
		Privacy:  profiles[2].Privacy,
		ShowName: profiles[2].ShowName,
	}
	checkEditProfile(t, profiles[2], params)

	req := require.New(t)
	var list []*models.Friend

	list = checkNameFollowers(t, userIDs[0], profiles[2].Name, 0, 100, 2)
	req.Equal(profiles[1].ID, list[0].ID)
	req.Equal(profiles[0].ID, list[1].ID)

	list = checkNameFollowers(t, userIDs[2], profiles[2].Name, 0, 100, 2)
	req.Equal(profiles[1].ID, list[0].ID)
	req.Equal(profiles[0].ID, list[1].ID)

	checkUnfollow(t, userIDs[0], userIDs[2])
	checkUnfollow(t, userIDs[1], userIDs[2])
}

func checkTopUsers(t *testing.T, top string, size int) []*models.Friend {
	get := api.UsersGetUsersHandler.Handle
	params := users.GetUsersParams{
		Top: &top,
	}
	resp := get(params, userIDs[0])
	body, ok := resp.(*users.GetUsersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, top, list.Top)
	require.Empty(t, list.Query)

	return list.Users
}

func TestTopUsers(t *testing.T) {
	list := checkTopUsers(t, "new", 4)

	req := require.New(t)
	req.Equal(int64(1), list[3].ID)

	checkTopUsers(t, "rank", 4)

	list = checkTopUsers(t, "waiting", 1)
	req.Equal(userIDs[3].ID, list[0].ID)
}

func checkSearchUsers(t *testing.T, query string, size int) []*models.Friend {
	get := api.UsersGetUsersHandler.Handle
	params := users.GetUsersParams{
		Query: &query,
	}
	resp := get(params, userIDs[0])
	body, ok := resp.(*users.GetUsersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, query, list.Query)
	require.Empty(t, list.Top)

	return list.Users
}

func TestSearchUsers(t *testing.T) {
	checkSearchUsers(t, "testo", 4)
	checkSearchUsers(t, "mind", 1)
	checkSearchUsers(t, "psychotherapist", 0)
}
