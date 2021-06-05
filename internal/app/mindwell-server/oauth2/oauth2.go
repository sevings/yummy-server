package oauth2

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"github.com/go-openapi/runtime/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/patrickmn/go-cache"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/oauth2"
	"github.com/sevings/mindwell-server/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	webIP = srv.ConfigString("web.ip")

	apiSecret := srv.ConfigBytes("server.api_secret")

	srv.API.APIKeyHeaderAuth = utils.NewKeyAuth(srv.DB, apiSecret)
	srv.API.NoAPIKeyAuth = utils.NoApiKeyAuth
	srv.API.OAuth2PasswordAuth = utils.NewOAuth2User(srv, srv.DB, utils.PasswordFlow)
	srv.API.OAuth2CodeAuth = utils.NewOAuth2User(srv, srv.DB, utils.CodeFlow)

	srv.API.Oauth2PostOauth2AllowHandler = oauth2.PostOauth2AllowHandlerFunc(newOAuth2Allow(srv))
	srv.API.Oauth2GetOauth2DenyHandler = oauth2.GetOauth2DenyHandlerFunc(newOAuth2Deny(srv))
	srv.API.Oauth2PostOauth2UpgradeHandler = oauth2.PostOauth2UpgradeHandlerFunc(newOAuth2Upgrade(srv))

	srv.API.Oauth2PostOauth2TokenHandler = oauth2.PostOauth2TokenHandlerFunc(newOAuth2Token(srv))
	srv.API.Oauth2GetOauth2AppsIDHandler = oauth2.GetOauth2AppsIDHandlerFunc(newAppLoader(srv))
}

type authData struct {
	appID       int64
	secretHash  []byte
	redirectUri string
	userID      int64
	userName    string
	scope       uint32
	challenge   string
	method      string
	access      string
	refresh     string
}

var authCache = cache.New(15*time.Minute, time.Hour)

type hasher interface {
	AppSecretHash(secret string) []byte
	AppTokenHash(secret string) []byte
	AccessTokenHash(secret string) []byte
	RefreshTokenHash(secret string) []byte
	PasswordHash(password string) []byte
}

func createTokens(h hasher, tx *utils.AutoTx, appID, userID int64, scope uint32, userName string) *models.OAuth2Token {
	token := &models.OAuth2Token{
		AccessToken:  userName + "." + utils.GenerateString(utils.AccessTokenLength),
		ExpiresIn:    utils.AccessTokenLifetime,
		RefreshToken: strconv.FormatInt(userID, 32) + "." + utils.GenerateString(utils.RefreshTokenLength),
		Scope:        utils.ScopeToString(scope),
		TokenType:    models.OAuth2TokenTokenTypeBearer,
	}

	accessHash := h.AccessTokenHash(token.AccessToken)
	refreshHash := h.RefreshTokenHash(token.RefreshToken)
	accessThru := time.Now().Add(utils.AccessTokenLifetime * time.Second)
	refreshThru := time.Now().Add(utils.RefreshTokenLifetime * time.Second)

	const query = `
INSERT INTO sessions(app_id, user_id, scope, access_hash, refresh_hash, access_thru, refresh_thru)
VALUES($1, $2, $3, $4, $5, $6, $7)
`

	tx.Exec(query, appID, userID, scope, accessHash[:], refreshHash[:], accessThru, refreshThru)
	if tx.Error() != nil {
		return nil
	}

	return token
}

func createAppToken(h hasher, tx *utils.AutoTx, appID int64, appName string) (string, error) {
	const query = `
INSERT INTO app_tokens(app_id, token_hash, valid_thru)
VALUES($1, $2, $3)
`

	token := appName + "+" + utils.GenerateString(utils.AppTokenLength)
	hash := h.AppTokenHash(token)
	thru := time.Now().Add(utils.AppTokenLifetime * time.Second)

	tx.Exec(query, appID, hash[:], thru)

	return token, tx.Error()
}

