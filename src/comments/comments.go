package comments

import (
	"database/sql"
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/restapi/operations/comments"

	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations"
	"github.com/sevings/yummy-server/src/entries"
	"github.com/sevings/yummy-server/src/users"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.CommentsGetCommentsIDHandler = comments.GetCommentsIDHandlerFunc(newCommentLoader(db))
	api.CommentsPutCommentsIDHandler = comments.PutCommentsIDHandlerFunc(newCommentEditor(db))
	api.CommentsDeleteCommentsIDHandler = comments.DeleteCommentsIDHandlerFunc(newCommentDeleter(db))
}

func loadComment(tx *sql.Tx, userID, commentID int64) (*models.Comment, error) {
	const q = `
		SELECT entry_id,
			created_at, content, rating,
			votes.positive AS vote,
			author_id, name, show_name, 
			is_online,
			name_color, avatar_color, avatar
		FROM comments
		LEFT JOIN (SELECT comment_id, positive FROM comment_votes WHERE user_id = $1) AS votes 
			ON comments.id = votes.comment_id
		WHERE comments.id = $2 AND comments.author_id = short_users.id`

	var vote sql.NullBool
	comment := models.Comment {
		ID: commentID,
		Author: &models.User{}
	}
	
	err := tx.QueryRow(q, userID, commentID).Scan(&comment.EntryID,
		&comment.CreatedAt, &comment.Content, &comment.Rating,
		&vote,
		&comment.Author.ID, &comment.Author.Name, &comment.Author.ShowName,
		&comment.Author.IsOnline, 
		&comment.Author.NameColor, &comment.Author.AvatarColor, &comment.Author.Avatar)
	
	if userID > 0 {
		switch {
		case !vote.Valid:
			comment.Vote = "not"
		case vote.Bool:
			comment.Vote = "pos"
		default:
			comment.Vote = "neg"
		}		
	}

	return &comment, err
}

func newCommentLoader(db *sql.DB) func(comments.GetCommentsIDParams) middleware.Responder {
	return func(params comments.GetCommentsIDParams) middleware.Responder {
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		
		userID, _ := users.FindAuthUser(tx, params.XUserKey)
		comment, err := loadComment(tx, userID, params.ID)
		if err != nil {
			if err != sql.ErrNoRows {
				log.print(err)
			}

			tx.Rollback()
			return comments.NewGetCommentsIDNotFound()
		}

		canView := entries.CanViewEntry(tx, userID, params.ID)
		if !canView {
			tx.Rollback()
			return comments.NewGetCommentsIDNotFound()
		}

		tx.Commit()
		return comments.NewGetCommentsIDOK().WithPayload(comment)
	}
}

func editComment(tx *sql.Tx, commentID int64, content string) error {
	const q = `
		UPDATE comments
		SET content = $2
		WHERE id = $1`

	_, err := tx.Exec(q, commentID, content)
	return err
}

func newCommentEditor(db *sql.DB) func(comments.PutCommentsIDParams) middleware.Responder {
	return func(params comments.PutCommentsIDParams) middleware.Responder {
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		
		userID, found := users.FindAuthUser(tx, &params.XUserKey)
		if !found {
			tx.Rollback()
			return comments.NewGetCommentsIDForbidden()
		}

		comment, err := loadComment(tx, userID, params.ID)
		if err != nil {
			if err != sql.ErrNoRows {
				log.print(err)
			}

			tx.Rollback()
			return comments.NewGetCommentsIDNotFound()
		}

		if comment.Author.ID != userID {
			tx.Rollback()
			return comments.NewGetCommentsIDForbidden()
		}

		err = editComment(tx, params.ID, *params.Content)
		if err != nil {
			log.Print(err)
			tx.Rollback()
			return comments.NewGetCommentsIDNotFound()
		}

		tx.Commit()
		comment.Content = *params.Content
		return comments.NewPutCommentsIDOK().WithPayload(comment)
	}
}

func commentAuthor(tx *sql.Tx, commentID int64) (int64, bool) {
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

func deleteComment(tx *sql.Tx, commentID int64) error {
	const q = `
		DELETE FROM comments
		WHERE id = $1`

	_, err := tx.Exec(q, commentID)
	return err
}

func newCommentDeleter(db *sql.DB) func(comments.DeleteCommentsIDParams) middleware.Responder {
	return func(params comments.DeleteCommentsIDParams) middleware.Responder {
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		
		userID, found := users.FindAuthUser(tx, &params.XUserKey)
		if !found {
			tx.Rollback()
			return comments.NewDeleteCommentsIDForbidden()
		}

		authorID, found := commentAuthor(tx, params.ID)
		if !found {
			tx.Rollback()
			return comments.NewDeleteCommentsIDNotFound()			
		}
		if authorID != userID {
			tx.Rollback()
			return comments.NewDeleteCommentsIDForbidden()			
		}

		err = deleteComment(tx, params.ID)
		if err != nil {
			log.Print(err)
			tx.Rollback()
			return comments.NewDeleteCommentsIDNotFound()
		}

		tx.Commit()
		return comments.NewDeleteCommentsIDOK()
	}
}
