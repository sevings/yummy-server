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

func checkMyFollowers(t *testing.T, user *models.UserID, after, before string, limit int64, size int) *models.FriendList {
	get := api.MeGetMeFollowersHandler.Handle
	params := me.GetMeFollowersParams{
		After:  &after,
		Before: &before,
		Limit:  &limit,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersNameFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, user.ID, list.Subject.ID)
	require.Equal(t, models.FriendListRelationFollowers, list.Relation)

	return list
}

func checkMyFollowings(t *testing.T, user *models.UserID, after, before string, limit int64, size int) *models.FriendList {
	get := api.MeGetMeFollowingsHandler.Handle
	params := me.GetMeFollowingsParams{
		After:  &after,
		Before: &before,
		Limit:  &limit,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersNameFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, user.ID, list.Subject.ID)
	require.Equal(t, models.FriendListRelationFollowings, list.Relation)

	return list
}

func checkMyIgnored(t *testing.T, user *models.UserID, after, before string, limit int64, size int) *models.FriendList {
	get := api.MeGetMeIgnoredHandler.Handle
	params := me.GetMeIgnoredParams{
		After:  &after,
		Before: &before,
		Limit:  &limit,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersNameFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, user.ID, list.Subject.ID)
	require.Equal(t, models.FriendListRelationIgnored, list.Relation)

	return list
}

func checkMyHidden(t *testing.T, user *models.UserID, after, before string, limit int64, size int) *models.FriendList {
	get := api.MeGetMeHiddenHandler.Handle
	params := me.GetMeHiddenParams{
		After:  &after,
		Before: &before,
		Limit:  &limit,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersNameFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, user.ID, list.Subject.ID)
	require.Equal(t, models.FriendListRelationHidden, list.Relation)

	return list
}

func checkMyInvited(t *testing.T, user *models.UserID, after, before string, limit int64, size int) *models.FriendList {
	get := api.MeGetMeInvitedHandler.Handle
	params := me.GetMeInvitedParams{
		After:  &after,
		Before: &before,
		Limit:  &limit,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersNameFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, user.ID, list.Subject.ID)
	require.Equal(t, models.FriendListRelationInvited, list.Relation)

	return list
}

func checkNameFollowersForbidden(t *testing.T, user *models.UserID, name string) {
	get := api.UsersGetUsersNameFollowersHandler.Handle
	params := users.GetUsersNameFollowersParams{
		After:  new(string),
		Before: new(string),
		Limit:  new(int64),
		Name:   name,
	}
	resp := get(params, user)
	_, ok := resp.(*users.GetUsersNameFollowersForbidden)

	require.True(t, ok)
}

func checkNameFollowers(t *testing.T, user *models.UserID, name, after, before string, limit int64, size int) *models.FriendList {
	get := api.UsersGetUsersNameFollowersHandler.Handle
	params := users.GetUsersNameFollowersParams{
		After:  &after,
		Before: &before,
		Limit:  &limit,
		Name:   name,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersNameFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, strings.ToLower(name), strings.ToLower(list.Subject.Name))
	require.Equal(t, models.FriendListRelationFollowers, list.Relation)

	return list
}

func checkNameFollowings(t *testing.T, user *models.UserID, name, after, before string, limit int64, size int) *models.FriendList {
	get := api.UsersGetUsersNameFollowingsHandler.Handle
	params := users.GetUsersNameFollowingsParams{
		After:  &after,
		Before: &before,
		Limit:  &limit,
		Name:   name,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersNameFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, strings.ToLower(name), strings.ToLower(list.Subject.Name))
	require.Equal(t, models.FriendListRelationFollowings, list.Relation)

	return list
}

func checkNameInvited(t *testing.T, user *models.UserID, name, after, before string, limit int64, size int) *models.FriendList {
	get := api.UsersGetUsersNameInvitedHandler.Handle
	params := users.GetUsersNameInvitedParams{
		After:  &after,
		Before: &before,
		Limit:  &limit,
		Name:   name,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersNameFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, strings.ToLower(name), strings.ToLower(list.Subject.Name))
	require.Equal(t, models.FriendListRelationInvited, list.Relation)

	return list
}

