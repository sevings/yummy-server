package entries

import (
	"database/sql"
	"regexp"

	"github.com/golang-commonmark/markdown"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/restapi/operations/entries"

	"github.com/sevings/mindwell-server/internal/app/mindwell-server/comments"
	"github.com/sevings/mindwell-server/internal/app/mindwell-server/users"
	"github.com/sevings/mindwell-server/internal/app/mindwell-server/watchings"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/utils"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	srv.API.EntriesPostEntriesUsersMeHandler = entries.PostEntriesUsersMeHandlerFunc(newMyTlogPoster(srv))

	srv.API.EntriesGetEntriesIDHandler = entries.GetEntriesIDHandlerFunc(newEntryLoader(srv))
	srv.API.EntriesPutEntriesIDHandler = entries.PutEntriesIDHandlerFunc(newEntryEditor(srv))
	srv.API.EntriesDeleteEntriesIDHandler = entries.DeleteEntriesIDHandlerFunc(newEntryDeleter(srv))

	srv.API.EntriesGetEntriesLiveHandler = entries.GetEntriesLiveHandlerFunc(newLiveLoader(srv))
	srv.API.EntriesGetEntriesAnonymousHandler = entries.GetEntriesAnonymousHandlerFunc(newAnonymousLoader(srv))
	srv.API.EntriesGetEntriesBestHandler = entries.GetEntriesBestHandlerFunc(newBestLoader(srv))
	srv.API.EntriesGetEntriesUsersIDHandler = entries.GetEntriesUsersIDHandlerFunc(newTlogLoader(srv))
	srv.API.EntriesGetEntriesUsersMeHandler = entries.GetEntriesUsersMeHandlerFunc(newMyTlogLoader(srv))
	srv.API.EntriesGetEntriesFriendsHandler = entries.GetEntriesFriendsHandlerFunc(newFriendsFeedLoader(srv))
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

func createEntry(srv *utils.MindwellServer, tx *utils.AutoTx, userID int64, title, content, privacy string, isVotable bool) *models.Entry {
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
	RETURNING id, extract(epoch from created_at)`

	tx.Query(q, userID, title, entry.Content, entry.EditContent, entry.WordCount,
		privacy, isVotable, category).Scan(&entry.ID, &entry.CreatedAt)

	watchings.AddWatching(tx, userID, entry.ID)
	author := users.LoadUserByID(srv, tx, userID)
	entry.Author = author

	return &entry
}

func newMyTlogPoster(srv *utils.MindwellServer) func(entries.PostEntriesUsersMeParams, *models.UserID) middleware.Responder {
	return func(params entries.PostEntriesUsersMeParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			entry := createEntry(srv, tx, int64(*uID),
				*params.Title, params.Content, params.Privacy, *params.IsVotable)

			if tx.Error() != nil {
				return entries.NewPostEntriesUsersMeForbidden()
			}

			return entries.NewPostEntriesUsersMeCreated().WithPayload(entry)
		})
	}
}

func editEntry(srv *utils.MindwellServer, tx *utils.AutoTx, entryID, userID int64, title, content, privacy string, isVotable bool) *models.Entry {
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
		IsVotable:   isVotable,
	}

	category := entryCategory(&entry)

	const q = `
	UPDATE entries
	SET title = $1, content = $2, edit_content = $3, word_count = $4, 
	visible_for = (SELECT id FROM entry_privacy WHERE type = $5), 
	is_votable = $6,
	category = (SELECT id from categories WHERE type = $7)
	WHERE id = $8 AND author_id = $9
	RETURNING extract(epoch from created_at)`

	tx.Query(q, title, entry.Content, entry.EditContent, entry.WordCount,
		privacy, isVotable, category, entryID, userID).Scan(&entry.CreatedAt)

	watchings.AddWatching(tx, userID, entry.ID)

	author := users.LoadUserByID(srv, tx, userID)
	entry.Author = author

	return &entry
}

func newEntryEditor(srv *utils.MindwellServer) func(entries.PutEntriesIDParams, *models.UserID) middleware.Responder {
	return func(params entries.PutEntriesIDParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			entry := editEntry(srv, tx, params.ID, int64(*uID),
				*params.Title, params.Content, params.Privacy, *params.IsVotable)

			if tx.Error() != nil {
				return entries.NewPutEntriesIDForbidden()
			}

			return entries.NewPutEntriesIDOK().WithPayload(entry)
		})
	}
}

func entryVoteStatus(authorID, userID int64, vote sql.NullFloat64) string {
	switch {
	case authorID == userID:
		return models.EntryVoteBan
	case !vote.Valid:
		return models.EntryVoteNot
	case vote.Float64 > 0:
		return models.EntryVotePos
	default:
		return models.EntryVoteNeg
	}
}

func loadEntry(srv *utils.MindwellServer, tx *utils.AutoTx, entryID, userID int64) *models.Entry {
	const q = tlogFeedQueryStart + `
		WHERE entries.id = $2
			AND (entries.author_id = $1
				OR entry_privacy.type = 'all' 
				OR (entry_privacy.type = 'some' 
					AND EXISTS(SELECT 1 from entries_privacy WHERE user_id = $1 AND entry_id = entries.id)))
		`

	var entry models.Entry
	var author models.User
	var vote sql.NullFloat64
	var avatar string
	tx.Query(q, userID, entryID).Scan(&entry.ID, &entry.CreatedAt, &entry.Rating, &entry.Votes,
		&entry.Title, &entry.Content, &entry.EditContent, &entry.WordCount,
		&entry.Privacy,
		&entry.IsVotable, &entry.CommentCount,
		&author.ID, &author.Name, &author.ShowName,
		&author.IsOnline,
		&avatar,
		&vote, &entry.IsFavorited, &entry.IsWatching)

	if author.ID != userID {
		entry.EditContent = ""
	}

	entry.Vote = entryVoteStatus(author.ID, userID, vote)

	author.Avatar = srv.NewAvatar(avatar)
	entry.Author = &author

	cmt := comments.LoadEntryComments(srv, tx, userID, entryID, 5, "", "")
	entry.Comments = cmt

	return &entry
}

func newEntryLoader(srv *utils.MindwellServer) func(entries.GetEntriesIDParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesIDParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			entry := loadEntry(srv, tx, params.ID, int64(*uID))

			if entry.ID == 0 {
				return entries.NewGetEntriesIDNotFound()
			}

			return entries.NewGetEntriesIDOK().WithPayload(entry)
		})
	}
}

func deleteEntry(tx *utils.AutoTx, entryID, userID int64) bool {
	var authorID int64
	tx.Query("SELECT author_id FROM entries WHERE id = $1", entryID).Scan(&authorID)
	if authorID != userID {
		return false
	}

	tx.Exec("DELETE from entries WHERE id = $1", entryID)
	return true
}

func newEntryDeleter(srv *utils.MindwellServer) func(entries.DeleteEntriesIDParams, *models.UserID) middleware.Responder {
	return func(params entries.DeleteEntriesIDParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			ok := deleteEntry(tx, params.ID, int64(*uID))
			if ok {
				return entries.NewDeleteEntriesIDOK()
			}

			if tx.Error() == sql.ErrNoRows {
				return entries.NewDeleteEntriesIDNotFound()
			}

			return entries.NewDeleteEntriesIDForbidden()
		})
	}
}
