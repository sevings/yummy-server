package test

import (
	"fmt"
	"testing"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/entries"
	"github.com/sevings/mindwell-server/restapi/operations/me"
)

func BenchmarkLoadLive(b *testing.B) {
	post := api.MePostMeTlogHandler.Handle
	var title string
	votable := true
	live := true
	entryParams := me.PostMeTlogParams{
		Content:   "test test test",
		Title:     &title,
		Privacy:   models.EntryPrivacyAll,
		IsVotable: &votable,
		InLive:    &live,
	}
	for i := 0; i < 1000; i++ {
		title = fmt.Sprintf("Entry %d", i)
		post(entryParams, userIDs[0])
	}

	var limit int64 = 30
	before := "0"
	after := "0"
	section := "entries"
	params := entries.GetEntriesLiveParams{
		Limit:   &limit,
		Before:  &before,
		After:   &after,
		Section: &section,
	}

	load := api.EntriesGetEntriesLiveHandler.Handle

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		load(params, userIDs[1])
	}
}
