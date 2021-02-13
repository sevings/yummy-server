package test

import (
	"database/sql"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	accountImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/account"
	admImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/adm"
	chatsImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/chats"
	commentsImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/comments"
	complainsImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/complains"
	designImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/design"
	entriesImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/entries"
	favoritesImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/favorites"
	notificationsImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/notifications"
	oauth2Impl "github.com/sevings/mindwell-server/internal/app/mindwell-server/oauth2"
	relationsImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/relations"
	tagsImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/tags"
	usersImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/users"
	votesImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/votes"
	watchingsImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/watchings"
	"github.com/sevings/mindwell-server/utils"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations"
	"github.com/sevings/mindwell-server/restapi/operations/account"
	"github.com/stretchr/testify/require"
)

var srv *utils.MindwellServer
var api *operations.MindwellAPI
var db *sql.DB
var userIDs []*models.UserID
var profiles []*models.AuthProfile
var esm EmailSenderMock
var ecm EmailCheckerMock

func TestMain(m *testing.M) {
	api = &operations.MindwellAPI{}
	srv = utils.NewMindwellServer(api, "../configs/server")
	db = srv.DB

	ecm.Trusted = []string{"example.com"}
	srv.Eac = &ecm
	srv.Ntf.Mail = &esm

	utils.ClearDatabase(db)

	accountImpl.ConfigureAPI(srv)
	admImpl.ConfigureAPI(srv)
	usersImpl.ConfigureAPI(srv)
	entriesImpl.ConfigureAPI(srv)
	votesImpl.ConfigureAPI(srv)
	favoritesImpl.ConfigureAPI(srv)
	watchingsImpl.ConfigureAPI(srv)
	commentsImpl.ConfigureAPI(srv)
	designImpl.ConfigureAPI(srv)
	relationsImpl.ConfigureAPI(srv)
	notificationsImpl.ConfigureAPI(srv)
	complainsImpl.ConfigureAPI(srv)
	chatsImpl.ConfigureAPI(srv)
	tagsImpl.ConfigureAPI(srv)
	oauth2Impl.ConfigureAPI(srv)

	userIDs, profiles = registerTestUsers(db)

	if len(esm.Emails) != 4 {
		log.Fatal("Email count")
	}

	for i := 0; i < 4; i++ {
		email := "test" + strconv.Itoa(i) + "@example.com"
		if esm.Emails[i] != email {
			log.Fatal("Greeting has not sent to ", email)
		}
	}

	esm.Clear()

	os.Exit(m.Run())
}

func checkEmail(t *testing.T, email string, success, free bool) {
	check := api.AccountGetAccountEmailEmailHandler.Handle
	resp := check(account.GetAccountEmailEmailParams{Email: email})
	body, ok := resp.(*account.GetAccountEmailEmailOK)

	require.Equal(t, success, ok)
	if !ok {
		return
	}

	require.Equal(t, email, body.Payload.Email)
	require.Equal(t, free, body.Payload.IsFree, email)
}

func TestCheckEmailFree(t *testing.T) {
	checkEmail(t, "123@example.com", true, true)
	checkEmail(t, "test", false, false)
	checkEmail(t, "test@test.com", false, false)
}

func checkName(t *testing.T, name string, free bool) {
	check := api.AccountGetAccountNameNameHandler.Handle
	resp := check(account.GetAccountNameNameParams{Name: name})
	body, ok := resp.(*account.GetAccountNameNameOK)

	require.True(t, ok, name)
	require.Equal(t, name, body.Payload.Name)
	require.Equal(t, free, body.Payload.IsFree, name)
}

func TestCheckNameFree(t *testing.T) {
	checkName(t, "mINDWell", false)
	checkName(t, "nAMe", true)
}

