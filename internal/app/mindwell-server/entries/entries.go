package entries

import (
	"database/sql"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/golang-commonmark/markdown"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/restapi/operations/entries"
	"github.com/sevings/mindwell-server/restapi/operations/me"
	usersAPI "github.com/sevings/mindwell-server/restapi/operations/users"

	"github.com/sevings/mindwell-server/internal/app/mindwell-server/comments"
	"github.com/sevings/mindwell-server/internal/app/mindwell-server/users"
	"github.com/sevings/mindwell-server/internal/app/mindwell-server/watchings"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/utils"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	srv.API.MePostMeTlogHandler = me.PostMeTlogHandlerFunc(newMyTlogPoster(srv))
	srv.API.MeGetMeTlogHandler = me.GetMeTlogHandlerFunc(newMyTlogLoader(srv))
	srv.API.UsersGetUsersNameTlogHandler = usersAPI.GetUsersNameTlogHandlerFunc(newTlogLoader(srv))

	srv.API.MeGetMeFavoritesHandler = me.GetMeFavoritesHandlerFunc(newMyFavoritesLoader(srv))
	srv.API.UsersGetUsersNameFavoritesHandler = usersAPI.GetUsersNameFavoritesHandlerFunc(newTlogFavoritesLoader(srv))

	srv.API.EntriesGetEntriesIDHandler = entries.GetEntriesIDHandlerFunc(newEntryLoader(srv))
	srv.API.EntriesPutEntriesIDHandler = entries.PutEntriesIDHandlerFunc(newEntryEditor(srv))
	srv.API.EntriesDeleteEntriesIDHandler = entries.DeleteEntriesIDHandlerFunc(newEntryDeleter(srv))

	srv.API.EntriesGetEntriesLiveHandler = entries.GetEntriesLiveHandlerFunc(newLiveLoader(srv))
	srv.API.EntriesGetEntriesAnonymousHandler = entries.GetEntriesAnonymousHandlerFunc(newAnonymousLoader(srv))
	srv.API.EntriesGetEntriesBestHandler = entries.GetEntriesBestHandlerFunc(newBestLoader(srv))
	srv.API.EntriesGetEntriesFriendsHandler = entries.GetEntriesFriendsHandlerFunc(newFriendsFeedLoader(srv))
	srv.API.EntriesGetEntriesWatchingHandler = entries.GetEntriesWatchingHandlerFunc(newWatchingLoader(srv))
}

var wordRe *regexp.Regexp
var imgRe *regexp.Regexp
var md *markdown.Markdown

func init() {
	wordRe = regexp.MustCompile("[a-zA-Zа-яА-ЯёЁ0-9]+")
	imgRe = regexp.MustCompile("!\\[[^\\]]*\\]\\([^\\)]+\\)")

	markdown.RegisterCoreRule(250, appendTargetToLinks)
	md = markdown.New(markdown.Typographer(false), markdown.Breaks(true), markdown.Tables(false))
}

func appendTargetToLinks(s *markdown.StateCore) {
	for _, token := range s.Tokens {
		inline, ok := token.(*markdown.Inline)
		if !ok {
			continue
		}

		for _, tok := range inline.Children {
			link, ok := tok.(*markdown.LinkOpen)
			if !ok {
				continue
			}

			link.Target = "_blank"
		}
	}
}

func wordCount(content, title string) int64 {
	var wc int64

	content = imgRe.ReplaceAllLiteralString(content, " ")
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

	images := imgRe.FindAllStringIndex(entry.EditContent, -1)
	if len(images) > 0 && entry.WordCount < 50 {
		return "media"
	}

	return "tweet"
}

