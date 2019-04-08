package test

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/votes"
)

func checkEntryVote(t *testing.T, user *models.UserID, entryID, eVotes int64, vote string) {
	load := api.VotesGetEntriesIDVoteHandler.Handle
	params := votes.GetEntriesIDVoteParams{
		ID: entryID,
	}
	resp := load(params, user)
	body, ok := resp.(*votes.GetEntriesIDVoteOK)
	req := require.New(t)
	req.True(ok)
	status := body.Payload
	req.Equal(entryID, status.ID)
	req.Equal(eVotes, status.UpCount-status.DownCount)
	req.Equal(vote, status.Vote)
}

func checkVoteForEntry(t *testing.T, user *models.UserID, success bool, entryID, eVotes int64, positive bool, vote string) {
	put := api.VotesPutEntriesIDVoteHandler.Handle
	params := votes.PutEntriesIDVoteParams{
		ID:       entryID,
		Positive: &positive,
	}
	resp := put(params, user)
	body, ok := resp.(*votes.PutEntriesIDVoteOK)
	req := require.New(t)
	req.Equal(success, ok)
	if !success {
		return
	}

	if !ok {
		badBody, ok := resp.(*votes.PutEntriesIDVoteForbidden)
		if ok {
			log.Fatal(badBody.Payload.Message)
		}

		log.Fatal("error vote for entry")
	}

	status := body.Payload
	req.Equal(entryID, status.ID)
	req.Equal(eVotes, status.UpCount-status.DownCount)
	req.Equal(vote, status.Vote)
}

func checkUnvoteEntry(t *testing.T, user *models.UserID, success bool, entryID, eVotes int64) {
	del := api.VotesDeleteEntriesIDVoteHandler.Handle
	params := votes.DeleteEntriesIDVoteParams{
		ID: entryID,
	}
	resp := del(params, user)
	body, ok := resp.(*votes.DeleteEntriesIDVoteOK)
	req := require.New(t)
	req.Equal(success, ok)
	if !success {
		return
	}

	if !ok {
		badBody, ok := resp.(*votes.DeleteEntriesIDVoteForbidden)
		if ok {
			log.Fatal(badBody.Payload.Message)
		}

		log.Fatal("error vote for entry")
	}

	status := body.Payload
	req.Equal(entryID, status.ID)
	req.Equal(eVotes, status.UpCount-status.DownCount)
	req.Equal(models.RatingVoteNot, status.Vote)
}

func TestEntryVotes(t *testing.T) {
	e := createTlogEntry(t, userIDs[0], models.EntryPrivacyAll, true, false)
	checkEntryVote(t, userIDs[0], e.ID, 0, models.RatingVoteBan)
	checkEntryVote(t, userIDs[1], e.ID, 0, models.RatingVoteNot)

	checkVoteForEntry(t, userIDs[1], true, e.ID, 1, true, models.RatingVotePos)
	checkVoteForEntry(t, userIDs[1], true, e.ID, -1, false, models.RatingVoteNeg)
	checkVoteForEntry(t, userIDs[2], true, e.ID, 0, true, models.RatingVotePos)
	checkEntryVote(t, userIDs[1], e.ID, 0, models.RatingVoteNeg)

	checkUnvoteEntry(t, userIDs[2], true, e.ID, -1)
	checkEntryVote(t, userIDs[2], e.ID, -1, models.RatingVoteNot)
	checkUnvoteEntry(t, userIDs[2], false, e.ID, -1)

	checkUnvoteEntry(t, userIDs[1], true, e.ID, 0)
	checkEntryVote(t, userIDs[1], e.ID, 0, models.RatingVoteNot)

	checkVoteForEntry(t, userIDs[0], false, e.ID, 0, false, "")
	checkUnvoteEntry(t, userIDs[0], false, e.ID, 0)

	e = createTlogEntry(t, userIDs[0], models.EntryPrivacyAll, false, false)
	checkEntryVote(t, userIDs[1], e.ID, 0, models.RatingVoteBan)

	checkVoteForEntry(t, userIDs[0], false, e.ID, 0, false, "")
	checkVoteForEntry(t, userIDs[1], false, e.ID, 0, false, "")
	checkUnvoteEntry(t, userIDs[2], false, e.ID, -1)

	e = createTlogEntry(t, userIDs[0], models.EntryPrivacyAnonymous, true, false)
	checkEntryVote(t, userIDs[0], e.ID, 0, models.RatingVoteBan)
	checkEntryVote(t, userIDs[1], e.ID, 0, models.RatingVoteBan)

	checkVoteForEntry(t, userIDs[0], false, e.ID, 0, false, "")
	checkVoteForEntry(t, userIDs[1], false, e.ID, 0, false, "")
	checkUnvoteEntry(t, userIDs[2], false, e.ID, -1)

	banVote(db, userIDs[1])
	checkVoteForEntry(t, userIDs[1], false, e.ID, 1, true, models.RatingVotePos)
	removeUserRestrictions(db, userIDs)
}
