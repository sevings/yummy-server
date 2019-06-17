package utils

import (
	crypto "crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"

	jwt "github.com/dgrijalva/jwt-go"
	goconf "github.com/zpatrick/go-config"

	"github.com/go-openapi/errors"
	"github.com/sevings/mindwell-server/models"

	// to use postgres
	_ "github.com/lib/pq"
)

var errUnauthorized = errors.New(401, "Unauthorized")
var htmlEsc *strings.Replacer = strings.NewReplacer(
	"<", "&lt;",
	">", "&gt;",
	"&", "&amp;",
	"\"", "&#34;",
	"'", "&#39;",
	"\n", "<br>",
	"\r", "",
)

// LoadConfig creates app config from file
func LoadConfig(fileName string) *goconf.Config {
	toml := goconf.NewTOMLFile(fileName + ".toml")
	loader := goconf.NewOnceLoader(toml)
	config := goconf.NewConfig([]goconf.Provider{loader})
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}

	return config
}

// CanViewEntry returns true if the user is allowed to read the entry.
func CanViewEntry(tx *AutoTx, userID, entryID int64) bool {
	return tx.QueryBool("SELECT can_view_entry($1, $2)", userID, entryID)
}

// IsOpenForMe returns true if you can see \param name tlog
func IsOpenForMe(tx *AutoTx, userID *models.UserID, name interface{}) bool {
	return tx.QueryBool("SELECT can_view_tlog($1, $2)", userID.ID, name)
}

func loadUserID(db *sql.DB, apiKey string) (*models.UserID, error) {
	const q = `
			SELECT id, name, followers_count, invited_by is not null, karma < -1,
				invite_ban > CURRENT_DATE, vote_ban > CURRENT_DATE, 
				comment_ban > CURRENT_DATE, live_ban > CURRENT_DATE
			FROM users
			WHERE api_key = $1 AND valid_thru > CURRENT_TIMESTAMP`

	var user models.UserID
	user.Ban = &models.UserIDBan{}
	err := db.QueryRow(q, apiKey).Scan(&user.ID, &user.Name, &user.FollowersCount, &user.IsInvited, &user.NegKarma,
		&user.Ban.Invite, &user.Ban.Vote, &user.Ban.Comment, &user.Ban.Live)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}

		return nil, errUnauthorized
	}

	user.Ban.Invite = user.Ban.Invite || !user.IsInvited
	user.Ban.Vote = user.Ban.Vote || !user.IsInvited
	user.Ban.Comment = user.Ban.Comment || !user.IsInvited
	user.Ban.Live = user.Ban.Live || !user.IsInvited

	return &user, nil
}

func readUserID(secret []byte, tokenString string) (*models.UserID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})

	if err != nil {
		log.Println(err)
		return nil, errUnauthorized
	}

	if !token.Valid {
		log.Printf("Invalid token: %s\n", tokenString)
		return nil, errUnauthorized

	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("Error get claims: %s\n", tokenString)
		return nil, errUnauthorized
	}

	now := time.Now().Unix()
	exp := int64(claims["exp"].(float64))
	if exp < now {
		return nil, errUnauthorized
	}

	id := models.UserID{
		ID:   int64(claims["id"].(float64)),
		Name: claims["name"].(string),
	}

	return &id, nil
}

func NewKeyAuth(db *sql.DB, secret []byte) func(apiKey string) (*models.UserID, error) {
	return func(apiKey string) (*models.UserID, error) {
		if len(apiKey) == 32 {
			return loadUserID(db, apiKey)
		}

		return readUserID(secret, apiKey)
	}
}

func BuildApiToken(secret []byte, userID *models.UserID) (string, int64) {
	now := time.Now().Unix()
	exp := now + 60*60*24*180
	thru := now + 60*60

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":  exp,
		"thru": thru,
		"id":   userID.ID,
		"name": userID.Name,
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		log.Print(err)
	}

	return tokenString, exp
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// GenerateString returns random string
func GenerateString(length int) string {
	bytesLen := length*6/8 + 1
	data := make([]byte, bytesLen)
	_, err := crypto.Read(data)
	if err == nil {
		str := base64.URLEncoding.EncodeToString(data)
		return str[:length]
	}

	log.Print(err)

	// fallback on error
	b := make([]byte, length)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := len(b)-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func CutText(text string, size int) (string, bool) {
	if len(text) <= size {
		return text, false
	}

	text = string([]rune(text)[:size])

	var isSpace bool
	trim := func(r rune) bool {
		if isSpace {
			return unicode.IsSpace(r)
		}

		isSpace = unicode.IsSpace(r)
		return true
	}
	text = strings.TrimRightFunc(text, trim)

	text += "..."

	return text, true
}

func ParseFloat(val string) float64 {
	res, err := strconv.ParseFloat(val, 64)
	if len(val) > 0 && err != nil {
		log.Printf("error parse float: '%s'", val)
	}

	return res
}

func FormatFloat(val float64) string {
	return strconv.FormatFloat(val, 'f', 6, 64)
}

func ReplaceToHtml(val string) string {
	return htmlEsc.Replace(val)
}

func HideEmail(email string) string {
	nameLen := strings.LastIndex(email, "@")

	if nameLen < 0 {
		return ""
	}

	if nameLen < 3 {
		return "***" + email[nameLen:]
	}

	return email[:1] + "***" + email[nameLen-1:]
}
