package test

import (
	"testing"
	"time"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/comments"
	"github.com/sevings/mindwell-server/utils"
	"github.com/stretchr/testify/require"
)

func checkComment(t *testing.T, cmt *models.Comment, entry *models.Entry, userID *models.UserID, author *models.AuthProfile, content string) {
	req := require.New(t)

	req.Equal(entry.ID, cmt.EntryID)
	req.Equal("<p>"+content+"</p>", cmt.Content)
	req.Equal(content, cmt.EditContent)

	req.Equal(author.ID, cmt.Author.ID)
	req.Equal(author.Name, cmt.Author.Name)
	req.Equal(author.ShowName, cmt.Author.ShowName)
	req.Equal(author.IsOnline, cmt.Author.IsOnline)
	req.Equal(author.Avatar, cmt.Author.Avatar)

	req.Equal(cmt.ID, cmt.Rating.ID)
	req.True(cmt.Rating.IsVotable)
	req.Zero(cmt.Rating.Rating)
	req.Zero(cmt.Rating.UpCount)
	req.Zero(cmt.Rating.DownCount)
	req.Zero(cmt.Rating.Vote)

	req.Equal(userID.ID == author.ID, cmt.Rights.Edit)
	req.Equal(userID.ID == author.ID, cmt.Rights.Delete || userID.ID == entry.Author.ID)
	req.Equal(userID.ID != author.ID && !userID.Ban.Vote, cmt.Rights.Vote)
	req.Equal(userID.ID != author.ID, cmt.Rights.Complain)
}

func checkLoadComment(t *testing.T, commentID int64, userID *models.UserID, success bool,
	author *models.AuthProfile, entry *models.Entry, content string) {

	load := api.CommentsGetCommentsIDHandler.Handle
	resp := load(comments.GetCommentsIDParams{ID: commentID}, userID)
	body, ok := resp.(*comments.GetCommentsIDOK)
	require.Equal(t, success, ok)
	if !success {
		return
	}

	cmt := body.Payload
	checkComment(t, cmt, entry, userID, author, content)
}

func checkPostComment(t *testing.T,
	entry *models.Entry, content string, success bool,
	author *models.AuthProfile, id *models.UserID) int64 {

	params := comments.PostEntriesIDCommentsParams{
		ID:      entry.ID,
		Content: content,
	}

	post := api.CommentsPostEntriesIDCommentsHandler.Handle
	resp := post(params, id)
	body, ok := resp.(*comments.PostEntriesIDCommentsCreated)
	require.Equal(t, success, ok)
	if !success {
		return 0
	}

	cmt := body.Payload
	checkComment(t, cmt, entry, id, author, params.Content)

	checkLoadComment(t, cmt.ID, id, true, author, entry, params.Content)

	return cmt.ID
}

func checkEditComment(t *testing.T,
	commentID int64, content string, entry *models.Entry, success bool,
	author *models.AuthProfile, id *models.UserID) {

	params := comments.PutCommentsIDParams{
		ID:      commentID,
		Content: content,
	}

	edit := api.CommentsPutCommentsIDHandler.Handle
	resp := edit(params, id)
	body, ok := resp.(*comments.PutCommentsIDOK)
	require.Equal(t, success, ok)
	if !success {
		return
	}

	cmt := body.Payload
	checkComment(t, cmt, entry, id, author, content)

	checkLoadComment(t, commentID, id, true, author, entry, content)
}

func checkDeleteComment(t *testing.T, commentID int64, userID *models.UserID, success bool) {
	del := api.CommentsDeleteCommentsIDHandler.Handle
	resp := del(comments.DeleteCommentsIDParams{ID: commentID}, userID)
	_, ok := resp.(*comments.DeleteCommentsIDOK)
	require.Equal(t, success, ok)
}

func TestOpenComments(t *testing.T) {
	postEntry(userIDs[0], models.EntryPrivacyAll, true)
	feed := checkLoadTlog(t, userIDs[0], userIDs[0], true, 10, "", "", "", "new", 1)
	entry := feed.Entries[0]

	var id int64

	id = checkPostComment(t, entry, "blabla", true, profiles[0], userIDs[0])
	checkEditComment(t, id, "edited comment", entry, true, profiles[0], userIDs[0])
	checkEntryWatching(t, userIDs[0], entry.ID, true, true)

	id = checkPostComment(t, entry, "blabla", true, profiles[1], userIDs[1])
	checkEditComment(t, id, "edited comment", entry, true, profiles[1], userIDs[1])
	checkEntryWatching(t, userIDs[1], entry.ID, true, true)
	checkDeleteComment(t, id, userIDs[0], true)

	id = checkPostComment(t, entry, "aaaa", true, profiles[1], userIDs[1])
	same := checkPostComment(t, entry, "aaaa", true, profiles[1], userIDs[1])
	require.Equal(t, id, same)

	checkDeleteComment(t, id, userIDs[2], false)
	checkDeleteComment(t, id, userIDs[1], true)
	id = checkPostComment(t, entry, "aaaa", true, profiles[1], userIDs[1])
	require.NotEqual(t, id, same)

	banComment(db, userIDs[0])
	checkPostComment(t, entry, "blabla", true, profiles[0], userIDs[0])
	removeUserRestrictions(db, userIDs)

	banComment(db, userIDs[1])
	checkPostComment(t, entry, "blabla", false, profiles[1], userIDs[1])
	removeUserRestrictions(db, userIDs)

	checkDeleteEntry(t, entry.ID, userIDs[0], true)
}

