package test

import (
	"database/sql"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations/account"
	"github.com/sevings/yummy-server/restapi/operations/entries"
)

func register(name string) (*models.UserID, *models.AuthProfile) {
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
	return &userID, body.Payload
}

func registerTestUsers(db *sql.DB) ([]*models.UserID, []*models.AuthProfile) {
	var userIDs []*models.UserID
	var profiles []*models.AuthProfile

	for i := 0; i < 3; i++ {
		id, profile := register("test" + strconv.Itoa(i))
		userIDs = append(userIDs, id)
		profiles = append(profiles, profile)

		time.Sleep(10 * time.Millisecond)
	}

	return userIDs, profiles
}

func createTlogEntry(t *testing.T, id *models.UserID, privacy string, votable bool) *models.Entry {
	title := ""
	params := entries.PostEntriesUsersMeParams{
		Content:   "test test test",
		Title:     &title,
		Privacy:   &privacy,
		IsVotable: &votable,
	}

	resp := api.EntriesPostEntriesUsersMeHandler.Handle(params, id)
	body, ok := resp.(*entries.PostEntriesUsersMeOK)
	require.True(t, ok)

	return body.Payload
}
