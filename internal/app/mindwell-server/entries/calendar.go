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

func baseCalendarQuery(start, end int64) *sqlf.Stmt {
	return sqlf.Select("entries.id, extract(epoch from entries.created_at) as created_at").
		Select("entries.title, entries.edit_content").
		From("entries").
		Where("entries.created_at >= to_timestamp(?)", start).
		Where("entries.created_at < to_timestamp(?)", end).
		OrderBy("entries.created_at ASC")
}

func myCalendarQuery(userID int64, start, end int64) *sqlf.Stmt {
	return baseCalendarQuery(start, end).
		Where("author_id = ?", userID)
}

func tlogCalendarQuery(userID, start, end int64, tlog string) *sqlf.Stmt {
	q := baseCalendarQuery(start, end).
		Join("users", "author_id = users.id").
		Where("lower(users.name) = lower(?)", tlog).
		Join("entry_privacy", "entries.visible_for = entry_privacy.id")
	return AddEntryOpenQuery(q, userID)
}

func getPeriod(createdAt float64, start, end int64) (int64, int64) {
	const maxDuration = 60*60*24*7*6 + 1 // six weeks
	minDate := int64(createdAt)
	maxDate := time.Now().Unix() + 1

	if start <= 0 && end <= 0 {
		start = maxDate - maxDuration
		end = maxDate
	} else if start <= 0 {
		start = end - maxDuration
	} else {
		end = start + maxDuration
	}

	if start < minDate {
		start = minDate
	}
	if end > maxDate {
		end = maxDate
	}

	if end-start > maxDuration {
		end = start + maxDuration
	}

	return start, end
}

func loadEmptyCalendar(tx *utils.AutoTx, q *sqlf.Stmt) *models.Calendar {
	cal := &models.Calendar{
		End: float64(time.Now().Unix()),
	}

	tx.QueryStmt(q)
	tx.Scan(&cal.Start)

	return cal
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
}

func loadTlogCalendar(tx *utils.AutoTx, userID *models.UserID, tlog string, start, end int64) *models.Calendar {
	if userID.Name == tlog {
		return loadMyCalendar(tx, userID, start, end)
	}

	createdAtQuery := sqlf.Select("extract(epoch FROM created_at)").
		From("users").
		Where("lower(name) = lower(?)", tlog)

	cal := loadEmptyCalendar(tx, createdAtQuery)
	start, end = getPeriod(cal.Start, start, end)
	if start >= end {
		return cal
	}

	q := tlogCalendarQuery(userID.ID, start, end, tlog)
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

			feed := loadTlogCalendar(tx, userID, params.Name, *params.Start, *params.End)
			return users.NewGetUsersNameCalendarOK().WithPayload(feed)
		})
	}
}

func loadMyCalendar(tx *utils.AutoTx, userID *models.UserID, start, end int64) *models.Calendar {
	createdAtQuery := sqlf.Select("extract(epoch FROM created_at)").
		From("users").
		Where("id = ?", userID.ID)

	cal := loadEmptyCalendar(tx, createdAtQuery)
	start, end = getPeriod(cal.Start, start, end)
	if start >= end {
		return cal
	}

	q := myCalendarQuery(userID.ID, start, end)
	tx.QueryStmt(q)
	loadCalendar(tx, cal)

	return cal
}

func newMyCalendarLoader(srv *utils.MindwellServer) func(me.GetMeCalendarParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeCalendarParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			feed := loadMyCalendar(tx, userID, *params.Start, *params.End)

			if tx.Error() != nil && tx.Error() != sql.ErrNoRows {
				err := srv.NewError(nil)
				return me.NewPutMeCoverBadRequest().WithPayload(err)
			}

			return me.NewGetMeCalendarOK().WithPayload(feed)
		})
	}
}
