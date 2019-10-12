package test

import (
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/comments"
	"github.com/sevings/mindwell-server/restapi/operations/entries"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEntryComplain(t *testing.T) {
	complain := func(entryID int64, userID *models.UserID, success bool) {
		post := api.EntriesPostEntriesIDComplainHandler.Handle
		params := entries.NewPostEntriesIDComplainParams()
		params.ID = entryID

		resp := post(params, userID)
		_, ok := resp.(*entries.PostEntriesIDComplainNoContent)

		require.Equal(t, success, ok)
	}

	entry := postEntry(userIDs[0], models.EntryPrivacyAll, true)
	complain(entry.ID, userIDs[0], false)
	complain(entry.ID, userIDs[1], true)

	esm.CheckEmail(t, "support")
}

func TestCommentComplain(t *testing.T) {
	complain := func(commentID int64, userID *models.UserID, success bool) {
		post := api.CommentsPostCommentsIDComplainHandler.Handle
		params := comments.NewPostCommentsIDComplainParams()
		params.ID = commentID

		resp := post(params, userID)
		_, ok := resp.(*comments.PostCommentsIDComplainNoContent)

		require.Equal(t, success, ok)
	}

	entry := postEntry(userIDs[1], models.EntryPrivacyAll, true)
	cmt := postComment(userIDs[0], entry.ID)
	complain(cmt, userIDs[0], false)
	complain(cmt, userIDs[1], true)

	esm.CheckEmail(t, "support")
}
