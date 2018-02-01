package favorites

import (
	"database/sql"
	"log"

	"github.com/go-openapi/runtime/middleware"
	yummy "github.com/sevings/yummy-server/internal/app/yummy-server"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"
	"github.com/sevings/yummy-server/restapi/operations/favorites"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.FavoritesGetEntriesIDFavoriteHandler = favorites.GetEntriesIDFavoriteHandlerFunc(newStatusLoader(db))
}

func favoriteStatus(tx yummy.AutoTx, userID, entryID int64) *models.FavoriteStatus {
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

func newStatusLoader(db *sql.DB) func(favorites.GetEntriesIDFavoriteParams, *models.UserID) middleware.Responder {
	return func(params favorites.GetEntriesIDFavoriteParams, uID *models.UserID) middleware.Responder {
		return yummy.Transact(db, func(tx yummy.AutoTx) (middleware.Responder, bool) {
			userID := int64(*uID)
			canView := yummy.CanViewEntry(tx, userID, params.ID)
			if !canView {
				return favorites.NewGetEntriesIDFavoriteNotFound(), false
			}

			status := favoriteStatus(tx, userID, params.ID)
			return favorites.NewGetEntriesIDFavoriteOK().WithPayload(status), true
		})
	}
}
