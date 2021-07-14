package entries

import (
	"database/sql"
	"github.com/go-openapi/runtime/middleware"
	"github.com/leporo/sqlf"
	"github.com/microcosm-cc/bluemonday"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/me"
	"github.com/sevings/mindwell-server/restapi/operations/users"
	"github.com/sevings/mindwell-server/utils"
	"strings"
	"time"
)

func baseCalendarQuery(cal *models.Calendar) *sqlf.Stmt {
	q := sqlf.Select("entries.id, extract(epoch from entries.created_at) as created_at").
		Select("entries.title, entries.edit_content").
		From("entries")

	if cal.Start > 0 {
		q.Where("entries.created_at >= to_timestamp(?)", cal.Start)
	}

	if cal.End > 0 {
		q.Where("entries.created_at < to_timestamp(?)", cal.End)
	}

	if cal.Start > 0 {
		q.OrderBy("entries.created_at ASC")
	} else {
		q.OrderBy("entries.created_at DESC")
	}

	return q.Limit(cal.Limit)
}

func myCalendarQuery(userID int64, cal *models.Calendar) *sqlf.Stmt {
	return baseCalendarQuery(cal).
		Where("author_id = ?", userID)
}

func tlogCalendarQuery(userID int64, tlog string, cal *models.Calendar) *sqlf.Stmt {
	q := baseCalendarQuery(cal).
		Join("users", "author_id = users.id").
		Where("lower(users.name) = lower(?)", tlog).
		Join("entry_privacy", "entries.visible_for = entry_privacy.id")
	return AddEntryOpenQuery(q, userID)
}

func loadEmptyCalendar(tx *utils.AutoTx, q *sqlf.Stmt, start, end, limit int64) *models.Calendar {
	var createdAt float64
	tx.QueryStmt(q)
	tx.Scan(&createdAt)

	const maxDuration = 60*60*24*7*6 + 1 // six weeks
	minDate := int64(createdAt)
	maxDate := time.Now().Unix() + 1

	if start > 0 && start < minDate {
		start = minDate
	}
	if end > 0 && end > maxDate {
		end = maxDate
	}

	if start > 0 && end > 0 && end-start > maxDuration {
		end = start + maxDuration
	}

	return &models.Calendar{
		Start: start,
		End:   end,
		Limit: limit,
	}
}

func loadCalendar(tx *utils.AutoTx, cal *models.Calendar) {
	for {
		var title, content string
		var entry models.CalendarEntriesItems0
		ok := tx.Scan(&entry.ID, &entry.CreatedAt,
			&title, &content)
		if !ok {
			break
		}

		title = strings.TrimSpace(title)
		if title != "" {
			title = bluemonday.StrictPolicy().Sanitize(title)
			entry.Title, _ = utils.CutText(title, 100)
		} else {
			content = strings.TrimSpace(content)
			content = md.RenderToString([]byte(content))
			content = utils.RemoveHTML(content)
			entry.Title, _ = utils.CutHtml(content, 1, 100)
		}

		cal.Entries = append(cal.Entries, &entry)
	}

	if cal.Start > 0 {
		list := cal.Entries
		for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
			list[i], list[j] = list[j], list[i]
		}
	}
}

func loadTlogCalendar(tx *utils.AutoTx, userID *models.UserID, tlog string, start, end, limit int64) *models.Calendar {
	if userID.Name == tlog {
		return loadMyCalendar(tx, userID, start, end, limit)
	}

	createdAtQuery := sqlf.Select("extract(epoch FROM created_at)").
		From("users").
		Where("lower(name) = lower(?)", tlog)

	cal := loadEmptyCalendar(tx, createdAtQuery, start, end, limit)
	if cal.End > 0 && cal.Start >= cal.End {
		return cal
	}

	q := tlogCalendarQuery(userID.ID, tlog, cal)
	tx.QueryStmt(q)
	loadCalendar(tx, cal)

	return cal
}

func newTlogCalendarLoader(srv *utils.MindwellServer) func(users.GetUsersNameCalendarParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersNameCalendarParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			canView := utils.IsOpenForMe(tx, userID, params.Name)
			if !canView {
				err := srv.StandardError("no_tlog")
				return users.NewGetUsersNameCalendarNotFound().WithPayload(err)
			}

			feed := loadTlogCalendar(tx, userID, params.Name, *params.Start, *params.End, *params.Limit)
			return users.NewGetUsersNameCalendarOK().WithPayload(feed)
		})
	}
}

func loadMyCalendar(tx *utils.AutoTx, userID *models.UserID, start, end, limit int64) *models.Calendar {
	createdAtQuery := sqlf.Select("extract(epoch FROM created_at)").
		From("users").
		Where("id = ?", userID.ID)

	cal := loadEmptyCalendar(tx, createdAtQuery, start, end, limit)
	if cal.End > 0 && cal.Start >= cal.End {
		return cal
	}

	q := myCalendarQuery(userID.ID, cal)
	tx.QueryStmt(q)
	loadCalendar(tx, cal)

	return cal
}

func newMyCalendarLoader(srv *utils.MindwellServer) func(me.GetMeCalendarParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeCalendarParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadMyCalendar(tx, userID, *params.Start, *params.End, *params.Limit)

			if tx.Error() != nil && tx.Error() != sql.ErrNoRows {
				err := srv.NewError(nil)
				return me.NewPutMeCoverBadRequest().WithPayload(err)
			}

			return me.NewGetMeCalendarOK().WithPayload(feed)
		})
	}
}
