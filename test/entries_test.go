package test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/entries"
	"github.com/sevings/mindwell-server/restapi/operations/favorites"
	"github.com/sevings/mindwell-server/restapi/operations/me"
	"github.com/sevings/mindwell-server/restapi/operations/users"
	"github.com/sevings/mindwell-server/utils"
	"github.com/stretchr/testify/require"
)

func checkEntry(t *testing.T, entry *models.Entry,
	user *models.AuthProfile, canEdit bool, vote string, watching bool,
	wc int64, privacy string, votable, live bool, title, content string) {

	req := require.New(t)
	req.NotEmpty(entry.CreatedAt)
	req.Equal("<p>"+content+"</p>\n", entry.Content)
	req.Equal(wc, entry.WordCount)
	req.Equal(privacy, entry.Privacy)
	req.Empty(entry.VisibleFor)
	req.Zero(entry.CommentCount)
	req.False(entry.IsFavorited)
	req.Equal(watching, entry.IsWatching)
	req.Equal(title, entry.Title)
	req.Empty(entry.CutTitle)
	req.Empty(entry.CutContent)
	req.False(entry.HasCut)
	req.Equal(live, entry.InLive)

	if canEdit {
		req.Equal(content, entry.EditContent)
	} else {
		req.Empty(entry.EditContent)
	}

	rating := entry.Rating
	req.Equal(entry.ID, rating.ID)
	req.Zero(rating.Rating)
	req.Equal(votable, rating.IsVotable)
	req.Equal(vote, rating.Vote)

	cmts := entry.Comments
	if cmts != nil {
		req.Empty(cmts.Data)
		req.False(cmts.HasAfter)
		req.False(cmts.HasBefore)
		req.Zero(cmts.NextAfter)
		req.Zero(cmts.NextBefore)
	}

	author := entry.Author
	req.Equal(user.ID, author.ID)
	req.Equal(user.Name, author.Name)
	req.Equal(user.ShowName, author.ShowName)
	req.Equal(user.IsOnline, author.IsOnline)
	req.Equal(user.Avatar, author.Avatar)
}

func checkLoadEntry(t *testing.T, entryID int64, userID *models.UserID, success bool,
	user *models.AuthProfile, canEdit bool, vote string, watching bool,
	wc int64, privacy string, votable, live bool, title, content string) {

	load := api.EntriesGetEntriesIDHandler.Handle
	resp := load(entries.GetEntriesIDParams{ID: entryID}, userID)
	body, ok := resp.(*entries.GetEntriesIDOK)
	require.Equal(t, success, ok)
	if !success {
		return
	}

	entry := body.Payload
	checkEntry(t, entry, user, true, models.RatingVoteBan, true, wc, privacy, votable, live, title, content)
}

func checkPostEntry(t *testing.T,
	params me.PostMeTlogParams,
	user *models.AuthProfile, id *models.UserID, success bool, wc int64) int64 {

	post := api.MePostMeTlogHandler.Handle
	resp := post(params, id)
	body, ok := resp.(*me.PostMeTlogCreated)
	require.Equal(t, success, ok)
	if !ok {
		return 0
	}

	entry := body.Payload
	checkEntry(t, entry, user, true, models.RatingVoteBan, true, wc, params.Privacy, *params.IsVotable, *params.InLive,
		*params.Title, params.Content)

	checkLoadEntry(t, entry.ID, id, true, user,
		true, models.RatingVoteBan, true, wc, params.Privacy, *params.IsVotable, *params.InLive,
		*params.Title, params.Content)

	return entry.ID
}

