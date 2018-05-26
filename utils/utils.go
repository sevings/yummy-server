package utils

import (
	"database/sql"
	"log"
	"math/rand"

	goconf "github.com/zpatrick/go-config"

	"github.com/go-openapi/errors"
	"github.com/sevings/mindwell-server/models"

	// to use postgres
	_ "github.com/lib/pq"
)

var cfg *goconf.Config

// LoadConfig creates app config from file
func LoadConfig(fileName string) *goconf.Config {
	toml := goconf.NewTOMLFile(fileName + ".toml")
	loader := goconf.NewOnceLoader(toml)
	config := goconf.NewConfig([]goconf.Provider{loader})
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	cfg = config
	return config
}

// NewError returns error object with some message
func NewError(msg string) *models.Error {
	return &models.Error{Message: msg}
}

// CanViewEntry returns true if the user is allowed to read the entry.
func CanViewEntry(tx *AutoTx, userID, entryID int64) bool {
	const q = `
		SELECT TRUE 
		FROM feed
		WHERE id = $2 AND (author_id = $1
			OR ((entry_privacy = 'all' 
				AND (author_privacy = 'all'
					OR EXISTS(SELECT 1 FROM relation, relations, entries
							  WHERE from_id = $1 AND to_id = entries.author_id
								  AND entries.id = $2
						 		  AND relation.type = 'followed'
						 		  AND relations.type = relation.id)))
			OR (entry_privacy = 'some' 
				AND EXISTS(SELECT 1 FROM entries_privacy
					WHERE user_id = $1 AND entry_id = $2))
			OR entry_privacy = 'anonymous'))`

	var allowed bool
	tx.Query(q, userID, entryID).Scan(&allowed)

	return allowed
}

func NewKeyAuth(db *sql.DB) func(apiKey string) (*models.UserID, error) {
	const q = `
		SELECT id
		FROM users
		WHERE api_key = $1 AND valid_thru > CURRENT_TIMESTAMP`

	return func(apiKey string) (*models.UserID, error) {
		var id int64
		err := db.QueryRow(q, apiKey).Scan(&id)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Print(err)
			}

			return nil, errors.New(401, "Unauthorized")
		}

		userID := models.UserID(id)
		return &userID, nil
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// GenerateString returns random string
func GenerateString(length int) string {
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

func NewAvatar(avatar string) *models.Avatar {
	base, err := cfg.String("images.base_url")
	if err != nil {
		log.Print(err)
	}

	return &models.Avatar{
		X42:  base + "42/" + avatar,
		X124: base + "124/" + avatar,
	}
}

func ImagesFolder() string {
	folder, err := cfg.String("images.folder")
	if err != nil {
		log.Print(err)
	}

	return folder
}

func DefaultCover() string {
	cover, err := cfg.String("images.cover")
	if err != nil {
		log.Print(err)
	}

	return cover
}
