package test

import (
	"fmt"
	"log"
	"strings"
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
	user *models.AuthProfile, canEdit bool, vote int64, watching bool,
	wc int64, privacy string, votable, live bool, title, content string, tags []string) {

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

	realTags := make([]string, 0, len(tags))
tagLoop:
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		tag = strings.ToLower(tag)
		if tag == "" {
			continue
		}

		for _, realTag := range realTags {
			if tag == realTag {
				continue tagLoop
			}
		}

		realTags = append(realTags, tag)
	}
	if len(realTags) > 0 {
		req.Equal(realTags, entry.Tags)
	} else {
		req.Nil(entry.Tags)
	}

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

	rights := entry.Rights
	req.Equal(canEdit, rights.Edit)
	req.Equal(canEdit, rights.Delete)
	req.Equal(true, rights.Comment)
	req.Equal(!canEdit && rating.IsVotable, rights.Vote)
	req.Equal(!canEdit, rights.Complain)
}

func checkLoadEntry(t *testing.T, entryID int64, userID *models.UserID, success bool,
	user *models.AuthProfile, canEdit bool, vote int64, watching bool,
	wc int64, privacy string, votable, live bool, title, content string, tags []string) {

	load := api.EntriesGetEntriesIDHandler.Handle
	resp := load(entries.GetEntriesIDParams{ID: entryID}, userID)
	body, ok := resp.(*entries.GetEntriesIDOK)
	require.Equal(t, success, ok)
	if !success {
		return
	}

	entry := body.Payload
	checkEntry(t, entry, user, canEdit, vote, watching, wc, privacy, votable, live, title, content, tags)
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
	checkEntry(t, entry, user, true, 0, true, wc, params.Privacy, *params.IsVotable, *params.InLive,
		*params.Title, params.Content, params.Tags)

	checkLoadEntry(t, entry.ID, id, true, user,
		true, 0, true, wc, params.Privacy, *params.IsVotable, *params.InLive,
		*params.Title, params.Content, params.Tags)

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
	checkEntry(t, entry, user, true, 0, true, wc, params.Privacy, *params.IsVotable, *params.InLive,
		*params.Title, params.Content, params.Tags)

	checkLoadEntry(t, entry.ID, id, true, user,
		true, 0, true, wc, params.Privacy, *params.IsVotable, *params.InLive,
		*params.Title, params.Content, params.Tags)
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

	params.Tags = []string{"tag1", "tag2"}

	id := checkPostEntry(t, params, profiles[0], userIDs[0], true, 5)
	checkEntryWatching(t, userIDs[0], id, true, true)

	req := require.New(t)
	idSame := checkPostEntry(t, params, profiles[0], userIDs[0], true, 5)
	req.Equal(id, idSame)

	votable = true
	checkPostEntry(t, params, profiles[0], userIDs[0], false, 5)

	votable = false
	checkPostEntry(t, params, profiles[3], userIDs[3], false, 5)
	votable = true
	id2 := checkPostEntry(t, params, profiles[3], userIDs[3], true, 5)

	var images []int64
	images = append(images, createImage(srv, db, userIDs[1]).ID)
	images = append(images, createImage(srv, db, userIDs[1]).ID)
	images = append(images, createImage(srv, db, userIDs[1]).ID)

	params.Images = images
	checkPostEntry(t, params, profiles[0], userIDs[0], false, 5)
	id3 := checkPostEntry(t, params, profiles[1], userIDs[1], true, 5)

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
		Tags:      []string{"tag1", "tag3"},
	}

	checkEditEntry(t, editParams, profiles[0], userIDs[0], true, 2)

	checkLoadEntry(t, id, userIDs[1], false, nil, false, 0, false, 0, "", false, false, "", "", []string{})

	editParams.ID = id2
	editParams.Privacy = models.EntryPrivacyAll
	checkEditEntry(t, editParams, profiles[3], userIDs[3], false, 2)
	votable = true
	checkEditEntry(t, editParams, profiles[3], userIDs[3], true, 2)

	images = images[1:]
	images = append(images, createImage(srv, db, userIDs[1]).ID)
	editParams.ID = id3
	editParams.Images = images
	checkEditEntry(t, editParams, profiles[1], userIDs[1], true, 2)

	checkDeleteEntry(t, id, userIDs[1], false)
	checkDeleteEntry(t, id, userIDs[0], true)
	checkDeleteEntry(t, id, userIDs[0], false)

	checkDeleteEntry(t, id2, userIDs[3], true)
	checkDeleteEntry(t, id3, userIDs[1], true)
}