func checkEditEntry(t *testing.T,
	params entries.PutEntriesIDParams,
	user *models.AuthProfile, id *models.UserID, success bool, wc int64) {

	edit := api.EntriesPutEntriesIDHandler.Handle
	resp := edit(params, id)
	body, ok := resp.(*entries.PutEntriesIDOK)
	require.Equal(t, success, ok)
	if !ok {
		return
	}

	entry := body.Payload
	checkEntry(t, entry, user, true, models.RatingVoteBan, true, wc, params.Privacy, *params.IsVotable, *params.InLive,
		*params.Title, params.Content)

	checkLoadEntry(t, entry.ID, id, true, user,
		true, models.RatingVoteBan, true, wc, params.Privacy, *params.IsVotable, *params.InLive,
		*params.Title, params.Content)
}

func checkDeleteEntry(t *testing.T, entryID int64, userID *models.UserID, success bool) {
	del := api.EntriesDeleteEntriesIDHandler.Handle
	resp := del(entries.DeleteEntriesIDParams{ID: entryID}, userID)
	_, ok := resp.(*entries.DeleteEntriesIDOK)
	require.Equal(t, success, ok)
}

func TestPostMyTlog(t *testing.T) {
	params := me.PostMeTlogParams{
		Content: "test content",
	}

	votable := false
	params.IsVotable = &votable

	live := true
	params.InLive = &live

	params.Privacy = models.EntryPrivacyAll

	title := "title title ti"
	params.Title = &title

	id := checkPostEntry(t, params, profiles[0], userIDs[0], true, 5)
	checkEntryWatching(t, userIDs[0], id, true, true)

	title = "title"
	votable = false
	live = false
	editParams := entries.PutEntriesIDParams{
		ID:        id,
		Content:   "content",
		Title:     &title,
		IsVotable: &votable,
		InLive:    &live,
		Privacy:   models.EntryPrivacyMe,
	}

	checkEditEntry(t, editParams, profiles[0], userIDs[0], true, 2)

	checkLoadEntry(t, id, userIDs[1], false, nil, false, "", false, 0, "", false, false, "", "")

	checkDeleteEntry(t, id, userIDs[1], false)
	checkDeleteEntry(t, id, userIDs[0], true)
	checkDeleteEntry(t, id, userIDs[0], false)
}

func TestLiveRestrictions(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)

	_, err := db.Exec("UPDATE users SET followers_count = 4 WHERE id = $1", userIDs[0].ID)
	if err != nil {
		log.Println(err)
	}

	votable := true
	live := true
	title := ""
	postParams := me.PostMeTlogParams{
		Content:   "test test test",
		Title:     &title,
		Privacy:   models.EntryPrivacyAll,
		IsVotable: &votable,
		InLive:    &live,
	}
	e0 := checkPostEntry(t, postParams, profiles[0], userIDs[0], true, 3)
	checkPostEntry(t, postParams, profiles[0], userIDs[0], true, 3)
	checkPostEntry(t, postParams, profiles[0], userIDs[0], false, 3)

	live = false
	e1 := checkPostEntry(t, postParams, profiles[0], userIDs[0], true, 3)

	live = true
	editParams := entries.PutEntriesIDParams{
		ID:        e0,
		Content:   "content",
		Title:     &title,
		IsVotable: &votable,
		InLive:    &live,
		Privacy:   models.EntryPrivacyAll,
	}
	checkEditEntry(t, editParams, profiles[0], userIDs[0], true, 1)

	live = false
	checkEditEntry(t, editParams, profiles[0], userIDs[0], true, 1)

	live = true
	checkEditEntry(t, editParams, profiles[0], userIDs[0], true, 1)

	editParams.ID = e1
	checkEditEntry(t, editParams, profiles[0], userIDs[0], false, 1)

	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)
}

func postEntry(id *models.UserID, privacy string, live bool) *models.Entry {
	votable := true
	title := ""
	params := me.PostMeTlogParams{
		Content:   "test test test",
		Title:     &title,
		Privacy:   privacy,
		IsVotable: &votable,
		InLive:    &live,
	}
	post := api.MePostMeTlogHandler.Handle
	resp := post(params, id)
	body := resp.(*me.PostMeTlogCreated)
	entry := body.Payload

	time.Sleep(10 * time.Millisecond)

	return entry
}

