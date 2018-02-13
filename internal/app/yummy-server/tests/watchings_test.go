package tests

import (
	"database/sql"
	"log"
	"os"
	"testing"

	entriesImpl "github.com/sevings/yummy-server/internal/app/yummy-server/entries"
	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	watchingsImpl "github.com/sevings/yummy-server/internal/app/yummy-server/watchings"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"
	"github.com/sevings/yummy-server/restapi/operations/watchings"
	"github.com/stretchr/testify/require"
)

var db *sql.DB
var userIDs []*models.UserID
var profiles []*models.AuthProfile

func TestMain(m *testing.M) {
	config := utils.LoadConfig("../../../../configs/server")
	db = utils.OpenDatabase(config)
	utils.ClearDatabase(db)

	userIDs, profiles = RegisterTestUsers(db)

	os.Exit(m.Run())
}

func checkEntryWatching(t *testing.T, user *models.UserID, entryID int64, watching, success bool) {
	api := operations.YummyAPI{}
	watchingsImpl.ConfigureAPI(db, &api)

	load := api.WatchingsGetEntriesIDWatchingHandler.Handle
	params := watchings.GetEntriesIDWatchingParams{
		ID: entryID,
	}
	resp := load(params, user)
	body, ok := resp.(*watchings.GetEntriesIDWatchingOK)
	req := require.New(t)
	req.Equal(success, ok)
	if !success {
		return
	}

	if !ok {
		badBody, ok := resp.(*watchings.GetEntriesIDWatchingNotFound)
		if ok {
			log.Fatal(badBody.Payload.Message)
		}

		log.Fatal("error load vote status")
	}

	status := body.Payload
	req.Equal(entryID, status.ID)
	req.Equal(watching, status.IsWatching)
}

func checkWatchEntry(t *testing.T, user *models.UserID, entryID int64, success bool) {
	api := operations.YummyAPI{}
	watchingsImpl.ConfigureAPI(db, &api)

	put := api.WatchingsPutEntriesIDWatchingHandler.Handle
	params := watchings.PutEntriesIDWatchingParams{
		ID: entryID,
	}
	resp := put(params, user)
	body, ok := resp.(*watchings.PutEntriesIDWatchingOK)
	req := require.New(t)
	req.Equal(success, ok)
	if !success {
		return
	}

	if !ok {
		badBody, ok := resp.(*watchings.PutEntriesIDWatchingNotFound)
		if ok {
			log.Fatal(badBody.Payload.Message)
		}

		log.Fatal("error vote for entry")
	}

	status := body.Payload
	req.Equal(entryID, status.ID)
	req.True(status.IsWatching)
}

func checkUnwatchEntry(t *testing.T, user *models.UserID, entryID int64, success bool) {
	api := operations.YummyAPI{}
	watchingsImpl.ConfigureAPI(db, &api)

	del := api.WatchingsDeleteEntriesIDWatchingHandler.Handle
	params := watchings.DeleteEntriesIDWatchingParams{
		ID: entryID,
	}
	resp := del(params, user)
	body, ok := resp.(*watchings.DeleteEntriesIDWatchingOK)
	req := require.New(t)
	req.Equal(success, ok)
	if !success {
		return
	}

	if !ok {
		badBody, ok := resp.(*watchings.DeleteEntriesIDWatchingNotFound)
		if ok {
			log.Fatal(badBody.Payload.Message)
		}

		log.Fatal("error vote for entry")
	}

	status := body.Payload
	req.Equal(entryID, status.ID)
	req.False(status.IsWatching)
}

func TestWatching(t *testing.T) {
	api := operations.YummyAPI{}
	entriesImpl.ConfigureAPI(db, &api)

	post := func(id *models.UserID, privacy string, votable bool) *models.Entry {
		return PostEntry(&api, id, privacy, votable)
	}

	e := post(userIDs[0], models.EntryPrivacyAll, true)
	checkEntryWatching(t, userIDs[0], e.ID, true, true)
	checkEntryWatching(t, userIDs[0], e.ID, true, true)
	checkEntryWatching(t, userIDs[1], e.ID, false, true)

	checkWatchEntry(t, userIDs[0], e.ID, true)
	checkEntryWatching(t, userIDs[0], e.ID, true, true)
	checkWatchEntry(t, userIDs[1], e.ID, true)
	checkEntryWatching(t, userIDs[1], e.ID, true, true)
	checkUnwatchEntry(t, userIDs[1], e.ID, true)
	checkEntryWatching(t, userIDs[1], e.ID, false, true)
	checkUnwatchEntry(t, userIDs[0], e.ID, true)
	checkUnwatchEntry(t, userIDs[0], e.ID, true)
	checkEntryWatching(t, userIDs[0], e.ID, false, true)

	e = post(userIDs[0], models.EntryPrivacyMe, true)
	checkEntryWatching(t, userIDs[0], e.ID, true, true)
	checkEntryWatching(t, userIDs[1], e.ID, true, false)
	checkWatchEntry(t, userIDs[1], e.ID, false)
	checkEntryWatching(t, userIDs[1], e.ID, false, false)
	checkUnwatchEntry(t, userIDs[1], e.ID, false)
	checkEntryWatching(t, userIDs[1], e.ID, false, false)
}
