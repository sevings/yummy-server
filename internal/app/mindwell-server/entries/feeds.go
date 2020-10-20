package entries

import (
	"database/sql"
	"github.com/go-openapi/runtime/middleware"
	"github.com/leporo/sqlf"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/entries"
	"github.com/sevings/mindwell-server/restapi/operations/me"
	"github.com/sevings/mindwell-server/restapi/operations/users"
	"github.com/sevings/mindwell-server/utils"
	"strings"
)

type addQuery func(stmt *sqlf.Stmt)

func baseFeedQuery(userID, limit int64) *sqlf.Stmt {
	return sqlf.Select("entries.id, extract(epoch from entries.created_at) as created_at").
		Select("rating, entries.up_votes, entries.down_votes").
		Select("entries.title, edit_content").
		Select("word_count, entry_privacy.type as privacy").
		Select("is_votable, in_live, entries.comments_count").
		Select("author_id, users.name, users.show_name").
		Select("is_online(users.last_seen_at) as is_online").
		Select("users.avatar").
		From("entries").
		Join("users", "entries.author_id = users.id").
		Join("entry_privacy", "entries.visible_for = entry_privacy.id").
		With("my_favorites",
			sqlf.Select("entry_id").From("favorites").Where("user_id = ?", userID)).
		LeftJoin("my_favorites", "my_favorites.entry_id = entries.id").
		With("my_watching",
			sqlf.Select("entry_id").From("watching").Where("user_id = ?", userID)).
		LeftJoin("my_watching", "my_watching.entry_id = entries.id").
		Limit(limit)
}

func addTagQuery(q *sqlf.Stmt, tag string) *sqlf.Stmt {
	tag = strings.TrimSpace(tag)
	tag = strings.ToLower(tag)

	if tag == "" {
		return q
	}

	return q.Join("entry_tags", "entries.id = entry_tags.entry_id").
		Join("tags", "entry_tags.tag_id = tags.id").
		Where("tags.tag = ?", tag)
}

func addSearchQuery(q *sqlf.Stmt, query string) *sqlf.Stmt {
	query = strings.TrimSpace(query)

	if query == "" {
		return q
	}

	sub := q.
		Select("to_search_vector(entries.title, entries.edit_content) <=> to_tsquery('russian', ?) AS rum_dist", query).
		Where("to_search_vector(entries.title, entries.edit_content) @@ to_tsquery('russian', ?)", query).
		OrderBy("rum_dist")

	return sqlf.Select("id, created_at").
		Select("rating, up_votes, down_votes").
		Select("title, edit_content").
		Select("word_count, privacy").
		Select("is_votable, in_live, comments_count").
		Select("author_id, name, show_name").
		Select("is_online").
		Select("avatar").
		Select("vote, is_favorite, is_watching").
		From("").SubQuery("(", ") AS entries", sub)
}

func myTlogQuery(userID, limit int64, tag string) *sqlf.Stmt {
	q := baseFeedQuery(userID, limit)
	addTagQuery(q, tag)
	return q.
		Select("NULL as vote").
		Select("my_favorites.entry_id IS NOT NULL as is_favorite").
		Select("my_watching.entry_id IS NOT NULL as is_watching").
		Where("entries.author_id = ?", userID)
}

func feedQuery(userID, limit int64) *sqlf.Stmt {
	return baseFeedQuery(userID, limit).
		Select("my_votes.vote").
		Select("my_favorites.entry_id IS NOT NULL as is_favorite").
		Select("my_watching.entry_id IS NOT NULL as is_watching").
		With("my_votes",
			sqlf.Select("entry_id, vote").From("entry_votes").Where("user_id = ?", userID)).
		LeftJoin("my_votes", "my_votes.entry_id = entries.id").
		Join("user_privacy", "users.privacy = user_privacy.id")
}

func addRelationsToMeQuery(q *sqlf.Stmt, userID int64) *sqlf.Stmt {
	return q.
		With("relations_to_me",
			sqlf.Select("relation.type, relations.from_id").
				From("relations").
				Join("relation", "relations.type = relation.id").
				Where("relations.to_id = ?", userID)).
		LeftJoin("relations_to_me", "relations_to_me.from_id = entries.author_id")
}

