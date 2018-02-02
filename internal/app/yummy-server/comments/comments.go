package comments

import (
	"database/sql"
	"log"

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
	case userID <= 0 || userID == authorID:
		return ""
	case !vote.Valid:
		return models.CommentVoteNot
	case vote.Bool:
		return models.CommentVotePos
	default:
		return models.CommentVoteNeg
	}
}

func loadComment(tx utils.AutoTx, userID, commentID int64) (*models.Comment, error) {
	const q = commentQuery + " WHERE comments.id = $2"

	var vote sql.NullBool
	comment := models.Comment{
		Author: &models.User{},
	}

	err := tx.QueryRow(q, userID, commentID).Scan(&comment.ID, &comment.EntryID,
		&comment.CreatedAt, &comment.Content, &comment.Rating,
		&vote,
		&comment.Author.ID, &comment.Author.Name, &comment.Author.ShowName,
		&comment.Author.IsOnline,
		&comment.Author.Avatar)

	comment.Vote = commentVote(userID, comment.Author.ID, vote)
	return &comment, err
}

func newCommentLoader(db *sql.DB) func(comments.GetCommentsIDParams, *models.UserID) middleware.Responder {
	return func(params comments.GetCommentsIDParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx utils.AutoTx) (middleware.Responder, bool) {
			userID := int64(*uID)
			comment, err := loadComment(tx, userID, params.ID)
			if err != nil {
				if err != sql.ErrNoRows {
					log.Print(err)
				}

				return comments.NewGetCommentsIDNotFound(), false
			}

			canView := utils.CanViewEntry(tx, userID, comment.EntryID)
			if !canView {
				return comments.NewGetCommentsIDNotFound(), false
			}

			return comments.NewGetCommentsIDOK().WithPayload(comment), true
		})
	}
}

func editComment(tx utils.AutoTx, commentID int64, content string) error {
	const q = `
		UPDATE comments
		SET content = $2
		WHERE id = $1`

	_, err := tx.Exec(q, commentID, content)
	return err
}

func newCommentEditor(db *sql.DB) func(comments.PutCommentsIDParams, *models.UserID) middleware.Responder {
	return func(params comments.PutCommentsIDParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx utils.AutoTx) (middleware.Responder, bool) {
			userID := int64(*uID)
			comment, err := loadComment(tx, userID, params.ID)
			if err != nil {
				if err != sql.ErrNoRows {
					log.Print(err)
				}

				return comments.NewGetCommentsIDNotFound(), false
			}

			if comment.Author.ID != userID {
				return comments.NewGetCommentsIDForbidden(), false
			}

			err = editComment(tx, params.ID, params.Content)
			if err != nil {
				log.Print(err)
				return comments.NewGetCommentsIDNotFound(), false
			}

			comment.Content = params.Content
			return comments.NewPutCommentsIDOK().WithPayload(comment), true
		})
	}
}

func commentAuthor(tx utils.AutoTx, commentID int64) (int64, bool) {
	const q = `
		SELECT author_id
		FROM comments
		WHERE id = $1`

	var authorID int64
	err := tx.QueryRow(q, commentID).Scan(&authorID)
	if err == nil {
		return authorID, true
	}

	if err != sql.ErrNoRows {
		log.Print(err)
	}

	return 0, false
}

func deleteComment(tx utils.AutoTx, commentID int64) error {
	const q = `
		DELETE FROM comments
		WHERE id = $1`

	_, err := tx.Exec(q, commentID)
	return err
}

func newCommentDeleter(db *sql.DB) func(comments.DeleteCommentsIDParams, *models.UserID) middleware.Responder {
	return func(params comments.DeleteCommentsIDParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx utils.AutoTx) (middleware.Responder, bool) {
			userID := int64(*uID)
			authorID, found := commentAuthor(tx, params.ID)
			if !found {
				return comments.NewDeleteCommentsIDNotFound(), false
			}
			if authorID != userID {
				return comments.NewDeleteCommentsIDForbidden(), false
			}

			err := deleteComment(tx, params.ID)
			if err != nil {
				log.Print(err)
				return comments.NewDeleteCommentsIDNotFound(), false
			}

			return comments.NewDeleteCommentsIDOK(), true
		})
	}
}

// LoadEntryComments loads comments for entry.
func LoadEntryComments(tx utils.AutoTx, userID, entryID, limit, offset int64) ([]*models.Comment, error) {
	const q = commentQuery + `
		WHERE entry_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	var list []*models.Comment
	rows, err := tx.Query(q, userID, entryID, limit, offset)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		comment := models.Comment{Author: &models.User{}}
		var vote sql.NullBool
		rows.Scan(&comment.ID, &comment.EntryID,
			&comment.CreatedAt, &comment.Content, &comment.Rating,
			&vote,
			&comment.Author.ID, &comment.Author.Name, &comment.Author.ShowName,
			&comment.Author.IsOnline,
			&comment.Author.Avatar)

		comment.Vote = commentVote(userID, comment.Author.ID, vote)
		list = append(list, &comment)
	}

	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}

	return list, rows.Err()
}

func newEntryCommentsLoader(db *sql.DB) func(comments.GetEntriesIDCommentsParams, *models.UserID) middleware.Responder {
	return func(params comments.GetEntriesIDCommentsParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx utils.AutoTx) (middleware.Responder, bool) {
			userID := int64(*uID)
			canView := utils.CanViewEntry(tx, userID, params.ID)
			if !canView {
				return comments.NewGetEntriesIDCommentsNotFound(), false
			}

			list, err := LoadEntryComments(tx, userID, params.ID, *params.Limit, *params.Skip)
			if err != nil {
				log.Print(err)
				return comments.NewGetEntriesIDCommentsNotFound(), false
			}

			res := models.CommentList{Comments: list}
			return comments.NewGetEntriesIDCommentsOK().WithPayload(&res), true
		})
	}
}

func postComment(tx utils.AutoTx, author *models.User, entryID int64, content string) (*models.Comment, bool) {
	const q = `
		INSERT INTO comments (author_id, entry_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	comment := models.Comment{
		Author:  author,
		Content: content,
		EntryID: entryID,
	}

	err := tx.QueryRow(q, author.ID, entryID, content).Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		log.Print(err)
		return nil, false
	}

	return &comment, true
}

func newCommentPoster(db *sql.DB) func(comments.PostEntriesIDCommentsParams, *models.UserID) middleware.Responder {
	return func(params comments.PostEntriesIDCommentsParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx utils.AutoTx) (middleware.Responder, bool) {
			userID := int64(*uID)
			canView := utils.CanViewEntry(tx, userID, params.ID)
			if !canView {
				return comments.NewPostEntriesIDCommentsNotFound(), false
			}

			user, _ := users.LoadUserByID(tx, userID)
			comment, ok := postComment(tx, user, params.ID, params.Content)
			if !ok {
				return comments.NewPostEntriesIDCommentsNotFound(), false
			}

			return comments.NewPostEntriesIDCommentsOK().WithPayload(comment), true
		})
	}
}
