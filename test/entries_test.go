package test

import (
	"testing"
	"time"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/entries"
	"github.com/sevings/mindwell-server/restapi/operations/me"
	"github.com/sevings/mindwell-server/restapi/operations/users"
	"github.com/sevings/mindwell-server/utils"
	"github.com/stretchr/testify/require"
)

func checkEntry(t *testing.T, entry *models.Entry,
	user *models.AuthProfile, canEdit bool, vote string, watching bool,
	wc int64, privacy string, votable bool, title, content string) {

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
	wc int64, privacy string, votable bool, title, content string) {

	load := api.EntriesGetEntriesIDHandler.Handle
	resp := load(entries.GetEntriesIDParams{ID: entryID}, userID)
	body, ok := resp.(*entries.GetEntriesIDOK)
	require.Equal(t, success, ok)
	if !success {
		return
	}

	entry := body.Payload
	checkEntry(t, entry, user, true, models.RatingVoteBan, true, wc, privacy, votable, title, content)
}

func checkPostEntry(t *testing.T,
	params me.PostMeTlogParams,
	user *models.AuthProfile, id *models.UserID, wc int64) int64 {

	post := api.MePostMeTlogHandler.Handle
	resp := post(params, id)
	body, ok := resp.(*me.PostMeTlogCreated)
	if !ok {
		t.Fatal("error post entry")
	}

	entry := body.Payload
	checkEntry(t, entry, user, true, models.RatingVoteBan, true, wc, params.Privacy, *params.IsVotable, *params.Title, params.Content)

	checkLoadEntry(t, entry.ID, id, true, user,
		true, models.RatingVoteBan, true, wc, params.Privacy, *params.IsVotable, *params.Title, params.Content)

	return entry.ID
}

func checkEditEntry(t *testing.T,
	params entries.PutEntriesIDParams,
	user *models.AuthProfile, id *models.UserID, wc int64) {

	edit := api.EntriesPutEntriesIDHandler.Handle
	resp := edit(params, id)
	body, ok := resp.(*entries.PutEntriesIDOK)
	if !ok {
		badBody, ok := resp.(*entries.PutEntriesIDForbidden)
		if ok {
			t.Fatal(badBody.Payload.Message)
		}

		t.Fatal("error edit entry")
	}

	entry := body.Payload
	checkEntry(t, entry, user, true, models.RatingVoteBan, true, wc, params.Privacy, *params.IsVotable, *params.Title, params.Content)

	checkLoadEntry(t, entry.ID, id, true, user,
		true, models.RatingVoteBan, true, wc, params.Privacy, *params.IsVotable, *params.Title, params.Content)
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

	params.Privacy = models.EntryPrivacyAll

	title := "title title ti"
	params.Title = &title

	id := checkPostEntry(t, params, profiles[0], userIDs[0], 5)
	checkEntryWatching(t, userIDs[0], id, true, true)

	title = "title"
	votable = true
	editParams := entries.PutEntriesIDParams{
		ID:        id,
		Content:   "content",
		Title:     &title,
		IsVotable: &votable,
		Privacy:   models.EntryPrivacyMe,
	}

	checkEditEntry(t, editParams, profiles[0], userIDs[0], 2)

	checkLoadEntry(t, id, userIDs[1], false, nil, false, "", false, 0, "", false, "", "")

	checkDeleteEntry(t, id, userIDs[1], false)
	checkDeleteEntry(t, id, userIDs[0], true)
	checkDeleteEntry(t, id, userIDs[0], false)
}

func postEntry(id *models.UserID, privacy string) {
	post := api.MePostMeTlogHandler.Handle
	votable := true
	title := ""
	params := me.PostMeTlogParams{
		Content:   "test test test",
		Title:     &title,
		Privacy:   privacy,
		IsVotable: &votable,
	}
	post(params, id)

	time.Sleep(10 * time.Millisecond)
}