func checkCodeGrant(tx *utils.AutoTx, appID int64) (authData, bool, error) {
	const query = `
SELECT secret_hash, redirect_uri, flow, ban
FROM apps
WHERE id = $1
`

	auth := authData{appID: appID}
	var ban bool
	var f utils.AuthFlow
	tx.Query(query, appID).Scan(&auth.secretHash, &auth.redirectUri, &f, &ban)

	granted := !ban && f&utils.CodeFlow == utils.CodeFlow
	return auth, granted, tx.Error()
}

func checkPasswordGrant(h hasher, tx *utils.AutoTx, appID int64, appSecret string) (bool, error) {
	const grantQuery = `
SELECT flow, ban
FROM apps
WHERE id = $1 AND secret_hash = $2
`
	var ban bool
	var f utils.AuthFlow
	secretHash := h.AppSecretHash(appSecret)
	tx.Query(grantQuery, appID, secretHash).Scan(&f, &ban)

	granted := !ban && f&utils.PasswordFlow == utils.PasswordFlow
	return granted, tx.Error()
}

func checkAppGrant(h hasher, tx *utils.AutoTx, appID int64, appSecret string) (string, bool, error) {
	const query = `
SELECT name, flow, ban
FROM apps
WHERE id = $1 AND secret_hash = $2
`
	var name string
	var ban bool
	var f utils.AuthFlow
	secretHash := h.AppSecretHash(appSecret)
	tx.Query(query, appID, secretHash).Scan(&name, &f, &ban)

	granted := !ban && f&utils.AppFlow == utils.AppFlow
	return name, granted, tx.Error()
}

func checkRefreshGrant(h hasher, tx *utils.AutoTx, appID int64, appSecret string) (bool, error) {
	const query = `
SELECT ban
FROM apps
WHERE id = $1 AND secret_hash = $2
`

	secretHash := h.AppSecretHash(appSecret)
	ban := tx.QueryBool(query, appID, secretHash)
	return !ban, tx.Error()
}

var webIP string

func checkWebRequest(req *http.Request) (string, bool) {
	if req == nil {
		return "", true
	}

	ip := strings.Split(req.RemoteAddr, ":")[0]
	if ip == webIP {
		return "", true
	}

	return models.OAuth2ErrorErrorUnauthorizedClient, false
}

func postAllowBadRequest(err string) middleware.Responder {
	body := models.OAuth2Error{Error: err}
	return oauth2.NewPostOauth2AllowBadRequest().WithPayload(&body)
}

func newOAuth2Allow(srv *utils.MindwellServer) func(oauth2.PostOauth2AllowParams, *models.UserID) middleware.Responder {
	return func(params oauth2.PostOauth2AllowParams, userID *models.UserID) middleware.Responder {
		if err, ok := checkWebRequest(params.HTTPRequest); !ok {
			return postAllowBadRequest(err)
		}

		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			auth, granted, err := checkCodeGrant(tx, params.ClientID)
			if err != nil {
				return postAllowBadRequest(models.OAuth2ErrorErrorUnrecognizedClient)
			}
			if auth.redirectUri != params.RedirectURI {
				return postAllowBadRequest(models.OAuth2ErrorErrorInvalidRedirect)
			}
			if !granted {
				return postAllowBadRequest(models.OAuth2ErrorErrorInvalidGrant)
			}

			scope, err := utils.ScopeFromString(params.Scope)
			if err != nil {
				return postAllowBadRequest(models.OAuth2ErrorErrorInvalidScope)
			}

			if len(auth.secretHash) == 0 && params.CodeChallenge == nil {
				return postAllowBadRequest(models.OAuth2ErrorErrorInvalidRequest)
			}

			resp := &oauth2.PostOauth2AllowOKBody{
				Code: utils.GenerateString(utils.CodeLength),
			}

			if params.State != nil {
				resp.State = *params.State
			}

			auth.userID = userID.ID
			auth.userName = userID.Name
			auth.scope = scope

			if params.CodeChallenge != nil {
				auth.challenge = *params.CodeChallenge
			}

			if params.CodeChallengeMethod != nil {
				auth.method = *params.CodeChallengeMethod
			}

			authCache.SetDefault(resp.Code, &auth)

			return oauth2.NewPostOauth2AllowOK().WithPayload(resp)
		})
	}
}

