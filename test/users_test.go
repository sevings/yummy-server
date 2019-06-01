package test

import (
	"strings"
	"testing"

	"github.com/sevings/mindwell-server/restapi/operations/me"
	"github.com/sevings/mindwell-server/restapi/operations/users"

	"github.com/stretchr/testify/require"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/utils"
)

func TestKeyAuth(t *testing.T) {
	auth := api.APIKeyHeaderAuth
	req := require.New(t)

	for _, user := range profiles {
		id, err := auth(user.Account.APIKey)
		req.Nil(err)
		req.Equal(id.ID, int64(user.ID))
	}

	_, err := auth("12345678901234567890123456789012")
	req.NotNil(err)
}

func TestGetMe(t *testing.T) {
	req := require.New(t)

	get := func(i int) *models.AuthProfile {
		load := api.MeGetMeHandler.Handle
		resp := load(me.GetMeParams{}, userIDs[i])
		body, ok := resp.(*me.GetMeOK)
		if !ok {
			t.Fatal("error get me")
		}

		return body.Payload
	}

	for i, user := range profiles {
		me := get(i)

		req.Equal(user.ID, me.ID)
		req.Equal(user.Name, me.Name)

		req.Zero(me.Ban.Invite)
		req.Zero(me.Ban.Vote)
		req.Zero(me.Ban.Comment)
		req.Zero(me.Ban.Live)
	}

	banInvite(db, userIDs[0])
	me := get(0)
	req.NotZero(me.Ban.Invite)
	req.Zero(me.Ban.Vote)
	removeUserRestrictions(db, userIDs)

	banVote(db, userIDs[0])
	me = get(0)
	req.Zero(me.Ban.Invite)
	req.NotZero(me.Ban.Vote)
	removeUserRestrictions(db, userIDs)

	banComment(db, userIDs[0])
	me = get(0)
	req.Zero(me.Ban.Invite)
	req.Zero(me.Ban.Vote)
	req.NotZero(me.Ban.Comment)
	removeUserRestrictions(db, userIDs)

	banLive(db, userIDs[0])
	me = get(0)
	req.Zero(me.Ban.Invite)
	req.Zero(me.Ban.Vote)
	req.Zero(me.Ban.Comment)
	req.NotZero(me.Ban.Live)
	removeUserRestrictions(db, userIDs)
}

func compareUsers(t *testing.T, user *models.AuthProfile, profile *models.Profile) {
	req := require.New(t)

	req.Equal(user.ID, profile.ID)
	req.Equal(user.Name, profile.Name)
	req.Equal(user.ShowName, profile.ShowName)
	req.Equal(user.IsOnline, profile.IsOnline)
	req.Equal(user.Avatar, profile.Avatar)

	req.Equal(user.Gender, profile.Gender)
	req.Equal(user.IsDaylog, profile.IsDaylog)
	req.Equal(user.Privacy, profile.Privacy)
	req.Equal(user.Title, profile.Title)
	req.Equal(user.Rank, profile.Rank)
	req.Equal(user.CreatedAt, profile.CreatedAt)
	req.Equal(user.LastSeenAt, profile.LastSeenAt)
	req.Equal(user.InvitedBy, profile.InvitedBy)
	req.Equal(user.AgeLowerBound, profile.AgeLowerBound)
	req.Equal(user.AgeUpperBound, profile.AgeUpperBound)
	req.Equal(user.Country, profile.Country)
	req.Equal(user.City, profile.City)
	req.Equal(user.Cover, profile.Cover)
	req.NotEmpty(user.Cover)
}

func TestGetUser(t *testing.T) {
	get := api.UsersGetUsersNameHandler.Handle
	for i, user := range profiles {
		resp := get(users.GetUsersNameParams{Name: strings.ToUpper(user.Name)}, userIDs[i])
		body, ok := resp.(*users.GetUsersNameOK)
		if !ok {
			badBody, ok := resp.(*users.GetUsersNameNotFound)
			if ok {
				t.Fatal(badBody.Payload.Message)
			}

			t.Fatalf("error get user by name %s", user.Name)
		}

		compareUsers(t, user, body.Payload)
	}

	resp := get(users.GetUsersNameParams{Name: "trolol not found"}, userIDs[0])
	_, ok := resp.(*users.GetUsersNameNotFound)
	require.True(t, ok)
}

