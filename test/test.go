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
	"github.com/sevings/mindwell-server/restapi/operations/comments"
	"github.com/sevings/mindwell-server/restapi/operations/me"
)

type EmailSenderMock struct {
	Emails []string
	Codes  []string
	Dates  []int64
}

func (esm *EmailSenderMock) SendGreeting(address, name, code string) {
	esm.Emails = append(esm.Emails, address)
	esm.Codes = append(esm.Codes, code)
}

func (esm *EmailSenderMock) SendResetPassword(address, name, gender, code string, date int64) {
	esm.Emails = append(esm.Emails, address)
	esm.Codes = append(esm.Codes, code)
	esm.Dates = append(esm.Dates, date)
}

func (esm *EmailSenderMock) SendNewComment(address, fromGender, toShowName, entryTitle string, cmt *models.Comment) {
	esm.Emails = append(esm.Emails, address)
}

func (esm *EmailSenderMock) SendNewFollower(address, fromName, fromShowName, fromGender string, toPrivate bool, toShowName string) {
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

func register(name, inviteWord string) (*models.UserID, *models.AuthProfile) {
	params := account.PostAccountRegisterParams{
		Name:     name,
		Email:    name + "@example.com",
		Password: "test123",
		Invite:   inviteWord + " " + inviteWord + " " + inviteWord,
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

	inviteWords := []string{"acknown", "aery", "affectioned"}

	for i := 0; i < 3; i++ {
		id, profile := register("test"+strconv.Itoa(i), inviteWords[i])
		userIDs = append(userIDs, id)
		profiles = append(profiles, profile)

		time.Sleep(10 * time.Millisecond)
	}

	// remove user restrictions
	_, err := db.Exec("UPDATE users SET followers_count = 100")
	if err != nil {
		log.Println(err)
	}

	return userIDs, profiles
}

func createTlogEntry(t *testing.T, id *models.UserID, privacy string, votable, live bool) *models.Entry {
	title := ""
	params := me.PostMeTlogParams{
		Content:   "test test test",
		Title:     &title,
		Privacy:   privacy,
		IsVotable: &votable,
		InLive:    &live,
	}

	resp := api.MePostMeTlogHandler.Handle(params, id)
	body, ok := resp.(*me.PostMeTlogCreated)
	require.True(t, ok)

	return body.Payload
}

func createComment(t *testing.T, id *models.UserID, entryID int64) *models.Comment {
	params := comments.PostEntriesIDCommentsParams{
		ID:      entryID,
		Content: "test comment",
	}

	post := api.CommentsPostEntriesIDCommentsHandler.Handle
	resp := post(params, id)
	body, ok := resp.(*comments.PostEntriesIDCommentsCreated)
	require.True(t, ok)

	return body.Payload
}
