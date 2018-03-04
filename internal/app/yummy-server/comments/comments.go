package comments

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/internal/app/yummy-server/users"
	"github.com/sevings/yummy-server/restapi/operations/comments"

	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.CommentsGetCommentsIDHandler = comments.GetCommentsIDHandlerFunc(newCommentLoader(db))
	api.CommentsPutCommentsIDHandler = comments.PutCommentsIDHandlerFunc(newCommentEditor(db))
	api.CommentsDeleteCommentsIDHandler = comments.DeleteCommentsIDHandlerFunc(newCommentDeleter(db))

	api.CommentsGetEntriesIDCommentsHandler = comments.GetEntriesIDCommentsHandlerFunc(newEntryCommentsLoader(db))
	api.CommentsPostEntriesIDCommentsHandler = comments.PostEntriesIDCommentsHandlerFunc(newCommentPoster(db))
}

const commentQuery = `
	SELECT comments.id, entry_id,
		created_at, content, rating,
		votes.positive AS vote,
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
		return models.CommentVoteBan
	case !vote.Valid:
		return models.CommentVoteNot
	case vote.Bool:
		return models.CommentVotePos
	default:
		return models.CommentVoteNeg
	}
}

func loadComment(tx *utils.AutoTx, userID, commentID int64) *models.Comment {
	const q = commentQuery + " WHERE comments.id = $2"

	var vote sql.NullBool
	comment := models.Comment{
		Author: &models.User{},
	}

	tx.Query(q, userID, commentID).Scan(&comment.ID, &comment.EntryID,
		&comment.CreatedAt, &comment.Content, &comment.Rating,
		&vote,
		&comment.Author.ID, &comment.Author.Name, &comment.Author.ShowName,
		&comment.Author.IsOnline,
		&comment.Author.Avatar)

	comment.Vote = commentVote(userID, comment.Author.ID, vote)
	return &comment
}

func newCommentLoader(db *sql.DB) func(comments.GetCommentsIDParams, *models.UserID) middleware.Responder {
	return func(params comments.GetCommentsIDParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			comment := loadComment(tx, userID, params.ID)
			if tx.Error() != nil {
				return comments.NewGetCommentsIDNotFound()
			}

			canView := utils.CanViewEntry(tx, userID, comment.EntryID)
			if !canView {
				return comments.NewGetCommentsIDNotFound()
			}

			return comments.NewGetCommentsIDOK().WithPayload(comment)
		})
	}
}

func editComment(tx *utils.AutoTx, commentID int64, content string) {
	const q = `
		UPDATE comments
		SET content = $2
		WHERE id = $1`

	tx.Exec(q, commentID, content)
}

func newCommentEditor(db *sql.DB) func(comments.PutCommentsIDParams, *models.UserID) middleware.Responder {
	return func(params comments.PutCommentsIDParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			comment := loadComment(tx, userID, params.ID)
			if tx.Error() != nil {
				return comments.NewGetCommentsIDNotFound()
			}

			if comment.Author.ID != userID {
				return comments.NewGetCommentsIDForbidden()
			}

			editComment(tx, params.ID, params.Content)
			if tx.Error() != nil {
				return comments.NewGetCommentsIDNotFound()
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

func newCommentDeleter(db *sql.DB) func(comments.DeleteCommentsIDParams, *models.UserID) middleware.Responder {
	return func(params comments.DeleteCommentsIDParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			authorID := commentAuthor(tx, params.ID)
			if tx.Error() != nil {
				return comments.NewDeleteCommentsIDNotFound()
			}
			if authorID != userID {
				return comments.NewDeleteCommentsIDForbidden()
			}

			deleteComment(tx, params.ID)
			if tx.Error() != nil {
				return comments.NewDeleteCommentsIDNotFound()
			}

			return comments.NewDeleteCommentsIDOK()
		})
	}
}

// LoadEntryComments loads comments for entry.
func LoadEntryComments(tx *utils.AutoTx, userID, entryID, limit, offset int64) []*models.Comment {
	const q = commentQuery + `
		WHERE entry_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	var list []*models.Comment
	tx.Query(q, userID, entryID, limit, offset)

	for {
		comment := models.Comment{Author: &models.User{}}
		var vote sql.NullBool
		ok := tx.Scan(&comment.ID, &comment.EntryID,
			&comment.CreatedAt, &comment.Content, &comment.Rating,
			&vote,
			&comment.Author.ID, &comment.Author.Name, &comment.Author.ShowName,
			&comment.Author.IsOnline,
			&comment.Author.Avatar)
		if !ok {
			break
		}

		comment.Vote = commentVote(userID, comment.Author.ID, vote)
		list = append(list, &comment)
	}

	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}

	return list
}

func newEntryCommentsLoader(db *sql.DB) func(comments.GetEntriesIDCommentsParams, *models.UserID) middleware.Responder {
	return func(params comments.GetEntriesIDCommentsParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			canView := utils.CanViewEntry(tx, userID, params.ID)
			if !canView {
				return comments.NewGetEntriesIDCommentsNotFound()
			}

			list := LoadEntryComments(tx, userID, params.ID, *params.Limit, *params.Skip)
			if tx.Error() != nil {
				return comments.NewGetEntriesIDCommentsNotFound()
			}

			res := models.CommentList{Comments: list}
			return comments.NewGetEntriesIDCommentsOK().WithPayload(&res)
		})
	}
}

func postComment(tx *utils.AutoTx, author *models.User, entryID int64, content string) *models.Comment {
	const q = `
		INSERT INTO comments (author_id, entry_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	comment := models.Comment{
		Author:  author,
		Content: content,
		EntryID: entryID,
	}

	tx.Query(q, author.ID, entryID, content).Scan(&comment.ID, &comment.CreatedAt)

	return &comment
}

func newCommentPoster(db *sql.DB) func(comments.PostEntriesIDCommentsParams, *models.UserID) middleware.Responder {
	return func(params comments.PostEntriesIDCommentsParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			userID := int64(*uID)
			canView := utils.CanViewEntry(tx, userID, params.ID)
			if !canView {
				return comments.NewPostEntriesIDCommentsNotFound()
			}

			user := users.LoadUserByID(tx, userID)
			comment := postComment(tx, user, params.ID, params.Content)
			if tx.Error() != nil {
				return comments.NewPostEntriesIDCommentsNotFound()
			}

			return comments.NewPostEntriesIDCommentsCreated().WithPayload(comment)
		})
	}
}
