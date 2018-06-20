package entries

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/entries"
	"github.com/sevings/mindwell-server/utils"
)

const feedQueryStart = `
SELECT entries.id, extract(epoch from entries.created_at), rating, entries.votes,
entries.title, cut_title, content, cut_content, edit_content, 
has_cut, word_count, entry_privacy.type,
is_votable, entries.comments_count,
users.id, users.name, users.show_name,
now() - users.last_seen_at < interval '15 minutes',
users.avatar, `

const tlogFeedQueryStart = feedQueryStart + `
votes.vote,
EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND entry_id = entries.id),
EXISTS(SELECT 1 FROM watching WHERE user_id = $1 AND entry_id = entries.id) 
FROM entries
INNER JOIN users ON entries.author_id = users.id
INNER JOIN entry_privacy ON entries.visible_for = entry_privacy.id
INNER JOIN user_privacy ON users.privacy = user_privacy.id
LEFT JOIN (SELECT entry_id, vote FROM entry_votes WHERE user_id = $1) AS votes ON entries.id = votes.entry_id`

const liveFeedQueryWhere = `
WHERE entry_privacy.type = 'all' 
	AND user_privacy.type = 'all' `

const liveFeedQuery = tlogFeedQueryStart + liveFeedQueryWhere

const anonymousFeedQuery = `
SELECT entries.id, extract(epoch from entries.created_at), 0, 
entries.title, content, edit_content, word_count,
entry_privacy.type,
false, entries.comments_count,
0, 'anonymous', 'Аноним',
true,
'', NULL,
EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND entry_id = entries.id),
EXISTS(SELECT 1 FROM watching WHERE user_id = $1 AND entry_id = entries.id) 
FROM entries
INNER JOIN entry_privacy ON entries.visible_for = entry_privacy.id
WHERE entry_privacy.type = 'anonymous' 
ORDER BY entries.created_at DESC LIMIT $2 OFFSET $3`

const bestFeedQueryWhere = liveFeedQueryWhere + " AND entries.rating > 5 "

const bestFeedQuery = tlogFeedQueryStart + bestFeedQueryWhere

const tlogFeedQueryWhere = liveFeedQueryWhere + " AND entries.author_id = $2 "

const tlogFeedQuery = tlogFeedQueryStart + tlogFeedQueryWhere

const myTlogFeedQueryWhere = " WHERE entries.author_id = $1 "

const myTlogFeedQuery = feedQueryStart + `
NULL, 
EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND entry_id = entries.id),
true
FROM entries
INNER JOIN users ON entries.author_id = users.id
INNER JOIN entry_privacy ON entries.visible_for = entry_privacy.id
` + myTlogFeedQueryWhere

const friendsFeedQueryWhere = `
WHERE (users.id = $1 
		OR EXISTS(SELECT 1 FROM relations WHERE from_id = $1 AND to_id = users.id 
			AND type = (SELECT id FROM relation WHERE type = 'followed')))
	AND (entry_privacy.type = 'all' 
		OR (entry_privacy.type = 'some' 
			AND (users.id = $1
				OR EXISTS(SELECT 1 from entries_privacy WHERE user_id = $1 AND entry_id = entries.id))))
`

const friendsFeedQuery = tlogFeedQueryStart + friendsFeedQueryWhere

func parseFloat(val string) float64 {
	res, err := strconv.ParseFloat(val, 64)
	if len(val) > 0 && err != nil {
		log.Printf("error parse float: '%s'", val)
	}

	return res
}

func formatFloat(val float64) string {
	return strconv.FormatFloat(val, 'f', 6, 64)
}

func loadFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID int64) *models.Feed {
	feed := models.Feed{}

	for {
		var entry models.Entry
		var author models.User
		var vote sql.NullFloat64
		var avatar string
		ok := tx.Scan(&entry.ID, &entry.CreatedAt, &entry.Rating, &entry.Votes,
			&entry.Title, &entry.CutTitle, &entry.Content, &entry.CutContent, &entry.EditContent,
			&entry.HasCut, &entry.WordCount, &entry.Privacy,
			&entry.IsVotable, &entry.CommentCount,
			&author.ID, &author.Name, &author.ShowName,
			&author.IsOnline,
			&avatar,
			&vote, &entry.IsFavorited, &entry.IsWatching)
		if !ok {
			break
		}

		if author.ID != userID {
			entry.EditContent = ""
		}

		entry.Vote = entryVoteStatus(author.ID, userID, vote)
		author.Avatar = srv.NewAvatar(avatar)
		entry.Author = &author
		feed.Entries = append(feed.Entries, &entry)
	}

	return &feed
}

