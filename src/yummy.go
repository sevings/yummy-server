package yummy

import (
	"database/sql"
	"log"

	"github.com/sevings/yummy-server/gen/models"
	"github.com/go-openapi/runtime/middleware"
)

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
