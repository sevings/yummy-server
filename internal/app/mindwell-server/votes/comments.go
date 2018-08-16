package votes

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/votes"
	"github.com/sevings/mindwell-server/utils"
)

func commentRating(tx *utils.AutoTx, userID, commentID int64) *models.Rating {
	const q = `
		WITH votes AS (
			SELECT comment_id, vote
			FROM comment_votes
			WHERE user_id = $1
		)
		SELECT comments.author_id,
			rating, up_votes, down_votes, vote
		FROM comments
		LEFT JOIN votes on votes.comment_id = comments.id
		WHERE comments.id = $2`

	var status = models.Rating{
		ID:        commentID,
		IsVotable: true,
	}

	var authorID int64
	var vote sql.NullFloat64
	tx.Query(q, userID, commentID).Scan(&authorID,
		&status.Rating, &status.UpCount, &status.DownCount, &vote)

	switch {
	case authorID == userID:
		status.Vote = models.RatingVoteBan
	case !vote.Valid:
		status.Vote = models.RatingVoteNot
	case vote.Float64 > 0:
		status.Vote = models.RatingVotePos
	default:
		status.Vote = models.RatingVoteNeg
	}

	return &status
}

func canViewComment(tx *utils.AutoTx, userID, commentID int64) bool {
	const q = `
		SELECT entry_id
		FROM comments
		WHERE id = $1
	`

	var entryID int64
	tx.Query(q, commentID).Scan(&entryID)

	return utils.CanViewEntry(tx, userID, entryID)
}

func newCommentVoteLoader(srv *utils.MindwellServer) func(votes.GetCommentsIDVoteParams, *models.UserID) middleware.Responder {
	return func(params votes.GetCommentsIDVoteParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := uID.ID
			canView := canViewComment(tx, userID, params.ID)
			if !canView {
				err := srv.StandardError("no_entry")
				return votes.NewGetCommentsIDVoteNotFound().WithPayload(err)
			}

			status := commentRating(tx, userID, params.ID)
			if tx.Error() != nil {
				err := srv.StandardError("no_entry")
				return votes.NewGetCommentsIDVoteNotFound().WithPayload(err)
			}

			return votes.NewGetCommentsIDVoteOK().WithPayload(status)
		})
	}
}

func loadCommentRating(tx *utils.AutoTx, commentID int64) *models.Rating {
	const q = `
		SELECT up_votes, down_votes, rating
		FROM comments
		WHERE id = $1`

	rating := &models.Rating{
		ID:        commentID,
		IsVotable: true,
	}

	tx.Query(q, commentID).Scan(&rating.UpCount, &rating.DownCount, &rating.Rating)
	return rating
}

func voteForComment(tx *utils.AutoTx, userID, commentID int64, positive bool) *models.Rating {
	const q = `
		INSERT INTO comment_votes (user_id, comment_id, vote)
		VALUES ($1, $2, (
			SELECT GREATEST(0.001, weight) * $3
			FROM vote_weights
			WHERE vote_weights.user_id = $1 AND vote_weights.category = 
					(SELECT id FROM categories WHERE "type" = 'comment')
		))
		ON CONFLICT ON CONSTRAINT unique_comment_vote
		DO UPDATE SET vote = EXCLUDED.vote`

	var vote int64
	if positive {
		vote = 1
	} else {
		vote = -1
	}
	tx.Exec(q, userID, commentID, vote)

	rating := loadCommentRating(tx, commentID)

	switch {
	case positive:
		rating.Vote = models.RatingVotePos
	default:
		rating.Vote = models.RatingVoteNeg
	}

	return rating
}

func newCommentVoter(srv *utils.MindwellServer) func(votes.PutCommentsIDVoteParams, *models.UserID) middleware.Responder {
	return func(params votes.PutCommentsIDVoteParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := uID.ID
			canVote := canViewComment(tx, userID, params.ID)
			if tx.Error() != nil {
				err := srv.StandardError("no_entry")
				return votes.NewPutCommentsIDVoteNotFound().WithPayload(err)
			}

			if !canVote {
				err := srv.NewError(&i18n.Message{ID: "cant_vote", Other: "You are not allowed to vote for this comment."})
				return votes.NewPutCommentsIDVoteForbidden().WithPayload(err)
			}

			status := voteForComment(tx, userID, params.ID, *params.Positive)
			if tx.Error() != nil {
				err := srv.NewError(nil)
				return votes.NewPutCommentsIDVoteNotFound().WithPayload(err)
			}

			return votes.NewPutCommentsIDVoteOK().WithPayload(status)
		})
	}
}

func unvoteComment(tx *utils.AutoTx, userID, commentID int64) (*models.Rating, bool) {
	const q = `
		DELETE FROM comment_votes
		WHERE user_id = $1 AND comment_id = $2`

	tx.Exec(q, userID, commentID)
	if tx.RowsAffected() != 1 {
		return nil, false
	}

	rating := loadCommentRating(tx, commentID)
	rating.Vote = models.RatingVoteNot

	return rating, true
}

func newCommentUnvoter(srv *utils.MindwellServer) func(votes.DeleteCommentsIDVoteParams, *models.UserID) middleware.Responder {
	return func(params votes.DeleteCommentsIDVoteParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := uID.ID
			canVote := canViewComment(tx, userID, params.ID)
			if tx.Error() != nil {
				err := srv.StandardError("no_entry")
				return votes.NewDeleteCommentsIDVoteNotFound().WithPayload(err)
			}

			if !canVote {
				err := srv.NewError(&i18n.Message{ID: "cant_vote", Other: "You are not allowed to vote for this comment."})
				return votes.NewDeleteCommentsIDVoteForbidden().WithPayload(err)
			}

			status, ok := unvoteComment(tx, userID, params.ID)
			if !ok || tx.Error() != nil {
				err := srv.NewError(nil)
				return votes.NewDeleteCommentsIDVoteNotFound().WithPayload(err)
			}

			return votes.NewDeleteCommentsIDVoteOK().WithPayload(status)
		})
	}
}
