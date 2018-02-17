package votes

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	entriesImpl "github.com/sevings/yummy-server/internal/app/yummy-server/entries"
	"github.com/sevings/yummy-server/internal/app/yummy-server/tests"
	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"
	"github.com/sevings/yummy-server/restapi/operations/votes"
)

var db *sql.DB
var userIDs []*models.UserID
var profiles []*models.AuthProfile

func TestMain(m *testing.M) {
	config := utils.LoadConfig("../../../../configs/server")
	db = utils.OpenDatabase(config)
	utils.ClearDatabase(db)

	userIDs, profiles = tests.RegisterTestUsers(db)

	os.Exit(m.Run())
}

// NewPostEntry returns func creating entries
func NewPostEntry(db *sql.DB) func(id *models.UserID, privacy string, votable bool) *models.Entry {
	api := operations.YummyAPI{}
	entriesImpl.ConfigureAPI(db, &api)

	return func(id *models.UserID, privacy string, votable bool) *models.Entry {
		return tests.PostEntry(&api, id, privacy, votable)
	}
}

func checkEntryVote(t *testing.T, user *models.UserID, entryID, eVotes int64, vote string) {
	load := newEntryVoteLoader(db)
	params := votes.GetEntriesIDVoteParams{
		ID: entryID,
	}
	resp := load(params, user)
	body, ok := resp.(*votes.GetEntriesIDVoteOK)
	if !ok {
		badBody, ok := resp.(*votes.GetEntriesIDVoteForbidden)
		if ok {
			log.Fatal(badBody.Payload.Message)
		}

		log.Fatal("error load vote status")
	}

	status := body.Payload
	req := require.New(t)
	req.Equal(entryID, status.ID)
	req.Equal(eVotes, status.Votes)
	req.Equal(vote, status.Vote)
}

func checkVoteForEntry(t *testing.T, user *models.UserID, success bool, entryID, eVotes int64, positive bool, vote string) {
	put := newEntryVoter(db)
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
	req.Equal(eVotes, status.Votes)
	req.Equal(vote, status.Vote)
}

func checkUnvoteEntry(t *testing.T, user *models.UserID, success bool, entryID, eVotes int64) {
	del := newEntryUnvoter(db)
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
	req.Equal(eVotes, status.Votes)
	req.Equal(models.EntryVoteNot, status.Vote)
}

func TestEntryVotes(t *testing.T) {
	post := NewPostEntry(db)

	e := post(userIDs[0], models.EntryPrivacyAll, true)
	checkEntryVote(t, userIDs[0], e.ID, 0, models.EntryVoteBan)
	checkEntryVote(t, userIDs[1], e.ID, 0, models.EntryVoteNot)

	checkVoteForEntry(t, userIDs[1], true, e.ID, 1, true, models.EntryVotePos)
	checkVoteForEntry(t, userIDs[1], true, e.ID, -1, false, models.EntryVoteNeg)
	checkVoteForEntry(t, userIDs[2], true, e.ID, 0, true, models.EntryVotePos)
	checkEntryVote(t, userIDs[1], e.ID, 0, models.EntryVoteNeg)

	checkUnvoteEntry(t, userIDs[2], true, e.ID, -1)
	checkEntryVote(t, userIDs[2], e.ID, -1, models.EntryVoteNot)
	checkUnvoteEntry(t, userIDs[2], false, e.ID, -1)

	checkUnvoteEntry(t, userIDs[1], true, e.ID, 0)
	checkEntryVote(t, userIDs[1], e.ID, 0, models.EntryVoteNot)

	checkVoteForEntry(t, userIDs[0], false, e.ID, 0, false, "")
	checkUnvoteEntry(t, userIDs[0], false, e.ID, 0)

	e = post(userIDs[0], models.EntryPrivacyAll, false)
	checkEntryVote(t, userIDs[1], e.ID, 0, models.EntryVoteBan)

	checkVoteForEntry(t, userIDs[0], false, e.ID, 0, false, "")
	checkVoteForEntry(t, userIDs[1], false, e.ID, 0, false, "")
	checkUnvoteEntry(t, userIDs[2], false, e.ID, -1)

	e = post(userIDs[0], models.EntryPrivacyAnonymous, true)
	checkEntryVote(t, userIDs[0], e.ID, 0, models.EntryVoteBan)
	checkEntryVote(t, userIDs[1], e.ID, 0, models.EntryVoteBan)

	checkVoteForEntry(t, userIDs[0], false, e.ID, 0, false, "")
	checkVoteForEntry(t, userIDs[1], false, e.ID, 0, false, "")
	checkUnvoteEntry(t, userIDs[2], false, e.ID, -1)
}
