package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/sevings/mindwell-server/models"
	"log"
	"strconv"
	"strings"
	"time"
)

const userIDQuery = `
			SELECT id, name, followers_count, 
				invited_by is not null, karma < -1, verified,
				invite_ban > CURRENT_DATE, vote_ban > CURRENT_DATE, 
				comment_ban > CURRENT_DATE, live_ban > CURRENT_DATE
			FROM users `

func LoadUserIDByID(tx *AutoTx, id int64) (*models.UserID, error) {
	const q = userIDQuery + "WHERE id = $1"
	tx.Query(q, id)
	return scanUserID(tx)
}

func LoadUserIDByName(tx *AutoTx, name string) (*models.UserID, error) {
	const q = userIDQuery + "WHERE lower(name) = lower($1)"
	tx.Query(q, name)
	return scanUserID(tx)
}

func LoadUserIDByApiKey(tx *AutoTx, apiKey string) (*models.UserID, error) {
	const q = userIDQuery + "WHERE api_key = $1 AND valid_thru > CURRENT_TIMESTAMP"
	tx.Query(q, apiKey)
	return scanUserID(tx)
}

func scanUserID(tx *AutoTx) (*models.UserID, error) {
	var user models.UserID
	user.Ban = &models.UserIDBan{}
	tx.Scan(&user.ID, &user.Name, &user.FollowersCount,
		&user.IsInvited, &user.NegKarma, &user.Verified,
		&user.Ban.Invite, &user.Ban.Vote, &user.Ban.Comment, &user.Ban.Live)
	if tx.Error() != nil {
		return nil, errUnauthorized
	}

	user.Ban.Invite = user.Ban.Invite || !user.IsInvited || !user.Verified
	user.Ban.Vote = user.Ban.Vote || !user.IsInvited || user.NegKarma || !user.Verified
	user.Ban.Comment = user.Ban.Comment || !user.IsInvited || !user.Verified
	user.Ban.Live = user.Ban.Live || !user.Verified

	return &user, nil
}

func readUserID(secret []byte, tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})

	if err != nil {
		log.Println(err)
		return 0, errUnauthorized
	}

	if !token.Valid {
		log.Printf("Invalid token: %s\n", tokenString)
		return 0, errUnauthorized

	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("Error get claims: %s\n", tokenString)
		return 0, errUnauthorized
	}

	if claims.Valid() != nil {
		return 0, errUnauthorized
	}

	id, err := strconv.ParseInt(claims["sub"].(string), 32, 64)
	if err != nil {
		log.Println(err)
		return 0, errUnauthorized
	}

	return id, nil
}

func NewKeyAuth(db *sql.DB, secret []byte) func(apiKey string) (*models.UserID, error) {
	return func(apiKey string) (*models.UserID, error) {
		if len(apiKey) < 32 {
			return nil, fmt.Errorf("api token is invalid: %s", apiKey)
		}

		tx := NewAutoTx(db)
		defer tx.Finish()

		if len(apiKey) == 32 {
			return LoadUserIDByApiKey(tx, apiKey)
		}

		id, err := readUserID(secret, apiKey)
		if err != nil {
			return nil, err
		}

		return LoadUserIDByID(tx, id)
	}
}

func NoApiKeyAuth(string) (*models.UserID, error) {
	return &models.UserID{
		Ban: &models.UserIDBan{
			Comment: true,
			Invite:  true,
			Live:    true,
			Vote:    true,
		},
	}, nil
}

func BuildApiToken(secret []byte, userID int64) (string, int64) {
	now := time.Now().Unix()
	exp := now + 60*60*24*365
	sub := strconv.FormatInt(userID, 32)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat": now,
		"exp": exp,
		"sub": sub,
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		log.Print(err)
	}

	return tokenString, exp
}

type AuthFlow uint8

const (
	AppFlow      AuthFlow = 1
	CodeFlow     AuthFlow = 2
	PasswordFlow AuthFlow = 4
)

