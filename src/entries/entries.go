package entries

import (
	"database/sql"
	"log"
	"regexp"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/restapi/operations/entries"
	"github.com/sevings/yummy-server/gen/restapi/operations/me"
	yummy "github.com/sevings/yummy-server/src"

	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations"
	"github.com/sevings/yummy-server/src/users"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.MePostUsersMeEntriesHandler = me.PostUsersMeEntriesHandlerFunc(newMyTlogPoster(db))
	api.EntriesGetEntriesLiveHandler = entries.GetEntriesLiveHandlerFunc(newLiveLoader(db))
	api.EntriesGetEntriesAnonymousHandler = entries.GetEntriesAnonymousHandlerFunc(newAnonymousLoader(db))
	api.EntriesGetEntriesBestHandler = entries.GetEntriesBestHandlerFunc(newBestLoader(db))
}

var wordRe *regexp.Regexp

func init() {
	wordRe = regexp.MustCompile("[a-zA-Zа-яА-ЯёЁ0-9]+")
}

const postEntryQuery = `
INSERT INTO entries (author_id, title, content, word_count, visible_for, is_votable)
VALUES ($1, $2, $3, $4, (SELECT id FROM entry_privacy WHERE type = $5), $6)
RETURNING id, created_at`

func createEntry(tx *sql.Tx, apiKey string, title, content, privacy string, isVotable bool) (*models.Entry, bool) {
	author, found := users.LoadAuthUser(tx, &apiKey)
	if !found {
		return nil, false
	}

	var wordCount int64
	contentWords := wordRe.FindAllStringIndex(content, -1)
	wordCount += int64(len(contentWords))

	titleWords := wordRe.FindAllStringIndex(title, -1)
	wordCount += int64(len(titleWords))

	if privacy == "followers" {
		privacy = models.EntryPrivacySome //! \todo add users to list
	}

	entry := models.Entry{
		Title:     title,
		Content:   content,
		WordCount: wordCount,
		Privacy:   privacy,
		Author:    author,
	}

	err := tx.QueryRow(postEntryQuery, author.ID, title, content, wordCount,
		privacy, isVotable).Scan(&entry.ID, &entry.CreatedAt)
	if err != nil {
		log.Print(err)
		return nil, false
	}

	return &entry, true
}

func newMyTlogPoster(db *sql.DB) func(me.PostUsersMeEntriesParams) middleware.Responder {
	return func(params me.PostUsersMeEntriesParams) middleware.Responder {
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}

		entry, created := createEntry(tx, params.XUserKey,
			*params.Title, params.Content, *params.Privacy, *params.IsVotable)

		if !created {
			tx.Rollback()
			return me.NewPostUsersMeEntriesForbidden()
		}

		tx.Commit()
		return me.NewPostUsersMeEntriesOK().WithPayload(entry)
	}
}

// CanViewEntry returns true if the user is allowed to read the entry.
func CanViewEntry(tx yummy.AutoTx, userID, entryID int64) bool {
	const q = `
		SELECT TRUE 
		FROM feed
		WHERE id = $2 AND (author_id = $1
			OR ((entry_privacy = 'all' 
				AND (author_privacy = 'all'
					OR (author_privacy = 'registered' AND $1 > 0)
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
	err := tx.QueryRow(q, userID, entryID).Scan(&allowed)
	if err != nil && err != sql.ErrNoRows {
		log.Print(err)
	}

	return allowed
}