func checkLoadLive(t *testing.T, id *models.UserID, limit int64, section, before, after string, size int) *models.Feed {
	params := entries.GetEntriesLiveParams{
		Limit:   &limit,
		Before:  &before,
		After:   &after,
		Section: &section,
	}

	load := api.EntriesGetEntriesLiveHandler.Handle
	resp := load(params, id)
	body, ok := resp.(*entries.GetEntriesLiveOK)
	if !ok {
		t.Fatal("error load live")
	}

	feed := body.Payload
	require.Equal(t, size, len(feed.Entries))

	return feed
}

func TestLoadLive(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)

	postEntry(userIDs[0], models.EntryPrivacyAll, true)
	postEntry(userIDs[0], models.EntryPrivacyAll, false)
	postEntry(userIDs[0], models.EntryPrivacySome, true)
	postEntry(userIDs[1], models.EntryPrivacyMe, true)
	postEntry(userIDs[1], models.EntryPrivacyAll, true)
	postEntry(userIDs[2], models.EntryPrivacyAll, false)
	postEntry(userIDs[2], models.EntryPrivacyAll, true)

	feed := checkLoadLive(t, userIDs[0], 10, "entries", "", "", 3)
	checkEntry(t, feed.Entries[0], profiles[2], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[1], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, true, "", "test test test")
	checkEntry(t, feed.Entries[2], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, true, "", "test test test")

	req := require.New(t)
	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 1, "entries", "", "", 1)
	checkEntry(t, feed.Entries[0], profiles[2], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, true, "", "test test test")

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 5, "entries", feed.NextBefore, "", 2)
	checkEntry(t, feed.Entries[0], profiles[1], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, true, "", "test test test")

	req.False(feed.HasBefore)
	req.True(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 2, "entries", "", "", 2)
	checkEntry(t, feed.Entries[0], profiles[2], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[1], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, true, "", "test test test")

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 5, "entries", feed.NextBefore, "", 1)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, true, "", "test test test")

	req.False(feed.HasBefore)
	req.True(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 5, "entries", "", feed.NextAfter, 2)
	checkEntry(t, feed.Entries[0], profiles[2], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[1], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, true, "", "test test test")

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	checkLoadLive(t, userIDs[0], 1, "entries", "", feed.NextAfter, 0)
	checkLoadLive(t, userIDs[0], 0, "entries", "", feed.NextAfter, 0)
}

func checkLoadTlog(t *testing.T, tlog, user *models.UserID, limit int64, before, after string, size int) *models.Feed {
	params := users.GetUsersNameTlogParams{
		Name:   tlog.Name,
		Limit:  &limit,
		Before: &before,
		After:  &after,
	}

	load := api.UsersGetUsersNameTlogHandler.Handle
	resp := load(params, user)
	body, ok := resp.(*users.GetUsersNameTlogOK)
	if !ok {
		t.Fatal("error load tlog")
	}

	feed := body.Payload
	require.Equal(t, size, len(feed.Entries))

	return feed
}

func TestLoadTlog(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)

	postEntry(userIDs[0], models.EntryPrivacyAll, true)
	postEntry(userIDs[0], models.EntryPrivacySome, true)
	postEntry(userIDs[0], models.EntryPrivacyMe, true)
	postEntry(userIDs[0], models.EntryPrivacyAll, false)

	feed := checkLoadTlog(t, userIDs[0], userIDs[1], 10, "", "", 2)
	checkEntry(t, feed.Entries[0], profiles[0], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, false, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, true, "", "test test test")

	req := require.New(t)
	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadTlog(t, userIDs[0], userIDs[0], 10, "", "", 4)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, false, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyMe, false, false, "", "test test test")
	checkEntry(t, feed.Entries[2], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, true, "", "test test test")
	checkEntry(t, feed.Entries[3], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, true, "", "test test test")

	checkLoadTlog(t, userIDs[1], userIDs[0], 10, "", "", 0)

	feed = checkLoadTlog(t, userIDs[0], userIDs[0], 3, "", "", 3)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, false, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyMe, false, false, "", "test test test")
	checkEntry(t, feed.Entries[2], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, true, "", "test test test")

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadTlog(t, userIDs[0], userIDs[0], 3, feed.NextBefore, "", 1)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, true, "", "test test test")

	req.False(feed.HasBefore)
	req.True(feed.HasAfter)
}

