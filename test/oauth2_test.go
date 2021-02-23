package test

import (
	"crypto/sha256"
	"encoding/base64"
	"github.com/sevings/mindwell-server/restapi/operations/oauth2"
	"github.com/sevings/mindwell-server/utils"
	"github.com/stretchr/testify/require"
	"log"
	"math/rand"
	"testing"
	"time"
)

type appData struct {
	id          int64
	secret      string
	redirectUri string
	devID       int64
	flow        uint8
	name        string
	showName    string
	platform    string
	info        string
	ban         bool
}

func createOauth2AppSecret(flow uint8, genSecret bool) *appData {
	const query = `
INSERT INTO apps(id, secret_hash, redirect_uri, developer_id, flow, name, show_name, platform, info)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

	app := &appData{
		id:          int64(rand.Int31()),
		redirectUri: "test://login",
		devID:       1,
		flow:        flow,
		name:        "TestApp",
		showName:    "Test app",
		platform:    "test",
		info:        "Test application.",
	}

	var secretHash []byte

	if genSecret {
		app.secret = utils.GenerateString(64)
		secretHash = srv.AppSecretHash(app.secret)
	}

	_, err := db.Exec(query, app.id, secretHash, app.redirectUri, app.devID, app.flow,
		app.name, app.showName, app.platform, app.info)
	if err != nil {
		log.Println(err)
		return nil
	}

	return app
}

func createOauth2App(flow uint8) *appData {
	return createOauth2AppSecret(flow, true)
}

func removeOAuth2App(app *appData) {
	const query = `
DELETE FROM apps
WHERE id = $1
`

	_, err := db.Exec(query, app.id)
	if err != nil {
		log.Println(err)
	}
}

func TestLoadApp(t *testing.T) {
	req := require.New(t)

	data := createOauth2App(1)
	req.NotNil(data)

	load := api.Oauth2GetOauth2AppsIDHandler.Handle
	params := oauth2.GetOauth2AppsIDParams{ID: data.id}
	resp := load(params, userIDs[0])
	body, ok := resp.(*oauth2.GetOauth2AppsIDOK)
	req.True(ok)

	app := body.Payload
	req.Equal(data.id, app.ID)
	req.Equal(data.name, app.Name)
	req.Equal(data.showName, app.ShowName)
	req.Equal(data.platform, app.Platform)
	req.Equal(data.info, app.Info)

	params.ID = data.id + 1
	resp = load(params, userIDs[0])
	_, ok = resp.(*oauth2.GetOauth2AppsIDOK)
	req.False(ok)

	removeOAuth2App(data)
}

func loadOAuth2Token(t *testing.T, params oauth2.PostOauth2TokenParams, success bool) *oauth2.PostOauth2TokenOKBody {
	req := require.New(t)
	post := api.Oauth2PostOauth2TokenHandler.Handle
	resp := post(params)
	body, ok := resp.(*oauth2.PostOauth2TokenOK)
	req.Equal(success, ok)
	if !ok {
		return nil
	}

	token := body.Payload

	req.NotEmpty(token.AccessToken)
	req.NotZero(token.ExpiresIn)

	if params.GrantType != "client_credentials" {
		req.NotEmpty(token.RefreshToken)
		req.NotEmpty(token.Scope)
	}

	return token
}

func TestPasswordToken(t *testing.T) {
	app := createOauth2App(4)

	name := "test0"
	pass := "test123"
	params := oauth2.PostOauth2TokenParams{
		GrantType:    "password",
		ClientID:     app.id,
		ClientSecret: &app.secret,
		Username:     &name,
		Password:     &pass,
	}

	load := func(success bool) *oauth2.PostOauth2TokenOKBody {
		return loadOAuth2Token(t, params, success)
	}

	load(true)

	params.ClientID++
	load(false)
	params.ClientID--

	params.ClientSecret = &name
	load(false)
	params.ClientSecret = &app.secret

	name = "TesT0"
	load(true)

	name = "tesT0@example.com"
	load(true)

	pass = "wrong password"
	load(false)

	removeOAuth2App(app)

	app = createOauth2App(3)
	params.ClientID = app.id
	load(false)
	removeOAuth2App(app)
}

func TestAppToken(t *testing.T) {
	app := createOauth2App(1)

	params := oauth2.PostOauth2TokenParams{
		GrantType:    "client_credentials",
		ClientID:     app.id,
		ClientSecret: &app.secret,
	}

	load := func(success bool) *oauth2.PostOauth2TokenOKBody {
		return loadOAuth2Token(t, params, success)
	}

	load(true)

	params.ClientID++
	load(false)
	params.ClientID--

	params.ClientSecret = nil
	load(false)

	params.ClientSecret = &app.platform
	load(false)

	removeOAuth2App(app)
}

func TestCodeToken(t *testing.T) {
	req := require.New(t)
	app := createOauth2App(2)

	var scope []string
	loadCode := func(success bool) string {
		state := "test state"
		params := oauth2.GetOauth2AuthParams{
			ClientID:     app.id,
			RedirectURI:  app.redirectUri,
			ResponseType: "code",
			Scope:        scope,
			State:        &state,
		}
		get := api.Oauth2GetOauth2AuthHandler.Handle
		resp := get(params, userIDs[0])
		body, ok := resp.(*oauth2.GetOauth2AuthOK)
		require.Equal(t, success, ok)
		if !ok {
			return ""
		}

		req.Equal(state, body.Payload.State)
		req.NotEmpty(body.Payload.Code)

		return body.Payload.Code
	}

	app.id++
	loadCode(false)
	app.id--

	uri := app.redirectUri
	app.redirectUri = "wrong://uri"
	loadCode(false)
	app.redirectUri = uri

	scope = []string{"wrong scope"}
	loadCode(false)

	scope[0] = "entries:read"
	code := loadCode(true)

	loadToken := func(success bool) *oauth2.PostOauth2TokenOKBody {
		params := oauth2.PostOauth2TokenParams{
			GrantType:    "authorization_code",
			ClientID:     app.id,
			ClientSecret: &app.secret,
			Code:         &code,
			RedirectURI:  &app.redirectUri,
		}

		return loadOAuth2Token(t, params, success)
	}

	app.id++
	loadToken(false)
	app.id--

	secret := app.secret
	app.secret = app.platform
	loadToken(false)
	app.secret = secret

	codeAct := code
	code = "wrong code"
	loadToken(false)
	code = codeAct

	app.redirectUri = "wrong://uri"
	loadCode(false)
	app.redirectUri = uri

	token := loadToken(true)
	req.Equal(scope, token.Scope)

	_, err := srv.API.OAuth2CodeAuth(token.AccessToken, []string{})
	req.Nil(err)

	loadToken(false)

	_, err = srv.API.OAuth2CodeAuth(token.AccessToken, []string{})
	req.NotNil(err)

	removeOAuth2App(app)

	noApp := createOauth2App(1)
	app.id = noApp.id
	loadCode(false)
	removeOAuth2App(noApp)
}

func TestCodeChallengeToken(t *testing.T) {
	req := require.New(t)
	app := createOauth2AppSecret(2, false)

	verifier := utils.GenerateString(32)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.URLEncoding.EncodeToString(sum[:])

	scope := []string{"entries:read"}
	loadCode := func(success bool) string {
		state := "test state"
		method := "S256"
		params := oauth2.GetOauth2AuthParams{
			ClientID:            app.id,
			RedirectURI:         app.redirectUri,
			ResponseType:        "code",
			Scope:               scope,
			State:               &state,
			CodeChallenge:       &challenge,
			CodeChallengeMethod: &method,
		}
		get := api.Oauth2GetOauth2AuthHandler.Handle
		resp := get(params, userIDs[0])
		body, ok := resp.(*oauth2.GetOauth2AuthOK)
		require.Equal(t, success, ok)
		if !ok {
			return ""
		}

		req.Equal(state, body.Payload.State)
		req.NotEmpty(body.Payload.Code)

		return body.Payload.Code
	}

	code := loadCode(true)

	loadToken := func(success bool) *oauth2.PostOauth2TokenOKBody {
		params := oauth2.PostOauth2TokenParams{
			GrantType:    "authorization_code",
			ClientID:     app.id,
			Code:         &code,
			RedirectURI:  &app.redirectUri,
			CodeVerifier: &verifier,
		}

		return loadOAuth2Token(t, params, success)
	}

	ver := verifier
	verifier = "wrong verifier"
	loadToken(false)
	verifier = ver

	token := loadToken(true)
	req.Equal(scope, token.Scope)

	loadToken(false)

	removeOAuth2App(app)
}

func TestRefreshToken(t *testing.T) {
	app := createOauth2App(4)

	name := "test0"
	pass := "test123"
	params := oauth2.PostOauth2TokenParams{
		GrantType:    "password",
		ClientID:     app.id,
		ClientSecret: &app.secret,
		Username:     &name,
		Password:     &pass,
	}

	token := loadOAuth2Token(t, params, true)

	params.GrantType = "refresh_token"
	params.RefreshToken = &token.RefreshToken
	params.Username = nil
	params.Password = nil

	time.Sleep(10 * time.Millisecond)

	params.ClientSecret = &name
	loadOAuth2Token(t, params, false)
	params.ClientSecret = &app.secret

	token2 := loadOAuth2Token(t, params, true)

	req := require.New(t)
	req.NotEqual(token.AccessToken, token2.AccessToken)
	req.NotEqual(token.RefreshToken, token2.RefreshToken)
	req.Equal(token.ExpiresIn, token2.ExpiresIn)

	loadOAuth2Token(t, params, false)

	removeOAuth2App(app)
}

func TestPasswordFlow(t *testing.T) {
	req := require.New(t)
	app := createOauth2App(4)

	name := "test0"
	pass := "test123"
	params := oauth2.PostOauth2TokenParams{
		GrantType:    "password",
		ClientID:     app.id,
		ClientSecret: &app.secret,
		Username:     &name,
		Password:     &pass,
	}

	token := loadOAuth2Token(t, params, true)
	auth := srv.API.OAuth2PasswordAuth

	id, err := auth(token.AccessToken, []string{"entries:read"})
	req.Nil(err)
	req.Equal(userIDs[0].ID, id.ID)

	id, err = auth(token.AccessToken, []string{"wrong scope"})
	req.NotNil(err)

	authCode := srv.API.OAuth2CodeAuth
	id, err = authCode(token.AccessToken, []string{"entries:read"})
	req.NotNil(err)

	params.GrantType = "refresh_token"
	params.RefreshToken = &token.RefreshToken
	params.Username = nil
	params.Password = nil

	token = loadOAuth2Token(t, params, true)

	id, err = auth(token.AccessToken, []string{"entries:read"})
	req.Nil(err)
	req.Equal(userIDs[0].ID, id.ID)

	removeOAuth2App(app)
}

func TestCodeFlow(t *testing.T) {
	req := require.New(t)
	app := createOauth2App(2)

	codeParams := oauth2.GetOauth2AuthParams{
		ClientID:     app.id,
		RedirectURI:  app.redirectUri,
		ResponseType: "code",
		Scope:        []string{"entries:read"},
	}
	get := api.Oauth2GetOauth2AuthHandler.Handle
	resp := get(codeParams, userIDs[0])
	body, ok := resp.(*oauth2.GetOauth2AuthOK)
	req.True(ok)
	code := body.Payload.Code

	params := oauth2.PostOauth2TokenParams{
		GrantType:    "authorization_code",
		ClientID:     app.id,
		ClientSecret: &app.secret,
		Code:         &code,
		RedirectURI:  &app.redirectUri,
	}

	token := loadOAuth2Token(t, params, true)
	auth := srv.API.OAuth2CodeAuth

	id, err := auth(token.AccessToken, []string{"entries:read"})
	req.Nil(err)
	req.Equal(userIDs[0].ID, id.ID)

	id, err = auth(token.AccessToken, []string{"wrong scope"})
	req.NotNil(err)

	authPass := srv.API.OAuth2PasswordAuth
	id, err = authPass(token.AccessToken, []string{"entries:read"})
	req.NotNil(err)

	params.GrantType = "refresh_token"
	params.RefreshToken = &token.RefreshToken
	params.Username = nil
	params.Password = nil

	token = loadOAuth2Token(t, params, true)

	id, err = auth(token.AccessToken, []string{"entries:read"})
	req.Nil(err)
	req.Equal(userIDs[0].ID, id.ID)

	removeOAuth2App(app)
}