func NewEntry(title, content string) *models.Entry {
	title = strings.TrimSpace(title)
	title = bluemonday.StrictPolicy().Sanitize(title)

	const titleLength = 100
	const titleFormat = "%.100s"
	cutTitle, isTitleCut := utils.CutText(title, titleFormat, titleLength)

	content = strings.TrimSpace(content)

	const contentLength = 500
	const contentFormat = "%.500s"
	cutContent, isContentCut := utils.CutText(content, contentFormat, contentLength)

	hasCut := isTitleCut || isContentCut
	if !hasCut {
		cutTitle = ""
		cutContent = ""
	}

	entry := models.Entry{
		Title:       title,
		CutTitle:    cutTitle,
		Content:     md.RenderToString([]byte(content)),
		CutContent:  md.RenderToString([]byte(cutContent)),
		EditContent: content,
		HasCut:      hasCut,
		WordCount:   wordCount(content, title),
	}

	return &entry
}

func myEntry(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, title, content, privacy string, isVotable, inLive bool) *models.Entry {
	if privacy == "followers" {
		privacy = models.EntryPrivacySome //! \todo add users to list
	}

	if privacy == models.EntryPrivacyMe {
		isVotable = false
		inLive = false
	}

	entry := NewEntry(title, content)

	entry.Author = users.LoadUserByID(srv, tx, userID.ID)
	entry.Privacy = privacy
	entry.IsWatching = true
	entry.InLive = inLive
	entry.Rating = &models.Rating{
		IsVotable: isVotable,
	}
	entry.Rights = &models.EntryRights{
		Edit:    true,
		Delete:  true,
		Comment: true,
		Vote:    false,
	}

	return entry
}

func createEntry(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, title, content, privacy string, isVotable, inLive bool) *models.Entry {
	entry := myEntry(srv, tx, userID, title, content, privacy, isVotable, inLive)

	category := entryCategory(entry)

	const q = `
	INSERT INTO entries (author_id, title, cut_title, content, cut_content, edit_content, 
		has_cut, word_count, visible_for, is_votable, in_live, category)
	VALUES ($1, $2, $3, $4, $5, $6,$7, $8,
		(SELECT id FROM entry_privacy WHERE type = $9), 
		$10, $11, (SELECT id from categories WHERE type = $12))
	RETURNING id, extract(epoch from created_at)`

	tx.Query(q, userID.ID, entry.Title, entry.CutTitle, entry.Content, entry.CutContent, entry.EditContent,
		entry.HasCut, entry.WordCount, entry.Privacy, entry.Rating.IsVotable, entry.InLive, category).
		Scan(&entry.ID, &entry.CreatedAt)

	entry.Rating.ID = entry.ID
	watchings.AddWatching(tx, userID.ID, entry.ID)

	return entry
}

func allowedInLive(followersCount, entryCount int64) bool {
	switch {
	case followersCount < 3:
		return entryCount < 1
	case followersCount < 10:
		return entryCount < 2
	case followersCount < 50:
		return entryCount < 3
	default:
		return true
	}
}

func allowedWithoutVoting(srv *utils.MindwellServer, userID *models.UserID) *models.Error {
	if userID.IsInvited {
		return nil
	}

	return srv.NewError(&i18n.Message{ID: "post_wo_voting", Other: "You're not allowed to post without voting."})
}

func canPostInLive(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID) *models.Error {
	if userID.Ban.Live {
		return srv.NewError(&i18n.Message{ID: "post_in_live", Other: "You're not allowed to post in live."})
	}

	if userID.NegKarma {
		return srv.NewError(&i18n.Message{ID: "post_in_live_karma", Other: "You're not allowed to post in live."})
	}

	var entryCount int64
	const countQ = `SELECT count(*) FROM entries WHERE author_id = $1 
		AND date_trunc('day', created_at) = CURRENT_DATE AND in_live
	`
	tx.Query(countQ, userID.ID).Scan(&entryCount)

	if !allowedInLive(userID.FollowersCount, entryCount) {
		return srv.NewError(&i18n.Message{ID: "post_in_live_followers", Other: "You can't post in live anymore today."})
	}

	return nil
}

