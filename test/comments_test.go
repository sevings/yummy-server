package test

import (
	"testing"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/comments"
	"github.com/stretchr/testify/require"
)

func checkComment(t *testing.T, cmt *models.Comment, entryID int64, mine bool, author *models.AuthProfile, content string) {
	req := require.New(t)

	req.Equal(entryID, cmt.EntryID)
	req.Equal(content, cmt.Content)
	req.Equal(mine, cmt.IsMine)

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

	if mine {
		req.Equal(models.RatingVoteBan, cmt.Rating.Vote)
	} else {
		req.Equal(models.RatingVoteNot, cmt.Rating.Vote)
	}
}

func checkLoadComment(t *testing.T, commentID int64, userID *models.UserID, success bool,
	author *models.AuthProfile, entryID int64, content string) {

	load := api.CommentsGetCommentsIDHandler.Handle
	resp := load(comments.GetCommentsIDParams{ID: commentID}, userID)
	body, ok := resp.(*comments.GetCommentsIDOK)
	require.Equal(t, success, ok)
	if !success {
		return
	}

	cmt := body.Payload
	checkComment(t, cmt, entryID, author.ID == userID.ID, author, content)
}

func checkPostComment(t *testing.T,
	entryID int64, content string, success bool,
	author *models.AuthProfile, id *models.UserID) int64 {

	params := comments.PostEntriesIDCommentsParams{
		ID:      entryID,
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
	checkComment(t, cmt, params.ID, true, author, params.Content)

	checkLoadComment(t, cmt.ID, id, true, author, params.ID, params.Content)

	return cmt.ID
}

func checkEditComment(t *testing.T,
	commentID int64, content string, entryID int64, success bool,
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
	checkComment(t, cmt, entryID, true, author, content)

	checkLoadComment(t, commentID, id, true, author, entryID, content)
}

func checkDeleteComment(t *testing.T, commentID int64, userID *models.UserID, success bool) {
	del := api.CommentsDeleteCommentsIDHandler.Handle
	resp := del(comments.DeleteCommentsIDParams{ID: commentID}, userID)
	_, ok := resp.(*comments.DeleteCommentsIDOK)
	require.Equal(t, success, ok)
}

func TestOpenComments(t *testing.T) {
	postEntry(userIDs[0], models.EntryPrivacyAll, true)
	feed := checkLoadTlog(t, userIDs[0], userIDs[0], 10, "", "", 1)
	entry := feed.Entries[0]

	var id int64

	id = checkPostComment(t, entry.ID, "blabla", true, profiles[0], userIDs[0])
	checkEditComment(t, id, "edited comment", entry.ID, true, profiles[0], userIDs[0])
	checkEntryWatching(t, userIDs[0], entry.ID, true, true)

	id = checkPostComment(t, entry.ID, "blabla", true, profiles[1], userIDs[1])
	checkEditComment(t, id, "edited comment", entry.ID, true, profiles[1], userIDs[1])
	checkEntryWatching(t, userIDs[1], entry.ID, true, true)

	checkDeleteEntry(t, entry.ID, userIDs[0], true)
}

func TestPrivateComments(t *testing.T) {
	postEntry(userIDs[0], models.EntryPrivacyMe, true)
	feed := checkLoadTlog(t, userIDs[0], userIDs[0], 10, "", "", 1)
	entry := feed.Entries[0]

	var id int64

	id = checkPostComment(t, entry.ID, "blabla", true, profiles[0], userIDs[0])
	checkEditComment(t, id, "edited comment", entry.ID, true, profiles[0], userIDs[0])
	checkEntryWatching(t, userIDs[0], entry.ID, true, true)

	checkEditComment(t, id, "edited comment", entry.ID, false, profiles[1], userIDs[1])
	id = checkPostComment(t, entry.ID, "blabla", false, profiles[1], userIDs[1])
	checkEntryWatching(t, userIDs[1], entry.ID, false, false)

	checkDeleteEntry(t, entry.ID, userIDs[0], true)
}
