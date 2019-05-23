package votes

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/votes"
	"github.com/sevings/mindwell-server/utils"
)

func entryRating(tx *utils.AutoTx, userID, entryID int64) *models.Rating {
	const q = `
		WITH votes AS (
			SELECT entry_id, vote
			FROM entry_votes
			WHERE user_id = $1
		)
		SELECT entries.author_id, entry_privacy.type, is_votable, 
			rating, up_votes, down_votes, vote
		FROM entries
		LEFT JOIN votes on votes.entry_id = entries.id
		JOIN entry_privacy on entry_privacy.id = entries.visible_for
		WHERE entries.id = $2`

	var status = models.Rating{ID: entryID}

	var authorID int64
	var privacy string
	var votable bool
	var vote sql.NullFloat64
	tx.Query(q, userID, entryID).Scan(&authorID, &privacy, &votable,
		&status.Rating, &status.UpCount, &status.DownCount, &vote)

	switch {
	case !vote.Valid:
		status.Vote = 0
	case vote.Float64 > 0:
		status.Vote = 1
	default:
		status.Vote = -1
	}

	return &status
}

func newEntryVoteLoader(srv *utils.MindwellServer) func(votes.GetEntriesIDVoteParams, *models.UserID) middleware.Responder {
	return func(params votes.GetEntriesIDVoteParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := uID.ID
			canView := utils.CanViewEntry(tx, userID, params.ID)
			if !canView {
				err := srv.StandardError("no_entry")
				return votes.NewGetEntriesIDVoteNotFound().WithPayload(err)
			}

			status := entryRating(tx, userID, params.ID)
			if tx.Error() != nil {
				err := srv.StandardError("no_entry")
				return votes.NewGetEntriesIDVoteNotFound().WithPayload(err)
			}

			return votes.NewGetEntriesIDVoteOK().WithPayload(status)
		})
	}
}

func canVoteForEntry(tx *utils.AutoTx, userID *models.UserID, entryID int64) bool {
	if userID.Ban.Vote {
		return false
	}

	var authorID int64
	var votable bool
	var privacy string

	const q = `
		SELECT author_id, is_votable, entry_privacy.type
		FROM entries
		INNER JOIN entry_privacy ON entries.visible_for = entry_privacy.id
		WHERE entries.id = $1
	`

	tx.Query(q, entryID).Scan(&authorID, &votable, &privacy)
	if authorID == userID.ID || !votable || privacy == models.EntryPrivacyAnonymous {
		return false
	}

	return utils.CanViewEntry(tx, userID.ID, entryID)
}

func loadEntryRating(tx *utils.AutoTx, entryID int64) *models.Rating {
	const q = `
		SELECT up_votes, down_votes, rating
		FROM entries
		WHERE id = $1`

	rating := &models.Rating{
		ID: entryID,
	}
	tx.Query(q, entryID).Scan(&rating.UpCount, &rating.DownCount, &rating.Rating)
	return rating
}

func voteForEntry(tx *utils.AutoTx, userID, entryID int64, positive bool) *models.Rating {
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

	rating := loadEntryRating(tx, entryID)

	switch {
	case positive:
		rating.Vote = 1
	default:
		rating.Vote = -1
	}

	return rating
}

func newEntryVoter(srv *utils.MindwellServer) func(votes.PutEntriesIDVoteParams, *models.UserID) middleware.Responder {
	return func(params votes.PutEntriesIDVoteParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			canVote := canVoteForEntry(tx, userID, params.ID)
			if tx.Error() != nil {
				err := srv.StandardError("no_entry")
				return votes.NewPutEntriesIDVoteNotFound().WithPayload(err)
			}

			if !canVote {
				err := srv.NewError(&i18n.Message{ID: "cant_vote", Other: "You are not allowed to vote for this entry."})
				return votes.NewPutEntriesIDVoteForbidden().WithPayload(err)
			}

			status := voteForEntry(tx, userID.ID, params.ID, *params.Positive)
			if tx.Error() != nil {
				err := srv.NewError(nil)
				return votes.NewPutEntriesIDVoteNotFound().WithPayload(err)
			}

			return votes.NewPutEntriesIDVoteOK().WithPayload(status)
		})
	}
}

func unvoteEntry(tx *utils.AutoTx, userID, entryID int64) (*models.Rating, bool) {
	const q = `
		DELETE FROM entry_votes
		WHERE user_id = $1 AND entry_id = $2`

	tx.Exec(q, userID, entryID)
	if tx.RowsAffected() != 1 {
		return nil, false
	}

	rating := loadEntryRating(tx, entryID)
	rating.Vote = 0

	return rating, true
}

func newEntryUnvoter(srv *utils.MindwellServer) func(votes.DeleteEntriesIDVoteParams, *models.UserID) middleware.Responder {
	return func(params votes.DeleteEntriesIDVoteParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			canVote := canVoteForEntry(tx, userID, params.ID)
			if tx.Error() != nil {
				err := srv.StandardError("no_entry")
				return votes.NewDeleteEntriesIDVoteNotFound().WithPayload(err)
			}

			if !canVote {
				err := srv.NewError(&i18n.Message{ID: "cant_vote", Other: "You are not allowed to vote for this entry."})
				return votes.NewDeleteEntriesIDVoteForbidden().WithPayload(err)
			}

			status, ok := unvoteEntry(tx, userID.ID, params.ID)
			if !ok || tx.Error() != nil {
				err := srv.NewError(nil)
				return votes.NewDeleteEntriesIDVoteNotFound().WithPayload(err)
			}

			return votes.NewDeleteEntriesIDVoteOK().WithPayload(status)
		})
	}
}
