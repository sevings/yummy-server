package entries

import (
	"database/sql"
	"log"

	"github.com/sevings/yummy-server/src"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations/entries"
	"github.com/sevings/yummy-server/src/comments"
)

const feedQueryStart = `
SELECT id, created_at, rating, 
title, content, word_count,
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
	AND (feed.author_privacy = 'all' OR feed.author_privacy = 'registered') `

const feedQueryEnd = " ORDER BY created_at DESC LIMIT $2 OFFSET $3"

const liveFeedQuery = tlogFeedQueryStart + feedQueryEnd

const anonymousFeedQuery = feedQueryStart + `,
false,
EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND entry_id = feed.id) AS favorited,
EXISTS(SELECT 1 FROM watching WHERE user_id = $1 AND entry_id = feed.id) AS watching 
FROM feed
WHERE feed.entry_privacy = 'anonymous'` + feedQueryEnd

const bestFeedQuery = tlogFeedQueryStart + " AND feed.rating > 5 " + feedQueryEnd

func reverse(feed *models.Feed) {
	for i, j := 0, len(feed.Entries)-1; i < j; i, j = i+1, j-1 {
		feed.Entries[i], feed.Entries[j] = feed.Entries[j], feed.Entries[i]
	}
}

func loadComments(tx yummy.AutoTx, userID int64, feed *models.Feed) {
	for _, entry := range feed.Entries {
		cmt, err := comments.LoadEntryComments(tx, userID, entry.ID, 5, 0)
		if err != nil {
			log.Print(err)
		}

		entry.Comments = cmt
	}
}

func loadFeed(tx yummy.AutoTx, query string, uID *models.UserID, limit, offset int64) (*models.Feed, error) {
	var feed models.Feed
	userID := int64(*uID)
	rows, err := tx.Query(query, userID, limit, offset)
	if err != nil {
		return &feed, err
	}

	for rows.Next() {
		var entry models.Entry
		var author models.User
		var vote sql.NullBool
		rows.Scan(&entry.ID, &entry.CreatedAt, &entry.Rating,
			&entry.Title, &entry.Content, &entry.WordCount,
			&entry.Privacy,
			&entry.IsVotable, &entry.CommentCount,
			&author.ID, &author.Name, &author.ShowName,
			&author.IsOnline,
			&author.Avatar,
			&vote, &entry.IsFavorited, &entry.IsWatching)

		switch {
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

	reverse(&feed)
	loadComments(tx, userID, &feed)

	return &feed, rows.Err()
}

func loadLiveFeed(tx yummy.AutoTx, userID *models.UserID, limit, offset int64) (*models.Feed, error) {
	return loadFeed(tx, liveFeedQuery, userID, limit, offset)
}

func newLiveLoader(db *sql.DB) func(entries.GetEntriesLiveParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesLiveParams, userID *models.UserID) middleware.Responder {
		return yummy.Transact(db, func(tx yummy.AutoTx) (middleware.Responder, bool) {
			feed, err := loadLiveFeed(tx, userID, *params.Limit, *params.Skip)
			if err != nil {
				log.Print(err)
				return entries.NewGetEntriesLiveOK(), false
			}

			return entries.NewGetEntriesLiveOK().WithPayload(feed), true
		})
	}
}

func loadAnonymousFeed(tx yummy.AutoTx, userID *models.UserID, limit, offset int64) (*models.Feed, error) {
	//! \todo do not load authors
	return loadFeed(tx, anonymousFeedQuery, userID, limit, offset)
}

func newAnonymousLoader(db *sql.DB) func(entries.GetEntriesAnonymousParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesAnonymousParams, userID *models.UserID) middleware.Responder {
		return yummy.Transact(db, func(tx yummy.AutoTx) (middleware.Responder, bool) {
			feed, err := loadAnonymousFeed(tx, userID, *params.Limit, *params.Skip)
			if err != nil {
				log.Print(err)
				return entries.NewGetEntriesAnonymousOK(), false
			}

			return entries.NewGetEntriesAnonymousOK().WithPayload(feed), true
		})
	}
}

func loadBestFeed(tx yummy.AutoTx, userID *models.UserID, limit, offset int64) (*models.Feed, error) {
	return loadFeed(tx, bestFeedQuery, userID, limit, offset)
}

func newBestLoader(db *sql.DB) func(entries.GetEntriesBestParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesBestParams, userID *models.UserID) middleware.Responder {
		return yummy.Transact(db, func(tx yummy.AutoTx) (middleware.Responder, bool) {
			feed, err := loadBestFeed(tx, userID, *params.Limit, *params.Skip)
			if err != nil {
				log.Print(err)
				return entries.NewGetEntriesBestOK(), false
			}

			return entries.NewGetEntriesBestOK().WithPayload(feed), true
		})
	}
}
