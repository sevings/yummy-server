package yummy

import (
	"database/sql"
	"log"

	goconf "github.com/zpatrick/go-config"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/models"

	// to use postgres
	_ "github.com/lib/pq"
)

// LoadConfig creates app config from file
func LoadConfig(fileName string) *goconf.Config {
	toml := goconf.NewTOMLFile(fileName + ".toml")
	loader := goconf.NewOnceLoader(toml)
	config := goconf.NewConfig([]goconf.Provider{loader})
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	return config
}

// OpenDatabase returns db opened from config.
func OpenDatabase(config *goconf.Config) *sql.DB {
	driver, err := config.StringOr("database.driver", "postgres")
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

	db, err := sql.Open(driver, "user="+user+" password="+pass+" dbname="+name)
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

// NewError returns error object with some message
func NewError(msg string) *models.Error {
	return &models.Error{Message: msg}
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

// CanViewEntry returns true if the user is allowed to read the entry.
func CanViewEntry(tx AutoTx, userID, entryID int64) bool {
	const q = `
		SELECT TRUE 
		FROM feed
		WHERE id = $2 AND (author_id = $1
			OR ((entry_privacy = 'all' 
				AND (author_privacy = 'all'
					OR (author_privacy = 'registered' AND $1 > 0)
					OR EXISTS(SELECT 1 FROM relation, relations, entries
							  WHERE from_id = $1 AND to_id = entries.author_id
								  AND entries.id = $2
						 		  AND relation.type = 'followed'
						 		  AND relations.type = relation.id)))
			OR (entry_privacy = 'some' 
				AND EXISTS(SELECT 1 FROM entries_privacy
					WHERE user_id = $1 AND entry_id = $2))
			OR entry_privacy = 'anonymous'))`

	var allowed bool
	err := tx.QueryRow(q, userID, entryID).Scan(&allowed)
	if err != nil && err != sql.ErrNoRows {
		log.Print(err)
	}

	return allowed
}
