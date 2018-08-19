package test

import (
	"testing"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/favorites"
	"github.com/stretchr/testify/require"
)

func checkEntryFavorite(t *testing.T, user *models.UserID, entryID int64, fav, success bool) {
	load := api.FavoritesGetEntriesIDFavoriteHandler.Handle
	params := favorites.GetEntriesIDFavoriteParams{
		ID: entryID,
	}
	resp := load(params, user)
	body, ok := resp.(*favorites.GetEntriesIDFavoriteOK)
	req := require.New(t)
	req.Equal(success, ok)
	if !success {
		return
	}

	status := body.Payload
	req.Equal(entryID, status.ID)
	req.Equal(fav, status.IsFavorited)
}

func checkFavoriteEntry(t *testing.T, user *models.UserID, entryID int64, success bool) {
	put := api.FavoritesPutEntriesIDFavoriteHandler.Handle
	params := favorites.PutEntriesIDFavoriteParams{
		ID: entryID,
	}
	resp := put(params, user)
	body, ok := resp.(*favorites.PutEntriesIDFavoriteOK)
	req := require.New(t)
	req.Equal(success, ok)
	if !success {
		return
	}

	status := body.Payload
	req.Equal(entryID, status.ID)
	req.True(status.IsFavorited)
}

func checkUnfavoriteEntry(t *testing.T, user *models.UserID, entryID int64, success bool) {
	del := api.FavoritesDeleteEntriesIDFavoriteHandler.Handle
	params := favorites.DeleteEntriesIDFavoriteParams{
		ID: entryID,
	}
	resp := del(params, user)
	body, ok := resp.(*favorites.DeleteEntriesIDFavoriteOK)
	req := require.New(t)
	req.Equal(success, ok)
	if !success {
		return
	}

	status := body.Payload
	req.Equal(entryID, status.ID)
	req.False(status.IsFavorited)
}

func TestFavorite(t *testing.T) {
	e := createTlogEntry(t, userIDs[0], models.EntryPrivacyAll, true, false)
	checkEntryFavorite(t, userIDs[0], e.ID, false, true)
	checkEntryFavorite(t, userIDs[0], e.ID, false, true)
	checkEntryFavorite(t, userIDs[1], e.ID, false, true)

	checkFavoriteEntry(t, userIDs[0], e.ID, true)
	checkEntryFavorite(t, userIDs[0], e.ID, true, true)
	checkFavoriteEntry(t, userIDs[1], e.ID, true)
	checkEntryFavorite(t, userIDs[1], e.ID, true, true)
	checkUnfavoriteEntry(t, userIDs[1], e.ID, true)
	checkEntryFavorite(t, userIDs[1], e.ID, false, true)
	checkUnfavoriteEntry(t, userIDs[0], e.ID, true)
	checkUnfavoriteEntry(t, userIDs[0], e.ID, true)
	checkEntryFavorite(t, userIDs[0], e.ID, false, true)

	e = createTlogEntry(t, userIDs[0], models.EntryPrivacyMe, true, true)
	checkEntryFavorite(t, userIDs[0], e.ID, false, true)
	checkEntryFavorite(t, userIDs[1], e.ID, false, false)
	checkFavoriteEntry(t, userIDs[1], e.ID, false)
	checkEntryFavorite(t, userIDs[1], e.ID, false, false)
	checkUnfavoriteEntry(t, userIDs[1], e.ID, false)
	checkEntryFavorite(t, userIDs[1], e.ID, false, false)
	checkFavoriteEntry(t, userIDs[0], e.ID, true)
	checkEntryFavorite(t, userIDs[0], e.ID, true, true)
	checkUnfavoriteEntry(t, userIDs[0], e.ID, true)
	checkEntryFavorite(t, userIDs[0], e.ID, false, true)
}