func newMyTlogPoster(srv *utils.MindwellServer) func(me.PostMeTlogParams, *models.UserID) middleware.Responder {
	return func(params me.PostMeTlogParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			if *params.InLive && params.Privacy == models.EntryPrivacyAll {
				err := canPostInLive(srv, tx, userID)
				if err != nil {
					return me.NewPostMeTlogForbidden().WithPayload(err)
				}
			}

			if !*params.IsVotable {
				err := allowedWithoutVoting(srv, userID)
				if err != nil {
					return me.NewPostMeTlogForbidden().WithPayload(err)
				}
			}

			entry := createEntry(srv, tx, userID,
				*params.Title, params.Content, params.Privacy, *params.IsVotable, *params.InLive)

			if tx.Error() != nil {
				err := srv.NewError(nil)
				return me.NewPostMeTlogForbidden().WithPayload(err)
			}

			return me.NewPostMeTlogCreated().WithPayload(entry)
		})
	}
}

func editEntry(srv *utils.MindwellServer, tx *utils.AutoTx, entryID int64, userID *models.UserID, title, content, privacy string, isVotable, inLive bool) *models.Entry {
	entry := myEntry(srv, tx, userID, title, content, privacy, isVotable, inLive)

	category := entryCategory(entry)

	const q = `
	UPDATE entries
	SET title = $1, cut_title = $2, content = $3, cut_content = $4, edit_content = $5, has_cut = $6, 
	word_count = $7, 
	visible_for = (SELECT id FROM entry_privacy WHERE type = $8), 
	is_votable = $9, in_live = $10,
	category = (SELECT id from categories WHERE type = $11)
	WHERE id = $12 AND author_id = $13
	RETURNING extract(epoch from created_at)`

	tx.Query(q, entry.Title, entry.CutTitle, entry.Content, entry.CutContent, entry.EditContent, entry.HasCut,
		entry.WordCount, entry.Privacy, entry.Rating.IsVotable, entry.InLive, category, entryID, userID.ID).
		Scan(&entry.CreatedAt)

	entry.ID = entryID
	entry.Rating.ID = entryID
	watchings.AddWatching(tx, userID.ID, entry.ID)

	return entry
}

func canEditInLive(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, entryID int64) *models.Error {
	var inLive bool
	const entryQ = "SELECT in_live FROM entries WHERE id = $1"
	tx.Query(entryQ, entryID).Scan(&inLive)
	if inLive {
		return nil
	}

	if userID.Ban.Live {
		return srv.NewError(&i18n.Message{ID: "edit_in_live", Other: "You are not allowed to post in live."})
	}

	if userID.NegKarma {
		return srv.NewError(&i18n.Message{ID: "edit_in_live_karma", Other: "You are not allowed to post in live."})
	}

	var entryCount int64
	const countQ = `
		SELECT count(*)
		FROM entries, 
			(
				SELECT created_at
				FROM entries
				WHERE id = $2
			) AS entry
		WHERE author_id = $1 
			AND date_trunc('day', entries.created_at) = date_trunc('day', entry.created_at)
			AND in_live
	`
	tx.Query(countQ, userID.ID, entryID).Scan(&entryCount)

	if !allowedInLive(userID.FollowersCount, entryCount) {
		return srv.NewError(&i18n.Message{ID: "edit_in_live_followers", Other: "You can't post in live anymore on this day."})
	}

	return nil
}

func newEntryEditor(srv *utils.MindwellServer) func(entries.PutEntriesIDParams, *models.UserID) middleware.Responder {
	return func(params entries.PutEntriesIDParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			if *params.InLive && params.Privacy == models.EntryPrivacyAll {
				err := canEditInLive(srv, tx, uID, params.ID)
				if err != nil {
					return entries.NewPutEntriesIDForbidden().WithPayload(err)
				}
			}

			if !*params.IsVotable {
				err := allowedWithoutVoting(srv, uID)
				if err != nil {
					return entries.NewPutEntriesIDForbidden().WithPayload(err)
				}
			}

			entry := editEntry(srv, tx, params.ID, uID,
				*params.Title, params.Content, params.Privacy, *params.IsVotable, *params.InLive)

			if tx.Error() != nil {
				err := srv.NewError(&i18n.Message{ID: "edit_not_your_entry", Other: "You can't edit someone else's entries."})
				return entries.NewPutEntriesIDForbidden().WithPayload(err)
			}

			return entries.NewPutEntriesIDOK().WithPayload(entry)
		})
	}
}