func checkLoadMyTlog(t *testing.T, user *models.UserID, limit int64, before, after string, size int) *models.Feed {
	params := me.GetMeTlogParams{
		Limit:  &limit,
		Before: &before,
		After:  &after,
	}

	load := api.MeGetMeTlogHandler.Handle
	resp := load(params, user)
	body, ok := resp.(*me.GetMeTlogOK)
	if !ok {
		t.Fatal("error load tlog")
	}

	feed := body.Payload
	require.Equal(t, size, len(feed.Entries))

	return feed
}

func TestLoadMyTlog(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)

	postEntry(userIDs[0], models.EntryPrivacyAll, true)
	postEntry(userIDs[0], models.EntryPrivacySome, true)
	postEntry(userIDs[0], models.EntryPrivacyMe, true)
	postEntry(userIDs[0], models.EntryPrivacyAll, false)

	feed := checkLoadMyTlog(t, userIDs[0], 10, "", "", 4)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, false, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyMe, false, false, "", "test test test")
	checkEntry(t, feed.Entries[2], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, true, "", "test test test")
	checkEntry(t, feed.Entries[3], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, true, "", "test test test")

	req := require.New(t)
	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	checkLoadMyTlog(t, userIDs[1], 10, "", "", 0)

	feed = checkLoadMyTlog(t, userIDs[0], 1, "", "", 1)

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadMyTlog(t, userIDs[0], 4, feed.NextBefore, "", 3)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyMe, false, false, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, true, "", "test test test")
	checkEntry(t, feed.Entries[2], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, true, "", "test test test")
}

func checkLoadFriendsFeed(t *testing.T, user *models.UserID, limit int64, before, after string, size int) *models.Feed {
	params := entries.GetEntriesFriendsParams{
		Limit:  &limit,
		Before: &before,
		After:  &after,
	}

	load := api.EntriesGetEntriesFriendsHandler.Handle
	resp := load(params, user)
	body, ok := resp.(*entries.GetEntriesFriendsOK)
	if !ok {
		t.Fatal("error load tlog")
	}

	feed := body.Payload
	require.Equal(t, size, len(feed.Entries))

	return feed
}

func TestLoadFriendsFeed(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)
	esm.Clear()

	checkFollow(t, userIDs[0], userIDs[1], profiles[1], models.RelationshipRelationFollowed)

	postEntry(userIDs[0], models.EntryPrivacyAll, true)
	postEntry(userIDs[0], models.EntryPrivacySome, true)
	postEntry(userIDs[0], models.EntryPrivacyMe, true)
	postEntry(userIDs[0], models.EntryPrivacyAll, false)

	postEntry(userIDs[1], models.EntryPrivacyAll, true)
	postEntry(userIDs[1], models.EntryPrivacySome, true)
	postEntry(userIDs[1], models.EntryPrivacyMe, true)

	postEntry(userIDs[2], models.EntryPrivacyAll, true)
	postEntry(userIDs[2], models.EntryPrivacySome, true)
	postEntry(userIDs[2], models.EntryPrivacyMe, true)

	feed := checkLoadFriendsFeed(t, userIDs[0], 10, "", "", 4)
	checkEntry(t, feed.Entries[0], profiles[1], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, false, "", "test test test")
	checkEntry(t, feed.Entries[2], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, true, "", "test test test")
	checkEntry(t, feed.Entries[3], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, true, "", "test test test")

	req := require.New(t)
	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadFriendsFeed(t, userIDs[1], 10, "", "", 2)
	checkEntry(t, feed.Entries[0], profiles[1], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[1], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, true, "", "test test test")

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadFriendsFeed(t, userIDs[0], 1, "", "", 1)

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadFriendsFeed(t, userIDs[0], 4, feed.NextBefore, "", 3)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, false, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, true, "", "test test test")
	checkEntry(t, feed.Entries[2], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, true, "", "test test test")

	checkUnfollow(t, userIDs[0], userIDs[1])
}

