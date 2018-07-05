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

func checkMyFollowers(t *testing.T, user *models.UserID, skip, limit int64, size int) models.FriendListUsers {
	get := api.MeGetUsersMeFollowersHandler.Handle
	params := me.GetUsersMeFollowersParams{
		Skip:  &skip,
		Limit: &limit,
	}
	resp := get(params, user)
	body, ok := resp.(*me.GetUsersMeFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, int64(*user), list.Subject.ID)
	require.Equal(t, models.FriendListRelationFollowers, list.Relation)

	return list.Users
}

func checkMyFollowings(t *testing.T, user *models.UserID, skip, limit int64, size int) models.FriendListUsers {
	get := api.MeGetUsersMeFollowingsHandler.Handle
	params := me.GetUsersMeFollowingsParams{
		Skip:  &skip,
		Limit: &limit,
	}
	resp := get(params, user)
	body, ok := resp.(*me.GetUsersMeFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, int64(*user), list.Subject.ID)
	require.Equal(t, models.FriendListRelationFollowings, list.Relation)

	return list.Users
}

func checkMyIgnored(t *testing.T, user *models.UserID, skip, limit int64, size int) models.FriendListUsers {
	get := api.MeGetUsersMeIgnoredHandler.Handle
	params := me.GetUsersMeIgnoredParams{
		Skip:  &skip,
		Limit: &limit,
	}
	resp := get(params, user)
	body, ok := resp.(*me.GetUsersMeFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, int64(*user), list.Subject.ID)
	require.Equal(t, models.FriendListRelationIgnored, list.Relation)

	return list.Users
}

func checkMyInvited(t *testing.T, user *models.UserID, skip, limit int64, size int) models.FriendListUsers {
	get := api.MeGetUsersMeInvitedHandler.Handle
	params := me.GetUsersMeInvitedParams{
		Skip:  &skip,
		Limit: &limit,
	}
	resp := get(params, user)
	body, ok := resp.(*me.GetUsersMeFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, int64(*user), list.Subject.ID)
	require.Equal(t, models.FriendListRelationInvited, list.Relation)

	return list.Users
}

func checkIDFollowers(t *testing.T, user *models.UserID, id, skip, limit int64, size int) models.FriendListUsers {
	get := api.UsersGetUsersIDFollowersHandler.Handle
	params := users.GetUsersIDFollowersParams{
		Skip:  &skip,
		Limit: &limit,
		ID:    id,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersIDFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, id, list.Subject.ID)
	require.Equal(t, models.FriendListRelationFollowers, list.Relation)

	return list.Users
}

func checkIDFollowings(t *testing.T, user *models.UserID, id, skip, limit int64, size int) models.FriendListUsers {
	get := api.UsersGetUsersIDFollowingsHandler.Handle
	params := users.GetUsersIDFollowingsParams{
		Skip:  &skip,
		Limit: &limit,
		ID:    id,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersIDFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, id, list.Subject.ID)
	require.Equal(t, models.FriendListRelationFollowings, list.Relation)

	return list.Users
}

func checkIDInvited(t *testing.T, user *models.UserID, id, skip, limit int64, size int) models.FriendListUsers {
	get := api.UsersGetUsersIDInvitedHandler.Handle
	params := users.GetUsersIDInvitedParams{
		Skip:  &skip,
		Limit: &limit,
		ID:    id,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersIDFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, id, list.Subject.ID)
	require.Equal(t, models.FriendListRelationInvited, list.Relation)

	return list.Users
}

func checkNameFollowers(t *testing.T, user *models.UserID, name string, skip, limit int64, size int) models.FriendListUsers {
	get := api.UsersGetUsersByNameNameFollowersHandler.Handle
	params := users.GetUsersByNameNameFollowersParams{
		Skip:  &skip,
		Limit: &limit,
		Name:  name,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersIDFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, strings.ToLower(name), strings.ToLower(list.Subject.Name))
	require.Equal(t, models.FriendListRelationFollowers, list.Relation)

	return list.Users
}

