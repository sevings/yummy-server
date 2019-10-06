package complains

import (
	"database/sql"
	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/comments"
	"github.com/sevings/mindwell-server/restapi/operations/entries"
	"github.com/sevings/mindwell-server/utils"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	srv.API.EntriesPostEntriesIDComplainHandler = entries.PostEntriesIDComplainHandlerFunc(newEntryComplainer(srv))
	srv.API.CommentsPostCommentsIDComplainHandler = comments.PostCommentsIDComplainHandlerFunc(newCommentComplainer(srv))
}

const selectPrevQuery = `
    SELECT id 
    FROM complains 
    WHERE user_id = $1 AND subject_id = $2 
        AND type = (SELECT id FROM complain_type WHERE type = $3)
`

const updateQuery = `
    UPDATE complains
    SET content = $2
    WHERE id = $1
`

const createQuery = `
    INSERT INTO complains(user_id, type, subject_id, content)
    VALUES($1, $2, $3, $4)    
`

func newEntryComplainer(srv *utils.MindwellServer) func(entries.PostEntriesIDComplainParams, *models.UserID) middleware.Responder {
	return func(params entries.PostEntriesIDComplainParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			allowed := utils.CanViewEntry(tx, userID.ID, params.ID)
			if !allowed {
				err := srv.StandardError("no_entry")
				return entries.NewPostEntriesIDComplainNotFound().WithPayload(err)
			}

			id := tx.QueryInt64(selectPrevQuery, userID.ID, params.ID, "entry")
			if tx.Error() != sql.ErrNoRows {
				if tx.Error() != nil {
					err := srv.StandardError("no_entry")
					return entries.NewPostEntriesIDComplainNotFound().WithPayload(err)
				}

				tx.Exec(updateQuery, id, params.Content)
				return entries.NewPostEntriesIDComplainNoContent()
			}

			// send email
			tx.Exec(createQuery, userID.ID, "entry", params.ID, params.Content)
			return entries.NewPostEntriesIDComplainNoContent()
		})
	}
}

func newCommentComplainer(srv *utils.MindwellServer) func(comments.PostCommentsIDComplainParams, *models.UserID) middleware.Responder {
	return func(params comments.PostCommentsIDComplainParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			entryID := tx.QueryInt64("SELECT entry_id FROM comments WHERE id = $1", params.ID)
			allowed := utils.CanViewEntry(tx, userID.ID, entryID)
			if !allowed {
				err := srv.StandardError("no_comment")
				return entries.NewPostEntriesIDComplainNotFound().WithPayload(err)
			}

			id := tx.QueryInt64(selectPrevQuery, userID.ID, params.ID, "comment")
			if tx.Error() != sql.ErrNoRows {
				if tx.Error() != nil {
					err := srv.StandardError("no_comment")
					return entries.NewPostEntriesIDComplainNotFound().WithPayload(err)
				}

				tx.Exec(updateQuery, id, params.Content)
				return entries.NewPostEntriesIDComplainNoContent()
			}

			// send email
			tx.Exec(createQuery, userID.ID, "comment", params.ID, params.Content)
			return entries.NewPostEntriesIDComplainNoContent()
		})
	}
}
