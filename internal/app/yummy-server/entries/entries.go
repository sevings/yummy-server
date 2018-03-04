package entries

import (
	"database/sql"
	"regexp"

	"github.com/golang-commonmark/markdown"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/restapi/operations/entries"

	"github.com/sevings/yummy-server/internal/app/yummy-server/users"
	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	"github.com/sevings/yummy-server/internal/app/yummy-server/watchings"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.EntriesPostEntriesUsersMeHandler = entries.PostEntriesUsersMeHandlerFunc(newMyTlogPoster(db))
	api.EntriesPutEntriesIDHandler = entries.PutEntriesIDHandlerFunc(newEntryEditor(db))

	api.EntriesGetEntriesLiveHandler = entries.GetEntriesLiveHandlerFunc(newLiveLoader(db))
	api.EntriesGetEntriesAnonymousHandler = entries.GetEntriesAnonymousHandlerFunc(newAnonymousLoader(db))
	api.EntriesGetEntriesBestHandler = entries.GetEntriesBestHandlerFunc(newBestLoader(db))
	api.EntriesGetEntriesUsersIDHandler = entries.GetEntriesUsersIDHandlerFunc(newTlogLoader(db))
	api.EntriesGetEntriesUsersMeHandler = entries.GetEntriesUsersMeHandlerFunc(newMyTlogLoader(db))
	api.EntriesGetEntriesFriendsHandler = entries.GetEntriesFriendsHandlerFunc(newFriendsFeedLoader(db))
}

var wordRe *regexp.Regexp
var md *markdown.Markdown

func init() {
	wordRe = regexp.MustCompile("[a-zA-Zа-яА-ЯёЁ0-9]+")
	md = markdown.New(markdown.Typographer(false), markdown.Breaks(true))
}

func wordCount(content, title string) int64 {
	var wc int64
	contentWords := wordRe.FindAllStringIndex(content, -1)
	wc += int64(len(contentWords))

	titleWords := wordRe.FindAllStringIndex(title, -1)
	wc += int64(len(titleWords))

	return wc
}

func entryCategory(entry *models.Entry) string {
	if entry.WordCount > 100 {
		return "longread"
	}

	// media
	return "tweet"
}

func createEntry(tx *utils.AutoTx, userID int64, title, content, privacy string, isVotable bool) *models.Entry {
	if privacy == "followers" {
		privacy = models.EntryPrivacySome //! \todo add users to list
	}

	entry := models.Entry{
		Title:       title,
		Content:     md.RenderToString([]byte(content)),
		EditContent: content,
		WordCount:   wordCount(content, title),
		Privacy:     privacy,
		Vote:        models.EntryVoteBan,
		IsWatching:  true,
	}

	category := entryCategory(&entry)

	const q = `
	INSERT INTO entries (author_id, title, content, edit_content, word_count, visible_for, is_votable, category)
	VALUES ($1, $2, $3, $4, $5, (SELECT id FROM entry_privacy WHERE type = $6), 
		$7, (SELECT id from categories WHERE type = $8))
	RETURNING id, created_at`

	tx.Query(q, userID, title, entry.Content, entry.EditContent, entry.WordCount,
		privacy, isVotable, category).Scan(&entry.ID, &entry.CreatedAt)

	watchings.AddWatching(tx, userID, entry.ID)
	author := users.LoadUserByID(tx, userID)
	entry.Author = author

	return &entry
}

func newMyTlogPoster(db *sql.DB) func(entries.PostEntriesUsersMeParams, *models.UserID) middleware.Responder {
	return func(params entries.PostEntriesUsersMeParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			entry := createEntry(tx, int64(*uID),
				*params.Title, params.Content, params.Privacy, *params.IsVotable)

			if tx.Error() != nil {
				return entries.NewPostEntriesUsersMeForbidden()
			}

			return entries.NewPostEntriesUsersMeOK().WithPayload(entry)
		})
	}
}

func editEntry(tx *utils.AutoTx, entryID, userID int64, title, content, privacy string, isVotable bool) *models.Entry {
	if privacy == "followers" {
		privacy = models.EntryPrivacySome //! \todo add users to list
	}

	entry := models.Entry{
		ID:          entryID,
		Title:       title,
		Content:     md.RenderToString([]byte(content)),
		EditContent: content,
		WordCount:   wordCount(content, title),
		Privacy:     privacy,
		Vote:        models.EntryVoteBan,
		IsWatching:  true,
	}

	category := entryCategory(&entry)

	const q = `
	UPDATE entries
	SET title = $1, content = $2, edit_content = $3, word_count = $4, 
	visible_for = (SELECT id FROM entry_privacy WHERE type = $5), 
	is_votable = $6,
	category = (SELECT id from categories WHERE type = $7)
	WHERE id = $8 AND author_id = $9
	RETURNING created_at`

	tx.Query(q, title, entry.Content, entry.EditContent, entry.WordCount,
		privacy, isVotable, category, entryID, userID).Scan(&entry.CreatedAt)

	watchings.AddWatching(tx, userID, entry.ID)

	author := users.LoadUserByID(tx, userID)
	entry.Author = author

	return &entry
}

func newEntryEditor(db *sql.DB) func(entries.PutEntriesIDParams, *models.UserID) middleware.Responder {
	return func(params entries.PutEntriesIDParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			entry := editEntry(tx, params.ID, int64(*uID),
				*params.Title, params.Content, params.Privacy, *params.IsVotable)

			if tx.Error() != nil {
				return entries.NewPutEntriesIDForbidden()
			}

			return entries.NewPutEntriesIDOK().WithPayload(entry)
		})
	}
}
