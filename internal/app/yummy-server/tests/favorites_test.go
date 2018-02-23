package tests

import (
	"database/sql"
	"os"
	"testing"

	entriesImpl "github.com/sevings/yummy-server/internal/app/yummy-server/entries"
	favoritesImpl "github.com/sevings/yummy-server/internal/app/yummy-server/favorites"
	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"
	"github.com/sevings/yummy-server/restapi/operations/favorites"
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

func checkEntryFavorite(t *testing.T, user *models.UserID, entryID int64, fav, success bool) {
	api := operations.YummyAPI{}
	favoritesImpl.ConfigureAPI(db, &api)

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
	api := operations.YummyAPI{}
	favoritesImpl.ConfigureAPI(db, &api)

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
	api := operations.YummyAPI{}
	favoritesImpl.ConfigureAPI(db, &api)

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
	api := operations.YummyAPI{}
	entriesImpl.ConfigureAPI(db, &api)

	post := func(id *models.UserID, privacy string, votable bool) *models.Entry {
		return PostEntry(&api, id, privacy, votable)
	}

	e := post(userIDs[0], models.EntryPrivacyAll, true)
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

	e = post(userIDs[0], models.EntryPrivacyMe, true)
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
