package entries

import (
	"database/sql"
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/entries"
	"github.com/sevings/mindwell-server/restapi/operations/me"
	"github.com/sevings/mindwell-server/restapi/operations/users"
	"github.com/sevings/mindwell-server/utils"
)

const feedQueryStart = `
SELECT entries.id, extract(epoch from entries.created_at) as created_at, 
rating, entries.up_votes, entries.down_votes,
entries.title, cut_title, content, cut_content, edit_content, 
has_cut, word_count, entry_privacy.type,
is_votable, in_live, entries.comments_count,
users.id, users.name, users.show_name,
is_online(users.last_seen_at),
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

const isInvitedQueryWhere = " (SELECT invited_by IS NOT NULL FROM users WHERE id = $1) "

const relationFromMeQuery = `
COALESCE((SELECT relation.type
	FROM relations
	INNER JOIN relation ON relations.type = relation.id
	WHERE from_id = $1 AND to_id = author_id), 'none')
`

const relationToMeQuery = `
COALESCE((SELECT relation.type
	FROM relations
	INNER JOIN relation ON relations.type = relation.id
	WHERE from_id = author_id AND to_id = $1), 'none')
`

const isNotIgnoredQueryWhere = `
(` + relationToMeQuery + ` <> 'ignored'
AND ` + relationFromMeQuery + ` <> 'ignored')
`

const liveFeedQueryWhere = `
WHERE entry_privacy.type = 'all' 
	AND in_live
	AND (user_privacy.type = 'all' 
		OR (user_privacy.type = 'invited' 
			AND ` + isInvitedQueryWhere + `))
	AND ` + relationToMeQuery + ` <> 'ignored'
	AND ` + relationFromMeQuery + ` NOT IN ('ignored', 'hidden')
`

const liveInvitedFeedQueryWhere = liveFeedQueryWhere + "AND users.invited_by IS NOT NULL "
const liveWaitingFeedQueryWhere = liveFeedQueryWhere + "AND users.invited_by IS NULL "

const liveInvitedFeedQuery = tlogFeedQueryStart + liveInvitedFeedQueryWhere
const liveWaitingFeedQuery = tlogFeedQueryStart + liveWaitingFeedQueryWhere

const commentsFeedQueryEnd = `
	AND users.invited_by IS NOT NULL
	AND entries.comments_count > 0
ORDER BY last_comment DESC
LIMIT $2
`

const liveCommentsFeedQuery = tlogFeedQueryStart + liveFeedQueryWhere + commentsFeedQueryEnd

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

const tlogFeedQueryWhere = `
	WHERE lower(users.name) = lower($2)
		AND ` + isEntryOpenQueryWhere

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

const isEntryOpenQueryWhere = `
(entry_privacy.type = 'all' 
	OR (entry_privacy.type = 'some' 
		AND (users.id = $1
			OR EXISTS(SELECT 1 from entries_privacy WHERE user_id = $1 AND entry_id = entries.id))))
