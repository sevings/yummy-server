package helper

import (
	"database/sql"
	"log"

	"github.com/sevings/mindwell-server/internal/app/mindwell-server/entries"
)

type entry struct {
	id        int64
	title     string
	content   string
	hasAttach bool
}

func UpdateEntries(tx *sql.Tx) {
	rows, err := tx.Query(`
		SELECT id, title, edit_content,
			EXISTS(SELECT 1 FROM entry_images WHERE entry_id = entries.id)
		FROM entries
	`)
	if err != nil {
		log.Println(err)
	}

	var feed []entry
	for rows.Next() {
		var e entry
		rows.Scan(&e.id, &e.title, &e.content, &e.hasAttach)
		feed = append(feed, e)
	}

	log.Printf("Updating %d entries...\n", len(feed))

	const q = `
		UPDATE entries
		SET title = $2, cut_title = $3, 
			content = $4, cut_content = $5, edit_content = $6, 
			has_cut = $7
		WHERE id = $1
	`

	for _, e := range feed {
		entry := entries.NewEntry(e.title, e.content, e.hasAttach)
		_, err = tx.Exec(q, e.id, entry.Title, entry.CutTitle, entry.Content, entry.CutContent, entry.EditContent, entry.HasCut)
		if err != nil {
			log.Println(err)
		}
	}

	log.Println("Completed.")
}
