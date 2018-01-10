package yummy

import (
	"database/sql"
	"log"

	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations"
	"github.com/sevings/yummy-server/gen/restapi/operations/account"
)

func dropTable(tx *sql.Tx, table string) {
	_, err := tx.Exec("delete from " + table)
	if err != nil {
		tx.Rollback()
		log.Fatal("cannot clear table " + table + ": " + err.Error())
	}
}

// ClearDatabase drops user data tables and then creates default user
func ClearDatabase(db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("cannot begin tx")
	}

	_, err = tx.Exec("delete from users where id != 1")
	if err != nil {
		tx.Rollback()
		log.Fatal("cannot clear table users: " + err.Error())
	}

	dropTable(tx, "comment_votes")
	dropTable(tx, "comments")
	dropTable(tx, "entries")
	dropTable(tx, "entries_privacy")
	dropTable(tx, "entry_tags")
	dropTable(tx, "entry_votes")
	dropTable(tx, "favorites")
	dropTable(tx, "invites")
	dropTable(tx, "relations")
	dropTable(tx, "tags")
	dropTable(tx, "watching")

	for i := 0; i < 3; i++ {
		_, err = tx.Exec("INSERT INTO invites (referrer_id, word1, word2, word3) VALUES(1, 1, 1, 1);")
		if err != nil {
			tx.Rollback()
			log.Fatal("cannot create invite")
		}
	}

	tx.Commit()
}

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
	account.ConfigureAPI(db, &api)

	var userIDs []*models.UserID
	var profiles []*models.AuthProfile

	for i := 0; i < 3; i++ {
		id, profile := register(&api, "test"+string(i))
		userIDs = append(userIDs, id)
		profiles = append(profiles, profile)
	}

	return userIDs, profiles
}
