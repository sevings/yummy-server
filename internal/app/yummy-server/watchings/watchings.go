package watchings

import (
	"database/sql"
	"log"

	"github.com/go-openapi/runtime/middleware"
	yummy "github.com/sevings/yummy-server/internal/app/yummy-server"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"
	"github.com/sevings/yummy-server/restapi/operations/watchings"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.WatchingsGetEntriesIDWatchingHandler = watchings.GetEntriesIDWatchingHandlerFunc(newStatusLoader(db))
}

func watchingStatus(tx yummy.AutoTx, userID, entryID int64) *models.WatchingStatus {
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
		return yummy.Transact(db, func(tx yummy.AutoTx) (middleware.Responder, bool) {
			userID := int64(*uID)
			canView := yummy.CanViewEntry(tx, userID, params.ID)
			if !canView {
				return watchings.NewGetEntriesIDWatchingNotFound(), false
			}

			status := watchingStatus(tx, userID, params.ID)
			return watchings.NewGetEntriesIDWatchingOK().WithPayload(status), true
		})
	}
}
