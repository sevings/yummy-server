package test

import (
	"fmt"
	"testing"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/entries"
	"github.com/sevings/mindwell-server/restapi/operations/favorites"
	"github.com/sevings/mindwell-server/restapi/operations/me"
	"github.com/sevings/mindwell-server/restapi/operations/users"
)

func BenchmarkLoadLive(b *testing.B) {
	post := api.MePostMeTlogHandler.Handle
	var title string
	votable := true
	live := true
	entryParams := me.PostMeTlogParams{
		Title:     &title,
		Privacy:   models.EntryPrivacyAll,
		IsVotable: &votable,
		InLive:    &live,
	}
	for i := 0; i < 1000; i++ {
		title = fmt.Sprintf("Entry %d", i)
		entryParams.Content = fmt.Sprintf("test test test %d", i)
		post(entryParams, userIDs[0])
	}

	var limit int64 = 30
	before := "0"
	after := "0"
	section := "entries"
	query := ""
	tag := ""
	params := entries.GetEntriesLiveParams{
		Limit:   &limit,
		Before:  &before,
		After:   &after,
		Section: &section,
		Query:   &query,
		Tag:     &tag,
	}

	load := api.EntriesGetEntriesLiveHandler.Handle

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		load(params, userIDs[1])
	}
}

func BenchmarkLoadFavorite(b *testing.B) {
	post := api.MePostMeTlogHandler.Handle
	var title string
	votable := true
	live := true
	entryParams := me.PostMeTlogParams{
		Title:     &title,
		Privacy:   models.EntryPrivacyAll,
		IsVotable: &votable,
		InLive:    &live,
	}
	var ids []int64
	for i := 0; i < 1000; i++ {
		title = fmt.Sprintf("Entry %d", i)
		entryParams.Content = fmt.Sprintf("test test test %d", i)
		resp := post(entryParams, userIDs[0])
		body := resp.(*me.PostMeTlogCreated)
		id := body.Payload.ID
		ids = append(ids, id)
	}

	fav := api.FavoritesPutEntriesIDFavoriteHandler.Handle
	for _, id := range ids {
		favParams := favorites.PutEntriesIDFavoriteParams{ID: id}
		fav(favParams, userIDs[1])
	}

	var limit int64 = 30
	before := "0"
	after := "0"
	query := ""
	params := users.GetUsersNameFavoritesParams{
		Limit:  &limit,
		Before: &before,
		After:  &after,
		Name:   userIDs[1].Name,
		Query:  &query,
	}

	load := api.UsersGetUsersNameFavoritesHandler.Handle

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		load(params, userIDs[1])
	}
}
