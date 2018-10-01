package main

import (
	"log"

	"github.com/sevings/mindwell-server/internal/app/mindwell-server/comments"
	"github.com/sevings/mindwell-server/utils"
)

type comment struct {
	id   int64
	text string
}

func main() {
	cfg := utils.LoadConfig("configs/server")
	db := utils.OpenDatabase(cfg)
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Println(err)
	}

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

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	log.Println("Completed.")
}
