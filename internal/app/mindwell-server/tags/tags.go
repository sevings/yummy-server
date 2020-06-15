package tags

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/leporo/sqlf"
	entriesImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/entries"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/entries"
	"github.com/sevings/mindwell-server/restapi/operations/me"
	"github.com/sevings/mindwell-server/restapi/operations/users"
	"github.com/sevings/mindwell-server/utils"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	srv.API.MeGetMeTagsHandler = me.GetMeTagsHandlerFunc(newMyTagsLoader(srv))
	srv.API.UsersGetUsersNameTagsHandler = users.GetUsersNameTagsHandlerFunc(newUserTagsLoader(srv))
	srv.API.EntriesGetEntriesTagsHandler = entries.GetEntriesTagsHandlerFunc(newLiveTagsLoader(srv))
}

func tagsQuery(limit int64) *sqlf.Stmt {
	return sqlf.Select("tags.tag, count(*) AS cnt").
		From("entries").
		Join("entry_tags", "entries.id = entry_tags.entry_id").
		Join("tags", "entry_tags.tag_id = tags.id").
		GroupBy("tags.id").
		Limit(limit)
}

func myTagsQuery(userID, limit int64) *sqlf.Stmt {
	return tagsQuery(limit).
		Where("entries.author_id = ?", userID).
		OrderBy("max(entries.created_at) DESC").
		OrderBy("cnt DESC")
}

func tlogTagsQuery(userID, limit int64, tlog string) *sqlf.Stmt {
	q := tagsQuery(limit).
		Join("entry_privacy", "entries.visible_for = entry_privacy.id").
		Where("entries.author_id = (SELECT id FROM users WHERE lower(name) = lower(?))", tlog).
		OrderBy("max(entries.created_at) DESC").
		OrderBy("cnt DESC")
	return entriesImpl.AddEntryOpenQuery(q, userID)
}

func liveTagsQuery(userID, limit int64) *sqlf.Stmt {
	q := tagsQuery(limit).
		Join("users", "entries.author_id = users.id").
		Join("entry_privacy", "entries.visible_for = entry_privacy.id").
		Join("user_privacy", "users.privacy = user_privacy.id").
		Where("age(entries.created_at) <= interval '1 month'").
		OrderBy("cnt DESC").
		OrderBy("max(entries.created_at) DESC")
	return entriesImpl.AddLiveInvitedQuery(q, userID, "")
}

func loadTags(tx *utils.AutoTx) *models.TagList {
	tags := &models.TagList{}

	var tag string
	var count int64
	for tx.Scan(&tag, &count) {
		item := &models.TagListDataItems0{
			Count: count,
			Tag:   tag,
		}

		tags.Data = append(tags.Data, item)
	}

	return tags
}

func newMyTagsLoader(srv *utils.MindwellServer) func(me.GetMeTagsParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeTagsParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			query := myTagsQuery(userID.ID, *params.Limit)

			tx.QueryStmt(query)
			tags := loadTags(tx)

			return me.NewGetMeTagsOK().WithPayload(tags)
		})
	}
}

func newUserTagsLoader(srv *utils.MindwellServer) func(users.GetUsersNameTagsParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersNameTagsParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			var query *sqlf.Stmt

			if userID.Name == params.Name {
				query = myTagsQuery(userID.ID, *params.Limit)
			} else {
				if !utils.IsOpenForMe(tx, userID, params.Name) {
					err := srv.StandardError("no_tlog")
					return users.NewGetUsersNameTagsNotFound().WithPayload(err)
				}

				query = tlogTagsQuery(userID.ID, *params.Limit, params.Name)
			}

			tx.QueryStmt(query)
			tags := loadTags(tx)

			return users.NewGetUsersNameTagsOK().WithPayload(tags)
		})
	}
}

func newLiveTagsLoader(srv *utils.MindwellServer) func(entries.GetEntriesTagsParams, *models.UserID) middleware.Responder {
	return func(params entries.GetEntriesTagsParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			query := liveTagsQuery(userID.ID, *params.Limit)

			tx.QueryStmt(query)
			tags := loadTags(tx)

			return entries.NewGetEntriesTagsOK().WithPayload(tags)
		})
	}
}