func TestPrivateComments(t *testing.T) {
	postEntry(userIDs[0], models.EntryPrivacyMe, true)
	feed := checkLoadTlog(t, userIDs[0], userIDs[0], true, 10, "", "", "", "new", 1)
	entry := feed.Entries[0]

	var id int64

	id = checkPostComment(t, entry, "blabla", true, profiles[0], userIDs[0])
	checkEditComment(t, id, "edited comment", entry, true, profiles[0], userIDs[0])
	checkEntryWatching(t, userIDs[0], entry.ID, true, true)

	checkEditComment(t, id, "edited comment", entry, false, profiles[1], userIDs[1])
	id = checkPostComment(t, entry, "blabla", false, profiles[1], userIDs[1])
	checkEntryWatching(t, userIDs[1], entry.ID, false, false)

	checkDeleteEntry(t, entry.ID, userIDs[0], true)
}

func postComment(id *models.UserID, entryID int64) int64 {
	params := comments.PostEntriesIDCommentsParams{
		ID:      entryID,
		Content: "test comment" + utils.GenerateString(5),
	}

	post := api.CommentsPostEntriesIDCommentsHandler.Handle
	resp := post(params, id)
	body := resp.(*comments.PostEntriesIDCommentsCreated)
	cmt := body.Payload

	time.Sleep(10 * time.Millisecond)

	return cmt.ID
}

func TestCommentHTML(t *testing.T) {
	req := require.New(t)
	entry := postEntry(userIDs[0], models.EntryPrivacyAll, false)

	post := func(content string) *models.Comment {
		params := comments.PostEntriesIDCommentsParams{
			ID:      entry.ID,
			Content: content,
		}

		post := api.CommentsPostEntriesIDCommentsHandler.Handle
		resp := post(params, userIDs[0])
		body, ok := resp.(*comments.PostEntriesIDCommentsCreated)
		req.True(ok)

		return body.Payload
	}

	content := "http://ex.com/im.jpg"
	cmt := post(content)

	req.Equal(content, cmt.EditContent)
	req.Equal("<p><img src=\""+content+"\"></p>", cmt.Content)

	edit := func(content string) *models.Comment {
		params := comments.PutCommentsIDParams{
			ID:      cmt.ID,
			Content: content,
		}

		edit := api.CommentsPutCommentsIDHandler.Handle
		resp := edit(params, userIDs[0])
		body, ok := resp.(*comments.PutCommentsIDOK)
		req.True(ok)
		return body.Payload
	}

	checkImage := func(content string) {
		cmt = edit(content)
		req.Equal(content, cmt.EditContent)
		req.Equal("<p><img src=\""+content+"\"></p>", cmt.Content)
	}

	checkImage("http://ex.com/im.jpg?trolo")
	checkImage("hTTps://ex.com/im.GIf?oooo#aaa")

	checkURL := func(content string) {
		cmt = edit(content)
		req.Equal(content, cmt.EditContent)
		req.Equal("<p><a href=\""+content+"\" target=\"_blank\" rel=\"noopener nofollow\">"+content+"</a></p>", cmt.Content)
	}

	checkURL("http://ex.com/im.ajpg")
	checkURL("tg://resolve?domain=telegram")
	checkURL("https://ex.com/im?oooo#aaa")

	checkText := func(content string) {
		cmt = edit(content)
		req.Equal(content, cmt.EditContent)
		req.Equal("<p>"+content+"</p>", cmt.Content)
	}

	checkText("http://")
	checkText("://a")
	checkText("aa:// a")

	content = "<>&\n\"'\t"
	cmt = edit(content)
	req.Equal("<p>&lt;&gt;&amp;<br>&#34;&#39;</p>", cmt.Content)

	content = "https://wiki.org/%D0%92%D0%B8%D0%BA%D0%B8%D0%BF%D0%B5%D0%B4%D0%B8%D1%8F"
	href := "https://wiki.org/Википедия"
	cmt = edit(content)
	req.Equal("<p><a href=\""+href+"\" target=\"_blank\" rel=\"noopener nofollow\">"+href+"</a></p>", cmt.Content)
}