var allScopes = [...]string{
	"account:read",
	"account:write",
	"adm:read",
	"adm:write",
	"comments:read",
	"comments:write",
	"entries:read",
	"entries:write",
	"favotites:write",
	"images:read",
	"images:write",
	"messages:read",
	"messages:write",
	"notifications:read",
	"relations:write",
	"settings:read",
	"settings:write",
	"users:read",
	"users:write",
	"votes:write",
	"watchings:write",
}

func findScope(scope string) (uint32, error) {
	for i, s := range allScopes {
		if scope == s {
			return 1 << i, nil
		}
	}

	return 0, fmt.Errorf("scope is invalid: %s", scope)
}

func ScopeFromString(scopes []string) (uint32, error) {
	var scope uint32

	for _, s := range scopes {
		n, err := findScope(s)
		if err != nil {
			return 0, err
		}

		scope += n
	}

	return scope, nil
}

func ScopeToString(scope uint32) []string {
	var scopes []string

	for i, s := range allScopes {
		if scope|1<<i == scope {
			scopes = append(scopes, s)
		}
	}

	return scopes
}

const AccessTokenLifetime = 60 * 60 * 24
const RefreshTokenLifetime = 60 * 60 * 24 * 30
const AppTokenLifetime = 60 * 60 * 24
const AccessTokenLength = 32
const RefreshTokenLength = 48
const AppTokenLength = 32
const CodeLength = 32

type AuthTokenHasher interface {
	AccessTokenHash(token string) []byte
	AppTokenHash(token string) []byte
}

func NewOAuth2User(h AuthTokenHasher, db *sql.DB, flowReq AuthFlow) func(string, []string) (*models.UserID, error) {
	const query = `
SELECT scope, flow, ban
FROM sessions
JOIN users ON users.id = user_id
JOIN apps ON apps.id = app_id
WHERE lower(users.name) = lower($1) 
	AND access_hash = $2
	AND access_thru > $3
`

	return func(token string, scopes []string) (*models.UserID, error) {
		scopeReq, err := ScopeFromString(scopes)
		if err != nil {
			return nil, err
		}

		nameToken := strings.Split(token, ".")
		if len(nameToken) < 2 {
			return nil, fmt.Errorf("access token is invalid: %s", token)
		}

		accessToken := nameToken[1]
		if len(accessToken) != AccessTokenLength {
			return nil, fmt.Errorf("access token is invalid: %s", token)
		}

		name := nameToken[0]
		hash := h.AccessTokenHash(token)
		now := time.Now()

		tx := NewAutoTx(db)
		defer tx.Finish()

		var scopeEx uint32
		var flowEx AuthFlow
		var ban bool
		tx.Query(query, name, hash[:], now).Scan(&scopeEx, &flowEx, &ban)
		if tx.Error() != nil {
			return nil, fmt.Errorf("access token is invalid: %s", token)
		}

		if ban || scopeEx&scopeReq != scopeReq || flowEx&flowReq != flowReq {
			return nil, errors.New("access denied")
		}

		return LoadUserIDByName(tx, name)
	}
}

func NewOAuth2App(h AuthTokenHasher, db *sql.DB) func(string) error {
	const query = `
SELECT ban, AuthFlow
FROM app_tokens
JOIN apps ON apps.id = app_id
WHERE lower(apps.name) = lower($1) 
	AND token_hash = $2
	AND valid_thru > $3
`

	return func(token string) error {
		nameToken := strings.Split(token, "+")
		if len(nameToken) < 2 {
			return fmt.Errorf("app token is invalid: %s", token)
		}

		appToken := nameToken[1]
		if len(appToken) != AppTokenLength {
			return fmt.Errorf("app token is invalid: %s", token)
		}

		name := nameToken[0]
		hash := h.AppTokenHash(token)
		now := time.Now()

		tx := NewAutoTx(db)
		defer tx.Finish()

		var ban bool
		var flowEx AuthFlow
		tx.Query(query, name, hash, now).Scan(&ban, &flowEx)
		if tx.Error() != nil {
			return fmt.Errorf("app token is invalid: %s", token)
		}

		if ban || flowEx&AppFlow != AppFlow {
			return errors.New("access denied")
		}

		return nil
	}
}