func TestOpenFriendLists(t *testing.T) {
	checkFollow(t, userIDs[0], userIDs[1], profiles[1], models.RelationshipRelationFollowed, true)
	time.Sleep(10 * time.Millisecond)
	checkFollow(t, userIDs[0], userIDs[2], profiles[2], models.RelationshipRelationFollowed, true)
	time.Sleep(10 * time.Millisecond)
	checkFollow(t, userIDs[1], userIDs[2], profiles[2], models.RelationshipRelationFollowed, true)

	req := require.New(t)
	var list *models.FriendList

	checkMyFollowers(t, userIDs[0], "", "", 100, 0)

	list = checkMyFollowers(t, userIDs[1], "", "", 100, 1)
	req.Equal(profiles[0].ID, list.Users[0].ID)

	checkMyFollowings(t, userIDs[2], "", "", 100, 0)

	list = checkMyFollowings(t, userIDs[0], "", "", 100, 2)
	req.Equal(profiles[2].ID, list.Users[0].ID)
	req.Equal(profiles[1].ID, list.Users[1].ID)

	list = checkMyFollowings(t, userIDs[0], "", "", 1, 1)
	req.Equal(profiles[2].ID, list.Users[0].ID)
	req.False(list.HasAfter)
	req.True(list.HasBefore)

	list = checkMyFollowings(t, userIDs[0], "", list.NextBefore, 10, 1)
	req.Equal(profiles[1].ID, list.Users[0].ID)
	req.True(list.HasAfter)
	req.False(list.HasBefore)

	checkNameFollowers(t, userIDs[0], profiles[0].Name, "", "", 100, 0)

	list = checkNameFollowers(t, userIDs[0], profiles[1].Name, "", "", 100, 1)
	req.Equal(profiles[0].ID, list.Users[0].ID)

	checkNameFollowings(t, userIDs[0], profiles[2].Name, "", "", 100, 0)

	list = checkNameFollowings(t, userIDs[0], strings.ToLower(profiles[0].Name), "", "", 1, 1)
	req.Equal(profiles[2].ID, list.Users[0].ID)
	req.False(list.HasAfter)
	req.True(list.HasBefore)

	list = checkNameFollowings(t, userIDs[0], strings.ToUpper(profiles[0].Name), "", list.NextBefore, 1, 1)
	req.Equal(profiles[1].ID, list.Users[0].ID)
	req.True(list.HasAfter)
	req.False(list.HasBefore)

	list = checkNameFollowings(t, userIDs[0], strings.ToLower(profiles[0].Name), list.NextAfter, "", 1, 1)
	req.Equal(profiles[2].ID, list.Users[0].ID)
	req.False(list.HasAfter)
	req.True(list.HasBefore)

	checkUnfollow(t, userIDs[0], userIDs[1])
	checkUnfollow(t, userIDs[0], userIDs[2])
	checkUnfollow(t, userIDs[1], userIDs[2])

	checkFollow(t, userIDs[0], userIDs[1], profiles[1], models.RelationshipRelationIgnored, true)
	time.Sleep(10 * time.Millisecond)
	checkFollow(t, userIDs[0], userIDs[2], profiles[2], models.RelationshipRelationIgnored, true)

	checkMyIgnored(t, userIDs[2], "", "", 100, 0)

	list = checkMyIgnored(t, userIDs[0], "", "", 100, 2)
	req.Equal(profiles[2].ID, list.Users[0].ID)
	req.Equal(profiles[1].ID, list.Users[1].ID)
	req.False(list.HasAfter)
	req.False(list.HasBefore)

	list = checkMyIgnored(t, userIDs[0], "", "", 1, 1)
	req.Equal(profiles[2].ID, list.Users[0].ID)

	checkUnfollow(t, userIDs[0], userIDs[1])
	checkUnfollow(t, userIDs[0], userIDs[2])

	checkMyInvited(t, userIDs[2], "", "", 100, 0)

	inviter := models.UserID{
		ID:   1,
		Name: "mindwell",
	}
	list = checkMyInvited(t, &inviter, "", "", 100, 4)
	req.Equal(profiles[2].ID, list.Users[0].ID)
	req.Equal(profiles[1].ID, list.Users[1].ID)
	req.Equal(profiles[0].ID, list.Users[2].ID)
	req.Equal(int64(1), list.Users[3].ID)

	list = checkMyInvited(t, &inviter, "", "", 2, 2)
	req.Equal(profiles[2].ID, list.Users[0].ID)
	req.Equal(profiles[1].ID, list.Users[1].ID)

	list = checkNameInvited(t, userIDs[0], "minDWEll", "", "", 100, 4)
	req.Equal(profiles[2].ID, list.Users[0].ID)
	req.Equal(profiles[1].ID, list.Users[1].ID)
	req.Equal(profiles[0].ID, list.Users[2].ID)
	req.Equal(int64(1), list.Users[3].ID)

	checkFollow(t, userIDs[0], userIDs[1], profiles[1], models.RelationshipRelationHidden, true)
	checkFollow(t, userIDs[0], userIDs[2], profiles[2], models.RelationshipRelationHidden, true)
	checkMyHidden(t, userIDs[0], "", "", 10, 2)
	checkUnfollow(t, userIDs[0], userIDs[1])
	checkUnfollow(t, userIDs[0], userIDs[2])
}

func TestPrivateFriendLists(t *testing.T) {
	checkFollow(t, userIDs[0], userIDs[2], profiles[2], models.RelationshipRelationFollowed, true)
	time.Sleep(10 * time.Millisecond)
	checkFollow(t, userIDs[1], userIDs[2], profiles[2], models.RelationshipRelationFollowed, true)

	setUserPrivacy(t, userIDs[2], "followers")

	req := require.New(t)
	var list *models.FriendList

	list = checkNameFollowers(t, userIDs[0], profiles[2].Name, "", "", 100, 2)
	req.Equal(profiles[1].ID, list.Users[0].ID)
	req.Equal(profiles[0].ID, list.Users[1].ID)

	list = checkNameFollowers(t, userIDs[2], profiles[2].Name, "", "", 100, 2)
	req.Equal(profiles[1].ID, list.Users[0].ID)
	req.Equal(profiles[0].ID, list.Users[1].ID)

	checkUnfollow(t, userIDs[0], userIDs[2])
	checkUnfollow(t, userIDs[1], userIDs[2])

	checkNameFollowersForbidden(t, userIDs[0], userIDs[2].Name)

	setUserPrivacy(t, userIDs[2], "all")
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
