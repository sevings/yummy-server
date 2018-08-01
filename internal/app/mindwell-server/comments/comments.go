package comments

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/go-openapi/runtime/middleware"
	"github.com/microcosm-cc/bluemonday"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sevings/mindwell-server/internal/app/mindwell-server/users"
	"github.com/sevings/mindwell-server/restapi/operations/comments"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/utils"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	srv.API.CommentsGetCommentsIDHandler = comments.GetCommentsIDHandlerFunc(newCommentLoader(srv))
	srv.API.CommentsPutCommentsIDHandler = comments.PutCommentsIDHandlerFunc(newCommentEditor(srv))
	srv.API.CommentsDeleteCommentsIDHandler = comments.DeleteCommentsIDHandlerFunc(newCommentDeleter(srv))

	srv.API.CommentsGetEntriesIDCommentsHandler = comments.GetEntriesIDCommentsHandlerFunc(newEntryCommentsLoader(srv))
	srv.API.CommentsPostEntriesIDCommentsHandler = comments.PostEntriesIDCommentsHandlerFunc(newCommentPoster(srv))
}

const commentQuery = `
	SELECT comments.id, entry_id,
		extract(epoch from created_at), content, rating,
		votes.positive, author_id = $1,
		author_id, name, show_name, 
		is_online,
		avatar
	FROM comments
	JOIN short_users ON comments.author_id = short_users.id
	LEFT JOIN (SELECT comment_id, positive FROM comment_votes WHERE user_id = $1) AS votes 
		ON comments.id = votes.comment_id 
`

func commentVote(userID, authorID int64, vote sql.NullBool) string {
	switch {
	case userID == authorID:
		return models.RatingVoteBan
	case !vote.Valid:
		return models.RatingVoteNot
	case vote.Bool:
		return models.RatingVotePos
	default:
		return models.RatingVoteNeg
	}
}

func loadComment(srv *utils.MindwellServer, tx *utils.AutoTx, userID, commentID int64) *models.Comment {
	const q = commentQuery + " WHERE comments.id = $2"

	var vote sql.NullBool
	var avatar string
	comment := models.Comment{
		Author: &models.User{},
		Rating: &models.Rating{
			IsVotable: true,
		},
	}

	tx.Query(q, userID, commentID).Scan(&comment.ID, &comment.EntryID,
		&comment.CreatedAt, &comment.Content, &comment.Rating.Rating,
		&vote, &comment.IsMine,
		&comment.Author.ID, &comment.Author.Name, &comment.Author.ShowName,
		&comment.Author.IsOnline,
		&avatar)

	comment.Rating.ID = comment.ID
	comment.Rating.Vote = commentVote(userID, comment.Author.ID, vote)
	comment.Author.Avatar = srv.NewAvatar(avatar)
	return &comment
}

func newCommentLoader(srv *utils.MindwellServer) func(comments.GetCommentsIDParams, *models.UserID) middleware.Responder {
	return func(params comments.GetCommentsIDParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := uID.ID
			comment := loadComment(srv, tx, userID, params.ID)
			if tx.Error() != nil {
				err := srv.StandardError("no_comment")
				return comments.NewGetCommentsIDNotFound().WithPayload(err)
			}

			canView := utils.CanViewEntry(tx, userID, comment.EntryID)
			if !canView {
				err := srv.StandardError("no_entry")
				return comments.NewGetCommentsIDNotFound().WithPayload(err)
			}

			return comments.NewGetCommentsIDOK().WithPayload(comment)
		})
	}
}

func editComment(tx *utils.AutoTx, commentID int64, content string) {
	content = bluemonday.StrictPolicy().Sanitize(content)

	const q = `
		UPDATE comments
		SET content = $2
		WHERE id = $1`

	tx.Exec(q, commentID, content)
}

