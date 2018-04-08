package test

import (
	"fmt"
	"testing"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/entries"
)

func BenchmarkLoadLive(b *testing.B) {
	post := api.EntriesPostEntriesUsersMeHandler.Handle
	var title string
	votable := true
	entryParams := entries.PostEntriesUsersMeParams{
		Content:   "test test test",
		Title:     &title,
		Privacy:   models.EntryPrivacyAll,
		IsVotable: &votable,
	}
	for i := 0; i < 1000; i++ {
		title = fmt.Sprintf("Entry %d", i)
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