func checkNameFollowings(t *testing.T, user *models.UserID, name string, skip, limit int64, size int) models.FriendListUsers {
	get := api.UsersGetUsersByNameNameFollowingsHandler.Handle
	params := users.GetUsersByNameNameFollowingsParams{
		Skip:  &skip,
		Limit: &limit,
		Name:  name,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersIDFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, strings.ToLower(name), strings.ToLower(list.Subject.Name))
	require.Equal(t, models.FriendListRelationFollowings, list.Relation)

	return list.Users
}

func checkNameInvited(t *testing.T, user *models.UserID, name string, skip, limit int64, size int) models.FriendListUsers {
	get := api.UsersGetUsersByNameNameInvitedHandler.Handle
	params := users.GetUsersByNameNameInvitedParams{
		Skip:  &skip,
		Limit: &limit,
		Name:  name,
	}
	resp := get(params, user)
	body, ok := resp.(*users.GetUsersIDFollowersOK)

	require.True(t, ok)

	list := body.Payload
	require.Equal(t, size, len(list.Users))
	require.Equal(t, strings.ToLower(name), strings.ToLower(list.Subject.Name))
	require.Equal(t, models.FriendListRelationInvited, list.Relation)

	return list.Users
}

func TestOpenFriendLists(t *testing.T) {
	checkFollow(t, userIDs[0], profiles[1], models.RelationshipRelationFollowed)
	time.Sleep(10 * time.Millisecond)
	checkFollow(t, userIDs[0], profiles[2], models.RelationshipRelationFollowed)
	time.Sleep(10 * time.Millisecond)
	checkFollow(t, userIDs[1], profiles[2], models.RelationshipRelationFollowed)

	req := require.New(t)
	var list models.FriendListUsers

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

	checkIDFollowers(t, userIDs[0], profiles[0].ID, 0, 100, 0)

	list = checkIDFollowers(t, userIDs[0], profiles[1].ID, 0, 100, 1)
	req.Equal(profiles[0].ID, list[0].ID)

	checkIDFollowings(t, userIDs[0], profiles[2].ID, 0, 100, 0)

	list = checkIDFollowings(t, userIDs[0], profiles[0].ID, 0, 1, 1)
	req.Equal(profiles[2].ID, list[0].ID)
	list = checkIDFollowings(t, userIDs[0], profiles[0].ID, 1, 1, 1)
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

	checkFollow(t, userIDs[0], profiles[1], models.RelationshipRelationIgnored)
	time.Sleep(10 * time.Millisecond)
	checkFollow(t, userIDs[0], profiles[2], models.RelationshipRelationIgnored)

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

	inviter := models.UserID(1)
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

	list = checkIDInvited(t, userIDs[0], 1, 0, 100, 4)
	req.Equal(profiles[2].ID, list[0].ID)
	req.Equal(profiles[1].ID, list[1].ID)
	req.Equal(profiles[0].ID, list[2].ID)
	req.Equal(int64(1), list[3].ID)

	list = checkNameInvited(t, userIDs[0], "haveANiceday", 0, 100, 4)
	req.Equal(profiles[2].ID, list[0].ID)
	req.Equal(profiles[1].ID, list[1].ID)
	req.Equal(profiles[0].ID, list[2].ID)
	req.Equal(int64(1), list[3].ID)
}

func TestPrivateFriendLists(t *testing.T) {
	checkFollow(t, userIDs[0], profiles[2], models.RelationshipRelationFollowed)
	time.Sleep(10 * time.Millisecond)
	checkFollow(t, userIDs[1], profiles[2], models.RelationshipRelationFollowed)

	profiles[2].Privacy = models.ProfileAllOf1PrivacyFollowers

	params := me.PutUsersMeParams{
		Privacy:  profiles[2].Privacy,
		ShowName: profiles[2].ShowName,
	}
	checkEditProfile(t, profiles[2], params)

	req := require.New(t)
	var list models.FriendListUsers

	list = checkIDFollowers(t, userIDs[0], profiles[2].ID, 0, 100, 2)
	req.Equal(profiles[1].ID, list[0].ID)
	req.Equal(profiles[0].ID, list[1].ID)

	list = checkNameFollowers(t, userIDs[0], profiles[2].Name, 0, 100, 2)
	req.Equal(profiles[1].ID, list[0].ID)
	req.Equal(profiles[0].ID, list[1].ID)

	checkUnfollow(t, userIDs[0], userIDs[2])
	checkUnfollow(t, userIDs[1], userIDs[2])
}
