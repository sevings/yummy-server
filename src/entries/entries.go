package entries

import (
	"log"
	"database/sql"
	"regexp"

	"github.com/sevings/yummy/gen/models"
	"github.com/sevings/yummy/gen/restapi/operations"
	"github.com/sevings/yummy/src/users"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {

}

wordRe := regexp.MustCompile("[a-zA-Zа-яА-ЯёЁ0-9]+")

const postEntryQuery = `
INSERT INTO entries (author_id, title, content, word_count, 
    (SELECT "type" from entry_privacy WHERE id = $5, is_votable)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at`

func createEntry(tx *sql.Tx, apiKey *string, title, content, privacy string, isVotable bool) (*models.Entry, bool) {
	authorID, found := users.FindAuthUser(tx, apiKey)
	if !found {
		return nil, false
	}

	var wordCount int
	words := wordRe.findAllStringIndex(content, -1)
	if words == nil {
		wordCount = 0
	} else {
		wordCount = len(words) / 2
	}

	var entryID int64
	var createdAt string
	err := tx.QueryRow(postEntryQuery, authorID, title, content, wordCount, 
		privacy, isVotable).Scan(&entryID, &createdAt)
	if err != nil {
		log.Print(err)
		return nil, false
	}

	var entry models.Entry {
		ID: entryID, 
		CreatedAt: createdAt, 
		Title: title, 
		Content: content, 
		WordCount: wordCount,
		VisibleFor: privacy }
}
