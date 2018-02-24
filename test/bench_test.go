package test

import (
	"testing"

	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations/entries"
)

func BenchmarkLoadLive(b *testing.B) {
	post := api.EntriesPostEntriesUsersMeHandler.Handle
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

	load := api.EntriesGetEntriesLiveHandler.Handle

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		load(params, userIDs[1])
	}
}
