package test

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	accountImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/account"
	commentsImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/comments"
	designImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/design"
	entriesImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/entries"
	favoritesImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/favorites"
	relationsImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/relations"
	usersImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/users"
	votesImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/votes"
	watchingsImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/watchings"
	"github.com/sevings/mindwell-server/utils"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations"
	"github.com/sevings/mindwell-server/restapi/operations/account"
	"github.com/stretchr/testify/require"
)

var api *operations.MindwellAPI
var db *sql.DB
var userIDs []*models.UserID
var profiles []*models.AuthProfile

func TestMain(m *testing.M) {
	config := utils.LoadConfig("../configs/server")
	db = utils.OpenDatabase(config)
	utils.ClearDatabase(db)

	api = &operations.MindwellAPI{}

	accountImpl.ConfigureAPI(db, api)
	usersImpl.ConfigureAPI(db, api)
	entriesImpl.ConfigureAPI(db, api)
	votesImpl.ConfigureAPI(db, api)
	favoritesImpl.ConfigureAPI(db, api)
	watchingsImpl.ConfigureAPI(db, api)
	commentsImpl.ConfigureAPI(db, api)
	designImpl.ConfigureAPI(db, api)
	relationsImpl.ConfigureAPI(db, api)

	userIDs, profiles = registerTestUsers(db)

	os.Exit(m.Run())
}

func checkEmail(t *testing.T, email string, free bool) {
	check := api.AccountGetAccountEmailEmailHandler.Handle
	resp := check(account.GetAccountEmailEmailParams{Email: email})
	body, ok := resp.(*account.GetAccountEmailEmailOK)

	require.True(t, ok, email)
	require.Equal(t, email, *body.Payload.Email)
	require.Equal(t, free, *body.Payload.IsFree, email)
}

func TestCheckEmail(t *testing.T) {
	checkEmail(t, "123", true)
}

func checkName(t *testing.T, name string, free bool) {
	check := api.AccountGetAccountNameNameHandler.Handle
	resp := check(account.GetAccountNameNameParams{Name: name})
	body, ok := resp.(*account.GetAccountNameNameOK)

	require.True(t, ok, name)
	require.Equal(t, name, *body.Payload.Name)
	require.Equal(t, free, *body.Payload.IsFree, name)
}

func TestCheckName(t *testing.T) {
	checkName(t, "HaveANICEDay", false)
	checkName(t, "nAMe", true)
}

func checkInvites(t *testing.T, userID int64, size int) {
	load := api.AccountGetAccountInvitesHandler.Handle
	id := models.UserID(userID)
	resp := load(account.GetAccountInvitesParams{}, &id)
	body, ok := resp.(*account.GetAccountInvitesOK)

	require.True(t, ok, "user %d", userID)
	require.Equal(t, size, len(body.Payload.Invites), "user %d", userID)
}

func checkLogin(t *testing.T, user *models.AuthProfile, name, password string) {
	params := account.PostAccountLoginParams{
		Name:     name,
		Password: password,
	}

	login := api.AccountPostAccountLoginHandler.Handle
	resp := login(params)
	body, ok := resp.(*account.PostAccountLoginOK)
	if !ok {
		badBody, ok := resp.(*account.PostAccountLoginBadRequest)
		if ok {
			t.Fatal(badBody.Payload.Message)
		}

		t.Fatal("login error")
	}

	require.Equal(t, user, body.Payload)
}

func changePassword(t *testing.T, userID int64, old, upd string, ok bool) {
	id := models.UserID(userID)
	params := account.PostAccountPasswordParams{
		OldPassword: old,
		NewPassword: upd,
	}

	update := api.AccountPostAccountPasswordHandler.Handle
	resp := update(params, &id)
	switch resp.(type) {
	case *account.PostAccountPasswordOK:
		require.True(t, ok)
		return
	case *account.PostAccountPasswordForbidden:
		body := resp.(*account.PostAccountPasswordForbidden)
		require.False(t, ok, body.Payload.Message)
	default:
		t.Fatalf("set password user %d", userID)
	}
}

