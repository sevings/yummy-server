package entries

import (
	"log"
	"database/sql"
	"regexp"

	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations"
	"github.com/sevings/yummy-server/src/users"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {

}

wordRe := regexp.MustCompile("[a-zA-Zа-яА-ЯёЁ0-9]+")

const postEntryQuery = `
INSERT INTO entries (author_id, title, content, word_count, 
    (SELECT "type" from entry_privacy WHERE id = $5, is_votable)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at`

func createEntry(tx *sql.Tx, apiKey *string, title, content, privacy string, isVotable bool) (*models.Entry, bool) {
	author, found := users.LoadAuthUser(tx, apiKey)
	if !found {
		return nil, false
	}

	wordCount := 0
	contentWords := wordRe.findAllStringIndex(content, -1)
	if contentWords != nil {
		wordCount += len(contentWords) / 2
	}
	titleWords := wordsRe.findAllStringIndex(title, -1)
	if titleWords != nil {
		wordCount += len(titleWords) / 2
	}

	var entryID int64
	var createdAt string
	err := tx.QueryRow(postEntryQuery, author.ID, title, content, wordCount, 
		privacy, isVotable).Scan(&entryID, &createdAt)
	if err != nil {
		log.Print(err)
		return nil, false
	}

	var entry models.Entry {
		ID: entryID, 
		CreatedAt: createdAt, 
		Title: title, 
		Content: content, 
		WordCount: wordCount,
		VisibleFor: privacy,
		Author: &author
	}

	return &entry, true
}

const feedQueryStart = `
SELECT id, created_at, rating, 
title, content, word_count,
entry_privacy.type AS privacy,
is_votable, comments_count,
author_id, author_name, author_show_name,
author_is_online,
author_name_color, author_avatar_color, author_avatar `

const feedQueryWhere = `
feed.entry_privacy = 'all' AND feed.author_privacy = 'all' `

const liveFeedQuery = feedQueryStart + "WHERE" + liveFeedQueryWhere + "FROM feed LIMIT $1 OFFSET $2"

const authFeedQueryStart = feedQueryStart + `,
entry_votes.positive AS vote,
EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND entry_id = entries.id) AS favorited,
EXISTS(SELECT 1 FROM watching WHERE user_id = $1 AND entry_id = entries.id) AS watching 
FROM feed
LEFT JOIN entry_votes ON entries.id = entry_votes.entry_id
WHERE entry_votes.user_id = $1  `

const authLiveFeedQueryWhere = `
	AND feed.entry_privacy = 'all' 
	AND (feed.author_privacy = 'all' OR feed.author_privacy = 'registered') `

const authLiveFeedQuery = authFeedQueryStart + authLiveFeedQueryWhere + " LIMIT $2 OFFSET $3"

func loadNotAuthFeed(tx *sql.Tx, query string, limit, offset int64}) (*models.Feed, error) {
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

func loadAuthFeed()(tx *sql.Tx, query string, userID int64, limit, offset int64) (*models.Feed, error) {
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

func loadFeed(tx *sql.Tx, authQuery, notAuthQuery string, apiKey *string, limit, offset int64) (*models.Feed, error) {
	userID, found := usres.FindAuthUser(tx, apiKey)
	if !found {
		return loadNotAuthFeed(tx, authQuery, limit, offset)
	}

	return loadAuthFeed(tx, notAuthQuery, userID, limit, offset)
}

func loadLiveFeed(tx *sql.Tx, apiKey *string, limit, offset int64) (*models.Feed, error) {
	return loadFeed(tx, authLiveFeedQuery, liveFeedQuery, apiKey, limit, offset)
}
