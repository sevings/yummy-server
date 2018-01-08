package watchings

import (
	"database/sql"
	"os"
	"testing"

	yummy "github.com/sevings/yummy-server/src"
)

var db *sql.DB

func TestMain(m *testing.M) {
	config := yummy.LoadConfig("../../config")
	db = yummy.OpenDatabase(config)
	yummy.ClearDatabase(db)

	os.Exit(m.Run())
}

func TestLoadWatching(t *testing.T) {

}
