package votes

import (
	"database/sql"
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations/votes"
	"github.com/sevings/yummy-server/src/entries"
	"github.com/sevings/yummy-server/src/users"
)

func entryVoteStatus(tx *sql.Tx, userID, entryID int64) (*models.VoteStatus, error) {
	const q = `
		WITH votes AS (
			SELECT entry_id, positive
			FROM entry_votes
			WHERE user_id = $1
		)
		SELECT rating, positive
		FROM entries
		LEFT JOIN votes on votes.entry_id = entries.id
		WHERE entries.id = $2`

	var status = models.VoteStatus{ID: entryID}

	var positive sql.NullBool
	err := tx.QueryRow(q, userID, entryID).Scan(&status.Rating, &positive)
	if err != nil {
		return nil, err
	}

	switch {
	case !positive.Valid:
		status.Vote = models.VoteStatusVoteNot
	case positive.Bool:
		status.Vote = models.VoteStatusVotePos
	default:
		status.Vote = models.VoteStatusVoteNeg
	}

	return &status, nil
}

func newEntryVoteLoader(db *sql.DB) func(votes.GetEntriesIDVoteParams) middleware.Responder {
	return func(params votes.GetEntriesIDVoteParams) middleware.Responder {
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		defer tx.Commit()

		userID, found := users.FindAuthUser(tx, &params.XUserKey)
		if !found {
			return votes.NewGetEntriesIDVoteForbidden()
		}

		canView := entries.CanViewEntry(tx, userID, params.ID)
		if !canView {
			return votes.NewGetEntriesIDVoteNotFound()
		}

		status, err := entryVoteStatus(tx, userID, params.ID)
		if err != nil {
			log.Print(err)
			return votes.NewGetEntriesIDVoteNotFound()
		}

		return votes.NewGetEntriesIDVoteOK().WithPayload(status)
	}
}

func canVoteForEntry(tx *sql.Tx, userID, entryID int64) (bool, error) {
	const q = `
	WITH allowed AS (
		SELECT id, TRUE AS vote
		FROM feed
		WHERE id = $2 AND author_id <> $1
			AND ((entry_privacy = 'all' 
				AND (author_privacy = 'all'
					OR (author_privacy = 'registered' AND $1 > 0)
					OR EXISTS(SELECT 1 FROM relation, relations, entries
							  WHERE from_id = $1 AND to_id = entries.author_id
								  AND entries.id = $2
						 		  AND relation.type = 'followed'
						 		  AND relations.type = relation.id)))
			OR (entry_privacy = 'some' 
				AND EXISTS(SELECT 1 FROM entries_privacy
					WHERE user_id = $1 AND entry_id = $2)))
	)
	SELECT entries.id, allowed.vote
	FROM entries
	LEFT JOIN allowed ON entries.id = allowed.id
	WHERE entries.id = $2`

	var id int64
	var allowed sql.NullBool
	err := tx.QueryRow(q, userID, entryID).Scan(&id, &allowed)
	if err != nil {
		return false, err
	}

	if allowed.Valid {
		return true, nil
	}

	return false, nil
}

func loadEntryRating(tx *sql.Tx, entryID int64) (int64, error) {
	const q = `
		SELECT rating
		FROM entries
		WHERE id = $1`

	var rating int64
	err := tx.QueryRow(q, entryID).Scan(&rating)
	return rating, err
}

func voteForEntry(tx *sql.Tx, userID, entryID int64, positive bool) (*models.VoteStatus, error) {
	const voteQ = `
		INSERT INTO entry_votes (user_id, entry_id, positive)
		VALUES ($1, $2, $3)
		ON CONFLICT ON CONSTRAINT unique_entry_vote
		DO UPDATE SET positive = EXCLUDED.positive`

	_, err := tx.Exec(voteQ, userID, entryID, positive)
	if err != nil {
		return nil, err
	}

	rating, err := loadEntryRating(tx, entryID)
	if err != nil {
		return nil, err
	}

	var status = models.VoteStatus{
		ID:     entryID,
		Rating: rating,
	}

	switch {
	case positive:
		status.Vote = models.VoteStatusVotePos
	default:
		status.Vote = models.VoteStatusVoteNeg
	}

	return &status, nil
}

func newEntryVoter(db *sql.DB) func(votes.PutEntriesIDVoteParams) middleware.Responder {
	return func(params votes.PutEntriesIDVoteParams) middleware.Responder {
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}

		userID, found := users.FindAuthUser(tx, &params.XUserKey)
		if !found {
			tx.Rollback()
			return votes.NewPutEntriesIDVoteForbidden()
		}

		canVote, err := canVoteForEntry(tx, userID, params.ID)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Print(err)
			}

			tx.Rollback()
			return votes.NewPutEntriesIDVoteNotFound()
		}

		if !canVote {
			tx.Rollback()
			return votes.NewPutEntriesIDVoteForbidden()
		}

		status, err := voteForEntry(tx, userID, params.ID, *params.Positive)
		if err != nil {
			log.Print(err)
			tx.Rollback()
			return votes.NewPutEntriesIDVoteNotFound()
		}

		tx.Commit()

		return votes.NewPutEntriesIDVoteOK().WithPayload(status)
	}
}

func unvoteEntry(tx *sql.Tx, userID, entryID int64) (*models.VoteStatus, error) {
	const q = `
		DELETE FROM entry_votes
		WHERE user_id = $1 AND entry_id = $2
		RETURNING positive`

	var pos bool
	err := tx.QueryRow(q, userID, entryID).Scan(&pos)
	if err != nil {
		return nil, err
	}

	rating, err := loadEntryRating(tx, entryID)
	if err != nil {
		return nil, err
	}

	var status = models.VoteStatus{
		ID:     entryID,
		Rating: rating,
		Vote:   models.EntryVoteNot,
	}

	return &status, nil
}

func newEntryUnvoter(db *sql.DB) func(votes.DeleteEntriesIDVoteParams) middleware.Responder {
	return func(params votes.DeleteEntriesIDVoteParams) middleware.Responder {
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}

		userID, found := users.FindAuthUser(tx, &params.XUserKey)
		if !found {
			tx.Rollback()
			return votes.NewDeleteEntriesIDVoteForbidden()
		}

		canVote, err := canVoteForEntry(tx, userID, params.ID)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Print(err)
			}

			tx.Rollback()
			return votes.NewDeleteEntriesIDVoteNotFound()
		}

		if !canVote {
			tx.Rollback()
			return votes.NewDeleteEntriesIDVoteForbidden()
		}

		status, err := unvoteEntry(tx, userID, params.ID)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Print(err)
			}

			tx.Rollback()
			return votes.NewDeleteEntriesIDVoteNotFound()
		}

		tx.Commit()

		return votes.NewDeleteEntriesIDVoteOK().WithPayload(status)
	}
}