func addRelationsFromMeQuery(q *sqlf.Stmt, userID int64) *sqlf.Stmt {
	return q.
		With("relations_from_me",
			sqlf.Select("relation.type, relations.to_id").
				From("relations").
				Join("relation", "relations.type = relation.id").
				Where("relations.from_id = ?", userID)).
		LeftJoin("relations_from_me", "relations_from_me.to_id = entries.author_id")
}

func addLiveQuery(q *sqlf.Stmt, userID int64, tag string) *sqlf.Stmt {
	addRelationsToMeQuery(q, userID)
	addRelationsFromMeQuery(q, userID)
	addTagQuery(q, tag)

	return q.
		Where("entry_privacy.type = 'all'").
		Where("entries.in_live").
		Where("(user_privacy.type = 'all' OR (user_privacy.type = 'invited' AND (SELECT invited_by IS NOT NULL FROM users WHERE id = ?)))", userID).
		Where("(relations_to_me.type IS NULL OR relations_to_me.type <> 'ignored')").
		Where("(relations_from_me.type IS NULL OR relations_from_me.type NOT IN ('ignored', 'hidden'))")
}

func AddLiveInvitedQuery(q *sqlf.Stmt, userID int64, tag string) *sqlf.Stmt {
	return addLiveQuery(q, userID, tag).
		Where("users.invited_by IS NOT NULL")
}

func liveInvitedQuery(userID, limit int64, tag string) *sqlf.Stmt {
	q := feedQuery(userID, limit)
	return AddLiveInvitedQuery(q, userID, tag)
}

func addLiveWaitingQuery(q *sqlf.Stmt, userID int64, tag string) *sqlf.Stmt {
	return addLiveQuery(q, userID, tag).
		Where("users.invited_by IS NULL")
}

func addLiveCommentsQuery(q *sqlf.Stmt, userID int64, tag string) *sqlf.Stmt {
	return addLiveQuery(q, userID, tag).
		Where("users.invited_by IS NOT NULL").
		Where("entries.comments_count > 0").
		OrderBy("last_comment DESC")
}

func liveCommentsQuery(userID, limit int64, tag string) *sqlf.Stmt {
	q := feedQuery(userID, limit)
	return addLiveCommentsQuery(q, userID, tag)
}

const isEntryOpenQueryWhere = `
(entry_privacy.type = 'all' 
	OR (entry_privacy.type = 'some' 
		AND (entries.author_id = ?
			OR EXISTS(SELECT 1 from entries_privacy WHERE user_id = ? AND entry_id = entries.id))))
`

func AddEntryOpenQuery(q *sqlf.Stmt, userID int64) *sqlf.Stmt {
	return q.Where(isEntryOpenQueryWhere, userID, userID)
}

func addTlogQuery(q *sqlf.Stmt, userID int64, tlog, tag string) *sqlf.Stmt {
	q.Where("lower(users.name) = lower(?)", tlog)
	addTagQuery(q, tag)
	return AddEntryOpenQuery(q, userID)
}

func tlogQuery(userID, limit int64, tlog, tag string) *sqlf.Stmt {
	q := feedQuery(userID, limit)
	return addTlogQuery(q, userID, tlog, tag)
}

func addFriendsQuery(q *sqlf.Stmt, userID int64, tag string) *sqlf.Stmt {
	addRelationsFromMeQuery(q, userID)
	AddEntryOpenQuery(q, userID)
	addTagQuery(q, tag)
	return q.
		Where("(users.id = ? OR relations_from_me.type = 'followed')", userID).
		Where("(user_privacy.type != 'invited' OR (SELECT invited_by IS NOT NULL FROM users WHERE id = ?))", userID)
}

func friendsQuery(userID, limit int64, tag string) *sqlf.Stmt {
	q := feedQuery(userID, limit)
	return addFriendsQuery(q, userID, tag)
}

const canViewEntryQueryWhere = `
(users.id = $1
	OR (` + isEntryOpenQueryWhere + `
		AND (user_privacy.type = 'all' 
			OR (user_privacy.type = 'followers' AND relations_from_me.type = 'followed')
			OR (user_privacy.type = 'invited' AND (SELECT invited_by IS NOT NULL FROM users WHERE id = ?))
	)))`