func entryVoteStatus(vote sql.NullFloat64) int64 {
	switch {
	case !vote.Valid:
		return 0
	case vote.Float64 > 0:
		return 1
	default:
		return -1
	}
}

func setEntryRights(entry *models.Entry, userID *models.UserID) {
	entry.Rights = &models.EntryRights{
		Edit:    entry.Author.ID == userID.ID,
		Delete:  entry.Author.ID == userID.ID,
		Comment: entry.Author.ID == userID.ID || !userID.Ban.Comment,
		Vote:    entry.Author.ID != userID.ID && !userID.Ban.Vote && entry.Rating.IsVotable,
	}
}

func LoadEntry(srv *utils.MindwellServer, tx *utils.AutoTx, entryID int64, userID *models.UserID) *models.Entry {
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
	var rating models.Rating
	tx.Query(q, userID.ID, entryID).Scan(&entry.ID, &entry.CreatedAt,
		&rating.Rating, &rating.UpCount, &rating.DownCount,
		&entry.Title, &entry.CutTitle, &entry.Content, &entry.CutContent, &entry.EditContent,
		&entry.HasCut, &entry.WordCount, &entry.Privacy,
		&rating.IsVotable, &entry.InLive, &entry.CommentCount,
		&author.ID, &author.Name, &author.ShowName,
		&author.IsOnline,
		&avatar,
		&vote, &entry.IsFavorited, &entry.IsWatching)

	if author.ID != userID.ID {
		entry.EditContent = ""
	}

	rating.Vote = entryVoteStatus(vote)

	rating.ID = entry.ID
	entry.Rating = &rating

	author.Avatar = srv.NewAvatar(avatar)
	entry.Author = &author

	cmt := comments.LoadEntryComments(srv, tx, userID, entryID, 5, "", "")
	entry.Comments = cmt

	setEntryRights(&entry, userID)

	return &entry
}

func newEntryLoader(srv *utils.MindwellServer) func(entries.GetEntriesIDParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesIDParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			entry := LoadEntry(srv, tx, params.ID, userID)

			if entry.ID == 0 {
				err := srv.StandardError("no_entry")
				return entries.NewGetEntriesIDNotFound().WithPayload(err)
			}

			return entries.NewGetEntriesIDOK().WithPayload(entry)
		})
	}
}

func deleteEntry(srv *utils.MindwellServer, tx *utils.AutoTx, entryID, userID int64) bool {
	var allowed bool
	tx.Query("SELECT author_id = $2 FROM entries WHERE id = $1", entryID, userID).Scan(&allowed)
	if !allowed {
		return false
	}

	tx.Query("SELECT id FROM comments WHERE entry_id = $1", entryID)

	commentIds := []int64{}

	var id int64
	for tx.Scan(&id) {
		commentIds = append(commentIds, id)
	}

	for _, id := range commentIds {
		srv.Ntf.NotifyRemove(tx, id, models.NotificationTypeComment)
	}

	tx.Exec("DELETE from entries WHERE id = $1", entryID)

	return true
}

func newEntryDeleter(srv *utils.MindwellServer) func(entries.DeleteEntriesIDParams, *models.UserID) middleware.Responder {
	return func(params entries.DeleteEntriesIDParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			ok := deleteEntry(srv, tx, params.ID, uID.ID)
			if ok {
				return entries.NewDeleteEntriesIDOK()
			}

			if tx.Error() == sql.ErrNoRows {
				err := srv.StandardError("no_entry")
				return entries.NewDeleteEntriesIDNotFound().WithPayload(err)
			}

			err := srv.NewError(&i18n.Message{ID: "delete_not_your_entry", Other: "You can't delete someone else's entries."})
			return entries.NewDeleteEntriesIDForbidden().WithPayload(err)
		})
	}
}