func checkEditProfile(t *testing.T, user *models.AuthProfile, params me.PutMeParams) {
	edit := api.MePutMeHandler.Handle
	id := models.UserID{
		ID:   user.ID,
		Name: user.Name,
	}
	resp := edit(params, &id)
	body, ok := resp.(*me.PutMeOK)
	require.True(t, ok)

	profile := body.Payload
	compareUsers(t, user, profile)
}

func setUserPrivacy(t *testing.T, userID *models.UserID, privacy string) {
	params := me.PutMeParams{
		Privacy: privacy,
		ShowName: userID.Name,
	}
	edit := api.MePutMeHandler.Handle
	resp := edit(params, userID)
	_, ok := resp.(*me.PutMeOK)
	require.True(t, ok)
}

func TestEditProfile(t *testing.T) {
	user := *profiles[0]
	user.AgeLowerBound = 30
	user.AgeUpperBound = 35
	user.Birthday = "1988-01-01T20:01:31.844+03:00"
	user.City = "city edit"
	user.Country = "country edit"
	user.Gender = "female"
	user.IsDaylog = true
	user.Privacy = "followers"
	user.Title = "title edit"
	user.ShowInTops = false
	user.ShowName = "showname edit"

	params := me.PutMeParams{
		Birthday:   &user.Birthday,
		City:       &user.City,
		Country:    &user.Country,
		Gender:     &user.Gender,
		IsDaylog:   &user.IsDaylog,
		Privacy:    user.Privacy,
		Title:      &user.Title,
		ShowInTops: &user.ShowInTops,
		ShowName:   user.ShowName,
	}

	checkEditProfile(t, &user, params)

	user.Privacy = "all"
	params.Privacy = user.Privacy
	checkEditProfile(t, &user, params)
}

func TestIsOpenForMe(t *testing.T) {
	req := require.New(t)

	check := func(userID *models.UserID, name string, res bool) {
		tx := utils.NewAutoTx(db)
		defer tx.Finish()
		req.Equal(res, utils.IsOpenForMe(tx, userID, name))
	}

	check(userIDs[0], userIDs[0].Name, true)
	check(userIDs[1], userIDs[0].Name, true)
	check(userIDs[2], userIDs[0].Name, true)
	check(userIDs[3], userIDs[0].Name, true)

	checkFollow(t, userIDs[1], userIDs[0], profiles[0], models.RelationshipRelationFollowed, true)
	setUserPrivacy(t, userIDs[0], "followers")
	checkFollow(t, userIDs[2], userIDs[0], profiles[0], models.RelationshipRelationRequested, true)

	check(userIDs[0], userIDs[0].Name, true)
	check(userIDs[1], userIDs[0].Name, true)
	check(userIDs[2], userIDs[0].Name, false)
	check(userIDs[3], userIDs[0].Name, false)

	setUserPrivacy(t, userIDs[0], "invited")

	check(userIDs[0], userIDs[0].Name, true)
	check(userIDs[1], userIDs[0].Name, true)
	check(userIDs[2], userIDs[0].Name, true)
	check(userIDs[3], userIDs[0].Name, false)

	checkFollow(t, userIDs[0], userIDs[1], profiles[1], models.RelationshipRelationIgnored, true)

	check(userIDs[0], userIDs[0].Name, true)
	check(userIDs[1], userIDs[0].Name, false)
	check(userIDs[2], userIDs[0].Name, true)
	check(userIDs[3], userIDs[0].Name, false)

	setUserPrivacy(t, userIDs[0], "all")

	check(userIDs[0], userIDs[0].Name, true)
	check(userIDs[1], userIDs[0].Name, false)
	check(userIDs[2], userIDs[0].Name, true)
	check(userIDs[3], userIDs[0].Name, true)

	checkUnfollow(t, userIDs[0], userIDs[1])
	checkUnfollow(t, userIDs[1], userIDs[0])
	checkUnfollow(t, userIDs[2], userIDs[0])
}