func TestRegister(t *testing.T) {
	{
		const q = "INSERT INTO invites(referrer_id, word1, word2, word3) VALUES(1, 1, 1, 1)"
		for i := 0; i < 3; i++ {
			db.Exec(q)
		}
	}

	checkInvites(t, 1, 3)
	checkName(t, "testtEst", true)
	checkEmail(t, "testeMAil", true)

	params := account.PostAccountRegisterParams{
		Name:     "testtest",
		Email:    "testemail",
		Password: "test123",
		Invite:   "acknown acknown acknown",
		Referrer: "HaveANiceDay",
	}

	register := api.AccountPostAccountRegisterHandler.Handle
	resp := register(params)
	body, ok := resp.(*account.PostAccountRegisterCreated)
	if !ok {
		badBody, ok := resp.(*account.PostAccountRegisterBadRequest)
		if ok {
			t.Fatal(badBody.Payload.Message)
		}

		t.Fatal("reg error")
	}

	user := body.Payload

	checkInvites(t, 1, 2)
	checkName(t, "testtEst", false)
	checkEmail(t, "testeMAil", false)
	checkLogin(t, user, params.Name, params.Password)
	checkLogin(t, user, strings.ToUpper(params.Name), params.Password)

	changePassword(t, user.ID, "test123", "new123", true)
	changePassword(t, user.ID, "test123", "new123", false)
	checkLogin(t, user, params.Name, "new123")

	req := require.New(t)
	req.Equal(params.Name, user.Name)
	req.Equal(params.Email, user.Account.Email)
	req.Equal(params.Referrer, user.InvitedBy.Name)

	req.Equal(user.Name, user.ShowName)
	req.True(user.IsOnline)
	// req.Empty(user.Avatar)

	req.Equal("not set", user.Gender)
	req.False(user.IsDaylog)
	req.Equal("all", user.Privacy)
	req.Empty(user.Title)
	req.Zero(user.Karma)
	req.NotEmpty(user.CreatedAt)
	req.Equal(user.CreatedAt, user.LastSeenAt)
	req.Zero(user.AgeLowerBound)
	req.Zero(user.AgeLowerBound)
	req.Empty(user.Country)
	req.Empty(user.City)

	cnt := user.Counts
	req.Zero(cnt.Entries)
	req.Zero(cnt.Followings)
	req.Zero(cnt.Followers)
	req.Zero(cnt.Ignored)
	req.Zero(cnt.Invited)
	req.Zero(cnt.Comments)
	req.Zero(cnt.Favorites)
	req.Zero(cnt.Tags)

	req.Empty(user.Birthday)
	req.False(user.ShowInTops)

	acc := user.Account
	req.Equal(32, len(acc.APIKey))
	req.NotEmpty(acc.ValidThru)
	req.False(acc.Verified)

	resp = register(params)
	_, ok = resp.(*account.PostAccountRegisterBadRequest)
	req.True(ok)
	checkInvites(t, 1, 2)

	gender := "female"
	city := "Moscow"
	country := "Russia"
	bday := "01.06.1992"

	params = account.PostAccountRegisterParams{
		Name:     "testtest2",
		Email:    "testemail2",
		Password: "test123",
		Gender:   &gender,
		City:     &city,
		Country:  &country,
		Birthday: &bday,
		Invite:   "acknown acknown acknown",
		Referrer: "HaveANiceDay",
	}

	resp = register(params)
	body, ok = resp.(*account.PostAccountRegisterCreated)
	req.True(ok)

	user = body.Payload

	checkInvites(t, 1, 1)
	checkName(t, "testtEst2", false)
	checkEmail(t, "testeMAil2", false)
	checkLogin(t, user, params.Name, params.Password)

	changePassword(t, user.ID, "test123", "new123", true)
	changePassword(t, user.ID, "test123", "new123", false)
	checkLogin(t, user, params.Name, "new123")

	req.Equal(gender, user.Gender)
	req.Equal(city, user.City)
	req.Equal(country, user.Country)

	req.Equal(int64(25), user.AgeLowerBound)
	req.Equal(int64(29), user.AgeUpperBound)
	req.Equal("1992-01-06T00:00:00Z", user.Birthday)

	req.NotEqual(acc.APIKey, user.Account.APIKey)
}
