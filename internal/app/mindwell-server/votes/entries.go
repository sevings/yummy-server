package votes

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/votes"
	"github.com/sevings/mindwell-server/utils"
)

func entryVoteStatus(tx *utils.AutoTx, userID, entryID int64) *models.VoteStatus {
	const q = `
		WITH votes AS (
			SELECT entry_id, vote
			FROM entry_votes
			WHERE user_id = $1
		)
		SELECT entries.author_id, entry_privacy.type, is_votable, 
			rating, votes, vote
		FROM entries
		LEFT JOIN votes on votes.entry_id = entries.id
		JOIN entry_privacy on entry_privacy.id = entries.visible_for
		WHERE entries.id = $2`

	var status = models.VoteStatus{ID: entryID}

	var authorID int64
	var privacy string
	var votable bool
	var vote sql.NullFloat64
	tx.Query(q, userID, entryID).Scan(&authorID, &privacy, &votable,
		&status.Rating, &status.Votes, &vote)

	switch {
	case authorID == userID || !votable || privacy == models.EntryPrivacyAnonymous:
		status.Vote = models.VoteStatusVoteBan
	case !vote.Valid:
		status.Vote = models.VoteStatusVoteNot
	case vote.Float64 > 0:
		status.Vote = models.VoteStatusVotePos
	default:
		status.Vote = models.VoteStatusVoteNeg
	}

	return &status
}

func newEntryVoteLoader(srv *utils.MindwellServer) func(votes.GetEntriesIDVoteParams, *models.UserID) middleware.Responder {
	return func(params votes.GetEntriesIDVoteParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
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

func loadEntryRating(tx *utils.AutoTx, entryID int64) (int64, float32) {
	const q = `
		SELECT votes, rating
		FROM entries
		WHERE id = $1`

	var votes int64
	var rating float32
	tx.Query(q, entryID).Scan(&votes, &rating)
	return votes, rating
}

func voteForEntry(tx *utils.AutoTx, userID, entryID int64, positive bool) *models.VoteStatus {
	const q = `
		INSERT INTO entry_votes (user_id, entry_id, vote)
		VALUES ($1, $2, (
			SELECT GREATEST(0.001, weight) * $3
			FROM entries, vote_weights
			WHERE vote_weights.user_id = $1 AND entries.id = $2
				AND entries.category = vote_weights.category
		))
		ON CONFLICT ON CONSTRAINT unique_entry_vote
		DO UPDATE SET vote = EXCLUDED.vote`

	var vote int64
	if positive {
		vote = 1
	} else {
		vote = -1
	}
	tx.Exec(q, userID, entryID, vote)

	votes, rating := loadEntryRating(tx, entryID)

	var status = models.VoteStatus{
		ID:     entryID,
		Votes:  votes,
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

func newEntryVoter(srv *utils.MindwellServer) func(votes.PutEntriesIDVoteParams, *models.UserID) middleware.Responder {
	return func(params votes.PutEntriesIDVoteParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
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

func unvoteEntry(tx *utils.AutoTx, userID, entryID int64) (*models.VoteStatus, bool) {
	const q = `
		DELETE FROM entry_votes
		WHERE user_id = $1 AND entry_id = $2`

	tx.Exec(q, userID, entryID)
	if tx.RowsAffected() != 1 {
		return nil, false
	}

	votes, rating := loadEntryRating(tx, entryID)

	var status = models.VoteStatus{
		ID:     entryID,
		Rating: rating,
		Votes:  votes,
		Vote:   models.EntryVoteNot,
	}

	return &status, true
}

func newEntryUnvoter(srv *utils.MindwellServer) func(votes.DeleteEntriesIDVoteParams, *models.UserID) middleware.Responder {
	return func(params votes.DeleteEntriesIDVoteParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			canVote := canVoteForEntry(tx, userID, params.ID)
			if tx.Error() != nil {
				return votes.NewDeleteEntriesIDVoteNotFound()
			}

			if !canVote {
				return votes.NewDeleteEntriesIDVoteForbidden()
			}

			status, ok := unvoteEntry(tx, userID, params.ID)
			if !ok || tx.Error() != nil {
				return votes.NewDeleteEntriesIDVoteNotFound()
			}

			return votes.NewDeleteEntriesIDVoteOK().WithPayload(status)
		})
	}
}