func checkLoadLive(t *testing.T, id *models.UserID, limit int64, before, after string, size int) *models.Feed {
	params := entries.GetEntriesLiveParams{
		Limit:  &limit,
		Before: &before,
		After:  &after,
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

	postEntry(userIDs[0], models.EntryPrivacyAll)
	postEntry(userIDs[0], models.EntryPrivacySome)
	postEntry(userIDs[1], models.EntryPrivacyMe)
	postEntry(userIDs[1], models.EntryPrivacyAll)
	postEntry(userIDs[2], models.EntryPrivacyAll)

	feed := checkLoadLive(t, userIDs[0], 10, "", "", 3)
	checkEntry(t, feed.Entries[0], profiles[2], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[1], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed.Entries[2], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")

	req := require.New(t)
	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 1, "", "", 1)
	checkEntry(t, feed.Entries[0], profiles[2], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, "", "test test test")

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 5, feed.NextBefore, "", 2)
	checkEntry(t, feed.Entries[0], profiles[1], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")

	req.False(feed.HasBefore)
	req.True(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 2, "", "", 2)
	checkEntry(t, feed.Entries[0], profiles[2], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[1], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, "", "test test test")

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 5, feed.NextBefore, "", 1)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")

	req.False(feed.HasBefore)
	req.True(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 5, "", feed.NextAfter, 2)
	checkEntry(t, feed.Entries[0], profiles[2], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[1], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, "", "test test test")

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	checkLoadLive(t, userIDs[0], 1, "", feed.NextAfter, 0)
	checkLoadLive(t, userIDs[0], 0, "", feed.NextAfter, 0)
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

	postEntry(userIDs[0], models.EntryPrivacyAll)
	postEntry(userIDs[0], models.EntryPrivacySome)
	postEntry(userIDs[0], models.EntryPrivacyMe)
	postEntry(userIDs[0], models.EntryPrivacyAll)

	feed := checkLoadTlog(t, userIDs[0], userIDs[1], 10, "", "", 2)
	checkEntry(t, feed.Entries[0], profiles[0], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, "", "test test test")

	req := require.New(t)
	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadTlog(t, userIDs[0], userIDs[0], 10, "", "", 4)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyMe, true, "", "test test test")
	checkEntry(t, feed.Entries[2], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, "", "test test test")
	checkEntry(t, feed.Entries[3], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")

	checkLoadTlog(t, userIDs[1], userIDs[0], 10, "", "", 0)

	feed = checkLoadTlog(t, userIDs[0], userIDs[0], 3, "", "", 3)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyMe, true, "", "test test test")
	checkEntry(t, feed.Entries[2], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, "", "test test test")

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadTlog(t, userIDs[0], userIDs[0], 3, feed.NextBefore, "", 1)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")

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

	postEntry(userIDs[0], models.EntryPrivacyAll)
	postEntry(userIDs[0], models.EntryPrivacySome)
	postEntry(userIDs[0], models.EntryPrivacyMe)
	postEntry(userIDs[0], models.EntryPrivacyAll)

	feed := checkLoadMyTlog(t, userIDs[0], 10, "", "", 4)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyMe, true, "", "test test test")
	checkEntry(t, feed.Entries[2], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, "", "test test test")
	checkEntry(t, feed.Entries[3], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")

	req := require.New(t)
	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	checkLoadMyTlog(t, userIDs[1], 10, "", "", 0)

	feed = checkLoadMyTlog(t, userIDs[0], 1, "", "", 1)

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadMyTlog(t, userIDs[0], 4, feed.NextBefore, "", 3)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyMe, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, "", "test test test")
	checkEntry(t, feed.Entries[2], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")
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

	postEntry(userIDs[0], models.EntryPrivacyAll)
	postEntry(userIDs[0], models.EntryPrivacySome)
	postEntry(userIDs[0], models.EntryPrivacyMe)
	postEntry(userIDs[0], models.EntryPrivacyAll)

	postEntry(userIDs[1], models.EntryPrivacyAll)
	postEntry(userIDs[1], models.EntryPrivacySome)
	postEntry(userIDs[1], models.EntryPrivacyMe)

	postEntry(userIDs[2], models.EntryPrivacyAll)
	postEntry(userIDs[2], models.EntryPrivacySome)
	postEntry(userIDs[2], models.EntryPrivacyMe)

	feed := checkLoadFriendsFeed(t, userIDs[0], 10, "", "", 4)
	checkEntry(t, feed.Entries[0], profiles[1], false, models.RatingVoteNot, false, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed.Entries[2], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, "", "test test test")
	checkEntry(t, feed.Entries[3], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")

	req := require.New(t)
	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadFriendsFeed(t, userIDs[1], 10, "", "", 2)
	checkEntry(t, feed.Entries[0], profiles[1], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[1], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadFriendsFeed(t, userIDs[0], 1, "", "", 1)

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadFriendsFeed(t, userIDs[0], 4, feed.NextBefore, "", 3)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, "", "test test test")
	checkEntry(t, feed.Entries[2], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")

	checkUnfollow(t, userIDs[0], userIDs[1])
}

func checkLoadFavorites(t *testing.T, tlog, user *models.UserID, limit int64, before, after string, size int) *models.Feed {
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
		t.Fatal("error load tlog")
	}

	feed := body.Payload
	require.Equal(t, size, len(feed.Entries))

	return feed
}

/*
func TestLoadFavorites(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)

	postEntry(userIDs[0], models.EntryPrivacyAll)
	postEntry(userIDs[0], models.EntryPrivacySome)
	postEntry(userIDs[0], models.EntryPrivacyMe)
	postEntry(userIDs[0], models.EntryPrivacyAll)

	tlog := checkLoadMyTlog(t, userIDs[0], 10, "", "", 4)

	checkFavoriteEntry(t, userIDs[0], tlog.Entries[0].ID, true)
	checkFavoriteEntry(t, userIDs[0], tlog.Entries[1].ID, true)
	checkFavoriteEntry(t, userIDs[0], tlog.Entries[2].ID, true)

	feed := checkLoadFavorites(t, userIDs[0], userIDs[0], 10, "", "", 2)

	req := require.New(t)
	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadFavorites(t, userIDs[0], userIDs[0], 10, "", "", 3)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyMe, true, "", "test test test")
	checkEntry(t, feed.Entries[2], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, "", "test test test")

	checkLoadFavorites(t, userIDs[1], userIDs[0], 10, "", "", 0)
	checkLoadFavorites(t, userIDs[0], userIDs[1], 10, "", "", 1)

	feed = checkLoadFavorites(t, userIDs[0], userIDs[0], 2, "", "", 2)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed.Entries[1], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacyMe, true, "", "test test test")

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadFavorites(t, userIDs[0], userIDs[0], 3, feed.NextBefore, "", 1)
	checkEntry(t, feed.Entries[0], profiles[0], true, models.RatingVoteBan, true, 3, models.EntryPrivacySome, true, "", "test test test")

	req.False(feed.HasBefore)
	req.True(feed.HasAfter)
}
*/
