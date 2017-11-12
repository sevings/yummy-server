package favorites

import (
	"database/sql"
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations"
	"github.com/sevings/yummy-server/gen/restapi/operations/favorites"
	"github.com/sevings/yummy-server/src/entries"
	"github.com/sevings/yummy-server/src/users"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.FavoritesGetEntriesIDFavoriteHandler = favorites.GetEntriesIDFavoriteHandlerFunc(newStatusLoader(db))
}

func favoriteStatus(tx *sql.Tx, userID, entryID int64) *models.FavoriteStatus {
	const q = `
		SELECT TRUE 
		FROM favorites
		WHERE user_id = $1 AND entry_id = $2`

	status := models.FavoriteStatus{ID: entryID}

	err := tx.QueryRow(q, userID, entryID).Scan(&status.IsFavorited)
	if err != nil && err != sql.ErrNoRows {
		log.Print(err)
	}

	return &status
}

func newStatusLoader(db *sql.DB) func(favorites.GetEntriesIDFavoriteParams) middleware.Responder {
	return func(params favorites.GetEntriesIDFavoriteParams) middleware.Responder {
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		defer tx.Commit()

		userID, found := users.FindAuthUser(tx, &params.XUserKey)
		if !found {
			return favorites.NewGetEntriesIDFavoriteForbidden()
		}

		canView := entries.CanViewEntry(tx, userID, params.ID)
		if !canView {
			return favorites.NewGetEntriesIDFavoriteNotFound()
		}

		status := favoriteStatus(tx, userID, params.ID)
		return favorites.NewGetEntriesIDFavoriteOK().WithPayload(status)
	}
}
