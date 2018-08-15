package test

import (
	"strings"
	"testing"

	"github.com/sevings/mindwell-server/restapi/operations/me"
	"github.com/sevings/mindwell-server/restapi/operations/users"

	"github.com/stretchr/testify/require"

	"github.com/sevings/mindwell-server/models"
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
	load := api.MeGetMeHandler.Handle
	req := require.New(t)

	for i, user := range profiles {
		resp := load(me.GetMeParams{}, userIDs[i])
		body, ok := resp.(*me.GetMeOK)
		if !ok {
			t.Fatal("error get me")
		}

		me := body.Payload

		req.Equal(user.ID, me.ID)
		req.Equal(user.Name, me.Name)
	}
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
	req.Equal(user.Karma, profile.Karma)
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
		body, ok := resp.(*users.GetUsersNameOK) // not GetUsersNameOK
		if !ok {
			badBody, ok := resp.(*users.GetUsersNameNotFound) // not GetUsersNameNotFound
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