func TestLiveRestrictions(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)

	userIDs[0].FollowersCount = 4
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

	postParams.Content = "test test test2"
	checkPostEntry(t, postParams, profiles[0], userIDs[0], true, 3)

	postParams.Content = "test test test3"
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

	banLive(db, userIDs[0])
	editParams.ID = e0
	checkEditEntry(t, editParams, profiles[0], userIDs[0], true, 1)
	live = false
	checkEditEntry(t, editParams, profiles[0], userIDs[0], true, 1)
	editParams.ID = e1
	checkEditEntry(t, editParams, profiles[0], userIDs[0], true, 1)
	live = true
	checkPostEntry(t, postParams, profiles[0], userIDs[0], false, 3)
	checkEditEntry(t, editParams, profiles[0], userIDs[0], false, 1)

	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)
}

func postEntry(id *models.UserID, privacy string, live bool) *models.Entry {
	votable := true
	title := ""
	params := me.PostMeTlogParams{
		Content:   "test test test" + utils.GenerateString(6),
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

func checkLoadLive(t *testing.T, id *models.UserID, limit int64, section, before, after, tag string, size int) *models.Feed {
	params := entries.GetEntriesLiveParams{
		Limit:   &limit,
		Before:  &before,
		After:   &after,
		Section: &section,
		Tag:     &tag,
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
	esm.Clear()

	e3 := postEntry(userIDs[3], models.EntryPrivacyAll, true)
	e2 := postEntry(userIDs[0], models.EntryPrivacyAll, true)
	postEntry(userIDs[0], models.EntryPrivacyAll, false)
	postEntry(userIDs[0], models.EntryPrivacySome, true)
	postEntry(userIDs[1], models.EntryPrivacyMe, true)
	e1 := postEntry(userIDs[1], models.EntryPrivacyAll, true)
	postEntry(userIDs[2], models.EntryPrivacyAll, false)
	e0 := postEntry(userIDs[2], models.EntryPrivacyAll, true)

	feed := checkLoadLive(t, userIDs[0], 10, "entries", "", "", "", 3)
	compareEntries(t, e0, feed.Entries[0], userIDs[0])
	compareEntries(t, e1, feed.Entries[1], userIDs[0])
	compareEntries(t, e2, feed.Entries[2], userIDs[0])

	req := require.New(t)
	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 1, "entries", "", "", "", 1)
	compareEntries(t, e0, feed.Entries[0], userIDs[0])

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 5, "entries", feed.NextBefore, "", "", 2)
	compareEntries(t, e1, feed.Entries[0], userIDs[0])
	compareEntries(t, e2, feed.Entries[1], userIDs[0])

	req.False(feed.HasBefore)
	req.True(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 2, "entries", "", "", "", 2)
	compareEntries(t, e0, feed.Entries[0], userIDs[0])
	compareEntries(t, e1, feed.Entries[1], userIDs[0])

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 5, "entries", feed.NextBefore, "", "", 1)
	compareEntries(t, e2, feed.Entries[0], userIDs[0])

	req.False(feed.HasBefore)
	req.True(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 5, "entries", "", feed.NextAfter, "", 2)
	compareEntries(t, e0, feed.Entries[0], userIDs[0])
	compareEntries(t, e1, feed.Entries[1], userIDs[0])

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	checkLoadLive(t, userIDs[0], 1, "entries", "", feed.NextAfter, "", 0)
	checkLoadLive(t, userIDs[0], 0, "entries", "", feed.NextAfter, "", 0)

	feed = checkLoadLive(t, userIDs[0], 10, "waiting", "", "", "", 1)
	compareEntries(t, e3, feed.Entries[0], userIDs[0])

	setUserPrivacy(t, userIDs[0], "invited")
	feed = checkLoadLive(t, userIDs[3], 10, "entries", "", "", "", 2)
	compareEntries(t, e0, feed.Entries[0], userIDs[3])
	compareEntries(t, e1, feed.Entries[1], userIDs[3])
	setUserPrivacy(t, userIDs[0], "all")

	checkFollow(t, userIDs[0], userIDs[2], profiles[2], models.RelationshipRelationIgnored, true)

	feed = checkLoadLive(t, userIDs[2], 10, "entries", "", "", "", 2)
	compareEntries(t, e0, feed.Entries[0], userIDs[2])
	compareEntries(t, e1, feed.Entries[1], userIDs[2])

	feed = checkLoadLive(t, userIDs[0], 10, "entries", "", "", "", 2)
	compareEntries(t, e1, feed.Entries[0], userIDs[0])
	compareEntries(t, e2, feed.Entries[1], userIDs[0])

	checkUnfollow(t, userIDs[0], userIDs[2])

	checkFollow(t, userIDs[0], userIDs[2], profiles[2], models.RelationshipRelationHidden, true)

	feed = checkLoadLive(t, userIDs[2], 10, "entries", "", "", "", 3)
	compareEntries(t, e0, feed.Entries[0], userIDs[2])
	compareEntries(t, e1, feed.Entries[1], userIDs[2])
	compareEntries(t, e2, feed.Entries[2], userIDs[2])

	feed = checkLoadLive(t, userIDs[0], 10, "entries", "", "", "", 2)
	compareEntries(t, e1, feed.Entries[0], userIDs[0])
	compareEntries(t, e2, feed.Entries[1], userIDs[0])

	checkUnfollow(t, userIDs[0], userIDs[2])
}

