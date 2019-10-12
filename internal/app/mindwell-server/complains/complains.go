package complains

import (
	"database/sql"
	"github.com/go-openapi/runtime/middleware"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/comments"
	"github.com/sevings/mindwell-server/restapi/operations/entries"
	"github.com/sevings/mindwell-server/utils"
)

var errCAY *models.Error

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	errCAY = srv.NewError(&i18n.Message{ID: "complain_against_yourself", Other: "You can't complain against yourself."})

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
    VALUES($1, (SELECT id FROM complain_type WHERE type = $2), $3, $4)    
`

func newEntryComplainer(srv *utils.MindwellServer) func(entries.PostEntriesIDComplainParams, *models.UserID) middleware.Responder {
	return func(params entries.PostEntriesIDComplainParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			allowed := utils.CanViewEntry(tx, userID.ID, params.ID)
			if !allowed {
				err := srv.StandardError("no_entry")
				return entries.NewPostEntriesIDComplainNotFound().WithPayload(err)
			}

			authorID := tx.QueryInt64("SELECT author_id FROM entries WHERE id = $1", params.ID)
			if authorID == userID.ID {
				return entries.NewGetEntriesIDForbidden().WithPayload(errCAY)
			}

			id := tx.QueryInt64(selectPrevQuery, userID.ID, params.ID, "entry")
			if tx.Error() != sql.ErrNoRows {
				if tx.Error() != nil {
					err := srv.StandardError("no_entry")
					return entries.NewPostEntriesIDComplainNotFound().WithPayload(err)
				}

				tx.Exec(updateQuery, id, *params.Content)
				return entries.NewPostEntriesIDComplainNoContent()
			}

			tx.Exec(createQuery, userID.ID, "entry", params.ID, *params.Content)
			srv.Ntf.SendNewEntryComplain(tx, params.ID, userID.Name, *params.Content)
			return entries.NewPostEntriesIDComplainNoContent()
		})
	}
}

func newCommentComplainer(srv *utils.MindwellServer) func(comments.PostCommentsIDComplainParams, *models.UserID) middleware.Responder {
	return func(params comments.PostCommentsIDComplainParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			var entryID, authorID int64
			tx.Query("SELECT entry_id, author_id FROM comments WHERE id = $1", params.ID)
			tx.Scan(&entryID, &authorID)

			allowed := utils.CanViewEntry(tx, userID.ID, entryID)
			if !allowed {
				err := srv.StandardError("no_comment")
				return comments.NewPostCommentsIDComplainNotFound().WithPayload(err)
			}

			if authorID == userID.ID {
				return comments.NewPostCommentsIDComplainForbidden().WithPayload(errCAY)
			}

			id := tx.QueryInt64(selectPrevQuery, userID.ID, params.ID, "comment")
			if tx.Error() != sql.ErrNoRows {
				if tx.Error() != nil {
					err := srv.StandardError("no_comment")
					return comments.NewPostCommentsIDComplainNotFound().WithPayload(err)
				}

				tx.Exec(updateQuery, id, *params.Content)
				return comments.NewPostCommentsIDComplainNoContent()
			}

			tx.Exec(createQuery, userID.ID, "comment", params.ID, *params.Content)
			srv.Ntf.SendNewCommentComplain(tx, params.ID, userID.Name, *params.Content)
			return comments.NewPostCommentsIDComplainNoContent()
		})
	}
}