func getDenyBadRequest(err string) middleware.Responder {
	body := models.OAuth2Error{Error: err}
	return oauth2.NewGetOauth2DenyBadRequest().WithPayload(&body)
}

func newOAuth2Deny(srv *utils.MindwellServer) func(oauth2.GetOauth2DenyParams) middleware.Responder {
	return func(params oauth2.GetOauth2DenyParams) middleware.Responder {
		if err, ok := checkWebRequest(params.HTTPRequest); !ok {
			return getDenyBadRequest(err)
		}

		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			auth, granted, err := checkCodeGrant(tx, params.ClientID)
			if err != nil {
				return getDenyBadRequest(models.OAuth2ErrorErrorUnrecognizedClient)
			}
			if auth.redirectUri != params.RedirectURI {
				return getDenyBadRequest(models.OAuth2ErrorErrorInvalidRedirect)
			}
			if !granted {
				return getDenyBadRequest(models.OAuth2ErrorErrorInvalidGrant)
			}

			return getDenyBadRequest(models.OAuth2ErrorErrorAccessDenied)
		})
	}
}

func postUpgradeBadRequest(err string) middleware.Responder {
	body := models.OAuth2Error{Error: err}
	return oauth2.NewPostOauth2UpgradeBadRequest().WithPayload(&body)
}

func newOAuth2Upgrade(srv *utils.MindwellServer) func(oauth2.PostOauth2UpgradeParams, *models.UserID) middleware.Responder {
	return func(params oauth2.PostOauth2UpgradeParams, userID *models.UserID) middleware.Responder {
		if err, ok := checkWebRequest(params.HTTPRequest); !ok {
			return postUpgradeBadRequest(err)
		}

		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			granted, err := checkPasswordGrant(srv, tx, params.ClientID, params.ClientSecret)
			if err != nil {
				return postUpgradeBadRequest(models.OAuth2ErrorErrorUnrecognizedClient)
			}
			if !granted {
				return postUpgradeBadRequest(models.OAuth2ErrorErrorInvalidGrant)
			}

			var scope uint32 = 1<<31 - 1
			resp := createTokens(srv, tx, params.ClientID, userID.ID, scope, userID.Name)
			if resp == nil {
				return postUpgradeBadRequest(models.OAuth2ErrorErrorServerError)
			}

			return oauth2.NewPostOauth2UpgradeOK().WithPayload(resp)
		})
	}
}

func loadUserByPassword(h hasher, tx *utils.AutoTx, name, password string) (int64, string) {
	name = strings.TrimSpace(name)
	password = strings.TrimSpace(password)
	hash := h.PasswordHash(password)

	const userIdQuery = `
SELECT id, name
FROM users
WHERE password_hash = $2
	AND (lower(name) = lower($1) OR lower(email) = lower($1))
`

	var userID int64
	var userName string
	tx.Query(userIdQuery, name, hash).Scan(&userID, &userName)

	return userID, userName
}

func postTokenBadRequest(err string) middleware.Responder {
	body := models.OAuth2Error{Error: err}
	return oauth2.NewPostOauth2TokenBadRequest().WithPayload(&body)
}

func requestPasswordToken(h hasher, tx *utils.AutoTx, appID int64, appSecret, name, password string) middleware.Responder {
	userID, userName := loadUserByPassword(h, tx, name, password)
	if userID == 0 {
		return postTokenBadRequest(models.OAuth2ErrorErrorInvalidRequest)
	}

	granted, err := checkPasswordGrant(h, tx, appID, appSecret)
	if err != nil {
		return postTokenBadRequest(models.OAuth2ErrorErrorUnrecognizedClient)
	}
	if !granted {
		return postTokenBadRequest(models.OAuth2ErrorErrorInvalidGrant)
	}

	var scope uint32 = 1<<31 - 1
	resp := createTokens(h, tx, appID, userID, scope, userName)
	if resp == nil {
		return postTokenBadRequest(models.OAuth2ErrorErrorServerError)
	}

	return oauth2.NewPostOauth2TokenOK().WithPayload(resp)
}

