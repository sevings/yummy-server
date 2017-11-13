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

			return comments.NewGetCommentsIDNotFound()
		}

		canView := entries.CanViewEntry(tx, userID, params.ID)
		if !canView {
			return comments.NewGetCommentsIDNotFound()
		}

		return comments.NewGetCommentsIDOK().WithPayload(comment)
	}
}
