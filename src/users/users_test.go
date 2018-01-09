package users

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/sevings/yummy-server/gen/restapi/operations/me"
	"github.com/sevings/yummy-server/gen/restapi/operations/users"

	"github.com/stretchr/testify/require"

	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations"
	"github.com/sevings/yummy-server/gen/restapi/operations/account"
	accountImpl "github.com/sevings/yummy-server/src/account"

	yummy "github.com/sevings/yummy-server/src"
)

var db *sql.DB
var userIDs []*models.UserID
var profiles []*models.AuthProfile

func TestMain(m *testing.M) {
	config := yummy.LoadConfig("../../config")
	db = yummy.OpenDatabase(config)
	yummy.ClearDatabase(db)

	api := operations.YummyAPI{}
	accountImpl.ConfigureAPI(db, &api)

	register(&api, "test1")
	register(&api, "test2")
	register(&api, "test3")

	os.Exit(m.Run())
}

func register(api *operations.YummyAPI, name string) {
	params := account.PostAccountRegisterParams{
		Name:     name,
		Email:    name,
		Password: "test123",
		Invite:   "acknown acknown acknown",
		Referrer: "HaveANiceDay",
	}

	resp := api.AccountPostAccountRegisterHandler.Handle(params)
	body, ok := resp.(*account.PostAccountRegisterOK)
	if !ok {
		badBody, ok := resp.(*account.PostAccountRegisterBadRequest)
		if ok {
			log.Fatal(badBody.Payload.Message)
		}

		log.Fatal("reg error")
	}

	userID := models.UserID(body.Payload.ID)
	userIDs = append(userIDs, &userID)
	profiles = append(profiles, body.Payload)
}

func TestKeyAuth(t *testing.T) {
	auth := newKeyAuth(db)
	req := require.New(t)

	for _, user := range profiles {
		id, err := auth(user.Account.APIKey)
		req.Nil(err)
		req.Equal(int64(*id), int64(user.ID))
	}

	_, err := auth("12345678901234567890123456789012")
	req.NotNil(err)
}

func TestGetMe(t *testing.T) {
	load := newMeLoader(db)
	req := require.New(t)

	for i, user := range profiles {
		resp := load(me.GetUsersMeParams{}, userIDs[i])
		body, ok := resp.(*me.GetUsersMeOK)
		if !ok {
			badBody, ok := resp.(*me.GetUsersMeForbidden)
			if ok {
				t.Fatal(badBody.Payload.Message)
			}

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
}

func TestGetUserByID(t *testing.T) {
	get := newUserLoader(db)
	for i, user := range profiles {
		resp := get(users.GetUsersIDParams{ID: user.ID}, userIDs[i])
		body, ok := resp.(*users.GetUsersIDOK)
		if !ok {
			badBody, ok := resp.(*users.GetUsersIDNotFound)
			if ok {
				t.Fatal(badBody.Payload.Message)
			}

			t.Fatalf("error get user by id %d", user.ID)
		}

		compareUsers(t, user, body.Payload)
	}

	resp := get(users.GetUsersIDParams{ID: 1e9}, userIDs[0])
	_, ok := resp.(*users.GetUsersIDNotFound)
	require.True(t, ok)
}

func TestGetUserByName(t *testing.T) {
	get := newUserLoaderByName(db)
	for i, user := range profiles {
		resp := get(users.GetUsersByNameNameParams{Name: user.Name}, userIDs[i])
		body, ok := resp.(*users.GetUsersIDOK) // not GetUsersByNameNameOK
		if !ok {
			badBody, ok := resp.(*users.GetUsersIDNotFound) // not GetUsersByNameNameNotFound
			if ok {
				t.Fatal(badBody.Payload.Message)
			}

			t.Fatalf("error get user by name %s", user.Name)
		}

		compareUsers(t, user, body.Payload)
	}

	resp := get(users.GetUsersByNameNameParams{Name: "trolol not found"}, userIDs[0])
	_, ok := resp.(*users.GetUsersIDNotFound)
	require.True(t, ok)
}
