package entries

import (
	"database/sql"
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations/entries"
	"github.com/sevings/yummy-server/src/users"
)

const feedQueryStart = `
SELECT id, created_at, rating, 
title, content, word_count,
entry_privacy,
is_votable, comments_count,
author_id, author_name, author_show_name,
author_is_online,
author_name_color, author_avatar_color, author_avatar `

const feedQueryWhere = `
feed.entry_privacy = 'all' AND feed.author_privacy = 'all' `

const feedQueryEnd = "LIMIT $1 OFFSET $2"

const liveFeedQuery = feedQueryStart + " FROM feed WHERE" + feedQueryWhere + feedQueryEnd

const authFeedQueryStart = feedQueryStart + `,
votes.positive AS vote,
EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND entry_id = feed.id) AS favorited,
EXISTS(SELECT 1 FROM watching WHERE user_id = $1 AND entry_id = feed.id) AS watching 
FROM feed
LEFT JOIN (SELECT entry_id, positive FROM entry_votes WHERE user_id = $1) AS votes ON feed.id = votes.entry_id
WHERE feed.entry_privacy = 'all' 
	AND (feed.author_privacy = 'all' OR feed.author_privacy = 'registered') `

const authFeedQueryEnd = " LIMIT $2 OFFSET $3"

const authLiveFeedQuery = authFeedQueryStart + authFeedQueryEnd

const anonymousFeedQueryWhere = " feed.entry_privacy = 'anonymous' "

const anonymousFeedQuery = feedQueryStart + " FROM feed WHERE" + anonymousFeedQueryWhere + feedQueryEnd

const anonymousAuthFeedQueryStart = feedQueryStart + `,
false,
EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND entry_id = feed.id) AS favorited,
EXISTS(SELECT 1 FROM watching WHERE user_id = $1 AND entry_id = feed.id) AS watching 
FROM feed`

const anonymousAuthFeedQuery = anonymousAuthFeedQueryStart + " WHERE " + anonymousFeedQueryWhere + authFeedQueryEnd

const bestfeedQueryWhere = " AND feed.rating > 5 "

const bestFeedQuery = feedQueryStart + " FROM feed WHERE " + feedQueryWhere + bestfeedQueryWhere + feedQueryEnd

const authBestFeedQuery = authFeedQueryStart + bestfeedQueryWhere + authFeedQueryEnd

func loadNotAuthFeed(tx *sql.Tx, query string, limit, offset int64) (*models.Feed, error) {
	var feed models.Feed
	rows, err := tx.Query(query, limit, offset)
	if err != nil {
		return &feed, err
	}

	for rows.Next() {
		var entry models.Entry
		var author models.User
		rows.Scan(&entry.ID, &entry.CreatedAt, &entry.Rating,
			&entry.Title, &entry.Content, &entry.WordCount,
			&entry.Privacy,
			&entry.IsVotable, &entry.CommentCount,
			&author.ID, &author.Name, &author.ShowName,
			&author.IsOnline,
			&author.NameColor, &author.AvatarColor, &author.Avatar)
		entry.Author = &author
		feed.Entries = append(feed.Entries, &entry)
	}

	return &feed, rows.Err()
}

func loadAuthFeed(tx *sql.Tx, query string, userID int64, limit, offset int64) (*models.Feed, error) {
	var feed models.Feed
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
			&author.NameColor, &author.AvatarColor, &author.Avatar,
			&vote, &entry.IsFavorited, &entry.IsWatching)

		switch {
		case !vote.Valid:
			entry.Vote = "not"
		case vote.Bool:
			entry.Vote = "pos"
		default:
			entry.Vote = "neg"
		}

		entry.Author = &author
		feed.Entries = append(feed.Entries, &entry)

		//! \todo load last comments
	}

	return &feed, rows.Err()
}

func loadFeed(db *sql.DB, authQuery, notAuthQuery string, apiKey *string, limit, offset int64) (*models.Feed, error) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Commit()

	userID, found := users.FindAuthUser(tx, apiKey)
	if !found {
		return loadNotAuthFeed(tx, notAuthQuery, limit, offset)
	}

	return loadAuthFeed(tx, authQuery, userID, limit, offset)
}

func loadLiveFeed(db *sql.DB, apiKey *string, limit, offset int64) (*models.Feed, error) {
	return loadFeed(db, authLiveFeedQuery, liveFeedQuery, apiKey, limit, offset)
}

func newLiveLoader(db *sql.DB) func(entries.GetEntriesLiveParams) middleware.Responder {
	return func(params entries.GetEntriesLiveParams) middleware.Responder {
		feed, err := loadLiveFeed(db, params.XUserKey, *params.Limit, *params.Skip)
		if err != nil {
			log.Print(err)
			return entries.NewGetEntriesLiveOK()
		}

		return entries.NewGetEntriesLiveOK().WithPayload(feed)
	}
}

func loadAnonymousFeed(db *sql.DB, apiKey *string, limit, offset int64) (*models.Feed, error) {
	//! \todo do not load authors
	return loadFeed(db, anonymousAuthFeedQuery, anonymousFeedQuery, apiKey, limit, offset)
}

func newAnonymousLoader(db *sql.DB) func(entries.GetEntriesAnonymousParams) middleware.Responder {
	return func(params entries.GetEntriesAnonymousParams) middleware.Responder {
		feed, err := loadAnonymousFeed(db, params.XUserKey, *params.Limit, *params.Skip)
		if err != nil {
			log.Print(err)
			return entries.NewGetEntriesAnonymousOK()
		}

		return entries.NewGetEntriesAnonymousOK().WithPayload(feed)
	}
}

func loadBestFeed(db *sql.DB, apiKey *string, limit, offset int64) (*models.Feed, error) {
	return loadFeed(db, authBestFeedQuery, bestFeedQuery, apiKey, limit, offset)
}

func newBestLoader(db *sql.DB) func(entries.GetEntriesBestParams) middleware.Responder {
	return func(params entries.GetEntriesBestParams) middleware.Responder {
		feed, err := loadBestFeed(db, params.XUserKey, *params.Limit, *params.Skip)
		if err != nil {
			log.Print(err)
			return entries.NewGetEntriesBestOK()
		}

		return entries.NewGetEntriesBestOK().WithPayload(feed)
	}
}
