package votes

import (
	"database/sql"
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations"
	"github.com/sevings/yummy-server/gen/restapi/operations/votes"
	"github.com/sevings/yummy-server/src/entries"
	"github.com/sevings/yummy-server/src/users"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.VotesGetEntriesIDVoteHandler = votes.GetEntriesIDVoteHandlerFunc(newEntryVoteLoader(db))
}

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
