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

	dropTable(tx, "vote_weights")
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

type AutoTx struct {
	tx   *sql.Tx
	rows *sql.Rows
	res  sql.Result
	err  error
}

func (tx *AutoTx) Query(query string, args ...interface{}) *AutoTx {
	if tx.err != nil && tx.err != sql.ErrNoRows {
		return tx
	}

	if tx.rows != nil {
		tx.rows.Close()
	}

	tx.rows, tx.err = tx.tx.Query(query, args...)

	return tx
}

func (tx *AutoTx) Scan(dest ...interface{}) bool {
	if tx.err != nil {
		return false
	}

	if !tx.rows.Next() {
		tx.err = tx.rows.Err()
		tx.rows = nil
		if tx.err == nil {
			tx.err = sql.ErrNoRows
		}

		return false
	}

	tx.err = tx.rows.Scan(dest...)

	return true
}

func (tx *AutoTx) Close() {
	if tx.rows != nil {
		tx.rows.Close()
		tx.rows = nil
	}
}

func (tx *AutoTx) Error() error {
	return tx.err
}

func (tx *AutoTx) Exec(query string, args ...interface{}) {
	if tx.err != nil && tx.err != sql.ErrNoRows {
		return
	}

	if tx.rows != nil {
		tx.rows.Close()
		tx.rows = nil
	}

	tx.res, tx.err = tx.tx.Exec(query, args...)
}

func (tx *AutoTx) RowsAffected() int64 {
	var cnt int64
	if tx.err == nil {
		cnt, tx.err = tx.res.RowsAffected()
	}

	return cnt
}

// Transact wraps func in an SQL transaction.
// Responder will be just passed through.
func Transact(db *sql.DB, txFunc func(*AutoTx) middleware.Responder) middleware.Responder {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	atx := &AutoTx{tx: tx}
	resp := txFunc(atx)
	p := recover()
	atx.Close()

	if p != nil {
		err = tx.Rollback()
		log.Print("Recovered in Transact:", p)
	} else if atx.Error() == nil {
		err = tx.Commit()
	} else {
		if atx.Error() != sql.ErrNoRows {
			log.Print(atx.Error())
		}

		err = tx.Rollback()
	}

	if err != nil {
		log.Print(err)
	}

	return resp
}
