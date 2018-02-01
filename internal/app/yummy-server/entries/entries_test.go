package entries

import (
	"database/sql"
	"os"
	"testing"

	yummy "github.com/sevings/yummy-server/internal/app/yummy-server"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations/entries"
	"github.com/stretchr/testify/require"
)

var db *sql.DB
var userIDs []*models.UserID
var profiles []*models.AuthProfile

func TestMain(m *testing.M) {
	config := yummy.LoadConfig("../../config")
	db = yummy.OpenDatabase(config)
	yummy.ClearDatabase(db)

	userIDs, profiles = yummy.RegisterTestUsers(db)

	os.Exit(m.Run())
}

func checkEntry(t *testing.T, entry *models.Entry,
	user *models.AuthProfile,
	wc int64, privacy string, votable bool, title, content string) {

	req := require.New(t)
	req.NotEmpty(entry.CreatedAt)
	req.Zero(entry.Rating)
	req.Equal(content, entry.Content)
	req.Equal(wc, entry.WordCount)
	req.Equal(privacy, entry.Privacy)
	req.Empty(entry.VisibleFor)
	req.Equal(votable, entry.IsVotable)
	req.Zero(entry.CommentCount)
	req.Equal(models.EntryVoteNot, entry.Vote)
	req.False(entry.IsFavorited)
	req.True(entry.IsWatching)
	req.Empty(entry.Comments)
	req.Equal(title, entry.Title)

	author := entry.Author
	req.Equal(user.ID, author.ID)
	req.Equal(user.Name, author.Name)
	req.Equal(user.ShowName, author.ShowName)
	req.Equal(user.IsOnline, author.IsOnline)
	req.Equal(user.Avatar, author.Avatar)
}

func checkPostEntry(t *testing.T,
	params entries.PostEntriesUsersMeParams,
	user *models.AuthProfile, id *models.UserID,
	wc int64, privacy string, votable bool, title string) {

	post := newMyTlogPoster(db)
	resp := post(params, id)
	body, ok := resp.(*entries.PostEntriesUsersMeOK)
	if !ok {
		badBody, ok := resp.(*entries.PostEntriesUsersMeForbidden)
		if ok {
			t.Fatal(badBody.Payload.Message)
		}

		t.Fatal("error post entry")
	}

	req := require.New(t)

	entry := body.Payload
	checkEntry(t, entry, user, wc, privacy, votable, title, params.Content)
}

func TestPostMyTlog(t *testing.T) {
	params := entries.PostEntriesUsersMeParams{
		Content: "test content",
	}

	checkPostEntry(t, params, profiles[0], userIDs[0], 2, models.EntryPrivacyAll, true, "")

	votable := false
	params.IsVotable = &votable

	privacy := models.EntryPrivacyAnonymous
	params.Privacy = &privacy

	title := "title title ti"
	params.Title = &title

	checkPostEntry(t, params, profiles[0], userIDs[0], 5, privacy, votable, title)
}

func postEntry(id *models.UserID, privacy string) {
	post := newMyTlogPoster(db)
	params := entries.PostEntriesUsersMeParams{
		Content: "test test test",
		Privacy: &privacy,
	}
	post(params, userIDs[0])
}

func TestLoadLive(t *testing.T) {
	yummy.ClearDatabase(db)
	userIDs, profiles = yummy.RegisterTestUsers(db)

	postEntry(userIDs[0], models.EntryPrivacyAll)
	postEntry(userIDs[0], models.EntryPrivacySome)
	postEntry(userIDs[0], models.EntryPrivacyMe)
	postEntry(userIDs[0], models.EntryPrivacyAll)

	params := entries.GetEntriesLiveParams{}

	load := newLiveLoader(db)
	resp := load(params, userIDs[0])
	body, ok := resp.(*entries.GetEntriesLiveOK)
	if !ok {
		t.Fatal("error load live")
	}

	feed := body.Payload.Entries
	require.Equal(t, 2, len(feed))
	checkEntry(t, feed[0], profiles[0], 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed[1], profiles[0], 3, models.EntryPrivacyAll, true, "", "test test test")
}

func loadTlog(t *testing.T, tlog, user *models.UserID) models.FeedEntries {
	params := entries.GetEntriesUsersIDParams{
		ID: int64(*tlog),
	}
	load := newTlogLoader(db)
	resp := load(params, user)
	body, ok := resp.(*entries.GetEntriesUsersIDOK)
	if !ok {
		t.Fatal("error load tlog")
	}

	return body.Payload.Entries
}

func TestLoadTlog(t *testing.T) {
	yummy.ClearDatabase(db)
	userIDs, profiles = yummy.RegisterTestUsers(db)

	postEntry(userIDs[0], models.EntryPrivacyAll)
	postEntry(userIDs[0], models.EntryPrivacySome)
	postEntry(userIDs[0], models.EntryPrivacyMe)
	postEntry(userIDs[0], models.EntryPrivacyAll)

	feed := loadTlog(t, userIDs[0], userIDs[1])
	require.Equal(t, 2, len(feed))
	checkEntry(t, feed[0], profiles[0], 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed[1], profiles[0], 3, models.EntryPrivacyAll, true, "", "test test test")

	feed = loadTlog(t, userIDs[0], userIDs[0])
	require.Equal(t, 4, len(feed))
	checkEntry(t, feed[0], profiles[0], 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed[1], profiles[0], 3, models.EntryPrivacySome, true, "", "test test test")
	checkEntry(t, feed[2], profiles[0], 3, models.EntryPrivacyMe, true, "", "test test test")
	checkEntry(t, feed[3], profiles[0], 3, models.EntryPrivacyAll, true, "", "test test test")

	feed = loadTlog(t, userIDs[1], userIDs[0])
	require.Empty(t, feed)
}

func loadMyTlog(t *testing.T, user *models.UserID) models.FeedEntries {
	params := entries.GetEntriesUsersMeParams{}
	load := newMyTlogLoader(db)
	resp := load(params, user)
	body, ok := resp.(*entries.GetEntriesUsersMeOK)
	if !ok {
		t.Fatal("error load tlog")
	}

	return body.Payload.Entries
}

func TestLoadMyTlog(t *testing.T) {
	yummy.ClearDatabase(db)
	userIDs, profiles = yummy.RegisterTestUsers(db)

	postEntry(userIDs[0], models.EntryPrivacyAll)
	postEntry(userIDs[0], models.EntryPrivacySome)
	postEntry(userIDs[0], models.EntryPrivacyMe)
	postEntry(userIDs[0], models.EntryPrivacyAll)

	feed := loadMyTlog(t, userIDs[0])
	require.Equal(t, 4, len(feed))
	checkEntry(t, feed[0], profiles[0], 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(t, feed[1], profiles[0], 3, models.EntryPrivacySome, true, "", "test test test")
	checkEntry(t, feed[2], profiles[0], 3, models.EntryPrivacyMe, true, "", "test test test")
	checkEntry(t, feed[3], profiles[0], 3, models.EntryPrivacyAll, true, "", "test test test")

	feed = loadMyTlog(t, userIDs[1])
	require.Empty(t, feed)
}