func requestAppToken(h hasher, tx *utils.AutoTx, appID int64, appSecret string) middleware.Responder {
	appName, granted, err := checkAppGrant(h, tx, appID, appSecret)
	if err != nil {
		return postTokenBadRequest(models.OAuth2ErrorErrorUnrecognizedClient)
	}
	if !granted {
		return postTokenBadRequest(models.OAuth2ErrorErrorInvalidGrant)
	}

	appToken, err := createAppToken(h, tx, appID, appName)
	if err != nil {
		return postTokenBadRequest(models.OAuth2ErrorErrorServerError)
	}

	resp := &models.OAuth2Token{
		AccessToken: appToken,
		ExpiresIn:   utils.AccessTokenLifetime,
		TokenType:   models.OAuth2TokenTokenTypeBearer,
	}

	return oauth2.NewPostOauth2TokenOK().WithPayload(resp)
}

func revokeTokens(h hasher, tx *utils.AutoTx, userID int64, access, refresh string) {
	const query = `
DELETE FROM sessions
WHERE user_id = $1
	AND access_hash = $2 AND refresh_hash = $3
`
	accessHash := h.AccessTokenHash(access)
	refreshHash := h.RefreshTokenHash(refresh)

	tx.Exec(query, userID, accessHash, refreshHash)
}

func requestAccessToken(h hasher, tx *utils.AutoTx, appID int64, code, redirectUri string, appSecret, verifier *string) middleware.Responder {
	authValue, found := authCache.Get(code)
	if !found {
		return postTokenBadRequest(models.OAuth2ErrorErrorInvalidRequest)
	}

	auth := authValue.(*authData)
	if auth.appID != appID {
		return postTokenBadRequest(models.OAuth2ErrorErrorUnrecognizedClient)
	}
	if auth.redirectUri != redirectUri {
		return postTokenBadRequest(models.OAuth2ErrorErrorInvalidRedirect)
	}

	if len(auth.secretHash) > 0 {
		if appSecret == nil {
			return postTokenBadRequest(models.OAuth2ErrorErrorInvalidRequest)
		}

		sum := h.AppSecretHash(*appSecret)
		if !bytes.Equal(auth.secretHash, sum) {
			return postTokenBadRequest(models.OAuth2ErrorErrorInvalidRequest)
		}
	}

	if auth.challenge != "" {
		if verifier == nil {
			return postTokenBadRequest(models.OAuth2ErrorErrorInvalidRequest)
		}

		switch auth.method {
		case "S256":
			sum := sha256.Sum256([]byte(*verifier))
			ch := base64.URLEncoding.EncodeToString(sum[:])
			if auth.challenge != ch {
				return postTokenBadRequest(models.OAuth2ErrorErrorInvalidRequest)
			}
		case "plain":
			fallthrough
		case "":
			if auth.challenge != *verifier {
				return postTokenBadRequest(models.OAuth2ErrorErrorInvalidRequest)
			}
		default:
			return postTokenBadRequest(models.OAuth2ErrorErrorInvalidRequest)
		}
	}

	if len(auth.access) > 0 {
		revokeTokens(h, tx, auth.userID, auth.access, auth.refresh)
		return postTokenBadRequest(models.OAuth2ErrorErrorInvalidRequest)
	}

	resp := createTokens(h, tx, auth.appID, auth.userID, auth.scope, auth.userName)
	if resp == nil {
		return postTokenBadRequest(models.OAuth2ErrorErrorServerError)
	}

	auth.access = resp.AccessToken
	auth.refresh = resp.RefreshToken

	return oauth2.NewPostOauth2TokenOK().WithPayload(resp)
}