func addCanViewEntryQuery(q *sqlf.Stmt, userID int64) *sqlf.Stmt {
	addRelationsFromMeQuery(q, userID)
	addRelationsToMeQuery(q, userID)
	return q.
		Where("(relations_to_me.type IS NULL OR relations_to_me.type <> 'ignored')").
		Where("(relations_from_me.type IS NULL OR relations_from_me.type <> 'ignored')").
		Where(canViewEntryQueryWhere, userID, userID, userID)
}

func watchingQuery(userID, limit int64) *sqlf.Stmt {
	q := feedQuery(userID, limit)
	addCanViewEntryQuery(q, userID)
	return q.Where("my_watching.entry_id IS NOT NULL").
		Where("users.invited_by IS NOT NULL").
		Where("entries.comments_count > 0").
		OrderBy("last_comment DESC")
}

func addFavoritesQuery(q *sqlf.Stmt, userID int64, tlog string) *sqlf.Stmt {
	addCanViewEntryQuery(q, userID)
	return q.Join("favorites", "entries.id = favorites.entry_id").
		Where("favorites.user_id = (SELECT id FROM users WHERE lower(name) = lower(?))", tlog)
}

func favoritesQuery(userID, limit int64, tlog string) *sqlf.Stmt {
	q := feedQuery(userID, limit)
	return addFavoritesQuery(q, userID, tlog)
}

func scrollQuery() *sqlf.Stmt {
	return sqlf.
		From("entries").
		Join("users", "entries.author_id = users.id").
		Join("entry_privacy", "entries.visible_for = entry_privacy.id").
		Join("user_privacy", "users.privacy = user_privacy.id").
		Limit(1)
}

func loadFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, reverse bool) *models.Feed {
	feed := models.Feed{}

	for {
		var entry models.Entry
		var author models.User
		var vote sql.NullFloat64
		var avatar string
		var rating models.Rating
		ok := tx.Scan(&entry.ID, &entry.CreatedAt,
			&rating.Rating, &rating.UpCount, &rating.DownCount,
			&entry.Title, &entry.EditContent,
			&entry.WordCount, &entry.Privacy,
			&rating.IsVotable, &entry.InLive, &entry.CommentCount,
			&author.ID, &author.Name, &author.ShowName,
			&author.IsOnline,
			&avatar,
			&vote, &entry.IsFavorited, &entry.IsWatching)
		if !ok {
			break
		}

		rating.Vote = entryVoteStatus(vote)
		entry.Rating = &rating
		rating.ID = entry.ID
		author.Avatar = srv.NewAvatar(avatar)
		entry.Author = &author

		feed.Entries = append(feed.Entries, &entry)
	}

	for _, entry := range feed.Entries {
		var images []int64
		var imageID int64
		tx.Query("SELECT image_id from entry_images WHERE entry_id = $1 ORDER BY image_id", entry.ID)
		for tx.Scan(&imageID) {
			images = append(images, imageID)
		}

		loadEntryImages(srv, tx, entry, images)
		loadEntryTags(tx, entry)
	}

	for _, entry := range feed.Entries {
		setEntryRights(entry, userID)
		setEntryTexts(entry, len(entry.Images) > 0)

		if entry.Author.ID != userID.ID {
			entry.EditContent = ""
		}
	}

	if reverse {
		list := feed.Entries
		for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
			list[i], list[j] = list[j], list[i]
		}
	}

	return &feed
}

func loadLiveFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, addQ addQuery,
	beforeS, afterS, search string, limit int64) *models.Feed {
	before := utils.ParseFloat(beforeS)
	after := utils.ParseFloat(afterS)

	query := feedQuery(userID.ID, limit)
	addQ(query)

	if search == "" {
		if after > 0 {
			query.Where("entries.created_at > to_timestamp(?)", after).
				OrderBy("entries.created_at ASC")
		} else if before > 0 {
			query.Where("entries.created_at < to_timestamp(?)", before).
				OrderBy("entries.created_at DESC")
		} else {
			query.OrderBy("entries.created_at DESC")
		}
	}

	query = addSearchQuery(query, search)
	tx.QueryStmt(query)
	feed := loadFeed(srv, tx, userID, after > 0)

	if search != "" || len(feed.Entries) == 0 {
		return feed
	}

	scrollQ := scrollQuery().Select("TRUE")
	addQ(scrollQ)
	defer scrollQ.Close()

	nextBefore := feed.Entries[len(feed.Entries)-1].CreatedAt
	feed.NextBefore = utils.FormatFloat(nextBefore)

	beforeQuery := scrollQ.Clone().Where("entries.created_at < to_timestamp(?)", nextBefore)

	tx.QueryStmt(beforeQuery)
	tx.Scan(&feed.HasBefore)

	nextAfter := feed.Entries[0].CreatedAt
	feed.NextAfter = utils.FormatFloat(nextAfter)

	afterQuery := scrollQ.Clone().Where("entries.created_at > to_timestamp(?)", nextAfter)

	tx.QueryStmt(afterQuery)
	tx.Scan(&feed.HasAfter)

	return feed
}

func loadLiveCommentsFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, tag string, limit int64) *models.Feed {
	query := liveCommentsQuery(userID.ID, limit, tag)
	tx.QueryStmt(query)
	return loadFeed(srv, tx, userID, false)
}

func newLiveLoader(srv *utils.MindwellServer) func(entries.GetEntriesLiveParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesLiveParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			var feed *models.Feed
			if *params.Section == "entries" {
				add := func(q *sqlf.Stmt) { AddLiveInvitedQuery(q, userID.ID, *params.Tag) }
				feed = loadLiveFeed(srv, tx, userID, add, *params.Before, *params.After, *params.Query, *params.Limit)
			} else if *params.Section == "comments" {
				feed = loadLiveCommentsFeed(srv, tx, userID, *params.Tag, *params.Limit)
			} else if *params.Section == "waiting" {
				add := func(q *sqlf.Stmt) { addLiveWaitingQuery(q, userID.ID, *params.Tag) }
				feed = loadLiveFeed(srv, tx, userID, add, *params.Before, *params.After, *params.Query, *params.Limit)
			}

			return entries.NewGetEntriesLiveOK().WithPayload(feed)
		})
	}
}

func newAnonymousLoader(srv *utils.MindwellServer) func(entries.GetEntriesAnonymousParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesAnonymousParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := &models.Feed{}
			return entries.NewGetEntriesAnonymousOK().WithPayload(feed)
		})
	}
}

func loadBestFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, category, tag string, limit int64) *models.Feed {
	var interval string
	if category == "month" {
		interval = "1 month"
	} else if category == "week" {
		interval = "7 days"
	} else {
		srv.LogApi().Sugar().Warn("Unknown best category:", category)
		interval = "1 month"
	}

	query := liveInvitedQuery(userID.ID, limit, "").
		Where("entries.created_at >= CURRENT_TIMESTAMP - interval '" + interval + "'").
		OrderBy("entries.rating DESC")

	query = sqlf.Select("*").From("").
		SubQuery("(", ") AS feed", query).
		OrderBy("feed.created_at DESC")

	addTagQuery(query, tag)
	tx.QueryStmt(query)

	feed := loadFeed(srv, tx, userID, false)

	return feed
}

func newBestLoader(srv *utils.MindwellServer) func(entries.GetEntriesBestParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesBestParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadBestFeed(srv, tx, userID, *params.Category, *params.Tag, *params.Limit)
			return entries.NewGetEntriesBestOK().WithPayload(feed)
		})
	}
}

func loadTlogFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, tlog, beforeS, afterS, tag, sort, search string, limit int64) *models.Feed {
	if userID.Name == tlog {
		return loadMyTlogFeed(srv, tx, userID, beforeS, afterS, tag, sort, search, limit)
	}

	before := utils.ParseFloat(beforeS)
	after := utils.ParseFloat(afterS)

	query := tlogQuery(userID.ID, limit, tlog, tag)
	reverse := false

	if search == "" {
		if sort == "new" || sort == "old" {
			if after > 0 {
				reverse = sort == "new"
				query.Where("entries.created_at > to_timestamp(?)", after).
					OrderBy("entries.created_at ASC")
			} else if before > 0 {
				reverse = sort == "old"
				query.Where("entries.created_at < to_timestamp(?)", before).
					OrderBy("entries.created_at DESC")
			} else {
				reverse = sort == "old"
				query.OrderBy("entries.created_at DESC")
			}
		} else {
			query.OrderBy("entries.rating DESC").
				OrderBy("entries.created_at DESC")
		}
	}

	query = addSearchQuery(query, search)
	tx.QueryStmt(query)
	feed := loadFeed(srv, tx, userID, reverse)

	if sort == "best" || search != "" {
		return feed
	}

	scrollQ := scrollQuery()
	addTlogQuery(scrollQ, userID.ID, tlog, tag)
	defer scrollQ.Close()

	if len(feed.Entries) == 0 {
		scrollQ.Select("extract(epoch from entries.created_at)").
			OrderBy("entries.created_at DESC")

		if before > 0 {
			query := scrollQ.Clone().
				Where("entries.created_at >= to_timestamp(?)", before)
			tx.QueryStmt(query)

			var nextAfter float64
			tx.Scan(&nextAfter)
			feed.NextAfter = utils.FormatFloat(nextAfter)
		}

		if after > 0 {
			query := scrollQ.Clone().
				Where("entries.created_at <= to_timestamp(?)", before)
			tx.QueryStmt(query)

			var nextBefore float64
			tx.Scan(&nextBefore)
			feed.NextBefore = utils.FormatFloat(nextBefore)
		}
	} else {
		scrollQ.Select("TRUE")

		oldest := feed.Entries[len(feed.Entries)-1].CreatedAt
		newest := feed.Entries[0].CreatedAt

		if sort == "old" {
			oldest, newest = newest, oldest
		}

		feed.NextBefore = utils.FormatFloat(oldest)
		beforeQuery := scrollQ.Clone().
			Where("entries.created_at < to_timestamp(?)", oldest)
		tx.QueryStmt(beforeQuery)
		tx.Scan(&feed.HasBefore)

		feed.NextAfter = utils.FormatFloat(newest)
		afterQuery := scrollQ.Clone().
			Where("entries.created_at > to_timestamp(?)", newest)
		tx.QueryStmt(afterQuery)
		tx.Scan(&feed.HasAfter)
	}

	return feed
}

func newTlogLoader(srv *utils.MindwellServer) func(users.GetUsersNameTlogParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersNameTlogParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			canView := utils.IsOpenForMe(tx, userID, params.Name)
			if !canView {
				err := srv.StandardError("no_tlog")
				return users.NewGetUsersNameTlogNotFound().WithPayload(err)
			}

			feed := loadTlogFeed(srv, tx, userID, params.Name, *params.Before, *params.After, *params.Tag, *params.Sort, *params.Query, *params.Limit)
			return users.NewGetUsersNameTlogOK().WithPayload(feed)
		})
	}
}

func loadMyTlogFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, beforeS, afterS, tag, sort, search string, limit int64) *models.Feed {
	before := utils.ParseFloat(beforeS)
	after := utils.ParseFloat(afterS)

	query := myTlogQuery(userID.ID, limit, tag)
	reverse := false

	if search == "" {
		if sort == "new" || sort == "old" {
			if after > 0 {
				reverse = sort == "new"
				query.Where("entries.created_at > to_timestamp(?)", after).
					OrderBy("entries.created_at ASC")
			} else if before > 0 {
				reverse = sort == "old"
				query.Where("entries.created_at < to_timestamp(?)", before).
					OrderBy("entries.created_at DESC")
			} else {
				reverse = sort == "old"
				query.OrderBy("entries.created_at DESC")
			}
		} else {
			query.OrderBy("entries.rating DESC").
				OrderBy("entries.created_at DESC")
		}
	}

	query = addSearchQuery(query, search)
	tx.QueryStmt(query)
	feed := loadFeed(srv, tx, userID, reverse)

	if sort == "best" || search != "" {
		return feed
	}

	scrollQ := sqlf.From("entries").
		Where("entries.author_id = ?", userID.ID).
		Limit(1)
	addTagQuery(scrollQ, tag)
	defer scrollQ.Close()

	if len(feed.Entries) == 0 {
		scrollQ.Select("extract(epoch from entries.created_at)").
			OrderBy("entries.created_at")

		if before > 0 {
			afterQuery := scrollQ.Clone().Where("created_at >= to_timestamp(?)", before)

			tx.QueryStmt(afterQuery)
			var nextAfter float64
			tx.Scan(&nextAfter)
			feed.NextAfter = utils.FormatFloat(nextAfter)
		}

		if after > 0 {
			beforeQuery := scrollQ.Clone().Where("created_at <= to_timestamp(?)", after)

			tx.QueryStmt(beforeQuery)
			var nextBefore float64
			tx.Scan(&nextBefore)
			feed.NextBefore = utils.FormatFloat(nextBefore)
		}
	} else {
		scrollQ.Select("TRUE")

		oldest := feed.Entries[len(feed.Entries)-1].CreatedAt
		newest := feed.Entries[0].CreatedAt

		if sort == "old" {
			oldest, newest = newest, oldest
		}

		feed.NextBefore = utils.FormatFloat(oldest)
		beforeQuery := scrollQ.Clone().Where("created_at < to_timestamp(?)", oldest)
		tx.QueryStmt(beforeQuery)
		tx.Scan(&feed.HasBefore)

		feed.NextAfter = utils.FormatFloat(newest)
		afterQuery := scrollQ.Clone().Where("created_at > to_timestamp(?)", newest)
		tx.QueryStmt(afterQuery)
		tx.Scan(&feed.HasAfter)
	}

	return feed
}