func checkLoadTlog(t *testing.T, tlog, user *models.UserID, success bool, limit int64, before, after, tag, sort string, size int) *models.Feed {
	params := users.GetUsersNameTlogParams{
		Name:   tlog.Name,
		Limit:  &limit,
		Before: &before,
		After:  &after,
		Tag:    &tag,
		Sort:   &sort,
	}

	load := api.UsersGetUsersNameTlogHandler.Handle
	resp := load(params, user)
	body, ok := resp.(*users.GetUsersNameTlogOK)
	require.Equal(t, success, ok)
	if !ok {
		return nil
	}

	feed := body.Payload
	require.Equal(t, size, len(feed.Entries))

	return feed
}

func TestLoadTlog(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)
	esm.Clear()

	e3 := postEntry(userIDs[0], models.EntryPrivacyAll, true)
	e2 := postEntry(userIDs[0], models.EntryPrivacySome, true)
	e1 := postEntry(userIDs[0], models.EntryPrivacyMe, true)
	e0 := postEntry(userIDs[0], models.EntryPrivacyAll, false)

	feed := checkLoadTlog(t, userIDs[0], userIDs[1], true, 10, "", "", "", "new", 2)
	compareEntries(t, e0, feed.Entries[0], userIDs[1])
	compareEntries(t, e3, feed.Entries[1], userIDs[1])

	req := require.New(t)
	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadTlog(t, userIDs[0], userIDs[0], true, 10, "", "", "", "new", 4)
	compareEntries(t, e0, feed.Entries[0], userIDs[0])
	compareEntries(t, e1, feed.Entries[1], userIDs[0])
	compareEntries(t, e2, feed.Entries[2], userIDs[0])
	compareEntries(t, e3, feed.Entries[3], userIDs[0])

	checkLoadTlog(t, userIDs[1], userIDs[0], true, 10, "", "", "", "new", 0)

	feed = checkLoadTlog(t, userIDs[0], userIDs[0], true, 3, "", "", "", "new", 3)
	compareEntries(t, e0, feed.Entries[0], userIDs[0])
	compareEntries(t, e1, feed.Entries[1], userIDs[0])
	compareEntries(t, e2, feed.Entries[2], userIDs[0])

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadTlog(t, userIDs[0], userIDs[0], true, 3, feed.NextBefore, "", "", "new", 1)
	compareEntries(t, e3, feed.Entries[0], userIDs[0])

	req.False(feed.HasBefore)
	req.True(feed.HasAfter)

	feed = checkLoadTlog(t, userIDs[0], userIDs[0], true, 10, "", feed.NextAfter, "", "old", 3)
	compareEntries(t, e0, feed.Entries[2], userIDs[0])
	compareEntries(t, e1, feed.Entries[1], userIDs[0])
	compareEntries(t, e2, feed.Entries[0], userIDs[0])

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	checkLoadTlog(t, userIDs[0], userIDs[0], true, 10, "", feed.NextAfter, "", "old", 0)
	checkLoadTlog(t, userIDs[0], userIDs[0], true, 10, feed.NextBefore, "", "", "old", 1)

	voteForEntry(userIDs[1], e0.ID, true)
	voteForEntry(userIDs[1], e3.ID, true)
	voteForEntry(userIDs[2], e0.ID, true)

	feed = checkLoadTlog(t, userIDs[0], userIDs[0], true, 10, "", "", "", "best", 4)
	req.Equal(e0.ID, feed.Entries[0].ID)
	req.Equal(e3.ID, feed.Entries[1].ID)
	req.Equal(e1.ID, feed.Entries[2].ID)
	req.Equal(e2.ID, feed.Entries[3].ID)

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	setUserPrivacy(t, userIDs[0], "followers")
	checkLoadTlog(t, userIDs[0], userIDs[1], false, 3, "", "", "", "new", 2)
	checkLoadTlog(t, userIDs[0], userIDs[3], false, 3, "", "", "", "new", 1)

	checkLoadEntry(t, feed.Entries[0].ID, userIDs[3], false, profiles[0], false, 0, false, 0, "", false, false, "", "", []string{})

	checkFollow(t, userIDs[1], userIDs[0], profiles[0], models.RelationshipRelationRequested, true)
	checkPermitFollow(t, userIDs[0], userIDs[1], true)

	checkLoadTlog(t, userIDs[0], userIDs[1], true, 3, "", "", "", "new", 2)

	setUserPrivacy(t, userIDs[0], "invited")
	checkLoadTlog(t, userIDs[0], userIDs[1], true, 3, "", "", "", "new", 2)
	checkLoadTlog(t, userIDs[0], userIDs[3], false, 3, "", "", "", "new", 1)

	checkLoadEntry(t, feed.Entries[0].ID, userIDs[3], false, profiles[0], false, 0, false, 0, "", false, false, "", "", []string{})

	checkFollow(t, userIDs[0], userIDs[1], profiles[1], models.RelationshipRelationIgnored, true)
	checkLoadTlog(t, userIDs[0], userIDs[1], false, 3, "", "", "", "new", 2)
	checkLoadTlog(t, userIDs[0], userIDs[2], true, 3, "", "", "", "new", 2)
	checkLoadTlog(t, userIDs[0], userIDs[3], false, 3, "", "", "", "new", 1)

	checkFollow(t, userIDs[0], userIDs[1], profiles[1], models.RelationshipRelationHidden, true)
	checkLoadTlog(t, userIDs[0], userIDs[1], true, 3, "", "", "", "new", 2)

	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)
	esm.Clear()
}

