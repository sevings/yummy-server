package tests

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"
	"github.com/sevings/yummy-server/restapi/operations/account"
	"github.com/sevings/yummy-server/restapi/operations/entries"

	accountImpl "github.com/sevings/yummy-server/internal/app/yummy-server/account"
	entriesImpl "github.com/sevings/yummy-server/internal/app/yummy-server/entries"
)

func register(api *operations.YummyAPI, name string) (*models.UserID, *models.AuthProfile) {
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

// RegisterTestUsers creates 3 test users: test1, test2, test3
func RegisterTestUsers(db *sql.DB) ([]*models.UserID, []*models.AuthProfile) {
	api := operations.YummyAPI{}
	accountImpl.ConfigureAPI(db, &api)

	var userIDs []*models.UserID
	var profiles []*models.AuthProfile

	for i := 0; i < 3; i++ {
		id, profile := register(&api, "test"+strconv.Itoa(i))
		userIDs = append(userIDs, id)
		profiles = append(profiles, profile)
	}

	return userIDs, profiles
}

func postEntry(api *operations.YummyAPI, id *models.UserID, privacy string, votable bool) *models.Entry {
	title := ""
	params := entries.PostEntriesUsersMeParams{
		Content:   "test test test",
		Title:     &title,
		Privacy:   &privacy,
		IsVotable: &votable,
	}

	resp := api.EntriesPostEntriesUsersMeHandler.Handle(params, id)
	body, ok := resp.(*entries.PostEntriesUsersMeOK)
	if !ok {
		badBody, ok := resp.(*entries.PostEntriesUsersMeForbidden)
		if ok {
			log.Fatal(badBody.Payload.Message)
		}

		log.Fatal("error post entry")
	}

	return body.Payload
}

// NewPostEntry returns func creating entries
func NewPostEntry(db *sql.DB) func(id *models.UserID, privacy string, votable bool) *models.Entry {
	api := operations.YummyAPI{}
	entriesImpl.ConfigureAPI(db, &api)

	return func(id *models.UserID, privacy string, votable bool) *models.Entry {
		return postEntry(&api, id, privacy, votable)
	}
}
