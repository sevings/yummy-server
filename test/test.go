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

func (esm *EmailSenderMock) SendPasswordChanged(address, name string) {
	esm.Emails = append(esm.Emails, address)
}

func (esm *EmailSenderMock) SendEmailChanged(address, name string) {
	esm.Emails = append(esm.Emails, address)
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

func (esm *EmailSenderMock) SendNewAccept(address, fromName, fromShowName, fromGender, toShowName string) {
	esm.Emails = append(esm.Emails, address)
}

func (esm *EmailSenderMock) SendNewWelcome(address, fromName, fromShowName, fromGender, toShowName string) {
	esm.Emails = append(esm.Emails, address)
}

func (esm *EmailSenderMock) SendNewInvite(address, name string) {
	esm.Emails = append(esm.Emails, address)
}

func (esm *EmailSenderMock) CheckEmail(t *testing.T, email string) {
	req := require.New(t)
	req.Equal(1, len(esm.Emails))
	req.Equal(email, esm.Emails[0])

	esm.Clear()
}

func (esm *EmailSenderMock) CheckEmail2(t *testing.T, email0, email1 string) {
	req := require.New(t)
	req.Equal(2, len(esm.Emails))
	req.Equal(email0, esm.Emails[0])
	req.Equal(email1, esm.Emails[1])

	esm.Clear()
}

func (esm *EmailSenderMock) Clear() {
	esm.Emails = nil
	esm.Codes = nil
	esm.Dates = nil
}

func register(name, inviteWord string) (*models.UserID, *models.AuthProfile) {
	invite := inviteWord + " " + inviteWord + " " + inviteWord
	params := account.PostAccountRegisterParams{
		Name:     name,
		Email:    name + "@example.com",
		Password: "test123",
		Invite:   &invite,
	}

	if len(inviteWord) == 0 {
		params.Invite = nil
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
		ID:             body.Payload.ID,
		Name:           body.Payload.Name,
		IsInvited:      true,
		NegKarma:       false,
		FollowersCount: 0,
		Ban: &models.UserIDBan{
			Invite:  false,
			Comment: false,
			Live:    false,
			Vote:    true,
		},
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

	{
		id, profile := register("test3", "")
		userIDs = append(userIDs, id)
		profiles = append(profiles, profile)
	}

	removeUserRestrictions(db, userIDs)

	return userIDs, profiles
}

func removeUserRestrictions(db *sql.DB, userIDs []*models.UserID) {
	_, err := db.Exec(`UPDATE users 
	SET followers_count = 100, 
		vote_ban = CURRENT_DATE, 
		invite_ban = CURRENT_DATE, 
		comment_ban = CURRENT_DATE,
		live_ban = CURRENT_DATE`)
	if err != nil {
		log.Println(err)
	}

	for _, user := range userIDs {
		user.FollowersCount = 100
		user.Ban.Invite = false
		user.Ban.Comment = false
		user.Ban.Live = false
		user.Ban.Vote = false
	}
}

func banVote(db *sql.DB, userID *models.UserID) {
	_, err := db.Exec("UPDATE users SET vote_ban = CURRENT_DATE + interval '1 day' WHERE id = $1", userID.ID)
	if err != nil {
		log.Println(err)
	}

	userID.Ban.Vote = true
}

func banInvite(db *sql.DB, userID *models.UserID) {
	_, err := db.Exec("UPDATE users SET invite_ban = CURRENT_DATE + interval '1 day' WHERE id = $1", userID.ID)
	if err != nil {
		log.Println(err)
	}

	userID.Ban.Invite = true
}

func banComment(db *sql.DB, userID *models.UserID) {
	_, err := db.Exec("UPDATE users SET comment_ban = CURRENT_DATE + interval '1 day' WHERE id = $1", userID.ID)
	if err != nil {
		log.Println(err)
	}

	userID.Ban.Comment = true
}

func banLive(db *sql.DB, userID *models.UserID) {
	_, err := db.Exec("UPDATE users SET live_ban = CURRENT_DATE + interval '1 day' WHERE id = $1", userID.ID)
	if err != nil {
		log.Println(err)
	}

	userID.Ban.Live = true
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