func checkLoadMyTlog(t *testing.T, user *models.UserID, limit int64, before, after, tag, sort string, size int) *models.Feed {
	params := me.GetMeTlogParams{
		Limit:  &limit,
		Before: &before,
		After:  &after,
		Tag:    &tag,
		Sort:   &sort,
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

	e3 := postEntry(userIDs[0], models.EntryPrivacyAll, true)
	e2 := postEntry(userIDs[0], models.EntryPrivacySome, true)
	e1 := postEntry(userIDs[0], models.EntryPrivacyMe, true)
	e0 := postEntry(userIDs[0], models.EntryPrivacyAll, false)

	feed := checkLoadMyTlog(t, userIDs[0], 10, "", "", "", "new", 4)
	compareEntries(t, e0, feed.Entries[0], userIDs[0])
	compareEntries(t, e1, feed.Entries[1], userIDs[0])
	compareEntries(t, e2, feed.Entries[2], userIDs[0])
	compareEntries(t, e3, feed.Entries[3], userIDs[0])

	req := require.New(t)
	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	checkLoadMyTlog(t, userIDs[1], 10, "", "", "", "new", 0)

	feed = checkLoadMyTlog(t, userIDs[0], 1, "", "", "", "new", 1)

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadMyTlog(t, userIDs[0], 4, feed.NextBefore, "", "", "new", 3)
	compareEntries(t, e1, feed.Entries[0], userIDs[0])
	compareEntries(t, e2, feed.Entries[1], userIDs[0])
	compareEntries(t, e3, feed.Entries[2], userIDs[0])

	feed = checkLoadMyTlog(t, userIDs[0], 10, "", "", "", "old", 4)
	compareEntries(t, e3, feed.Entries[0], userIDs[0])
	compareEntries(t, e2, feed.Entries[1], userIDs[0])
	compareEntries(t, e1, feed.Entries[2], userIDs[0])
	compareEntries(t, e0, feed.Entries[3], userIDs[0])

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)
}

