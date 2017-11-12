package votes

import (
	"database/sql"

	"github.com/sevings/yummy-server/gen/restapi/operations"
	"github.com/sevings/yummy-server/gen/restapi/operations/votes"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.VotesGetEntriesIDVoteHandler = votes.GetEntriesIDVoteHandlerFunc(newEntryVoteLoader(db))
	api.VotesPutEntriesIDVoteHandler = votes.PutEntriesIDVoteHandlerFunc(newEntryVoter(db))
	api.VotesDeleteEntriesIDVoteHandler = votes.DeleteEntriesIDVoteHandlerFunc(newEntryUnvoter(db))
}
