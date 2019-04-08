package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/votes"
)

func checkCommentVote(t *testing.T, user *models.UserID, commentID, cVotes int64, vote string) {
	load := api.VotesGetCommentsIDVoteHandler.Handle
	params := votes.GetCommentsIDVoteParams{
		ID: commentID,
	}
	resp := load(params, user)
	body, ok := resp.(*votes.GetCommentsIDVoteOK)
	req := require.New(t)
	req.True(ok)

	status := body.Payload
	req.Equal(commentID, status.ID)
	req.True(status.IsVotable)
	req.Equal(cVotes, status.UpCount-status.DownCount)
	req.Equal(vote, status.Vote)
}

func checkVoteForComment(t *testing.T, user *models.UserID, success bool, commentID, cVotes int64, positive bool, vote string) {
	put := api.VotesPutCommentsIDVoteHandler.Handle
	params := votes.PutCommentsIDVoteParams{
		ID:       commentID,
		Positive: &positive,
	}
	resp := put(params, user)
	body, ok := resp.(*votes.PutCommentsIDVoteOK)
	req := require.New(t)
	req.Equal(success, ok)
	if !success {
		return
	}

	status := body.Payload
	req.Equal(commentID, status.ID)
	req.True(status.IsVotable)
	req.Equal(cVotes, status.UpCount-status.DownCount)
	req.Equal(vote, status.Vote)
}

func checkUnvoteComment(t *testing.T, user *models.UserID, success bool, commentID, cVotes int64) {
	del := api.VotesDeleteCommentsIDVoteHandler.Handle
	params := votes.DeleteCommentsIDVoteParams{
		ID: commentID,
	}
	resp := del(params, user)
	body, ok := resp.(*votes.DeleteCommentsIDVoteOK)
	req := require.New(t)
	req.Equal(success, ok)
	if !success {
		return
	}

	status := body.Payload
	req.Equal(commentID, status.ID)
	req.True(status.IsVotable)
	req.Equal(cVotes, status.UpCount-status.DownCount)
	req.Equal(models.RatingVoteNot, status.Vote)
}

func TestCommentVotes(t *testing.T) {
	e := createTlogEntry(t, userIDs[0], models.EntryPrivacyAll, true, false)
	c := createComment(t, userIDs[0], e.ID)

	checkCommentVote(t, userIDs[0], c.ID, 0, models.RatingVoteBan)
	checkCommentVote(t, userIDs[1], c.ID, 0, models.RatingVoteNot)

	checkVoteForComment(t, userIDs[1], true, c.ID, 1, true, models.RatingVotePos)
	checkVoteForComment(t, userIDs[1], true, c.ID, -1, false, models.RatingVoteNeg)
	checkVoteForComment(t, userIDs[2], true, c.ID, 0, true, models.RatingVotePos)
	checkCommentVote(t, userIDs[1], c.ID, 0, models.RatingVoteNeg)

	checkUnvoteComment(t, userIDs[2], true, c.ID, -1)
	checkCommentVote(t, userIDs[2], c.ID, -1, models.RatingVoteNot)
	checkUnvoteComment(t, userIDs[2], false, c.ID, -1)

	checkUnvoteComment(t, userIDs[1], true, c.ID, 0)
	checkCommentVote(t, userIDs[1], c.ID, 0, models.RatingVoteNot)

	checkVoteForComment(t, userIDs[0], false, c.ID, 0, false, "")
	checkUnvoteComment(t, userIDs[0], false, c.ID, 0)

	banVote(db, userIDs[1])
	checkVoteForComment(t, userIDs[1], false, c.ID, 1, true, models.RatingVotePos)
	removeUserRestrictions(db, userIDs)
}
