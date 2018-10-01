package main

import (
	"log"
	"strings"

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

	htmlEsc := strings.NewReplacer(
		"&lt;", "<",
		"&gt;", ">",
		"&amp;", "&",
		"&quot;", "\"",
		"&#34;", "\"",
		"&#034;", "\"",
		"&#39;", "'",
		"&#039;", "'",
	)

	for _, cmt := range cmts {
		text := htmlEsc.Replace(cmt.text)
		html := comments.HtmlContent(text)
		_, err = tx.Exec("UPDATE comments SET content = $2, edit_content = $3 WHERE id = $1", cmt.id, html, text)
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