func requestRefreshToken(h hasher, tx *utils.AutoTx, appID int64, appSecret, token string) middleware.Responder {
	const query = `
DELETE FROM sessions
WHERE app_id = $1 AND user_id = $2 AND refresh_hash = $3
RETURNING scope, refresh_thru
`

	granted, err := checkRefreshGrant(h, tx, appID, appSecret)
	if err != nil {
		return postTokenBadRequest(models.OAuth2ErrorErrorUnrecognizedClient)
	}
	if !granted {
		return postTokenBadRequest(models.OAuth2ErrorErrorInvalidGrant)
	}

	idToken := strings.Split(token, ".")
	if len(idToken) != 2 {
		return postTokenBadRequest(models.OAuth2ErrorErrorInvalidToken)
	}

	userID, err := strconv.ParseInt(idToken[0], 32, 32)
	if err != nil {
		return postTokenBadRequest(models.OAuth2ErrorErrorInvalidToken)
	}

	hash := h.RefreshTokenHash(token)

	var scope uint32
	var thru time.Time
	tx.Query(query, appID, userID, hash[:]).Scan(&scope, &thru)

	if scope == 0 || time.Now().After(thru) {
		return postTokenBadRequest(models.OAuth2ErrorErrorInvalidToken)
	}

	userName := tx.QueryString("SELECT name FROM users WHERE id = $1", userID)
	resp := createTokens(h, tx, appID, userID, scope, userName)
	if resp == nil {
		return postTokenBadRequest(models.OAuth2ErrorErrorServerError)
	}

	return oauth2.NewPostOauth2TokenOK().WithPayload(resp)
}

func newOAuth2Token(srv *utils.MindwellServer) func(oauth2.PostOauth2TokenParams) middleware.Responder {
	return func(params oauth2.PostOauth2TokenParams) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			if params.GrantType == "password" {
				if params.Username == nil || params.Password == nil {
					return postTokenBadRequest(models.OAuth2ErrorErrorInvalidRequest)
				}

				return requestPasswordToken(srv, tx, params.ClientID, *params.ClientSecret, *params.Username, *params.Password)
			}

			if params.GrantType == "authorization_code" {
				if params.Code == nil || params.RedirectURI == nil {
					return postTokenBadRequest(models.OAuth2ErrorErrorInvalidRequest)
				}

				return requestAccessToken(srv, tx, params.ClientID, *params.Code, *params.RedirectURI, params.ClientSecret, params.CodeVerifier)
			}

			if params.GrantType == "client_credentials" {
				if *params.ClientSecret == "" {
					return postTokenBadRequest(models.OAuth2ErrorErrorInvalidRequest)
				}

				return requestAppToken(srv, tx, params.ClientID, *params.ClientSecret)
			}

			if params.GrantType == "refresh_token" {
				if params.RefreshToken == nil {
					return postTokenBadRequest(models.OAuth2ErrorErrorInvalidRequest)
				}

				return requestRefreshToken(srv, tx, params.ClientID, *params.ClientSecret, *params.RefreshToken)
			}

			return postTokenBadRequest(models.OAuth2ErrorErrorUnsupportedGrantType)
		})
	}
}

func loadApp(tx *utils.AutoTx, appID int64) (*models.App, bool) {
	const query = `
SELECT name, show_name, platform, info
FROM apps
WHERE id = $1
`

	app := &models.App{ID: appID}
	tx.Query(query, appID).Scan(&app.Name, &app.ShowName, &app.Platform, &app.Info)

	return app, tx.Error() == nil
}

func newAppLoader(srv *utils.MindwellServer) func(oauth2.GetOauth2AppsIDParams, *models.UserID) middleware.Responder {
	return func(params oauth2.GetOauth2AppsIDParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			app, ok := loadApp(tx, params.ID)
			if !ok {
				err := &i18n.Message{ID: "no_app", Other: "App not found."}
				return oauth2.NewGetOauth2AppsIDNotFound().WithPayload(srv.NewError(err))
			}

			return oauth2.NewGetOauth2AppsIDOK().WithPayload(app)
		})
	}
}