`

const friendsFeedQueryWhere = `
WHERE (users.id = $1 
		OR EXISTS(SELECT 1 FROM relations WHERE from_id = $1 AND to_id = users.id 
			AND type = (SELECT id FROM relation WHERE type = 'followed')))
	AND ` + isEntryOpenQueryWhere + `
	AND (user_privacy.type != 'invited' OR` + isInvitedQueryWhere + ")"

const friendsFeedQuery = tlogFeedQueryStart + friendsFeedQueryWhere

const canViewEntryQueryWhere = `
(users.id = $1
	OR (` + isEntryOpenQueryWhere + `
	AND (user_privacy.type = 'all' 
		OR (user_privacy.type = 'followers'
			AND EXISTS(SELECT 1 FROM relations WHERE from_id = $1 AND to_id = users.id 
				AND type = (SELECT id FROM relation WHERE type = 'followed')))
		OR (user_privacy.type = 'invited'
			AND ` + isInvitedQueryWhere + `)
	)))
	AND ` + isNotIgnoredQueryWhere

const watchingFeedQuery = tlogFeedQueryStart + `
INNER JOIN watching ON watching.entry_id = entries.id 
	AND watching.user_id = $1
WHERE ` + canViewEntryQueryWhere + commentsFeedQueryEnd

const tlogFavoritesQueryStart = tlogFeedQueryStart + `
	INNER JOIN favorites ON entries.id = favorites.entry_id
`

const tlogFavoritesQueryWhere = `
WHERE favorites.user_id = (SELECT id FROM users WHERE lower(name) = lower($2))
AND ` + canViewEntryQueryWhere

const tlogFavoritesQuery = tlogFavoritesQueryStart + tlogFavoritesQueryWhere

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
			&entry.Title, &entry.CutTitle, &entry.Content, &entry.CutContent, &entry.EditContent,
			&entry.HasCut, &entry.WordCount, &entry.Privacy,
			&rating.IsVotable, &entry.InLive, &entry.CommentCount,
			&author.ID, &author.Name, &author.ShowName,
			&author.IsOnline,
			&avatar,
			&vote, &entry.IsFavorited, &entry.IsWatching)
		if !ok {
			break
		}

		if author.ID != userID.ID {
			entry.EditContent = ""
		}

		rating.Vote = entryVoteStatus(vote)
		entry.Rating = &rating
		rating.ID = entry.ID
		author.Avatar = srv.NewAvatar(avatar)
		entry.Author = &author
		setEntryRights(&entry, userID)
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
	}

	if reverse {
		list := feed.Entries
		for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
			list[i], list[j] = list[j], list[i]
		}
	}

	return &feed
}

func loadLiveFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, query, queryWhere, beforeS, afterS string, limit int64) *models.Feed {
	before := utils.ParseFloat(beforeS)
	after := utils.ParseFloat(afterS)

	if after > 0 {
		q := query + " AND entries.created_at > to_timestamp($2) ORDER BY entries.created_at ASC LIMIT $3"
		tx.Query(q, userID.ID, after, limit)
	} else if before > 0 {
		q := query + " AND entries.created_at < to_timestamp($2) ORDER BY entries.created_at DESC LIMIT $3"
		tx.Query(q, userID.ID, before, limit)
	} else {
		q := query + " ORDER BY entries.created_at DESC LIMIT $2"
		tx.Query(q, userID.ID, limit)
	}

	feed := loadFeed(srv, tx, userID, after > 0)

	if len(feed.Entries) == 0 {
		return feed
	}

	nextBefore := feed.Entries[len(feed.Entries)-1].CreatedAt
	feed.NextBefore = utils.FormatFloat(nextBefore)

	beforeQuery := `SELECT EXISTS(
		SELECT 1 
		FROM entries
		INNER JOIN users ON entries.author_id = users.id
		INNER JOIN entry_privacy ON entries.visible_for = entry_privacy.id
		INNER JOIN user_privacy ON users.privacy = user_privacy.id
	` + queryWhere + " AND entries.created_at < to_timestamp($2))"

	tx.Query(beforeQuery, userID.ID, nextBefore)
	tx.Scan(&feed.HasBefore)

	afterQuery := `SELECT EXISTS(
		SELECT 1 
		FROM entries
		INNER JOIN users ON entries.author_id = users.id
		INNER JOIN entry_privacy ON entries.visible_for = entry_privacy.id
		INNER JOIN user_privacy ON users.privacy = user_privacy.id
	` + queryWhere + " AND entries.created_at > to_timestamp($2))"

	nextAfter := feed.Entries[0].CreatedAt
	feed.NextAfter = utils.FormatFloat(nextAfter)
	tx.Query(afterQuery, userID.ID, nextAfter)
	tx.Scan(&feed.HasAfter)

	return feed
}

func loadLiveCommentsFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, limit int64) *models.Feed {
	tx.Query(liveCommentsFeedQuery, userID.ID, limit)
	return loadFeed(srv, tx, userID, false)
}

func newLiveLoader(srv *utils.MindwellServer) func(entries.GetEntriesLiveParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesLiveParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			var feed *models.Feed
			if *params.Section == "entries" {
				feed = loadLiveFeed(srv, tx, userID, liveInvitedFeedQuery, liveInvitedFeedQueryWhere, *params.Before, *params.After, *params.Limit)
			} else if *params.Section == "comments" {
				feed = loadLiveCommentsFeed(srv, tx, userID, *params.Limit)
			} else if *params.Section == "waiting" {
				feed = loadLiveFeed(srv, tx, userID, liveWaitingFeedQuery, liveWaitingFeedQueryWhere, *params.Before, *params.After, *params.Limit)
			}

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

func loadBestFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, category string, limit int64) *models.Feed {
	var interval string
	if category == "month" {
		interval = "1 month"
	} else if category == "week" {
		interval = "7 days"
	} else {
		log.Printf("Unknown best category: %s", category)
		interval = "1 month"
	}

	q := "SELECT * FROM (" + liveInvitedFeedQuery + " AND entries.created_at >= CURRENT_TIMESTAMP - interval '" +
		interval + "' ORDER BY entries.rating DESC LIMIT $2) AS feed ORDER BY feed.created_at DESC"
	tx.Query(q, userID.ID, limit)

	feed := loadFeed(srv, tx, userID, false)

	return feed
}

func newBestLoader(srv *utils.MindwellServer) func(entries.GetEntriesBestParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesBestParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadBestFeed(srv, tx, userID, *params.Category, *params.Limit)
			return entries.NewGetEntriesBestOK().WithPayload(feed)
		})
	}
}

func loadTlogFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, tlog, beforeS, afterS string, limit int64) *models.Feed {
	if userID.Name == tlog {
		return loadMyTlogFeed(srv, tx, userID, beforeS, afterS, limit)
	}

	before := utils.ParseFloat(beforeS)
	after := utils.ParseFloat(afterS)

	if after > 0 {
		q := tlogFeedQuery + " AND entries.created_at > to_timestamp($3) ORDER BY entries.created_at ASC LIMIT $4"
		tx.Query(q, userID.ID, tlog, after, limit)
	} else if before > 0 {
		q := tlogFeedQuery + " AND entries.created_at < to_timestamp($3) ORDER BY entries.created_at DESC LIMIT $4"
		tx.Query(q, userID.ID, tlog, before, limit)
	} else {
		q := tlogFeedQuery + " ORDER BY entries.created_at DESC LIMIT $3"
		tx.Query(q, userID.ID, tlog, limit)
	}

	feed := loadFeed(srv, tx, userID, after > 0)

	const scrollQ = `FROM entries
		INNER JOIN users ON entries.author_id = users.id
		INNER JOIN entry_privacy ON entries.visible_for = entry_privacy.id
		INNER JOIN user_privacy ON users.privacy = user_privacy.id
		` + tlogFeedQueryWhere + " AND entries.created_at "

	if len(feed.Entries) == 0 {
		if before > 0 {
			const afterQuery = `SELECT extract(epoch from entries.created_at) ` + scrollQ +
				` >= to_timestamp($3)
				ORDER BY entries.created_at DESC LIMIT 1`

			tx.Query(afterQuery, userID.ID, tlog, before)
			var nextAfter float64
			tx.Scan(&nextAfter)
			feed.NextAfter = utils.FormatFloat(nextAfter)
		}

		if after > 0 {
			const beforeQuery = `SELECT extract(epoch from entries.created_at) ` + scrollQ +
				` <= to_timestamp($3)
				ORDER BY entries.created_at DESC LIMIT 1`

			tx.Query(beforeQuery, userID.ID, tlog, after)
			var nextBefore float64
			tx.Scan(&nextBefore)
			feed.NextBefore = utils.FormatFloat(nextBefore)
		}
	} else {
		const beforeQuery = "SELECT EXISTS(SELECT 1 " + scrollQ + "< to_timestamp($3))"

		nextBefore := feed.Entries[len(feed.Entries)-1].CreatedAt
		feed.NextBefore = utils.FormatFloat(nextBefore)
		tx.Query(beforeQuery, userID.ID, tlog, nextBefore)
		tx.Scan(&feed.HasBefore)

		const afterQuery = "SELECT EXISTS(SELECT 1 " + scrollQ + "> to_timestamp($3))"

		nextAfter := feed.Entries[0].CreatedAt
		feed.NextAfter = utils.FormatFloat(nextAfter)
		tx.Query(afterQuery, userID.ID, tlog, nextAfter)
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

			feed := loadTlogFeed(srv, tx, userID, params.Name, *params.Before, *params.After, *params.Limit)
			return users.NewGetUsersNameTlogOK().WithPayload(feed)
		})
	}
}

func loadMyTlogFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, beforeS, afterS string, limit int64) *models.Feed {
	before := utils.ParseFloat(beforeS)
	after := utils.ParseFloat(afterS)

	if after > 0 {
		q := myTlogFeedQuery + " AND entries.created_at > to_timestamp($2) ORDER BY entries.created_at ASC LIMIT $3"
		tx.Query(q, userID.ID, after, limit)
	} else if before > 0 {
		q := myTlogFeedQuery + " AND entries.created_at < to_timestamp($2) ORDER BY entries.created_at DESC LIMIT $3"
		tx.Query(q, userID.ID, before, limit)
	} else {
		q := myTlogFeedQuery + " ORDER BY entries.created_at DESC LIMIT $2"
		tx.Query(q, userID.ID, limit)
	}

	feed := loadFeed(srv, tx, userID, after > 0)

	const scrollQ = "FROM entries " + myTlogFeedQueryWhere + " AND created_at "

	if len(feed.Entries) == 0 {
		if before > 0 {
			const afterQuery = `SELECT extract(epoch from entries.created_at) ` + scrollQ +
				` >= to_timestamp($1)
				ORDER BY entries.created_at DESC LIMIT 1`

			tx.Query(afterQuery, userID.ID, before)
			var nextAfter float64
			tx.Scan(&nextAfter)
			feed.NextAfter = utils.FormatFloat(nextAfter)
		}

		if after > 0 {
			const beforeQuery = `SELECT extract(epoch from entries.created_at) ` + scrollQ +
				` <= to_timestamp($1)
				ORDER BY entries.created_at DESC LIMIT 1`

			tx.Query(beforeQuery, userID.ID, after)
			var nextBefore float64
			tx.Scan(&nextBefore)
			feed.NextBefore = utils.FormatFloat(nextBefore)
		}
	} else {
		const beforeQuery = "SELECT EXISTS(SELECT 1 " + scrollQ + "	< to_timestamp($2))"

		nextBefore := feed.Entries[len(feed.Entries)-1].CreatedAt
		feed.NextBefore = utils.FormatFloat(nextBefore)
		tx.Query(beforeQuery, userID.ID, nextBefore)
		tx.Scan(&feed.HasBefore)

		const afterQuery = "SELECT EXISTS(SELECT 1 " + scrollQ + " > to_timestamp($2))"

		nextAfter := feed.Entries[0].CreatedAt
		feed.NextAfter = utils.FormatFloat(nextAfter)
		tx.Query(afterQuery, userID.ID, nextAfter)
		tx.Scan(&feed.HasAfter)
	}

	return feed
}

func newMyTlogLoader(srv *utils.MindwellServer) func(me.GetMeTlogParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeTlogParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadMyTlogFeed(srv, tx, userID, *params.Before, *params.After, *params.Limit)

			if tx.Error() != nil && tx.Error() != sql.ErrNoRows {
				err := srv.NewError(nil)
				return me.NewPutMeCoverBadRequest().WithPayload(err)
			}

			return me.NewGetMeTlogOK().WithPayload(feed)
		})
	}
}

func loadFriendsFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, beforeS, afterS string, limit int64) *models.Feed {
	before := utils.ParseFloat(beforeS)
	after := utils.ParseFloat(afterS)

	if after > 0 {
		q := friendsFeedQuery + " AND entries.created_at > to_timestamp($2) ORDER BY entries.created_at ASC LIMIT $3"
		tx.Query(q, userID.ID, after, limit)
	} else if before > 0 {
		q := friendsFeedQuery + " AND entries.created_at < to_timestamp($2) ORDER BY entries.created_at DESC LIMIT $3"
		tx.Query(q, userID.ID, before, limit)
	} else {
		q := friendsFeedQuery + " ORDER BY entries.created_at DESC LIMIT $2"
		tx.Query(q, userID.ID, limit)
	}

	feed := loadFeed(srv, tx, userID, after > 0)

	const scrollQ = `FROM entries
		INNER JOIN users ON entries.author_id = users.id
		INNER JOIN entry_privacy ON entries.visible_for = entry_privacy.id
		INNER JOIN user_privacy ON users.privacy = user_privacy.id
		` + friendsFeedQueryWhere + " AND entries.created_at"

	if len(feed.Entries) == 0 {
		if before > 0 {
			const afterQuery = `SELECT extract(epoch from entries.created_at) ` + scrollQ +
				` >= to_timestamp($2)
				ORDER BY entries.created_at DESC LIMIT 1`

			tx.Query(afterQuery, userID.ID, before)
			var nextAfter float64
			tx.Scan(&nextAfter)
			feed.NextAfter = utils.FormatFloat(nextAfter)
		}

		if after > 0 {
			const beforeQuery = `SELECT extract(epoch from entries.created_at) ` + scrollQ +
				` <= to_timestamp($2)
				ORDER BY entries.created_at DESC LIMIT 1`

			tx.Query(beforeQuery, userID.ID, after)
			var nextBefore float64
			tx.Scan(&nextBefore)
			feed.NextBefore = utils.FormatFloat(nextBefore)
		}
	} else {
		const beforeQuery = "SELECT EXISTS(SELECT 1 " + scrollQ + " < to_timestamp($2))"

		nextBefore := feed.Entries[len(feed.Entries)-1].CreatedAt
		feed.NextBefore = utils.FormatFloat(nextBefore)
		tx.Query(beforeQuery, userID.ID, nextBefore)
		tx.Scan(&feed.HasBefore)

		const afterQuery = "SELECT EXISTS(SELECT 1 " + scrollQ + " > to_timestamp($2))"

		nextAfter := feed.Entries[0].CreatedAt
		feed.NextAfter = utils.FormatFloat(nextAfter)
		tx.Query(afterQuery, userID.ID, nextAfter)
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

func loadTlogFavorites(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, tlog, beforeS, afterS string, limit int64) *models.Feed {
	before := utils.ParseFloat(beforeS)
	after := utils.ParseFloat(afterS)

	if after > 0 {
		q := tlogFavoritesQuery + " AND favorites.date > to_timestamp($3) ORDER BY favorites.date ASC LIMIT $4"
		tx.Query(q, userID.ID, tlog, after, limit)
	} else if before > 0 {
		q := tlogFavoritesQuery + " AND favorites.date < to_timestamp($3) ORDER BY favorites.date DESC LIMIT $4"
		tx.Query(q, userID.ID, tlog, before, limit)
	} else {
		q := tlogFavoritesQuery + " ORDER BY favorites.date DESC LIMIT $3"
		tx.Query(q, userID.ID, tlog, limit)
	}

	feed := loadFeed(srv, tx, userID, after > 0)

	const scrollQ = `FROM entries
		INNER JOIN users ON entries.author_id = users.id
		INNER JOIN entry_privacy ON entries.visible_for = entry_privacy.id
		INNER JOIN user_privacy ON users.privacy = user_privacy.id
		INNER JOIN favorites ON entries.id = favorites.entry_id
		` + tlogFavoritesQueryWhere + " AND favorites.date "

	if len(feed.Entries) == 0 {
		if before > 0 {
			const afterQuery = `SELECT extract(epoch from favorites.date) ` + scrollQ +
				` >= to_timestamp($3)
				ORDER BY favorites.date DESC LIMIT 1`

			tx.Query(afterQuery, userID.ID, tlog, before)
			var nextAfter float64
			tx.Scan(&nextAfter)
			feed.NextAfter = utils.FormatFloat(nextAfter)
		}

		if after > 0 {
			const beforeQuery = `SELECT extract(epoch from favorites.date) ` + scrollQ +
				` <= to_timestamp($3)
				ORDER BY favorites.date DESC LIMIT 1`

			tx.Query(beforeQuery, userID.ID, tlog, after)
			var nextBefore float64
			tx.Scan(&nextBefore)
			feed.NextBefore = utils.FormatFloat(nextBefore)
		}
	} else {
		const dateQuery = `
			SELECT extract(epoch from date) 
			FROM favorites 
			WHERE user_id = (SELECT id FROM users WHERE lower(name) = lower($1)) 
				AND entry_id = $2`

		const queryEnd = ` (
			SELECT date 
			FROM favorites 
			WHERE user_id = (SELECT id FROM users WHERE lower(name) = lower($2)) 
				AND entry_id = $3))`

		const beforeQuery = "SELECT EXISTS(SELECT 1 " + scrollQ + " < " + queryEnd

		lastID := feed.Entries[len(feed.Entries)-1].ID
		tx.Query(beforeQuery, userID.ID, tlog, lastID).Scan(&feed.HasBefore)
		if feed.HasBefore {
			tx.Query(dateQuery, tlog, lastID)
			var nextBefore float64
			tx.Scan(&nextBefore)
			feed.NextBefore = utils.FormatFloat(nextBefore)
		}

		const afterQuery = "SELECT EXISTS(SELECT 1 " + scrollQ + " > " + queryEnd

		firstID := feed.Entries[0].ID
		tx.Query(afterQuery, userID.ID, tlog, firstID).Scan(&feed.HasAfter)
		if feed.HasAfter {
			tx.Query(dateQuery, tlog, firstID)
			var nextAfter float64
			tx.Scan(&nextAfter)
			feed.NextAfter = utils.FormatFloat(nextAfter)
		}
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

			feed := loadTlogFavorites(srv, tx, userID, params.Name, *params.Before, *params.After, *params.Limit)
			return users.NewGetUsersNameFavoritesOK().WithPayload(feed)
		})
	}
}

func newMyFavoritesLoader(srv *utils.MindwellServer) func(me.GetMeFavoritesParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeFavoritesParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadTlogFavorites(srv, tx, userID, userID.Name, *params.Before, *params.After, *params.Limit)
			return me.NewGetMeFavoritesOK().WithPayload(feed)
		})
	}
}

func loadWatchingFeed(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, limit int64) *models.Feed {
	tx.Query(watchingFeedQuery, userID.ID, limit)
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