func loadLiveFeed(srv *utils.MindwellServer, tx *utils.AutoTx, uID *models.UserID, beforeS, afterS string, limit int64) *models.Feed {
	before := parseFloat(beforeS)
	after := parseFloat(afterS)

	var q string
	var arg float64
	if before > 0 {
		q = liveFeedQuery + " AND entries.created_at < to_timestamp($2) ORDER BY entries.created_at DESC LIMIT $3"
		arg = before
	} else {
		q = liveFeedQuery + " AND entries.created_at > to_timestamp($2) ORDER BY entries.created_at DESC LIMIT $3"
		arg = after
	}

	userID := int64(*uID)
	tx.Query(q, userID, arg, limit)

	feed := loadFeed(srv, tx, userID)

	if len(feed.Entries) == 0 {
		return feed
	}

	nextBefore := feed.Entries[len(feed.Entries)-1].CreatedAt
	feed.NextBefore = formatFloat(nextBefore)

	const beforeQuery = `SELECT EXISTS(
		SELECT 1 
		FROM entries
		INNER JOIN users ON entries.author_id = users.id
		INNER JOIN entry_privacy ON entries.visible_for = entry_privacy.id
		INNER JOIN user_privacy ON users.privacy = user_privacy.id
	` + liveFeedQueryWhere + " AND entries.created_at < to_timestamp($1))"

	tx.Query(beforeQuery, nextBefore)
	tx.Scan(&feed.HasBefore)

	const afterQuery = `SELECT EXISTS(
		SELECT 1 
		FROM entries
		INNER JOIN users ON entries.author_id = users.id
		INNER JOIN entry_privacy ON entries.visible_for = entry_privacy.id
		INNER JOIN user_privacy ON users.privacy = user_privacy.id
	` + liveFeedQueryWhere + " AND entries.created_at > to_timestamp($1))"

	nextAfter := feed.Entries[0].CreatedAt
	feed.NextAfter = formatFloat(nextAfter)
	tx.Query(afterQuery, nextAfter)
	tx.Scan(&feed.HasAfter)

	return feed
}

func newLiveLoader(srv *utils.MindwellServer) func(entries.GetEntriesLiveParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesLiveParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadLiveFeed(srv, tx, userID, *params.Before, *params.After, *params.Limit)
			return entries.NewGetEntriesLiveOK().WithPayload(feed)
		})
	}
}

func loadAnonymousFeed(tx *utils.AutoTx, userID *models.UserID, beforeS, afterS string, limit int64) *models.Feed {
	//! \todo do not load authors
	return &models.Feed{}
}

func newAnonymousLoader(srv *utils.MindwellServer) func(entries.GetEntriesAnonymousParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesAnonymousParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadAnonymousFeed(tx, userID, *params.Before, *params.After, *params.Limit)
			return entries.NewGetEntriesAnonymousOK().WithPayload(feed)
		})
	}
}

func loadBestFeed(tx *utils.AutoTx, userID *models.UserID, beforeS, afterS string, limit int64) *models.Feed {
	return &models.Feed{}
}

func newBestLoader(srv *utils.MindwellServer) func(entries.GetEntriesBestParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesBestParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadBestFeed(tx, userID, *params.Before, *params.After, *params.Limit)
			return entries.NewGetEntriesBestOK().WithPayload(feed)
		})
	}
}

