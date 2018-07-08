package test

import (
	"database/sql"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/account"
	"github.com/sevings/mindwell-server/restapi/operations/me"
)

type EmailSenderMock struct {
	Emails []string
	Codes  []string
}

func (esm *EmailSenderMock) SendGreeting(address, name, code string) {
	esm.Emails = append(esm.Emails, address)
	esm.Codes = append(esm.Codes, code)
}

func (esm *EmailSenderMock) SendNewComment(address, name, gender, entryTitle string, cmt *models.Comment) {
	esm.Emails = append(esm.Emails, address)
}

func (esm *EmailSenderMock) SendNewFollower(address, name string, isPrivate bool, hisShowName, hisName, gender string) {
	esm.Emails = append(esm.Emails, address)
}

func (esm *EmailSenderMock) CheckEmail(t *testing.T, email string) {
	req := require.New(t)
	req.Equal(1, len(esm.Emails))
	req.Equal(email, esm.Emails[0])

	esm.Clear()
}

func (esm *EmailSenderMock) Clear() {
	esm.Emails = nil
	esm.Codes = nil
}

func register(name string) (*models.UserID, *models.AuthProfile) {
	params := account.PostAccountRegisterParams{
		Name:     name,
		Email:    name,
		Password: "test123",
		Invite:   "acknown acknown acknown",
		Referrer: "HaveANiceDay",
	}

	resp := api.AccountPostAccountRegisterHandler.Handle(params)
	body, ok := resp.(*account.PostAccountRegisterCreated)
	if !ok {
		badBody, ok := resp.(*account.PostAccountRegisterBadRequest)
		if ok {
			log.Fatal(badBody.Payload.Message)
		}

		log.Fatal("reg error")
	}

	userID := models.UserID{
		ID:   body.Payload.ID,
		Name: body.Payload.Name,
	}
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
	params := me.PostMeTlogParams{
		Content:   "test test test",
		Title:     &title,
		Privacy:   privacy,
		IsVotable: &votable,
	}

	resp := api.MePostMeTlogHandler.Handle(params, id)
	body, ok := resp.(*me.PostMeTlogCreated)
	require.True(t, ok)

	return body.Payload
}
