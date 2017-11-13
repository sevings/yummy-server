package watchings

import (
	"database/sql"
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations"
	"github.com/sevings/yummy-server/gen/restapi/operations/watchings"
	"github.com/sevings/yummy-server/src/entries"
	"github.com/sevings/yummy-server/src/users"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.WatchingsGetEntriesIDWatchingHandler = watchings.GetEntriesIDWatchingHandlerFunc(newStatusLoader(db))
}

func watchingStatus(tx *sql.Tx, userID, entryID int64) *models.WatchingStatus {
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

func newStatusLoader(db *sql.DB) func(watchings.GetEntriesIDWatchingParams) middleware.Responder {
	return func(params watchings.GetEntriesIDWatchingParams) middleware.Responder {
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		defer tx.Commit()

		userID, found := users.FindAuthUser(tx, &params.XUserKey)
		if !found {
			return watchings.NewGetEntriesIDWatchingForbidden()
		}

		canView := entries.CanViewEntry(tx, userID, params.ID)
		if !canView {
			return watchings.NewGetEntriesIWatchingNotFound()
		}

		status := watchingStatus(tx, userID, params.ID)
		return watchings.NewGetEntriesIDWatchingOK().WithPayload(status)
	}
}