func checkLoadFriendsFeed(t *testing.T, user *models.UserID, limit int64, before, after, tag string, size int) *models.Feed {
	params := entries.GetEntriesFriendsParams{
		Limit:  &limit,
		Before: &before,
		After:  &after,
		Tag:    &tag,
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

	checkFollow(t, userIDs[0], userIDs[1], profiles[1], models.RelationshipRelationFollowed, true)

	ea3 := postEntry(userIDs[0], models.EntryPrivacyAll, true)
	es2 := postEntry(userIDs[0], models.EntryPrivacySome, true)
	postEntry(userIDs[0], models.EntryPrivacyMe, true)
	ea2 := postEntry(userIDs[0], models.EntryPrivacyAll, false)

	ea1 := postEntry(userIDs[1], models.EntryPrivacyAll, true)
	es1 := postEntry(userIDs[1], models.EntryPrivacySome, true)
	postEntry(userIDs[1], models.EntryPrivacyMe, true)

	postEntry(userIDs[2], models.EntryPrivacyAll, true)
	postEntry(userIDs[2], models.EntryPrivacySome, true)
	postEntry(userIDs[2], models.EntryPrivacyMe, true)

	feed := checkLoadFriendsFeed(t, userIDs[0], 10, "", "", "", 4)
	compareEntries(t, ea1, feed.Entries[0], userIDs[0])
	compareEntries(t, ea2, feed.Entries[1], userIDs[0])
	compareEntries(t, es2, feed.Entries[2], userIDs[0])
	compareEntries(t, ea3, feed.Entries[3], userIDs[0])

	req := require.New(t)
	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadFriendsFeed(t, userIDs[1], 10, "", "", "", 2)
	compareEntries(t, es1, feed.Entries[0], userIDs[1])
	compareEntries(t, ea1, feed.Entries[1], userIDs[1])

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadFriendsFeed(t, userIDs[0], 1, "", "", "", 1)

	req.True(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadFriendsFeed(t, userIDs[0], 4, feed.NextBefore, "", "", 3)
	compareEntries(t, ea2, feed.Entries[0], userIDs[0])
	compareEntries(t, es2, feed.Entries[1], userIDs[0])
	compareEntries(t, ea3, feed.Entries[2], userIDs[0])

	checkUnfollow(t, userIDs[0], userIDs[1])

	checkFollow(t, userIDs[3], userIDs[1], profiles[1], models.RelationshipRelationFollowed, true)
	setUserPrivacy(t, userIDs[1], "invited")
	feed = checkLoadFriendsFeed(t, userIDs[3], 10, "", "", "", 0)
	setUserPrivacy(t, userIDs[0], "all")
	checkUnfollow(t, userIDs[3], userIDs[1])
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

func favoriteEntry(user *models.UserID, entryID int64) {
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

	tlog := checkLoadMyTlog(t, userIDs[0], 10, "", "", "", "new", 4)

	favoriteEntry(userIDs[0], tlog.Entries[2].ID)
	favoriteEntry(userIDs[0], tlog.Entries[1].ID)
	favoriteEntry(userIDs[0], tlog.Entries[0].ID)

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

	setUserPrivacy(t, userIDs[1], "invited")
	e4 := postEntry(userIDs[1], models.EntryPrivacyAll, true)
	favoriteEntry(userIDs[0], e4.ID)

	checkLoadFavorites(t, userIDs[0], userIDs[0], 10, "", "", 4)
	checkLoadFavorites(t, userIDs[3], userIDs[0], 10, "", "", 1)

	setUserPrivacy(t, userIDs[1], "all")

	feed = checkLoadTlog(t, userIDs[0], userIDs[1], true, 10, "", "", "", "new", 2)
	favoriteEntry(userIDs[1], feed.Entries[0].ID)
	favoriteEntry(userIDs[1], feed.Entries[1].ID)
	favoriteEntry(userIDs[1], e4.ID)

	checkFollow(t, userIDs[0], userIDs[2], profiles[2], models.RelationshipRelationIgnored, true)
	checkLoadFavorites(t, userIDs[2], userIDs[1], 10, "", "", 1)
	checkUnfollow(t, userIDs[0], userIDs[2])
}

func compareEntries(t *testing.T, exp, act *models.Entry, user *models.UserID) {
	req := require.New(t)

	req.Equal(exp.ID, act.ID)
	req.Equal(exp.Author, act.Author)
	req.Equal(exp.CommentCount, act.CommentCount)
	req.Equal(exp.Content, act.Content)
	req.Equal(exp.CreatedAt, act.CreatedAt)
	req.Equal(exp.CutContent, exp.CutContent)
	req.Equal(exp.CutTitle, act.CutTitle)
	req.Equal(exp.HasCut, act.HasCut)
	req.Equal(exp.InLive, act.InLive)
	req.Equal(exp.Privacy, act.Privacy)
	req.Equal(exp.Title, act.Title)
	req.Equal(exp.VisibleFor, act.VisibleFor)
	req.Equal(exp.WordCount, act.WordCount)

	req.Equal(exp.ID, exp.Rating.ID)
	req.Equal(exp.Rating.ID, act.Rating.ID)
	req.Equal(exp.Rating.DownCount, act.Rating.DownCount)
	req.Equal(exp.Rating.UpCount, act.Rating.UpCount)
	req.Equal(exp.Rating.Rating, act.Rating.Rating)
	req.Equal(exp.Rating.IsVotable, act.Rating.IsVotable)

	if exp.Author.ID == user.ID {
		req.NotEmpty(act.EditContent)
	} else {
		req.Empty(act.EditContent)
	}

	rights := act.Rights
	req.Equal(act.Author.ID == user.ID, rights.Edit)
	req.Equal(act.Author.ID == user.ID, rights.Delete)
	req.Equal(act.Author.ID == user.ID || !user.Ban.Comment, rights.Comment)
	req.Equal(act.Author.ID != user.ID && !user.Ban.Vote && act.Rating.IsVotable, rights.Vote)
}

func TestLoadLiveComments(t *testing.T) {
	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)
	esm.Clear()

	es := make([]*models.Entry, 6)

	es[0] = postEntry(userIDs[0], models.EntryPrivacyAll, true) // 2
	es[1] = postEntry(userIDs[0], models.EntryPrivacyAll, false)
	es[2] = postEntry(userIDs[0], models.EntryPrivacySome, true)
	es[3] = postEntry(userIDs[1], models.EntryPrivacyAll, true) // 1
	es[4] = postEntry(userIDs[1], models.EntryPrivacyAll, true)
	es[5] = postEntry(userIDs[1], models.EntryPrivacyAll, true) // 3

	// skip 4
	comments := make([]int64, 5)

	comments[0] = postComment(userIDs[0], es[5].ID)
	comments[1] = postComment(userIDs[0], es[0].ID)
	comments[2] = postComment(userIDs[0], es[3].ID)
	comments[3] = postComment(userIDs[0], es[1].ID)
	comments[4] = postComment(userIDs[0], es[2].ID)

	for _, e := range es {
		e.CommentCount = 1
		e.EditContent = ""
		e.IsWatching = false
		e.Rating.Vote = 0
	}

	feed := checkLoadLive(t, userIDs[2], 10, "comments", "", "", "", 3)

	compareEntries(t, es[3], feed.Entries[0], userIDs[2])
	compareEntries(t, es[0], feed.Entries[1], userIDs[2])
	compareEntries(t, es[5], feed.Entries[2], userIDs[2])

	req := require.New(t)
	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[2], 1, "comments", "", "", "", 1)
	compareEntries(t, es[3], feed.Entries[0], userIDs[2])

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	checkDeleteComment(t, comments[0], userIDs[0], true)
	checkDeleteComment(t, comments[3], userIDs[0], true)
	checkLoadLive(t, userIDs[2], 10, "comments", "", "", "", 2)

	checkLoadLive(t, userIDs[3], 10, "comments", "", "", "", 2)
	setUserPrivacy(t, userIDs[1], "invited")
	checkLoadLive(t, userIDs[3], 10, "comments", "", "", "", 1)
	setUserPrivacy(t, userIDs[1], "all")

	checkFollow(t, userIDs[0], userIDs[1], profiles[1], models.RelationshipRelationIgnored, true)
	checkLoadLive(t, userIDs[0], 10, "comments", "", "", "", 1)
	checkLoadLive(t, userIDs[1], 10, "comments", "", "", "", 1)
	checkUnfollow(t, userIDs[0], userIDs[1])

	checkFollow(t, userIDs[0], userIDs[1], profiles[1], models.RelationshipRelationHidden, true)
	checkLoadLive(t, userIDs[0], 10, "comments", "", "", "", 1)
	checkLoadLive(t, userIDs[1], 10, "comments", "", "", "", 2)
	checkUnfollow(t, userIDs[0], userIDs[1])
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
	esm.Clear()

	es := make([]*models.Entry, 4)

	es[0] = postEntry(userIDs[0], models.EntryPrivacyAll, true) // 2
	es[1] = postEntry(userIDs[1], models.EntryPrivacyAll, true) // 1
	es[2] = postEntry(userIDs[1], models.EntryPrivacyAll, true)
	es[3] = postEntry(userIDs[1], models.EntryPrivacyAll, true) // 3

	// skip 2
	postComment(userIDs[2], es[3].ID)
	postComment(userIDs[2], es[1].ID)
	postComment(userIDs[2], es[0].ID)
	postComment(userIDs[0], es[1].ID)

	for _, e := range es {
		e.CommentCount = 1
		e.EditContent = ""
		e.IsWatching = true
		e.Rating.Vote = 0
	}

	es[1].CommentCount = 2

	feed := checkLoadWatching(t, userIDs[2], 10, 3)

	req := require.New(t)
	compareEntries(t, es[1], feed.Entries[0], userIDs[2])
	compareEntries(t, es[0], feed.Entries[1], userIDs[2])
	compareEntries(t, es[3], feed.Entries[2], userIDs[2])

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadWatching(t, userIDs[2], 1, 1)
	compareEntries(t, es[1], feed.Entries[0], userIDs[2])

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	postComment(userIDs[3], es[0].ID)
	postComment(userIDs[3], es[1].ID)

	checkLoadWatching(t, userIDs[3], 10, 2)
	setUserPrivacy(t, userIDs[1], "invited")
	checkLoadWatching(t, userIDs[3], 10, 1)
	setUserPrivacy(t, userIDs[1], "all")

	checkFollow(t, userIDs[0], userIDs[2], profiles[2], models.RelationshipRelationIgnored, true)
	checkLoadWatching(t, userIDs[2], 10, 2)
	checkUnfollow(t, userIDs[0], userIDs[2])

	checkFollow(t, userIDs[2], userIDs[1], profiles[1], models.RelationshipRelationIgnored, true)
	checkLoadWatching(t, userIDs[2], 10, 1)
	checkUnfollow(t, userIDs[2], userIDs[1])

	checkFollow(t, userIDs[2], userIDs[1], profiles[1], models.RelationshipRelationHidden, true)
	checkLoadWatching(t, userIDs[2], 10, 3)
	checkUnfollow(t, userIDs[2], userIDs[1])
}

func TestRandomEntry(t *testing.T) {
	req := require.New(t)

	es := make([]*models.Entry, 0, 100)

	for i := 0; i < 100; i++ {
		var privacy string
		if i%20 == 3 {
			privacy = models.EntryPrivacyMe
		} else {
			privacy = models.EntryPrivacyAll
		}

		e := postEntry(userIDs[i%4], privacy, true)

		if i%20 == 13 {
			checkDeleteEntry(t, e.ID, userIDs[i%4], true)
			es = append(es, &models.Entry{})
		} else {
			es = append(es, e)
		}
	}

	load := func(success bool) bool {
		load := api.EntriesGetEntriesRandomHandler.Handle
		resp := load(entries.GetEntriesRandomParams{}, userIDs[0])
		body, ok := resp.(*entries.GetEntriesRandomOK)
		req.Equal(success, ok)
		if !ok {
			return false
		}

		entry := body.Payload

		found := false
		for _, e := range es {
			if e.ID == entry.ID {
				req.True(entry.Privacy == models.EntryPrivacyAll || entry.Author.ID == userIDs[0].ID)
				found = true
				break
			}
		}

		req.True(found)
		return true
	}

	ok := false
	for i := 0; i < 10; i++ {
		ok = ok || load(true)
	}
	req.True(ok)

	for i := 0; i < 100; i++ {
		if es[i].ID == 0 {
			continue
		}

		checkDeleteEntry(t, es[i].ID, userIDs[i%4], true)
	}

	load(false)
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

func TestCanViewEntry(t *testing.T) {
	req := require.New(t)

	check := func(userID, entryID int64, res bool) {
		tx := utils.NewAutoTx(db)
		defer tx.Finish()
		req.Equal(res, utils.CanViewEntry(tx, userID, entryID))
	}

	e1 := createTlogEntry(t, userIDs[0], models.EntryPrivacyAll, true, true)
	e2 := createTlogEntry(t, userIDs[0], models.EntryPrivacyMe, true, true)
	e3 := createTlogEntry(t, userIDs[0], models.EntryPrivacyAnonymous, true, true)

	check(userIDs[0].ID, e1.ID, true)
	check(userIDs[0].ID, e2.ID, true)
	check(userIDs[0].ID, e3.ID, true)

	check(userIDs[1].ID, e1.ID, true)
	check(userIDs[1].ID, e2.ID, false)
	check(userIDs[1].ID, e3.ID, true)

	setUserPrivacy(t, userIDs[0], "followers")

	check(userIDs[0].ID, e1.ID, true)
	check(userIDs[0].ID, e2.ID, true)
	check(userIDs[0].ID, e3.ID, true)

	check(userIDs[1].ID, e1.ID, false)
	check(userIDs[1].ID, e2.ID, false)
	check(userIDs[1].ID, e3.ID, true)

	checkFollow(t, userIDs[1], userIDs[0], profiles[0], models.RelationshipRelationRequested, true)
	checkPermitFollow(t, userIDs[0], userIDs[1], true)

	check(userIDs[1].ID, e1.ID, true)
	check(userIDs[1].ID, e2.ID, false)
	check(userIDs[1].ID, e3.ID, true)

	setUserPrivacy(t, userIDs[0], "invited")

	check(userIDs[1].ID, e1.ID, true)
	check(userIDs[1].ID, e2.ID, false)
	check(userIDs[1].ID, e3.ID, true)

	check(userIDs[2].ID, e1.ID, true)
	check(userIDs[2].ID, e2.ID, false)
	check(userIDs[2].ID, e3.ID, true)

	check(userIDs[3].ID, e1.ID, false)
	check(userIDs[3].ID, e2.ID, false)
	check(userIDs[3].ID, e3.ID, true)

	checkFollow(t, userIDs[0], userIDs[1], profiles[1], models.RelationshipRelationIgnored, true)

	check(userIDs[1].ID, e1.ID, false)
	check(userIDs[1].ID, e2.ID, false)
	check(userIDs[1].ID, e3.ID, true)

	setUserPrivacy(t, userIDs[0], "all")

	check(userIDs[0].ID, e1.ID, true)
	check(userIDs[0].ID, e2.ID, true)
	check(userIDs[0].ID, e3.ID, true)

	check(userIDs[1].ID, e1.ID, false)
	check(userIDs[1].ID, e2.ID, false)
	check(userIDs[1].ID, e3.ID, true)

	check(userIDs[2].ID, e1.ID, true)
	check(userIDs[2].ID, e2.ID, false)
	check(userIDs[2].ID, e3.ID, true)

	check(userIDs[3].ID, e1.ID, true)
	check(userIDs[3].ID, e2.ID, false)
	check(userIDs[3].ID, e3.ID, true)

	utils.ClearDatabase(db)
	userIDs, profiles = registerTestUsers(db)
	esm.Clear()
}

func checkPostTaggedEntry(t *testing.T, user *models.UserID, author *models.AuthProfile, content string, wc int64, tags []string) *models.Entry {
	title := ""
	votable := true
	live := true
	params := me.PostMeTlogParams{
		Content:   content,
		Title:     &title,
		Privacy:   "all",
		IsVotable: &votable,
		InLive:    &live,
		Tags:      tags,
	}

	resp := api.MePostMeTlogHandler.Handle(params, user)
	body, ok := resp.(*me.PostMeTlogCreated)
	require.True(t, ok)

	entry := body.Payload
	checkEntry(t, entry, author, true, 0, true, wc, params.Privacy, *params.IsVotable, *params.InLive,
		*params.Title, params.Content, params.Tags)

	checkLoadEntry(t, entry.ID, user, true, author,
		true, 0, true, wc, params.Privacy, *params.IsVotable, *params.InLive,
		*params.Title, params.Content, params.Tags)

	return body.Payload
}

func checkEditTaggedEntry(t *testing.T, entry *models.Entry, user *models.AuthProfile, id *models.UserID, tags []string) {
	params := entries.PutEntriesIDParams{
		ID:        entry.ID,
		Content:   entry.EditContent,
		InLive:    &entry.InLive,
		IsVotable: &entry.Rating.IsVotable,
		Privacy:   entry.Privacy,
		Tags:      tags,
		Title:     &entry.Title,
	}

	edit := api.EntriesPutEntriesIDHandler.Handle
	resp := edit(params, id)
	body, ok := resp.(*entries.PutEntriesIDOK)
	require.True(t, ok)

	edited := body.Payload
	checkEntry(t, edited, user, true, 0, true, entry.WordCount, params.Privacy, *params.IsVotable, *params.InLive,
		*params.Title, params.Content, params.Tags)

	checkLoadEntry(t, entry.ID, id, true, user,
		true, 0, true, entry.WordCount, params.Privacy, *params.IsVotable, *params.InLive,
		*params.Title, params.Content, params.Tags)
}

func TestEntryTags(t *testing.T) {
	e2 := checkPostTaggedEntry(t, userIDs[0], profiles[0], "test test test2", 3, []string{"aaa", "bbb"})
	e1 := checkPostTaggedEntry(t, userIDs[1], profiles[1], "test test test1", 3, []string{" aaa  ", " ccc", "  ", ""})
	e0 := checkPostTaggedEntry(t, userIDs[0], profiles[0], "test test test0", 3, []string{"bbb", "bbb"})

	req := require.New(t)
	req.NotEqual(e2.ID, e1.ID)
	req.NotEqual(e2.ID, e0.ID)
	req.NotEqual(e1.ID, e0.ID)

	feed := checkLoadLive(t, userIDs[0], 10, "entries", "", "", "", 3)

	compareEntries(t, e0, feed.Entries[0], userIDs[0])
	compareEntries(t, e1, feed.Entries[1], userIDs[0])
	compareEntries(t, e2, feed.Entries[2], userIDs[0])

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 10, "entries", "", "", "aaa", 2)

	compareEntries(t, e1, feed.Entries[0], userIDs[0])
	compareEntries(t, e2, feed.Entries[1], userIDs[0])

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 10, "entries", "", "", "bbb", 2)

	compareEntries(t, e0, feed.Entries[0], userIDs[0])
	compareEntries(t, e2, feed.Entries[1], userIDs[0])

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadLive(t, userIDs[0], 10, "entries", "", "", "ccc", 1)

	compareEntries(t, e1, feed.Entries[0], userIDs[0])

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadTlog(t, userIDs[0], userIDs[1], true, 10, "", "", "", "new", 2)

	compareEntries(t, e0, feed.Entries[0], userIDs[1])
	compareEntries(t, e2, feed.Entries[1], userIDs[1])

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadTlog(t, userIDs[0], userIDs[1], true, 10, "", "", "aaa", "new", 1)

	compareEntries(t, e2, feed.Entries[0], userIDs[1])

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	feed = checkLoadTlog(t, userIDs[0], userIDs[1], true, 10, "", "", "test", "new", 0)

	req.False(feed.HasBefore)
	req.False(feed.HasAfter)

	checkEditTaggedEntry(t, e0, profiles[0], userIDs[0], []string{"bbb", "ccc"})
	checkEditTaggedEntry(t, e0, profiles[0], userIDs[0], []string{})

	checkDeleteEntry(t, e0.ID, userIDs[0], true)
	checkDeleteEntry(t, e1.ID, userIDs[1], true)
	checkDeleteEntry(t, e2.ID, userIDs[0], true)
}
