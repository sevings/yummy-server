package account

import (
	"database/sql"
	"os"
	"testing"

	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations/account"
	"github.com/stretchr/testify/require"
)

var db *sql.DB

func TestMain(m *testing.M) {
	config := utils.LoadConfig("../../../../configs/server")
	db = utils.OpenDatabase(config)
	utils.ClearDatabase(db)

	os.Exit(m.Run())
}

func checkEmail(t *testing.T, email string, free bool) {
	check := newEmailChecker(db)
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
	check := newNameChecker(db)
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
	load := newInvitesLoader(db)
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

	login := newLoginer(db)
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

	update := newPasswordUpdater(db)
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
	checkInvites(t, 1, 3)
	checkName(t, "tEst", true)
	checkEmail(t, "eMAil", true)

	params := account.PostAccountRegisterParams{
		Name:     "test",
		Email:    "email",
		Password: "test123",
		Invite:   "acknown acknown acknown",
		Referrer: "HaveANiceDay",
	}

	register := newRegistrator(db)
	resp := register(params)
	body, ok := resp.(*account.PostAccountRegisterOK)
	if !ok {
		badBody, ok := resp.(*account.PostAccountRegisterBadRequest)
		if ok {
			t.Fatal(badBody.Payload.Message)
		}

		t.Fatal("reg error")
	}

	user := body.Payload

	checkInvites(t, 1, 2)
	checkName(t, "tEst", false)
	checkEmail(t, "eMAil", false)
	checkLogin(t, user, params.Name, params.Password)

	changePassword(t, user.ID, "test123", "new123", true)
	changePassword(t, user.ID, "test123", "new123", false)
	checkLogin(t, user, params.Name, "new123")

	req := require.New(t)
	req.Equal(params.Name, user.Name)
	req.Equal(params.Email, user.Account.Email)
	req.Equal(params.Referrer, user.InvitedBy.Name)

	req.Equal(user.Name, user.ShowName)
	req.True(user.IsOnline)
	req.Empty(user.Avatar)

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
		Name:     "test2",
		Email:    "email2",
		Password: "test123",
		Gender:   &gender,
		City:     &city,
		Country:  &country,
		Birthday: &bday,
		Invite:   "acknown acknown acknown",
		Referrer: "HaveANiceDay",
	}

	resp = register(params)
	body, ok = resp.(*account.PostAccountRegisterOK)

	user = body.Payload

	checkInvites(t, 1, 1)
	checkName(t, "tEst2", false)
	checkEmail(t, "eMAil2", false)
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