func newCommentEditor(srv *utils.MindwellServer) func(comments.PutCommentsIDParams, *models.UserID) middleware.Responder {
	return func(params comments.PutCommentsIDParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := uID.ID
			comment := loadComment(srv, tx, userID, params.ID)
			if tx.Error() != nil {
				err := srv.StandardError("no_comment")
				return comments.NewGetCommentsIDNotFound().WithPayload(err)
			}

			if comment.Author.ID != userID {
				err := srv.NewError(&i18n.Message{ID: "edit_not_your_comment", Other: "You can't edit someone else's comments."})
				return comments.NewGetCommentsIDForbidden().WithPayload(err)
			}

			editComment(tx, params.ID, params.Content)
			if tx.Error() != nil {
				err := srv.NewError(nil)
				return comments.NewGetCommentsIDNotFound().WithPayload(err)
			}

			comment.Content = params.Content
			return comments.NewPutCommentsIDOK().WithPayload(comment)
		})
	}
}

func commentAuthor(tx *utils.AutoTx, commentID int64) int64 {
	const q = `
		SELECT author_id
		FROM comments
		WHERE id = $1`

	var authorID int64
	tx.Query(q, commentID).Scan(&authorID)

	return authorID
}

func deleteComment(tx *utils.AutoTx, commentID int64) {
	const q = `
		DELETE FROM comments
		WHERE id = $1`

	tx.Exec(q, commentID)
}

func newCommentDeleter(srv *utils.MindwellServer) func(comments.DeleteCommentsIDParams, *models.UserID) middleware.Responder {
	return func(params comments.DeleteCommentsIDParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := uID.ID
			authorID := commentAuthor(tx, params.ID)
			if tx.Error() != nil {
				err := srv.NewError(nil)
				return comments.NewDeleteCommentsIDNotFound().WithPayload(err)
			}
			if authorID != userID {
				err := srv.NewError(&i18n.Message{ID: "delete_not_your_comment", Other: "You can't delete someone else's comments."})
				return comments.NewDeleteCommentsIDForbidden().WithPayload(err)
			}

			deleteComment(tx, params.ID)
			if tx.Error() != nil {
				err := srv.NewError(nil)
				return comments.NewDeleteCommentsIDNotFound().WithPayload(err)
			}

			return comments.NewDeleteCommentsIDOK()
		})
	}
}

// LoadEntryComments loads comments for entry.
func LoadEntryComments(srv *utils.MindwellServer, tx *utils.AutoTx, userID, entryID, limit int64, afterS, beforeS string) *models.CommentList {
	var list []*models.Comment

	before, err := strconv.ParseInt(beforeS, 10, 64)
	if len(beforeS) > 0 && err != nil {
		log.Printf("error parse before: %s", beforeS)
	}

	after, err := strconv.ParseInt(afterS, 10, 64)
	if len(afterS) > 0 && err != nil {
		log.Printf("error parse after: %s", afterS)
	}

	if after > 0 {
		const q = commentQuery + `
			WHERE entry_id = $2 AND comments.id > $3
			ORDER BY comments.id ASC
			LIMIT $4`

		tx.Query(q, userID, entryID, after, limit)
	} else if before > 0 {
		const q = commentQuery + `
			WHERE entry_id = $2 AND comments.id < $3
			ORDER BY comments.id DESC
			LIMIT $4`

		tx.Query(q, userID, entryID, before, limit)
	} else {
		const q = commentQuery + `
			WHERE entry_id = $2
			ORDER BY comments.id DESC
			LIMIT $3`

		tx.Query(q, userID, entryID, limit)
	}

	for {
		comment := models.Comment{
			Author: &models.User{},
			Rating: &models.Rating{
				IsVotable: true,
			},
		}
		var vote sql.NullBool
		var avatar string
		ok := tx.Scan(&comment.ID, &comment.EntryID,
			&comment.CreatedAt, &comment.Content, &comment.Rating.Rating,
			&vote, &comment.IsMine,
			&comment.Author.ID, &comment.Author.Name, &comment.Author.ShowName,
			&comment.Author.IsOnline,
			&avatar)
		if !ok {
			break
		}

		comment.Rating.ID = comment.ID
		comment.Rating.Vote = commentVote(userID, comment.Author.ID, vote)
		comment.Author.Avatar = srv.NewAvatar(avatar)
		list = append(list, &comment)
	}

	if after <= 0 {
		for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
			list[i], list[j] = list[j], list[i]
		}
	}

	comments := &models.CommentList{Data: list}

	if len(list) > 0 {
		nextBefore := list[0].ID
		var hasBefore bool
		tx.Query("SELECT EXISTS(SELECT 1 FROM comments WHERE entry_id = $1 AND comments.id < $2)", entryID, nextBefore)
		tx.Scan(&hasBefore)
		if hasBefore {
			comments.NextBefore = strconv.FormatInt(nextBefore, 10)
			comments.HasBefore = hasBefore
		}

		nextAfter := list[len(list)-1].ID
		comments.NextAfter = strconv.FormatInt(nextAfter, 10)
		tx.Query("SELECT EXISTS(SELECT 1 FROM comments WHERE entry_id = $1 AND comments.id > $2)", entryID, nextAfter)
		tx.Scan(&comments.HasAfter)
	}

	return comments
}