func loadTlogFeed(srv *utils.MindwellServer, tx *utils.AutoTx, uID *models.UserID, beforeS, afterS string, limit, tlog int64) *models.Feed {
	userID := int64(*uID)
	if userID == tlog {
		return loadMyTlogFeed(srv, tx, uID, beforeS, afterS, limit)
	}

	before := parseFloat(beforeS)
	after := parseFloat(afterS)

	var q string
	var arg float64
	if before > 0 {
		q = tlogFeedQuery + " AND entries.created_at < to_timestamp($3) ORDER BY entries.created_at DESC LIMIT $4"
		arg = before
	} else {
		q = tlogFeedQuery + " AND entries.created_at > to_timestamp($3) ORDER BY entries.created_at DESC LIMIT $4"
		arg = after
	}

	tx.Query(q, userID, tlog, arg, limit)

	feed := loadFeed(srv, tx, userID)

	const scrollQ = `FROM entries
		INNER JOIN users ON entries.author_id = users.id
		INNER JOIN entry_privacy ON entries.visible_for = entry_privacy.id
		INNER JOIN user_privacy ON users.privacy = user_privacy.id
		` + tlogFeedQueryWhere + " AND entries.created_at "

	if len(feed.Entries) == 0 {
		if before > 0 {
			const afterQuery = `SELECT extract(epoch from entries.created_at) ` + scrollQ +
				` >= to_timestamp($1)
				ORDER BY entries.created_at DESC LIMIT 1`

			tx.Query(afterQuery, before, tlog)
			var nextAfter float64
			tx.Scan(&nextAfter)
			feed.NextAfter = formatFloat(nextAfter)
		}

		if after > 0 {
			const beforeQuery = `SELECT extract(epoch from entries.created_at) ` + scrollQ +
				` <= to_timestamp($1)
				ORDER BY entries.created_at DESC LIMIT 1`

			tx.Query(beforeQuery, after, tlog)
			var nextBefore float64
			tx.Scan(&nextBefore)
			feed.NextBefore = formatFloat(nextBefore)
		}
	} else {
		const beforeQuery = "SELECT EXISTS(SELECT 1 " + scrollQ + "< to_timestamp($1))"

		nextBefore := feed.Entries[len(feed.Entries)-1].CreatedAt
		feed.NextBefore = formatFloat(nextBefore)
		tx.Query(beforeQuery, nextBefore, tlog)
		tx.Scan(&feed.HasBefore)

		const afterQuery = "SELECT EXISTS(SELECT 1 " + scrollQ + "> to_timestamp($1))"

		nextAfter := feed.Entries[0].CreatedAt
		feed.NextAfter = formatFloat(nextAfter)
		tx.Query(afterQuery, nextAfter, tlog)
		tx.Scan(&feed.HasAfter)
	}

	return feed
}

func newTlogLoader(srv *utils.MindwellServer) func(entries.GetEntriesUsersIDParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesUsersIDParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadTlogFeed(srv, tx, userID, *params.Before, *params.After, *params.Limit, params.ID)
			return entries.NewGetEntriesUsersIDOK().WithPayload(feed)
		})
	}
}

func loadMyTlogFeed(srv *utils.MindwellServer, tx *utils.AutoTx, uID *models.UserID, beforeS, afterS string, limit int64) *models.Feed {
	before := parseFloat(beforeS)
	after := parseFloat(afterS)

	var q string
	var arg float64
	if before > 0 {
		q = myTlogFeedQuery + " AND entries.created_at < to_timestamp($2) ORDER BY entries.created_at DESC LIMIT $3"
		arg = before
	} else {
		q = myTlogFeedQuery + " AND entries.created_at > to_timestamp($2) ORDER BY entries.created_at DESC LIMIT $3"
		arg = after
	}

	userID := int64(*uID)
	tx.Query(q, userID, arg, limit)

	feed := loadFeed(srv, tx, userID)

	const scrollQ = "FROM entries " + myTlogFeedQueryWhere + " AND created_at "

	if len(feed.Entries) == 0 {
		if before > 0 {
			const afterQuery = `SELECT extract(epoch from entries.created_at) ` + scrollQ +
				` >= to_timestamp($1)
				ORDER BY entries.created_at DESC LIMIT 1`

			tx.Query(afterQuery, userID, before)
			var nextAfter float64
			tx.Scan(&nextAfter)
			feed.NextAfter = formatFloat(nextAfter)
		}

		if after > 0 {
			const beforeQuery = `SELECT extract(epoch from entries.created_at) ` + scrollQ +
				` <= to_timestamp($1)
				ORDER BY entries.created_at DESC LIMIT 1`

			tx.Query(beforeQuery, userID, after)
			var nextBefore float64
			tx.Scan(&nextBefore)
			feed.NextBefore = formatFloat(nextBefore)
		}
	} else {
		const beforeQuery = "SELECT EXISTS(SELECT 1 " + scrollQ + "	< to_timestamp($2))"

		nextBefore := feed.Entries[len(feed.Entries)-1].CreatedAt
		feed.NextBefore = formatFloat(nextBefore)
		tx.Query(beforeQuery, userID, nextBefore)
		tx.Scan(&feed.HasBefore)

		const afterQuery = "SELECT EXISTS(SELECT 1 " + scrollQ + " > to_timestamp($2))"

		nextAfter := feed.Entries[0].CreatedAt
		feed.NextAfter = formatFloat(nextAfter)
		tx.Query(afterQuery, userID, nextAfter)
		tx.Scan(&feed.HasAfter)
	}

	return feed
}

