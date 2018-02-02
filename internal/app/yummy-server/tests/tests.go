package tests

import (
	"database/sql"
	"log"

	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"
	"github.com/sevings/yummy-server/restapi/operations/account"

	accountImpl "github.com/sevings/yummy-server/internal/app/yummy-server/account"
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
		id, profile := register(&api, "test"+string(i))
		userIDs = append(userIDs, id)
		profiles = append(profiles, profile)
	}

	return userIDs, profiles
}
