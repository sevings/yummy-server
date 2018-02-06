package votes

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations/votes"
)

func entryVoteStatus(tx *utils.AutoTx, userID, entryID int64) *models.VoteStatus {
	const q = `
		WITH votes AS (
			SELECT entry_id, positive
			FROM entry_votes
			WHERE user_id = $1
		)
		SELECT entries.author_id, entry_privacy.type, is_votable, rating, positive
		FROM entries
		LEFT JOIN votes on votes.entry_id = entries.id
		JOIN entry_privacy on entry_privacy.id = entries.visible_for
		WHERE entries.id = $2`

	var status = models.VoteStatus{ID: entryID}

	var authorID int64
	var privacy string
	var votable bool
	var positive sql.NullBool
	tx.Query(q, userID, entryID).Scan(&authorID, &privacy, &votable, &status.Rating, &positive)

	switch {
	case authorID == userID || !votable || privacy == models.EntryPrivacyAnonymous:
		status.Vote = models.VoteStatusVoteBan
	case !positive.Valid:
		status.Vote = models.VoteStatusVoteNot
	case positive.Bool:
		status.Vote = models.VoteStatusVotePos
	default:
		status.Vote = models.VoteStatusVoteNeg
	}

	return &status
}

func newEntryVoteLoader(db *sql.DB) func(votes.GetEntriesIDVoteParams, *models.UserID) middleware.Responder {
	return func(params votes.GetEntriesIDVoteParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			canView := utils.CanViewEntry(tx, userID, params.ID)
			if !canView {
				return votes.NewGetEntriesIDVoteNotFound()
			}

			status := entryVoteStatus(tx, userID, params.ID)
			if tx.Error() != nil {
				return votes.NewGetEntriesIDVoteNotFound()
			}

			return votes.NewGetEntriesIDVoteOK().WithPayload(status)
		})
	}
}

func canVoteForEntry(tx *utils.AutoTx, userID, entryID int64) bool {
	const q = `
	WITH allowed AS (
		SELECT id, TRUE AS vote
		FROM feed
		WHERE id = $2 AND author_id <> $1 AND is_votable
			AND ((entry_privacy = 'all' 
				AND (author_privacy = 'all'
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
	tx.Query(q, userID, entryID).Scan(&id, &allowed)

	return allowed.Valid
}

func loadEntryRating(tx *utils.AutoTx, entryID int64) int64 {
	const q = `
		SELECT rating
		FROM entries
		WHERE id = $1`

	var rating int64
	tx.Query(q, entryID).Scan(&rating)
	return rating
}

func voteForEntry(tx *utils.AutoTx, userID, entryID int64, positive bool) *models.VoteStatus {
	const voteQ = `
		INSERT INTO entry_votes (user_id, entry_id, positive)
		VALUES ($1, $2, $3)
		ON CONFLICT ON CONSTRAINT unique_entry_vote
		DO UPDATE SET positive = EXCLUDED.positive`

	tx.Exec(voteQ, userID, entryID, positive)

	rating := loadEntryRating(tx, entryID)

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

	return &status
}

func newEntryVoter(db *sql.DB) func(votes.PutEntriesIDVoteParams, *models.UserID) middleware.Responder {
	return func(params votes.PutEntriesIDVoteParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			canVote := canVoteForEntry(tx, userID, params.ID)
			if tx.Error() != nil {
				return votes.NewPutEntriesIDVoteNotFound()
			}

			if !canVote {
				return votes.NewPutEntriesIDVoteForbidden()
			}

			status := voteForEntry(tx, userID, params.ID, *params.Positive)
			if tx.Error() != nil {
				return votes.NewPutEntriesIDVoteNotFound()
			}

			return votes.NewPutEntriesIDVoteOK().WithPayload(status)
		})
	}
}

func unvoteEntry(tx *utils.AutoTx, userID, entryID int64) *models.VoteStatus {
	const q = `
		DELETE FROM entry_votes
		WHERE user_id = $1 AND entry_id = $2
		RETURNING positive`

	var pos bool
	tx.Query(q, userID, entryID).Scan(&pos)

	rating := loadEntryRating(tx, entryID)

	var status = models.VoteStatus{
		ID:     entryID,
		Rating: rating,
		Vote:   models.EntryVoteNot,
	}

	return &status
}

func newEntryUnvoter(db *sql.DB) func(votes.DeleteEntriesIDVoteParams, *models.UserID) middleware.Responder {
	return func(params votes.DeleteEntriesIDVoteParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			canVote := canVoteForEntry(tx, userID, params.ID)
			if tx.Error() != nil {
				return votes.NewDeleteEntriesIDVoteNotFound()
			}

			if !canVote {
				return votes.NewDeleteEntriesIDVoteForbidden()
			}

			status := unvoteEntry(tx, userID, params.ID)
			if tx.Error() != nil {
				return votes.NewDeleteEntriesIDVoteNotFound()
			}

			return votes.NewDeleteEntriesIDVoteOK().WithPayload(status)
		})
	}
}
