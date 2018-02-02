package watchings

import (
	"database/sql"
	"os"
	"testing"

	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
)

var db *sql.DB

func TestMain(m *testing.M) {
	config := utils.LoadConfig("../../../../configs/server")
	db = utils.OpenDatabase(config)
	utils.ClearDatabase(db)

	os.Exit(m.Run())
}

func TestLoadWatching(t *testing.T) {

}