func newMyTlogLoader(srv *utils.MindwellServer) func(entries.GetEntriesUsersMeParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesUsersMeParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadMyTlogFeed(srv, tx, userID, *params.Before, *params.After, *params.Limit)

			if tx.Error() != nil && tx.Error() != sql.ErrNoRows {
				return entries.NewGetEntriesUsersMeForbidden()
			}

			return entries.NewGetEntriesUsersMeOK().WithPayload(feed)
		})
	}
}

func loadFriendsFeed(srv *utils.MindwellServer, tx *utils.AutoTx, uID *models.UserID, beforeS, afterS string, limit int64) *models.Feed {
	before := parseFloat(beforeS)
	after := parseFloat(afterS)

	var q string
	var arg float64
	if before > 0 {
		q = friendsFeedQuery + " AND entries.created_at < to_timestamp($2) ORDER BY entries.created_at DESC LIMIT $3"
		arg = before
	} else {
		q = friendsFeedQuery + " AND entries.created_at > to_timestamp($2) ORDER BY entries.created_at DESC LIMIT $3"
		arg = after
	}

	userID := int64(*uID)
	tx.Query(q, userID, arg, limit)

	feed := loadFeed(srv, tx, userID)

	const scrollQ = `FROM entries
		INNER JOIN users ON entries.author_id = users.id
		INNER JOIN entry_privacy ON entries.visible_for = entry_privacy.id
		` + friendsFeedQueryWhere + " AND entries.created_at"

	if len(feed.Entries) == 0 {
		if before > 0 {
			const afterQuery = `SELECT extract(epoch from entries.created_at) ` + scrollQ +
				` >= to_timestamp($1)
				ORDER BY entries.created_at DESC LIMIT 1`

			tx.Query(afterQuery, userID, before)
			var nextAfter float64
			tx.Scan(&nextAfter)
			feed.NextAfter = formatFloat(nextAfter)
		}

		if after > 0 {
			const beforeQuery = `SELECT extract(epoch from entries.created_at) ` + scrollQ +
				` <= to_timestamp($1)
				ORDER BY entries.created_at DESC LIMIT 1`

			tx.Query(beforeQuery, userID, after)
			var nextBefore float64
			tx.Scan(&nextBefore)
			feed.NextBefore = formatFloat(nextBefore)
		}
	} else {
		const beforeQuery = "SELECT EXISTS(SELECT 1 " + scrollQ + " < to_timestamp($2))"

		nextBefore := feed.Entries[len(feed.Entries)-1].CreatedAt
		feed.NextBefore = formatFloat(nextBefore)
		tx.Query(beforeQuery, userID, nextBefore)
		tx.Scan(&feed.HasBefore)

		const afterQuery = "SELECT EXISTS(SELECT 1 " + scrollQ + " > to_timestamp($2))"

		nextAfter := feed.Entries[0].CreatedAt
		feed.NextAfter = formatFloat(nextAfter)
		tx.Query(afterQuery, userID, nextAfter)
		tx.Scan(&feed.HasAfter)
	}

	return feed
}

func newFriendsFeedLoader(srv *utils.MindwellServer) func(entries.GetEntriesFriendsParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesFriendsParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadFriendsFeed(srv, tx, userID, *params.Before, *params.After, *params.Limit)
			return entries.NewGetEntriesFriendsOK().WithPayload(feed)
		})
	}
}
