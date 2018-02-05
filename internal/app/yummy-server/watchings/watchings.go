package watchings

import (
	"database/sql"
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"
	"github.com/sevings/yummy-server/restapi/operations/watchings"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.WatchingsGetEntriesIDWatchingHandler = watchings.GetEntriesIDWatchingHandlerFunc(newStatusLoader(db))
}

func watchingStatus(tx utils.AutoTx, userID, entryID int64) *models.WatchingStatus {
	const q = `
		SELECT TRUE 
		FROM watching
		WHERE user_id = $1 AND entry_id = $2`

	status := models.WatchingStatus{ID: entryID}

	err := tx.QueryRow(q, userID, entryID).Scan(&status.IsWatching)
	if err != nil && err != sql.ErrNoRows {
		log.Print(err)
	}

	return &status
}

func newStatusLoader(db *sql.DB) func(watchings.GetEntriesIDWatchingParams, *models.UserID) middleware.Responder {
	return func(params watchings.GetEntriesIDWatchingParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx utils.AutoTx) (middleware.Responder, bool) {
			userID := int64(*uID)
			canView := utils.CanViewEntry(tx, userID, params.ID)
			if !canView {
				return watchings.NewGetEntriesIDWatchingNotFound(), false
			}

			status := watchingStatus(tx, userID, params.ID)
			return watchings.NewGetEntriesIDWatchingOK().WithPayload(status), true
		})
	}
}

func AddWatching(tx utils.AutoTx, userID, entryID int64) error {
	const q = `
		INSERT INTO watching(user_id, entry_id)
		VALUES($1, $2)
		ON CONFLICT ON CONSTRAINT unique_user_watching
		DO NOTHING`

	_, err := tx.Exec(q, userID, entryID)
	if err != nil {
		log.Print(err)
	}

	return err
}

func RemoveWatching(tx utils.AutoTx, userID, entryID int64) error {
	const q = `
		DELETE FROM watching
		WHERE user_id = $1 AND entry_id = $2`

	_, err := tx.Exec(q, userID, entryID)
	if err != nil && err != sql.ErrNoRows {
		log.Print(err)
		return err
	}

	return nil
}
