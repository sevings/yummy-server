package entries

import (
	"testing"

	"github.com/sevings/yummy-server/internal/app/yummy-server/tests"
	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations/entries"
)

func BenchmarkLoadLive(b *testing.B) {
	config := utils.LoadConfig("../../../../configs/server")
	db := utils.OpenDatabase(config)
	utils.ClearDatabase(db)

	userIDs, _ := tests.RegisterTestUsers(db)

	post := newMyTlogPoster(db)
	title := "title"
	privacy := models.EntryPrivacyAll
	votable := true
	entryParams := entries.PostEntriesUsersMeParams{
		Content:   "test test test",
		Title:     &title,
		Privacy:   &privacy,
		IsVotable: &votable,
	}
	for i := 0; i < 100; i++ {
		post(entryParams, userIDs[0])
	}

	var limit int64 = 50
	var skip int64 = 50
	params := entries.GetEntriesLiveParams{
		Limit: &limit,
		Skip:  &skip,
	}

	load := newLiveLoader(db)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		load(params, userIDs[1])
	}
}
