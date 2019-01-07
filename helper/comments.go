package helper

import (
	"database/sql"
	"log"

	"github.com/sevings/mindwell-server/internal/app/mindwell-server/comments"
)

type comment struct {
	id   int64
	text string
}

func UpdateComments(tx *sql.Tx) {
	rows, err := tx.Query("SELECT id, edit_content FROM comments")
	if err != nil {
		log.Println(err)
	}

	var cmts []comment
	for rows.Next() {
		var cmt comment
		rows.Scan(&cmt.id, &cmt.text)
		cmts = append(cmts, cmt)
	}

	log.Printf("Updating %d comments...\n", len(cmts))

	for _, cmt := range cmts {
		html := comments.HtmlContent(cmt.text)
		_, err = tx.Exec("UPDATE comments SET content = $2 WHERE id = $1", cmt.id, html)
		if err != nil {
			log.Println(err)
		}
	}

	log.Println("Completed.")
}
