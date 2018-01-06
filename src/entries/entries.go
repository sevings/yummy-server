package entries

import (
	"database/sql"
	"log"
	"regexp"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/restapi/operations/entries"
	"github.com/sevings/yummy-server/gen/restapi/operations/me"

	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations"
	yummy "github.com/sevings/yummy-server/src"
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

func createEntry(tx yummy.AutoTx, userID int64, title, content, privacy string, isVotable bool) (*models.Entry, bool) {
	author, _ := users.LoadUserByID(tx, userID)

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

func newMyTlogPoster(db *sql.DB) func(me.PostUsersMeEntriesParams, *models.UserID) middleware.Responder {
	return func(params me.PostUsersMeEntriesParams, uID *models.UserID) middleware.Responder {
		return yummy.Transact(db, func(tx yummy.AutoTx) (middleware.Responder, bool) {
			entry, created := createEntry(tx, int64(*uID),
				*params.Title, params.Content, *params.Privacy, *params.IsVotable)

			if !created {
				return me.NewPostUsersMeEntriesForbidden(), false
			}

			return me.NewPostUsersMeEntriesOK().WithPayload(entry), true
		})
	}
}