func newEntryCommentsLoader(srv *utils.MindwellServer) func(comments.GetEntriesIDCommentsParams, *models.UserID) middleware.Responder {
	return func(params comments.GetEntriesIDCommentsParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := uID.ID
			canView := utils.CanViewEntry(tx, userID, params.ID)
			if !canView {
				err := srv.StandardError("no_entry")
				return comments.NewGetEntriesIDCommentsNotFound().WithPayload(err)
			}

			data := LoadEntryComments(srv, tx, userID, params.ID, *params.Limit, *params.After, *params.Before)
			if tx.Error() != nil && tx.Error() != sql.ErrNoRows {
				err := srv.NewError(nil)
				return comments.NewGetEntriesIDCommentsNotFound().WithPayload(err)
			}

			return comments.NewGetEntriesIDCommentsOK().WithPayload(data)
		})
	}
}

func postComment(tx *utils.AutoTx, author *models.User, entryID int64, content string) *models.Comment {
	content = bluemonday.StrictPolicy().Sanitize(content)

	const q = `
		INSERT INTO comments (author_id, entry_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, extract(epoch from created_at)`

	comment := models.Comment{
		Author:  author,
		Content: content,
		EntryID: entryID,
		IsMine:  true,
	}

	tx.Query(q, author.ID, entryID, content).Scan(&comment.ID, &comment.CreatedAt)

	return &comment
}

func notifyNewComment(srv *utils.MindwellServer, tx *utils.AutoTx, cmt *models.Comment) {
	const titleQ = "SELECT title FROM entries WHERE id = $1"
	var title string
	tx.Query(titleQ, cmt.EntryID).Scan(&title)

	title, _ = utils.CutText(title, "%.60s", 60)

	const fromQ = `
		SELECT gender.type 
		FROM users, gender 
		WHERE users.id = $1 AND users.gender = gender.id
	`

	var fromGender string
	tx.Query(fromQ, cmt.Author.ID).Scan(&fromGender)

	const toQ = `
		SELECT show_name, email
		FROM users, watching 
		WHERE watching.entry_id = $1 AND watching.user_id = users.id 
			AND users.id <> $2 AND users.verified AND users.email_comments`

	tx.Query(toQ, cmt.EntryID, cmt.Author.ID)

	var toShowName, email string
	for tx.Scan(&toShowName, &email) {
		srv.Mail.SendNewComment(email, fromGender, toShowName, title, cmt)
	}
}

func newCommentPoster(srv *utils.MindwellServer) func(comments.PostEntriesIDCommentsParams, *models.UserID) middleware.Responder {
	return func(params comments.PostEntriesIDCommentsParams, uID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			userID := uID.ID
			canView := utils.CanViewEntry(tx, userID, params.ID)
			if !canView {
				err := srv.StandardError("no_entry")
				return comments.NewPostEntriesIDCommentsNotFound().WithPayload(err)
			}

			user := users.LoadUserByID(srv, tx, userID)
			comment := postComment(tx, user, params.ID, params.Content)
			if tx.Error() != nil {
				err := srv.NewError(nil)
				return comments.NewPostEntriesIDCommentsNotFound().WithPayload(err)
			}

			notifyNewComment(srv, tx, comment)

			return comments.NewPostEntriesIDCommentsCreated().WithPayload(comment)
		})
	}
}
