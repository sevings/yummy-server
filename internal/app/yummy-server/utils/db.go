package utils

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/go-openapi/runtime/middleware"
	goconf "github.com/zpatrick/go-config"
)

func dropTable(tx *sql.Tx, table string) {
	_, err := tx.Exec("delete from " + table)
	if err != nil {
		tx.Rollback()
		log.Fatal("cannot clear table " + table + ": " + err.Error())
	}
}

// ClearDatabase drops user data tables and then creates default user
func ClearDatabase(db *sql.DB) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("cannot begin tx")
	}

	dropTable(tx, "entries_privacy")
	// dropTable(tx, "entry_tags")
	dropTable(tx, "entry_votes")
	dropTable(tx, "comment_votes")
	dropTable(tx, "comments")
	dropTable(tx, "favorites")
	dropTable(tx, "invites")
	dropTable(tx, "relations")
	dropTable(tx, "tags")
	dropTable(tx, "watching")
	dropTable(tx, "entries")

	_, err = tx.Exec("delete from users where id != 1")
	if err != nil {
		tx.Rollback()
		log.Fatal("cannot clear table users: " + err.Error())
	}

	for i := 0; i < 3; i++ {
		_, err = tx.Exec("INSERT INTO invites (referrer_id, word1, word2, word3) VALUES(1, 1, 1, 1);")
		if err != nil {
			tx.Rollback()
			log.Fatal("cannot create invite")
		}
	}

	tx.Commit()
}

// OpenDatabase returns db opened from config.
func OpenDatabase(config *goconf.Config) *sql.DB {
	driver, err := config.StringOr("database.driver", "postgres")
	if err != nil {
		log.Print(err)
	}

	port, err := config.Int("database.port")
	if err != nil {
		log.Print(err)
	}

	user, err := config.String("database.user")
	if err != nil {
		log.Print(err)
	}

	pass, err := config.String("database.password")
	if err != nil {
		log.Print(err)
	}

	name, err := config.String("database.name")
	if err != nil {
		log.Print(err)
	}

	db, err := sql.Open(driver, "user="+user+" password="+pass+" dbname="+name+" port="+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}

	schema, err := config.String("database.schema")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("SET search_path = " + schema + ", public")
	if err != nil {
		log.Fatal(err)
	}

	return db
}

type AutoTx interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// Transact wraps func in an SQL transaction.
// Return true to commit or false to rollback. Responder will be just passed through.
func Transact(db *sql.DB, txFunc func(AutoTx) (middleware.Responder, bool)) middleware.Responder {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	resp, ok := txFunc(tx)
	if p := recover(); p != nil {
		tx.Rollback()
		panic(p) // re-throw panic after Rollback
	} else if ok {
		err = tx.Commit()
	} else {
		err = tx.Rollback()
	}

	if err != nil {
		log.Print(err)
	}

	return resp
}
