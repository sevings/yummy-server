package entries

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/internal/app/yummy-server/comments"
	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations/entries"
)

const feedQueryStart = `
SELECT id, created_at, rating, 
title, content, edit_content, word_count,
entry_privacy,
is_votable, comments_count,
author_id, author_name, author_show_name,
author_is_online,
author_avatar `

const tlogFeedQueryStart = feedQueryStart + `,
votes.positive AS vote,
EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND entry_id = feed.id) AS favorited,
EXISTS(SELECT 1 FROM watching WHERE user_id = $1 AND entry_id = feed.id) AS watching 
FROM feed
LEFT JOIN (SELECT entry_id, positive FROM entry_votes WHERE user_id = $1) AS votes ON feed.id = votes.entry_id
WHERE feed.entry_privacy = 'all' 
	AND feed.author_privacy = 'all' `

const feedQueryEnd = " ORDER BY created_at DESC LIMIT $2 OFFSET $3"

const liveFeedQuery = tlogFeedQueryStart + feedQueryEnd

const anonymousFeedQuery = feedQueryStart + `,
false,
EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND entry_id = feed.id) AS favorited,
EXISTS(SELECT 1 FROM watching WHERE user_id = $1 AND entry_id = feed.id) AS watching 
FROM feed
WHERE feed.entry_privacy = 'anonymous'` + feedQueryEnd

const bestFeedQuery = tlogFeedQueryStart + " AND feed.rating > 5 " + feedQueryEnd

const tlogFeedQuery = tlogFeedQueryStart + " AND feed.author_id = $4 " + feedQueryEnd

const myTlogFeedQuery = feedQueryStart + `,
NULL, 
EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND entry_id = feed.id) AS favorited,
true
FROM feed
WHERE feed.author_id = $1 ` + feedQueryEnd

func loadComments(tx *utils.AutoTx, userID int64, feed *models.Feed) {
	for _, entry := range feed.Entries {
		cmt := comments.LoadEntryComments(tx, userID, entry.ID, 5, 0)
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
		var vote sql.NullBool
		ok := tx.Scan(&entry.ID, &entry.CreatedAt, &entry.Rating,
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

		switch {
		case author.ID == userID:
			entry.Vote = models.EntryVoteBan
		case !vote.Valid:
			entry.Vote = models.EntryVoteNot
		case vote.Bool:
			entry.Vote = models.EntryVotePos
		default:
			entry.Vote = models.EntryVoteNeg
		}

		entry.Author = &author
		feed.Entries = append(feed.Entries, &entry)
	}

	loadComments(tx, userID, &feed)

	return &feed
}

func loadLiveFeed(tx *utils.AutoTx, userID *models.UserID, limit, offset int64) *models.Feed {
	return loadFeed(tx, liveFeedQuery, userID, limit, offset)
}

func newLiveLoader(db *sql.DB) func(entries.GetEntriesLiveParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesLiveParams, userID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			feed := loadLiveFeed(tx, userID, *params.Limit, *params.Skip)

			if tx.Error() != nil {
				return entries.NewGetEntriesLiveOK()
			}

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

			if tx.Error() != nil {
				return entries.NewGetEntriesAnonymousOK()
			}

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

			if tx.Error() != nil {
				return entries.NewGetEntriesBestOK()
			}

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

			if tx.Error() != nil {
				return entries.NewGetEntriesUsersIDNotFound()
			}

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

			if tx.Error() != nil {
				return entries.NewGetEntriesUsersMeForbidden()
			}

			return entries.NewGetEntriesUsersMeOK().WithPayload(feed)
		})
	}
}
