package entries

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/internal/app/mindwell-server/comments"
	"github.com/sevings/mindwell-server/internal/app/mindwell-server/utils"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/entries"
)

const feedQueryStart = `
SELECT entries.id, extract(epoch from entries.created_at), rating, entries.votes,
entries.title, content, edit_content, word_count,
entry_privacy.type,
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

const tlogFeedQueryWhere = `
WHERE entry_privacy.type = 'all' 
	AND user_privacy.type = 'all' `

const feedQueryEnd = " ORDER BY entries.created_at DESC LIMIT $2 OFFSET $3"

const liveFeedQuery = tlogFeedQueryStart + tlogFeedQueryWhere + feedQueryEnd

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

const bestFeedQuery = tlogFeedQueryStart + tlogFeedQueryWhere + " AND entries.rating > 5 " + feedQueryEnd

const tlogFeedQuery = tlogFeedQueryStart + tlogFeedQueryWhere + " AND entries.author_id = $4 " + feedQueryEnd

const myTlogFeedQuery = feedQueryStart + `
NULL, 
EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND entry_id = entries.id),
true
FROM entries
INNER JOIN users ON entries.author_id = users.id
INNER JOIN entry_privacy ON entries.visible_for = entry_privacy.id
WHERE entries.author_id = $1 ` + feedQueryEnd

const friendsFeedQuery = tlogFeedQueryStart + `
WHERE (users.id = $1 
		OR EXISTS(SELECT 1 FROM relations WHERE from_id = $1 AND to_id = users.id 
			AND type = (SELECT id FROM relation WHERE type = 'followed')))
	AND (entry_privacy.type = 'all' 
		OR (entry_privacy.type = 'some' 
			AND (users.id = $1
				OR EXISTS(SELECT 1 from entries_privacy WHERE user_id = $1 AND entry_id = entries.id))))
` + feedQueryEnd

func loadComments(tx *utils.AutoTx, userID int64, feed *models.Feed) {
	for _, entry := range feed.Entries {
		cmt := comments.LoadEntryComments(tx, userID, entry.ID, 5, "", "")
		entry.Comments = cmt
	}
}

func loadFeed(tx *utils.AutoTx, query string, uID *models.UserID, args ...interface{}) *models.Feed {
	var feed models.Feed
	userID := int64(*uID)
	tx.Query(query, append([]interface{}{userID}, args...)...)

	for {
		var entry models.Entry
		var author models.User
		var vote sql.NullFloat64
		ok := tx.Scan(&entry.ID, &entry.CreatedAt, &entry.Rating, &entry.Votes,
			&entry.Title, &entry.Content, &entry.EditContent, &entry.WordCount,
			&entry.Privacy,
			&entry.IsVotable, &entry.CommentCount,
			&author.ID, &author.Name, &author.ShowName,
			&author.IsOnline,
			&author.Avatar,
			&vote, &entry.IsFavorited, &entry.IsWatching)
		if !ok {
			break
		}

		if author.ID != userID {
			entry.EditContent = ""
		}

		entry.Vote = entryVoteStatus(author.ID, userID, vote)

		entry.Author = &author
		feed.Entries = append(feed.Entries, &entry)
	}

	// loadComments(tx, userID, &feed)

	return &feed
}

func loadLiveFeed(tx *utils.AutoTx, userID *models.UserID, limit, offset int64) *models.Feed {
	return loadFeed(tx, liveFeedQuery, userID, limit, offset)
}

func newLiveLoader(db *sql.DB) func(entries.GetEntriesLiveParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesLiveParams, userID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			feed := loadLiveFeed(tx, userID, *params.Limit, *params.Skip)
			return entries.NewGetEntriesLiveOK().WithPayload(feed)
		})
	}
}

func loadAnonymousFeed(tx *utils.AutoTx, userID *models.UserID, limit, offset int64) *models.Feed {
	//! \todo do not load authors
	return loadFeed(tx, anonymousFeedQuery, userID, limit, offset)
}

func newAnonymousLoader(db *sql.DB) func(entries.GetEntriesAnonymousParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesAnonymousParams, userID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			feed := loadAnonymousFeed(tx, userID, *params.Limit, *params.Skip)
			return entries.NewGetEntriesAnonymousOK().WithPayload(feed)
		})
	}
}

func loadBestFeed(tx *utils.AutoTx, userID *models.UserID, limit, offset int64) *models.Feed {
	return loadFeed(tx, bestFeedQuery, userID, limit, offset)
}

func newBestLoader(db *sql.DB) func(entries.GetEntriesBestParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesBestParams, userID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			feed := loadBestFeed(tx, userID, *params.Limit, *params.Skip)
			return entries.NewGetEntriesBestOK().WithPayload(feed)
		})
	}
}

func loadTlogFeed(tx *utils.AutoTx, userID *models.UserID, limit, offset, tlog int64) *models.Feed {
	if int64(*userID) == tlog {
		return loadMyTlogFeed(tx, userID, limit, offset)
	}

	return loadFeed(tx, tlogFeedQuery, userID, limit, offset, tlog)
}

func newTlogLoader(db *sql.DB) func(entries.GetEntriesUsersIDParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesUsersIDParams, userID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			feed := loadTlogFeed(tx, userID, *params.Limit, *params.Skip, params.ID)
			return entries.NewGetEntriesUsersIDOK().WithPayload(feed)
		})
	}
}

func loadMyTlogFeed(tx *utils.AutoTx, userID *models.UserID, limit, offset int64) *models.Feed {
	return loadFeed(tx, myTlogFeedQuery, userID, limit, offset)
}

func newMyTlogLoader(db *sql.DB) func(entries.GetEntriesUsersMeParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesUsersMeParams, userID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			feed := loadMyTlogFeed(tx, userID, *params.Limit, *params.Skip)

			if tx.Error() != nil && tx.Error() != sql.ErrNoRows {
				return entries.NewGetEntriesUsersMeForbidden()
			}

			return entries.NewGetEntriesUsersMeOK().WithPayload(feed)
		})
	}
}

func loadFriendsFeed(tx *utils.AutoTx, userID *models.UserID, limit, offset int64) *models.Feed {
	return loadFeed(tx, friendsFeedQuery, userID, limit, offset)
}

func newFriendsFeedLoader(db *sql.DB) func(entries.GetEntriesFriendsParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesFriendsParams, userID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			feed := loadFriendsFeed(tx, userID, *params.Limit, *params.Skip)
			return entries.NewGetEntriesFriendsOK().WithPayload(feed)
		})
	}
}