func checkInvites(t *testing.T, userID *models.UserID, size int) {
	load := api.AccountGetAccountInvitesHandler.Handle
	resp := load(account.GetAccountInvitesParams{}, userID)
	body, ok := resp.(*account.GetAccountInvitesOK)

	require.True(t, ok, "user %d", userID.ID)
	require.Equal(t, size, len(body.Payload.Invites), "user %d", userID.ID)
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

func changePassword(t *testing.T, userID *models.UserID, old, upd, email string, ok bool) {
	params := account.PostAccountPasswordParams{
		OldPassword: old,
		NewPassword: upd,
	}

	update := api.AccountPostAccountPasswordHandler.Handle
	resp := update(params, userID)
	switch resp.(type) {
	case *account.PostAccountPasswordOK:
		require.True(t, ok)
		if len(email) > 0 {
			esm.CheckEmail(t, email)
		}
		return
	case *account.PostAccountPasswordForbidden:
		body := resp.(*account.PostAccountPasswordForbidden)
		require.False(t, ok, body.Payload.Message)
	default:
		t.Fatalf("set password user %d", userID.ID)
	}
}

func changeEmail(t *testing.T, userID *models.UserID, oldEmail, newEmail, password string, ok bool) {
	params := account.PostAccountEmailParams{
		Email:    newEmail,
		Password: password,
	}

	req := require.New(t)

	upd := api.AccountPostAccountEmailHandler.Handle
	resp := upd(params, userID)
	switch resp.(type) {
	case *account.PostAccountEmailOK:
		req.True(ok)
		req.Equal(1, len(esm.Codes))
		esm.CheckEmail2(t, oldEmail, newEmail)
	case *account.PostAccountEmailForbidden:
		body := resp.(*account.PostAccountEmailForbidden)
		req.False(ok, body.Payload.Message)
	case *account.PostAccountEmailBadRequest:
		body := resp.(*account.PostAccountEmailBadRequest)
		req.False(ok, body.Payload.Message)
	default:
		t.Fatalf("set email user %s", userID.Name)
	}
}

func checkVerify(t *testing.T, userID *models.UserID, email string) {
	request := api.AccountPostAccountVerificationHandler.Handle
	resp := request(account.PostAccountVerificationParams{}, userID)
	_, ok := resp.(*account.PostAccountVerificationOK)

	req := require.New(t)
	req.True(ok, "user %d", userID.ID)
	req.Equal(1, len(esm.Emails))
	req.Equal(email, esm.Emails[0])
	req.Equal(1, len(esm.Codes))

	verify := api.AccountGetAccountVerificationEmailHandler.Handle
	params := account.GetAccountVerificationEmailParams{
		Code:  esm.Codes[0],
		Email: esm.Emails[0],
	}
	resp = verify(params)
	_, ok = resp.(*account.GetAccountVerificationEmailOK)

	req.True(ok)

	esm.Clear()
}

func checkResetPassword(t *testing.T, email string) {
	request := api.AccountPostAccountRecoverHandler.Handle
	resp := request(account.PostAccountRecoverParams{Email: email})
	_, ok := resp.(*account.PostAccountRecoverOK)

	req := require.New(t)
	req.True(ok)
	req.Equal(1, len(esm.Emails))
	req.Equal(email, esm.Emails[0])
	req.Equal(1, len(esm.Codes))
	req.Equal(1, len(esm.Dates))

	reset := api.AccountPostAccountRecoverPasswordHandler.Handle
	params := account.PostAccountRecoverPasswordParams{
		Code:     esm.Codes[0],
		Email:    esm.Emails[0],
		Date:     esm.Dates[0],
		Password: "test123",
	}
	resp = reset(params)
	_, ok = resp.(*account.PostAccountRecoverPasswordOK)

	req.True(ok)

	esm.Clear()
}

func TestInvites(t *testing.T) {
	utils.ClearDatabase(db)

	inviter := &models.UserID{
		ID:   1,
		Name: "Mindwell",
	}

	checkInvites(t, inviter, 3)

	userIDs, profiles = registerTestUsers(db)
	esm.Clear()
}

func TestRegister(t *testing.T) {
	checkName(t, "testtEst", true)
	checkEmail(t, "testeMAil@example.com", true, true)

	params := account.PostAccountRegisterParams{
		Name:     "testtest",
		Email:    "testemail@example.com",
		Password: "test123",
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
	userID := &models.UserID{
		ID:   user.ID,
		Name: user.Name,
	}

	checkName(t, "testtEst", false)
	checkEmail(t, "testeMAil@example.com", true, false)
	checkLogin(t, user, params.Name, params.Password)
	checkLogin(t, user, strings.ToUpper(params.Name), params.Password)
	checkLogin(t, user, params.Email, params.Password)
	checkLogin(t, user, strings.ToUpper(params.Email), params.Password)

	esm.CheckEmail(t, "testemail@example.com")
	checkVerify(t, userID, "testemail@example.com")
	user.Account.Verified = true
	checkResetPassword(t, "testemail@example.com")
	checkResetPassword(t, "testeMAil@example.com")
	checkLogin(t, user, params.Name, params.Password)

	changePassword(t, userID, "test123", "new123", "testemail@example.com", true)
	changePassword(t, userID, "test123", "new123", "", false)
	checkLogin(t, user, params.Name, "new123")

	changeEmail(t, userID, "testemail@example.com", "tEsteMail@example.com", "new123", false)
	changeEmail(t, userID, "testemail@example.com", "testemail0@example.com", "xvc", false)
	changeEmail(t, userID, "testemail@example.com", "test", "new123", false)
	changeEmail(t, userID, "testemail@example.com", "testemail0@test.com", "new123", false)
	changeEmail(t, userID, "testemail@example.com", "testemail0@example.com", "new123", true)
	user.Account.Email = utils.HideEmail("testemail0@example.com")
	user.Account.Verified = false
	checkLogin(t, user, "testemail0@example.com", "new123")
	checkVerify(t, userID, "testemail0@example.com")
	user.Account.Verified = true

	req := require.New(t)
	req.Equal(params.Name, user.Name)
	req.Equal(params.Email, "testemail@example.com")

	req.Equal(user.Name, user.ShowName)
	req.True(user.IsOnline)
	// req.Empty(user.Avatar)

	req.Equal("not set", user.Gender)
	req.False(user.IsDaylog)
	req.Equal("all", user.Privacy)
	req.Empty(user.Title)
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
	req.NotEmpty(acc.ValidThru)
	req.True(acc.Verified)

	resp = register(params)
	_, ok = resp.(*account.PostAccountRegisterBadRequest)
	req.True(ok)

	gender := "female"
	city := "Moscow"
	country := "Russia"
	bday := "01.06.1992"

	params = account.PostAccountRegisterParams{
		Name:     "testtest2",
		Email:    "testemail2@example.com",
		Password: "test123",
		Gender:   &gender,
		City:     &city,
		Country:  &country,
		Birthday: &bday,
	}

	resp = register(params)
	body, ok = resp.(*account.PostAccountRegisterCreated)
	req.True(ok)

	user = body.Payload
	userID = &models.UserID{
		ID:   user.ID,
		Name: user.Name,
	}

	checkName(t, "testtEst2", false)
	checkEmail(t, "testeMAil2@example.com", true, false)
	checkLogin(t, user, params.Name, params.Password)

	esm.CheckEmail(t, "testemail2@example.com")
	checkVerify(t, userID, "testemail2@example.com")
	user.Account.Verified = true
	checkLogin(t, user, params.Name, params.Password)

	changePassword(t, userID, "test123", "new123", "testemail2@example.com", true)
	changePassword(t, userID, "test123", "new123", "", false)
	checkLogin(t, user, params.Name, "new123")

	req.Equal(gender, user.Gender)
	req.Equal(city, user.City)
	req.Equal(country, user.Country)

	req.Equal(int64(25), user.AgeLowerBound)
	req.Equal(int64(29), user.AgeUpperBound)
	req.Equal("1992-01-06T00:00:00Z", user.Birthday)

	req.NotEqual(acc.APIKey, user.Account.APIKey)

	params = account.PostAccountRegisterParams{
		Name:     "testtest3",
		Email:    "testemail3@example.com",
		Password: "test123",
	}

	resp = register(params)
	body, ok = resp.(*account.PostAccountRegisterCreated)
	req.True(ok)

	user = body.Payload
	userID = &models.UserID{
		ID:   user.ID,
		Name: user.Name,
	}

	req.Nil(user.InvitedBy)

	checkName(t, "testtEst3", false)
	checkEmail(t, "testeMAil3@example.com", true, false)
	checkLogin(t, user, params.Name, params.Password)

	esm.CheckEmail(t, "testemail3@example.com")
	checkVerify(t, userID, "testemail3@example.com")
	user.Account.Verified = true
	checkLogin(t, user, params.Name, params.Password)

	changePassword(t, userID, "test123", "new123", "testemail3@example.com", true)
	changePassword(t, userID, "test123", "new123", "", false)
	checkLogin(t, user, params.Name, "new123")

	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)
	esm.Clear()
}

func getEmailSettings(t *testing.T, userID *models.UserID) *account.GetAccountSettingsEmailOKBody {
	load := api.AccountGetAccountSettingsEmailHandler.Handle
	resp := load(account.GetAccountSettingsEmailParams{}, userID)
	body, ok := resp.(*account.GetAccountSettingsEmailOK)
	require.True(t, ok, "user %d", userID.ID)
	return body.Payload
}

func checkEmailSettings(t *testing.T, userID *models.UserID, comments, followers, invites bool) {
	settings := getEmailSettings(t, userID)
	require.Equal(t, comments, settings.Comments)
	require.Equal(t, followers, settings.Followers)
	require.Equal(t, invites, settings.Invites)
}

func checkUpdateEmailSettings(t *testing.T, userID *models.UserID, comments, followers, invites bool) {
	load := api.AccountPutAccountSettingsEmailHandler.Handle

	settings := account.PutAccountSettingsEmailParams{
		Comments:  &comments,
		Followers: &followers,
		Invites:   &invites,
	}

	resp := load(settings, userID)
	_, ok := resp.(*account.PutAccountSettingsEmailOK)

	require.True(t, ok, "user %d", userID.ID)

	checkEmailSettings(t, userID, comments, followers, invites)
}

func TestEmailSettings(t *testing.T) {
	checkEmailSettings(t, userIDs[0], false, false, false)
	checkUpdateEmailSettings(t, userIDs[0], true, false, false)
	checkUpdateEmailSettings(t, userIDs[0], false, false, true)
	checkUpdateEmailSettings(t, userIDs[0], true, true, false)
}

func checkTelegramSettings(t *testing.T, userID *models.UserID, comments, followers, invites, messages bool) {
	load := api.AccountGetAccountSettingsTelegramHandler.Handle
	resp := load(account.GetAccountSettingsTelegramParams{}, userID)
	body, ok := resp.(*account.GetAccountSettingsTelegramOK)
	require.True(t, ok, "user %d", userID.ID)

	settings := body.Payload
	require.Equal(t, comments, settings.Comments)
	require.Equal(t, followers, settings.Followers)
	require.Equal(t, invites, settings.Invites)
	require.Equal(t, messages, settings.Messages)
}

func checkUpdateTelegramSettings(t *testing.T, userID *models.UserID, comments, followers, invites, messages bool) {
	load := api.AccountPutAccountSettingsTelegramHandler.Handle

	settings := account.PutAccountSettingsTelegramParams{
		Comments:  &comments,
		Followers: &followers,
		Invites:   &invites,
		Messages:  &messages,
	}

	resp := load(settings, userID)
	_, ok := resp.(*account.PutAccountSettingsTelegramOK)

	require.True(t, ok, "user %d", userID.ID)

	checkTelegramSettings(t, userID, comments, followers, invites, messages)
}

func TestTelegramSettings(t *testing.T) {
	checkTelegramSettings(t, userIDs[0], true, true, true, true)
	checkUpdateTelegramSettings(t, userIDs[0], true, false, false, false)
	checkUpdateTelegramSettings(t, userIDs[0], false, false, true, true)
	checkUpdateTelegramSettings(t, userIDs[0], true, true, false, false)
}

func TestConnectionToken(t *testing.T) {
	load := api.AccountGetAccountSubscribeTokenHandler.Handle
	resp := load(account.GetAccountSubscribeTokenParams{}, userIDs[1])
	body, ok := resp.(*account.GetAccountSubscribeTokenOK)

	req := require.New(t)
	req.True(ok)

	data := body.Payload
	req.NotEmpty(data.Token)
}

func TestTelegramToken(t *testing.T) {
	load := api.AccountGetAccountSubscribeTelegramHandler.Handle
	resp := load(account.GetAccountSubscribeTelegramParams{}, userIDs[1])
	body, ok := resp.(*account.GetAccountSubscribeTelegramOK)

	req := require.New(t)
	req.True(ok)

	data := body.Payload
	req.NotEmpty(data.Token)
	req.Equal(userIDs[1].ID, srv.Ntf.Tg.VerifyToken(data.Token))
}

func TestTelegramLogout(t *testing.T) {
	logout := api.AccountDeleteAccountSubscribeTelegramHandler.Handle
	resp := logout(account.DeleteAccountSubscribeTelegramParams{}, userIDs[1])
	_, ok := resp.(*account.DeleteAccountSubscribeTelegramNoContent)

	req := require.New(t)
	req.True(ok)
}

func TestHideEmail(t *testing.T) {
	he := utils.HideEmail
	req := require.New(t)

	req.Equal("", he(""))
	req.Equal("", he("mindwell.win"))
	req.Equal("***@ml.win", he("s@ml.win"))
	req.Equal("***@ml.win", he("sp@ml.win"))
	req.Equal("s***t@mindwell.win", he("support@mindwell.win"))
}

func TestCheckEmailAllowed(t *testing.T) {
	ec := utils.NewEmailChecker(srv)
	req := require.New(t)

	req.True(ec.IsAllowed("test@ya.ru"))
	req.False(ec.IsAllowed("test@mailinator.com"))
}
