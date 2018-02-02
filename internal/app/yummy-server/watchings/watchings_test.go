package watchings

import (
	"database/sql"
	"os"
	"testing"
)

var db *sql.DB

func TestMain(m *testing.M) {
	config := utils.LoadConfig("../../config")
	db = utils.OpenDatabase(config)
	utils.ClearDatabase(db)

	os.Exit(m.Run())
}

func TestLoadWatching(t *testing.T) {

}