func newMyTlogLoader(srv *utils.MindwellServer) func(me.GetMeTlogParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeTlogParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadMyTlogFeed(srv, tx, userID, *params.Before, *params.After, *params.Tag, *params.Sort, *params.Query, *params.Limit)

			if tx.Error() != nil && tx.Error() != sql.ErrNoRows {
				err := srv.NewError(nil)
				return me.NewPutMeCoverBadRequest().WithPayload(err)
			}

			return me.NewGetMeTlogOK().WithPayload(feed)
		})
	}
}

func loadFriendsFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, beforeS, afterS, tag, search string, limit int64) *models.Feed {
	before := utils.ParseFloat(beforeS)
	after := utils.ParseFloat(afterS)

	query := friendsQuery(userID.ID, limit, tag)

	if search == "" {
		if after > 0 {
			query.Where("entries.created_at > to_timestamp(?)", after).
				OrderBy("entries.created_at ASC")
		} else if before > 0 {
			query.Where("entries.created_at < to_timestamp(?)", before).
				OrderBy("entries.created_at DESC")
		} else {
			query.OrderBy("entries.created_at DESC")
		}
	}

	query = addSearchQuery(query, search)
	tx.QueryStmt(query)
	feed := loadFeed(srv, tx, userID, after > 0)

	if search != "" {
		return feed
	}

	scrollQ := scrollQuery()
	addFriendsQuery(scrollQ, userID.ID, tag)
	defer scrollQ.Close()

	if len(feed.Entries) == 0 {
		scrollQ.Select("extract(epoch from entries.created_at)").
			OrderBy("entries.created_at DESC")

		if before > 0 {
			afterQuery := scrollQ.Clone().
				Where("entries.created_at >= to_timestamp(?)", before)
			tx.QueryStmt(afterQuery)

			var nextAfter float64
			tx.Scan(&nextAfter)
			feed.NextAfter = utils.FormatFloat(nextAfter)
		}

		if after > 0 {
			beforeQuery := scrollQ.Clone().
				Where("entries.created_at <= to_timestamp(?)", before)
			tx.QueryStmt(beforeQuery)

			var nextBefore float64
			tx.Scan(&nextBefore)
			feed.NextBefore = utils.FormatFloat(nextBefore)
		}
	} else {
		scrollQ.Select("TRUE")

		nextBefore := feed.Entries[len(feed.Entries)-1].CreatedAt
		feed.NextBefore = utils.FormatFloat(nextBefore)

		beforeQuery := scrollQ.Clone().Where("entries.created_at < to_timestamp(?)", nextBefore)

		tx.QueryStmt(beforeQuery)
		tx.Scan(&feed.HasBefore)

		nextAfter := feed.Entries[0].CreatedAt
		feed.NextAfter = utils.FormatFloat(nextAfter)

		afterQuery := scrollQ.Clone().Where("entries.created_at > to_timestamp(?)", nextAfter)

		tx.QueryStmt(afterQuery)
		tx.Scan(&feed.HasAfter)
	}

	return feed
}

func newFriendsFeedLoader(srv *utils.MindwellServer) func(entries.GetEntriesFriendsParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesFriendsParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadFriendsFeed(srv, tx, userID, *params.Before, *params.After, *params.Tag, *params.Query, *params.Limit)
			return entries.NewGetEntriesFriendsOK().WithPayload(feed)
		})
	}
}

