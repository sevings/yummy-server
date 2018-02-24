package utils

import (
	"log"

	goconf "github.com/zpatrick/go-config"

	"github.com/sevings/yummy-server/models"

	// to use postgres
	_ "github.com/lib/pq"
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
