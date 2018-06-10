package favorites

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/favorites"
	"github.com/sevings/mindwell-server/utils"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	srv.API.FavoritesGetEntriesIDFavoriteHandler = favorites.GetEntriesIDFavoriteHandlerFunc(newStatusLoader(srv))
	srv.API.FavoritesPutEntriesIDFavoriteHandler = favorites.PutEntriesIDFavoriteHandlerFunc(newFavoriteAdder(srv))
	srv.API.FavoritesDeleteEntriesIDFavoriteHandler = favorites.DeleteEntriesIDFavoriteHandlerFunc(newFavoriteDeleter(srv))
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

func newStatusLoader(srv *utils.MindwellServer) func(favorites.GetEntriesIDFavoriteParams, *models.UserID) middleware.Responder {
	return func(params favorites.GetEntriesIDFavoriteParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
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

func newFavoriteAdder(srv *utils.MindwellServer) func(favorites.PutEntriesIDFavoriteParams, *models.UserID) middleware.Responder {
	return func(params favorites.PutEntriesIDFavoriteParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
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

func newFavoriteDeleter(srv *utils.MindwellServer) func(favorites.DeleteEntriesIDFavoriteParams, *models.UserID) middleware.Responder {
	return func(params favorites.DeleteEntriesIDFavoriteParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
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
