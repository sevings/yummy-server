package entries

import (
	"database/sql"
	"os"
	"testing"

	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations/entries"
	"github.com/sevings/yummy-server/gen/restapi/operations/me"
	yummy "github.com/sevings/yummy-server/src"
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

func checkEntry(entry *models.Entry,
	user *models.AuthProfile,
	wc int64, privacy string, votable bool, title, content string) {

	req.NotEmpty(entry.CreatedAt)
	req.Zero(entry.Rating)
	req.Equal(params.Content, entry.Content)
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

func checkPostEntry(params me.PostUsersMeEntriesParams,
	user *models.AuthProfile, id *models.UserID,
	wc int64, privacy string, votable bool, title string) {

	post := newMyTlogPoster(db)
	resp := post(params, id)
	body, ok := resp.(*me.PostUsersMeEntriesOK)
	if !ok {
		badBody, ok := resp.(*me.PostUsersMeEntriesForbidden)
		if ok {
			t.Fatal(badBody.Payload.Message)
		}

		t.Fatal("error post entry")
	}

	req := require.New(t)

	entry := body.Payload
	checkEntry(entry, user, wc, privacy, votable, title, params.Content)
}

func TestPostMyTlog(t *testing.T) {
	params := me.PostUsersMeEntriesParams{
		Content: "test content",
	}

	checkPostEntry(params, id, 2, models.EntryPrivacyAll, true, "")

	votable := false
	params.IsVotable = &votable

	privacy := models.EntryPrivacyAnonymous
	params.Privacy = &privacy

	title := "title title ti"
	params.Title = &title

	checkPostEntry(params, profiles[0], userIDs[0], 5, privacy, votable, title)
}

func postEntry(id *models.UserID, privacy string) {
	post := newMyTlogPoster(db)
	entryParams := me.PostUsersMeEntriesParams{
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
	checkEntry(feed[0], profiles[0], 3, models.EntryPrivacyAll, true, "", "test test test")
	checkEntry(feed[1], profiles[0], 3, models.EntryPrivacyAll, true, "", "test test test")
}
