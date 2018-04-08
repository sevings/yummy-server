package test

import (
	"testing"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/watchings"
	"github.com/stretchr/testify/require"
)

func checkEntryWatching(t *testing.T, user *models.UserID, entryID int64, watching, success bool) {
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

	status := body.Payload
	req.Equal(entryID, status.ID)
	req.Equal(watching, status.IsWatching)
}

func checkWatchEntry(t *testing.T, user *models.UserID, entryID int64, success bool) {
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

	status := body.Payload
	req.Equal(entryID, status.ID)
	req.True(status.IsWatching)
}

func checkUnwatchEntry(t *testing.T, user *models.UserID, entryID int64, success bool) {
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

	status := body.Payload
	req.Equal(entryID, status.ID)
	req.False(status.IsWatching)
}

func TestWatching(t *testing.T) {
	e := createTlogEntry(t, userIDs[0], models.EntryPrivacyAll, true)
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

	e = createTlogEntry(t, userIDs[0], models.EntryPrivacyMe, true)
	checkEntryWatching(t, userIDs[0], e.ID, true, true)
	checkEntryWatching(t, userIDs[1], e.ID, false, false)
	checkWatchEntry(t, userIDs[1], e.ID, false)
	checkEntryWatching(t, userIDs[1], e.ID, false, false)
	checkUnwatchEntry(t, userIDs[1], e.ID, false)
	checkEntryWatching(t, userIDs[1], e.ID, false, false)
}
