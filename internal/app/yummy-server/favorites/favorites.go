package favorites

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"
	"github.com/sevings/yummy-server/restapi/operations/favorites"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.FavoritesGetEntriesIDFavoriteHandler = favorites.GetEntriesIDFavoriteHandlerFunc(newStatusLoader(db))
	api.FavoritesPutEntriesIDFavoriteHandler = favorites.PutEntriesIDFavoriteHandlerFunc(newFavoriteAdder(db))
	api.FavoritesDeleteEntriesIDFavoriteHandler = favorites.DeleteEntriesIDFavoriteHandlerFunc(newFavoriteDeleter(db))
}

func favoriteStatus(tx *utils.AutoTx, userID, entryID int64) *models.FavoriteStatus {
	const q = `
		SELECT TRUE 
		FROM favorites
		WHERE user_id = $1 AND entry_id = $2`

	status := models.FavoriteStatus{ID: entryID}

	tx.Query(q, userID, entryID).Scan(&status.IsFavorited)

	return &status
}

func newStatusLoader(db *sql.DB) func(favorites.GetEntriesIDFavoriteParams, *models.UserID) middleware.Responder {
	return func(params favorites.GetEntriesIDFavoriteParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			canView := utils.CanViewEntry(tx, userID, params.ID)
			if !canView {
				return favorites.NewGetEntriesIDFavoriteNotFound()
			}

			status := favoriteStatus(tx, userID, params.ID)
			return favorites.NewGetEntriesIDFavoriteOK().WithPayload(status)
		})
	}
}

func addToFavorites(tx *utils.AutoTx, userID, entryID int64) *models.FavoriteStatus {
	const q = `
		INSERT INTO favorites(user_id, entry_id)
		VALUES($1, $2)
		ON CONFLICT ON CONSTRAINT unique_user_favorite
		DO NOTHING`

	tx.Exec(q, userID, entryID)

	status := models.FavoriteStatus{
		ID:          entryID,
		IsFavorited: true,
	}

	return &status
}

func newFavoriteAdder(db *sql.DB) func(favorites.PutEntriesIDFavoriteParams, *models.UserID) middleware.Responder {
	return func(params favorites.PutEntriesIDFavoriteParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			canView := utils.CanViewEntry(tx, userID, params.ID)
			if !canView {
				return favorites.NewPutEntriesIDFavoriteNotFound()
			}

			status := addToFavorites(tx, userID, params.ID)
			return favorites.NewPutEntriesIDFavoriteOK().WithPayload(status)
		})
	}
}

func removeFromFavorites(tx *utils.AutoTx, userID, entryID int64) *models.FavoriteStatus {
	const q = `
		DELETE FROM favorites
		WHERE user_id = $1 AND entry_id = $2`

	tx.Exec(q, userID, entryID)

	status := models.FavoriteStatus{
		ID:          entryID,
		IsFavorited: false,
	}

	return &status
}

func newFavoriteDeleter(db *sql.DB) func(favorites.DeleteEntriesIDFavoriteParams, *models.UserID) middleware.Responder {
	return func(params favorites.DeleteEntriesIDFavoriteParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			canView := utils.CanViewEntry(tx, userID, params.ID)
			if !canView {
				return favorites.NewDeleteEntriesIDFavoriteNotFound()
			}

			status := removeFromFavorites(tx, userID, params.ID)
			return favorites.NewDeleteEntriesIDFavoriteOK().WithPayload(status)
		})
	}
}
