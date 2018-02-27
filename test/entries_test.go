package test

import (
	"testing"

	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations/entries"
	"github.com/stretchr/testify/require"
)

func checkEntry(t *testing.T, entry *models.Entry,
	user *models.AuthProfile, canEdit bool, vote string, watching bool,
	wc int64, privacy string, votable bool, title, content string) {

	req := require.New(t)
	req.NotEmpty(entry.CreatedAt)
	req.Zero(entry.Rating)
	req.Equal("<p>"+content+"</p>\n", entry.Content)
	req.Equal(wc, entry.WordCount)
	req.Equal(privacy, entry.Privacy)
	req.Empty(entry.VisibleFor)
	req.Equal(votable, entry.IsVotable)
	req.Zero(entry.CommentCount)
	req.Equal(vote, entry.Vote)
	req.False(entry.IsFavorited)
	req.Equal(watching, entry.IsWatching)
	req.Empty(entry.Comments)
	req.Equal(title, entry.Title)

	if canEdit {
		req.Equal(content, entry.EditContent)
	} else {
		req.Empty(entry.EditContent)
	}

	author := entry.Author
	req.Equal(user.ID, author.ID)
	req.Equal(user.Name, author.Name)
	req.Equal(user.ShowName, author.ShowName)
	req.Equal(user.IsOnline, author.IsOnline)
	req.Equal(user.Avatar, author.Avatar)
}

func checkPostEntry(t *testing.T,
	params entries.PostEntriesUsersMeParams,
	user *models.AuthProfile, id *models.UserID, wc int64) int64 {

	post := api.EntriesPostEntriesUsersMeHandler.Handle
	resp := post(params, id)
	body, ok := resp.(*entries.PostEntriesUsersMeOK)
	if !ok {
		badBody, ok := resp.(*entries.PostEntriesUsersMeForbidden)
		if ok {
			t.Fatal(badBody.Payload.Message)
		}

		t.Fatal("error post entry")
	}

	entry := body.Payload
	checkEntry(t, entry, user, true, models.EntryVoteBan, true, wc, params.Privacy, *params.IsVotable, *params.Title, params.Content)

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
	checkEntry(t, entry, user, true, models.EntryVoteBan, true, wc, params.Privacy, *params.IsVotable, *params.Title, params.Content)
}

func TestPostMyTlog(t *testing.T) {
	params := entries.PostEntriesUsersMeParams{
		Content: "test content",
	}

	votable := false
	params.IsVotable = &votable

	privacy := models.EntryPrivacyAll
	params.Privacy = privacy

	title := "title title ti"
	params.Title = &title

	id := checkPostEntry(t, params, profiles[0], userIDs[0], 5)

	privacy = models.EntryPrivacyMe
	title = "title"
	editParams := entries.PutEntriesIDParams{
		ID:        id,
		Content:   "content",
		Title:     &title,
		IsVotable: &votable,
		Privacy:   privacy,
	}

	checkEditEntry(t, editParams, profiles[0], userIDs[0], 2)
}

func postEntry(id *models.UserID, privacy string) {
	post := api.EntriesPostEntriesUsersMeHandler.Handle
	votable := true
	title := ""
	params := entries.PostEntriesUsersMeParams{
		Content:   "test test test",
		Title:     &title,
		Privacy:   privacy,
		IsVotable: &votable,
	}
	post(params, id)
}

func checkLoadLive(t *testing.T, id *models.UserID, limit, skip int64, size int) models.FeedEntries {
	params := entries.GetEntriesLiveParams{
		Limit: &limit,
		Skip:  &skip,
	}

	load := api.EntriesGetEntriesLiveHandler.Handle
	resp := load(params, id)
	body, ok := resp.(*entries.GetEntriesLiveOK)
	if !ok {
		t.Fatal("error load live")
	}

	feed := body.Payload.Entries
	require.Equal(t, size, len(feed))

	return feed
}