func checkLoadFavorites(t *testing.T, user, tlog *models.UserID, limit int64, before, after string, size int) *models.Feed {
	params := users.GetUsersNameFavoritesParams{
		Name:   tlog.Name,
		Limit:  &limit,
		Before: &before,
		After:  &after,
	}

	load := api.UsersGetUsersNameFavoritesHandler.Handle
	resp := load(params, user)
	body, ok := resp.(*users.GetUsersNameFavoritesOK)
	if !ok {
		t.Fatal("error load favorites")
	}

	feed := body.Payload
	require.Equal(t, size, len(feed.Entries))

	return feed
}

func favoriteEntry(t *testing.T, user *models.UserID, entryID int64) {
	put := api.FavoritesPutEntriesIDFavoriteHandler.Handle
	params := favorites.PutEntriesIDFavoriteParams{
		ID: entryID,
	}
	put(params, user)

	time.Sleep(10 * time.Millisecond)
}

func TestLoadFavorites(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)
	esm.Clear()

	postEntry(userIDs[0], models.EntryPrivacyAll, true)
	postEntry(userIDs[0], models.EntryPrivacySome, true)
	postEntry(userIDs[0], models.EntryPrivacyMe, true)
	postEntry(userIDs[0], models.EntryPrivacyAll, false)

	tlog := checkLoadMyTlog(t, userIDs[0], 10, "", "", 4)

	favoriteEntry(t, userIDs[0], tlog.Entries[2].ID)
	favoriteEntry(t, userIDs[0], tlog.Entries[1].ID)
	favoriteEntry(t, userIDs[0], tlog.Entries[0].ID)

	req := require.New(t)

	feed := checkLoadFavorites(t, userIDs[0], userIDs[0], 10, "", "", 3)
	req.Equal(tlog.Entries[0].ID, feed.Entries[0].ID)
	req.Equal(tlog.Entries[1].ID, feed.Entries[1].ID)
	req.Equal(tlog.Entries[2].ID, feed.Entries[2].ID)

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	checkLoadFavorites(t, userIDs[1], userIDs[0], 10, "", "", 1)
	checkLoadFavorites(t, userIDs[0], userIDs[1], 10, "", "", 0)

	feed = checkLoadFavorites(t, userIDs[0], userIDs[0], 2, "", "", 2)
	req.Equal(tlog.Entries[0].ID, feed.Entries[0].ID)
	req.Equal(tlog.Entries[1].ID, feed.Entries[1].ID)

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadFavorites(t, userIDs[0], userIDs[0], 10, feed.NextBefore, "", 1)
	req.Equal(tlog.Entries[2].ID, feed.Entries[0].ID)

	req.False(feed.HasBefore)
	req.True(feed.HasAfter)

	feed = checkLoadFavorites(t, userIDs[1], userIDs[0], 2, "", "", 1)
	req.Equal(tlog.Entries[0].ID, feed.Entries[0].ID)

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)
}