func loadTlogFavorites(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, tlog, beforeS, afterS, search string, limit int64) *models.Feed {
	before := utils.ParseFloat(beforeS)
	after := utils.ParseFloat(afterS)

	query := favoritesQuery(userID.ID, limit, tlog)

	if search == "" {
		if after > 0 {
			query.Where("favorites.date > to_timestamp(?)", after).
				OrderBy("favorites.date ASC")
		} else if before > 0 {
			query.Where("favorites.date < to_timestamp(?)", before).
				OrderBy("favorites.date DESC")
		} else {
			query.OrderBy("favorites.date DESC")
		}
	}

	query = addSearchQuery(query, search)
	tx.QueryStmt(query)
	feed := loadFeed(srv, tx, userID, after > 0)

	if search != "" {
		return feed
	}

	scrollQ := scrollQuery()
	addFavoritesQuery(scrollQ, userID.ID, tlog)
	defer scrollQ.Close()

	if len(feed.Entries) == 0 {
		scrollQ.Select("extract(epoch from favorites.date)").
			OrderBy("favorites.date DESC")

		if before > 0 {
			afterQuery := scrollQ.Clone().
				Where("favorites.date >= to_timestamp(?)", before)
			tx.QueryStmt(afterQuery)

			var nextAfter float64
			tx.Scan(&nextAfter)
			feed.NextAfter = utils.FormatFloat(nextAfter)
		}

		if after > 0 {
			beforeQuery := scrollQ.Clone().
				Where("favorites.date <= to_timestamp(?)", after)
			tx.QueryStmt(beforeQuery)

			var nextBefore float64
			tx.Scan(&nextBefore)
			feed.NextBefore = utils.FormatFloat(nextBefore)
		}
	} else {
		scrollQ.Select("TRUE")

		dateQuery := sqlf.Select("extract(epoch from date)").
			From("favorites").
			Where("user_id = (SELECT id FROM users WHERE lower(name) = lower(?))", tlog)
		defer dateQuery.Close()

		lastID := feed.Entries[len(feed.Entries)-1].ID
		tx.QueryStmt(dateQuery.Clone().Where("entry_id = ?", lastID))
		var nextBefore float64
		tx.Scan(&nextBefore)
		feed.NextBefore = utils.FormatFloat(nextBefore)

		beforeQuery := scrollQ.Clone().Where("favorites.date < to_timestamp(?)", nextBefore)

		tx.QueryStmt(beforeQuery)
		tx.Scan(&feed.HasBefore)

		firstID := feed.Entries[0].ID
		tx.QueryStmt(dateQuery.Clone().Where("entry_id = ?", firstID))
		var nextAfter float64
		tx.Scan(&nextAfter)
		feed.NextAfter = utils.FormatFloat(nextAfter)

		afterQuery := scrollQ.Clone().Where("favorites.date > to_timestamp(?)", nextAfter)

		tx.QueryStmt(afterQuery)
		tx.Scan(&feed.HasAfter)
	}

	return feed
}

func newTlogFavoritesLoader(srv *utils.MindwellServer) func(users.GetUsersNameFavoritesParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersNameFavoritesParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			canView := utils.IsOpenForMe(tx, userID, params.Name)
			if !canView {
				err := srv.StandardError("no_tlog")
				return users.NewGetUsersNameFavoritesNotFound().WithPayload(err)
			}

			feed := loadTlogFavorites(srv, tx, userID, params.Name, *params.Before, *params.After, *params.Query, *params.Limit)
			return users.NewGetUsersNameFavoritesOK().WithPayload(feed)
		})
	}
}

func newMyFavoritesLoader(srv *utils.MindwellServer) func(me.GetMeFavoritesParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeFavoritesParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadTlogFavorites(srv, tx, userID, userID.Name, *params.Before, *params.After, *params.Query, *params.Limit)
			return me.NewGetMeFavoritesOK().WithPayload(feed)
		})
	}
}

func loadWatchingFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, limit int64) *models.Feed {
	query := watchingQuery(userID.ID, limit)
	tx.QueryStmt(query)
	return loadFeed(srv, tx, userID, false)
}

func newWatchingLoader(srv *utils.MindwellServer) func(entries.GetEntriesWatchingParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesWatchingParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadWatchingFeed(srv, tx, userID, *params.Limit)
			return entries.NewGetEntriesWatchingOK().WithPayload(feed)
		})
	}
}