func TestLoadLive(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)

	postEntry(userIDs[0], models.EntryPrivacyAll)
	postEntry(userIDs[0], models.EntryPrivacySome)
	postEntry(userIDs[1], models.EntryPrivacyMe)
	postEntry(userIDs[2], models.EntryPrivacyAll)

	feed := checkLoadLive(t, userIDs[0], 10, 0, 2)
	checkEntry(t, feed[0], profiles[2], false, models.EntryVoteNot, false, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed[1], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")

	feed = checkLoadLive(t, userIDs[0], 1, 0, 1)
	checkEntry(t, feed[0], profiles[2], false, models.EntryVoteNot, false, 3, models.EntryPrivacyAll, true, "", "test test test")

	feed = checkLoadLive(t, userIDs[0], 1, 1, 1)
	checkEntry(t, feed[0], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")

	checkLoadLive(t, userIDs[0], 1, 2, 0)
	checkLoadLive(t, userIDs[0], 10, 200, 0)
	checkLoadLive(t, userIDs[0], 0, 2, 0)
}

func checkLoadTlog(t *testing.T, tlog, user *models.UserID, limit, skip int64, size int) models.FeedEntries {
	params := entries.GetEntriesUsersIDParams{
		ID:    int64(*tlog),
		Limit: &limit,
		Skip:  &skip,
	}

	load := api.EntriesGetEntriesUsersIDHandler.Handle
	resp := load(params, user)
	body, ok := resp.(*entries.GetEntriesUsersIDOK)
	if !ok {
		t.Fatal("error load tlog")
	}

	feed := body.Payload.Entries
	require.Equal(t, size, len(feed))

	return feed
}

func TestLoadTlog(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)

	postEntry(userIDs[0], models.EntryPrivacyAll)
	postEntry(userIDs[0], models.EntryPrivacySome)
	postEntry(userIDs[0], models.EntryPrivacyMe)
	postEntry(userIDs[0], models.EntryPrivacyAll)

	feed := checkLoadTlog(t, userIDs[0], userIDs[1], 10, 0, 2)
	checkEntry(t, feed[0], profiles[0], false, models.EntryVoteNot, false, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed[1], profiles[0], false, models.EntryVoteNot, false, 3, models.EntryPrivacyAll, true, "", "test test test")

	feed = checkLoadTlog(t, userIDs[0], userIDs[0], 10, 0, 4)
	checkEntry(t, feed[0], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed[1], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacyMe, true, "", "test test test")
	checkEntry(t, feed[2], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacySome, true, "", "test test test")
	checkEntry(t, feed[3], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")

	checkLoadTlog(t, userIDs[1], userIDs[0], 10, 0, 0)

	feed = checkLoadTlog(t, userIDs[0], userIDs[0], 3, 0, 3)
	checkEntry(t, feed[0], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed[1], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacyMe, true, "", "test test test")
	checkEntry(t, feed[2], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacySome, true, "", "test test test")

	feed = checkLoadTlog(t, userIDs[0], userIDs[0], 3, 3, 1)
	checkEntry(t, feed[0], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")
}

func checkLoadMyTlog(t *testing.T, user *models.UserID, limit, skip int64, size int) models.FeedEntries {
	params := entries.GetEntriesUsersMeParams{
		Limit: &limit,
		Skip:  &skip,
	}

	load := api.EntriesGetEntriesUsersMeHandler.Handle
	resp := load(params, user)
	body, ok := resp.(*entries.GetEntriesUsersMeOK)
	if !ok {
		t.Fatal("error load tlog")
	}

	feed := body.Payload.Entries
	require.Equal(t, size, len(feed))

	return feed
}

func TestLoadMyTlog(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)

	postEntry(userIDs[0], models.EntryPrivacyAll)
	postEntry(userIDs[0], models.EntryPrivacySome)
	postEntry(userIDs[0], models.EntryPrivacyMe)
	postEntry(userIDs[0], models.EntryPrivacyAll)

	feed := checkLoadMyTlog(t, userIDs[0], 10, 0, 4)
	checkEntry(t, feed[0], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed[1], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacyMe, true, "", "test test test")
	checkEntry(t, feed[2], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacySome, true, "", "test test test")
	checkEntry(t, feed[3], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")

	checkLoadMyTlog(t, userIDs[1], 10, 0, 0)

	feed = checkLoadMyTlog(t, userIDs[0], 4, 1, 3)
	checkEntry(t, feed[0], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacyMe, true, "", "test test test")
	checkEntry(t, feed[1], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacySome, true, "", "test test test")
	checkEntry(t, feed[2], profiles[0], true, models.EntryVoteBan, true, 3, models.EntryPrivacyAll, true, "", "test test test")
}