func TestLoadLiveComments(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)

	entries := make([]*models.Entry, 6)

	entries[0] = postEntry(userIDs[0], models.EntryPrivacyAll, true) // 2
	entries[1] = postEntry(userIDs[0], models.EntryPrivacyAll, false)
	entries[2] = postEntry(userIDs[0], models.EntryPrivacySome, true)
	entries[3] = postEntry(userIDs[1], models.EntryPrivacyAll, true) // 1
	entries[4] = postEntry(userIDs[1], models.EntryPrivacyAll, true)
	entries[5] = postEntry(userIDs[1], models.EntryPrivacyAll, true) // 3

	// skip 4
	comments := make([]int64, 5)

	comments[0] = postComment(userIDs[0], entries[5].ID)
	comments[1] = postComment(userIDs[0], entries[0].ID)
	comments[2] = postComment(userIDs[0], entries[3].ID)
	comments[3] = postComment(userIDs[0], entries[1].ID)
	comments[4] = postComment(userIDs[0], entries[2].ID)

	for _, e := range entries {
		e.CommentCount = 1
		e.EditContent = ""
		e.IsWatching = false
		e.Rating.Vote = "not"
	}

	feed := checkLoadLive(t, userIDs[2], 10, "comments", "", "", 3)

	req := require.New(t)
	req.Equal(*entries[3], *feed.Entries[0])
	req.Equal(*entries[0], *feed.Entries[1])
	req.Equal(*entries[5], *feed.Entries[2])

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[2], 1, "comments", "", "", 1)
	req.Equal(*entries[3], *feed.Entries[0])

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	checkDeleteComment(t, comments[0], userIDs[0], true)
	checkDeleteComment(t, comments[3], userIDs[0], true)
	checkLoadLive(t, userIDs[2], 10, "comments", "", "", 2)

	esm.Clear()
}

func checkLoadWatching(t *testing.T, id *models.UserID, limit int64, size int) *models.Feed {
	params := entries.GetEntriesWatchingParams{
		Limit: &limit,
	}

	load := api.EntriesGetEntriesWatchingHandler.Handle
	resp := load(params, id)
	body, ok := resp.(*entries.GetEntriesWatchingOK)

	require.True(t, ok)

	feed := body.Payload
	require.Equal(t, size, len(feed.Entries))

	return feed
}

func TestLoadWatching(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)

	entries := make([]*models.Entry, 4)

	entries[0] = postEntry(userIDs[0], models.EntryPrivacyAll, true) // 2
	entries[1] = postEntry(userIDs[1], models.EntryPrivacyAll, true) // 1
	entries[2] = postEntry(userIDs[1], models.EntryPrivacyAll, true)
	entries[3] = postEntry(userIDs[1], models.EntryPrivacyAll, true) // 3

	// skip 2
	postComment(userIDs[2], entries[3].ID)
	postComment(userIDs[2], entries[1].ID)
	postComment(userIDs[2], entries[0].ID)
	postComment(userIDs[0], entries[1].ID)

	for _, e := range entries {
		e.CommentCount = 1
		e.EditContent = ""
		e.IsWatching = true
		e.Rating.Vote = "not"
	}

	entries[1].CommentCount = 2

	feed := checkLoadWatching(t, userIDs[2], 10, 3)

	req := require.New(t)
	req.Equal(*entries[1], *feed.Entries[0])
	req.Equal(*entries[0], *feed.Entries[1])
	req.Equal(*entries[3], *feed.Entries[2])

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadWatching(t, userIDs[2], 1, 1)
	req.Equal(*entries[1], *feed.Entries[0])

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	esm.Clear()
}

func TestEntryHTML(t *testing.T) {
	post := func(in, out string) {
		params := me.PostMeTlogParams{
			Content: in,
		}

		votable := false
		params.IsVotable = &votable

		live := false
		params.InLive = &live

		params.Privacy = models.EntryPrivacyAll

		title := "title title ti"
		params.Title = &title

		post := api.MePostMeTlogHandler.Handle
		resp := post(params, userIDs[0])
		body, ok := resp.(*me.PostMeTlogCreated)
		require.True(t, ok)
		if !ok {
			return
		}

		entry := body.Payload
		require.Equal(t, out, entry.Content)
	}

	linkify := func(url string) (string, string) {
		return url, fmt.Sprintf(`<p><a href="%s" target="_blank">%s</a></p>
`, url, url)
	}

	post(linkify("https://ya.ru"))
}
