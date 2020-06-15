package test

import (
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/entries"
	"github.com/sevings/mindwell-server/restapi/operations/me"
	"github.com/sevings/mindwell-server/restapi/operations/users"
	"github.com/sevings/mindwell-server/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

func postTaggedEntry(t *testing.T, user *models.UserID, privacy string, tags []string) *models.Entry {
	title := ""
	votable := true
	live := true
	params := me.PostMeTlogParams{
		Content:   "test tagged " + utils.GenerateString(8),
		Title:     &title,
		Privacy:   privacy,
		IsVotable: &votable,
		InLive:    &live,
		Tags:      tags,
	}

	resp := api.MePostMeTlogHandler.Handle(params, user)
	body, ok := resp.(*me.PostMeTlogCreated)
	require.True(t, ok)

	return body.Payload
}

func TestMyTags(t *testing.T) {
	e2 := postTaggedEntry(t, userIDs[0], "all", []string{"aaa", "ccc"})
	e1 := postTaggedEntry(t, userIDs[0], "all", []string{"aaa", "bbb", "ccc"})
	e0 := postTaggedEntry(t, userIDs[0], "me", []string{"aaa", "bbb"})

	aaa := &models.TagListDataItems0{Count: 3, Tag: "aaa"}
	bbb := &models.TagListDataItems0{Count: 2, Tag: "bbb"}
	ccc := &models.TagListDataItems0{Count: 2, Tag: "ccc"}

	req := require.New(t)

	load := func(userID *models.UserID, limit int64, exp []*models.TagListDataItems0) *models.TagList {
		params := me.GetMeTagsParams{Limit: &limit}
		get := api.MeGetMeTagsHandler.Handle
		resp := get(params, userID)
		tags, ok := resp.(*me.GetMeTagsOK)

		req.True(ok)
		req.Equal(exp, tags.Payload.Data)

		return tags.Payload
	}

	load(userIDs[1], 10, nil)
	load(userIDs[0], 10, []*models.TagListDataItems0{aaa, bbb, ccc})
	load(userIDs[0], 1, []*models.TagListDataItems0{aaa})

	checkDeleteEntry(t, e0.ID, userIDs[0], true)
	checkDeleteEntry(t, e1.ID, userIDs[0], true)
	checkDeleteEntry(t, e2.ID, userIDs[0], true)
}

func TestUserTags(t *testing.T) {
	e3 := postTaggedEntry(t, userIDs[0], "me", []string{"aaa", "ddd"})
	e2 := postTaggedEntry(t, userIDs[0], "all", []string{"aaa", "ccc"})
	e1 := postTaggedEntry(t, userIDs[0], "all", []string{"aaa", "bbb", "ccc"})
	e0 := postTaggedEntry(t, userIDs[0], "all", []string{"aaa", "bbb"})

	aaa := &models.TagListDataItems0{Count: 3, Tag: "aaa"}
	aaa4 := &models.TagListDataItems0{Count: 4, Tag: "aaa"}
	bbb := &models.TagListDataItems0{Count: 2, Tag: "bbb"}
	ccc := &models.TagListDataItems0{Count: 2, Tag: "ccc"}
	ddd := &models.TagListDataItems0{Count: 1, Tag: "ddd"}

	req := require.New(t)

	load := func(userID *models.UserID, tlog string, limit int64, success bool, exp []*models.TagListDataItems0) *models.TagList {
		params := users.GetUsersNameTagsParams{Limit: &limit, Name: tlog}
		get := api.UsersGetUsersNameTagsHandler.Handle
		resp := get(params, userID)
		tags, ok := resp.(*users.GetUsersNameTagsOK)

		req.Equal(success, ok)
		if !ok {
			return nil
		}
		req.Equal(exp, tags.Payload.Data)

		return tags.Payload
	}

	load(userIDs[0], userIDs[1].Name, 10, true, nil)
	load(userIDs[1], userIDs[0].Name, 10, true, []*models.TagListDataItems0{aaa, bbb, ccc})
	load(userIDs[0], userIDs[0].Name, 10, true, []*models.TagListDataItems0{aaa4, bbb, ccc, ddd})
	load(userIDs[1], userIDs[0].Name, 1, true, []*models.TagListDataItems0{aaa})

	setUserPrivacy(t, userIDs[0], "followers")

	load(userIDs[1], userIDs[0].Name, 10, false, nil)
	load(userIDs[0], userIDs[0].Name, 10, true, []*models.TagListDataItems0{aaa4, bbb, ccc, ddd})

	checkFollow(t, userIDs[1], userIDs[0], profiles[0], models.RelationshipRelationRequested, true)
	checkPermitFollow(t, userIDs[0], userIDs[1], true)

	load(userIDs[1], userIDs[0].Name, 10, true, []*models.TagListDataItems0{aaa, bbb, ccc})
	load(userIDs[0], userIDs[0].Name, 10, true, []*models.TagListDataItems0{aaa4, bbb, ccc, ddd})

	checkUnfollow(t, userIDs[1], userIDs[0])
	setUserPrivacy(t, userIDs[0], "all")

	checkDeleteEntry(t, e0.ID, userIDs[0], true)
	checkDeleteEntry(t, e1.ID, userIDs[0], true)
	checkDeleteEntry(t, e2.ID, userIDs[0], true)
	checkDeleteEntry(t, e3.ID, userIDs[0], true)
}

func TestLiveTags(t *testing.T) {
	e3 := postTaggedEntry(t, userIDs[0], "me", []string{"aaa", "ddd"})
	e2 := postTaggedEntry(t, userIDs[2], "all", []string{"aaa", "ccc"})
	e1 := postTaggedEntry(t, userIDs[1], "all", []string{"aaa", "bbb", "ccc"})
	e0 := postTaggedEntry(t, userIDs[0], "all", []string{"aaa", "bbb"})

	aaa := &models.TagListDataItems0{Count: 3, Tag: "aaa"}
	bbb := &models.TagListDataItems0{Count: 2, Tag: "bbb"}
	ccc := &models.TagListDataItems0{Count: 2, Tag: "ccc"}

	req := require.New(t)

	load := func(userID *models.UserID, limit int64, exp []*models.TagListDataItems0) *models.TagList {
		params := entries.GetEntriesTagsParams{Limit: &limit}
		get := api.EntriesGetEntriesTagsHandler.Handle
		resp := get(params, userID)
		tags, ok := resp.(*entries.GetEntriesTagsOK)

		req.True(ok)
		req.Equal(exp, tags.Payload.Data)

		return tags.Payload
	}

	load(userIDs[0], 10, []*models.TagListDataItems0{aaa, bbb, ccc})
	load(userIDs[1], 10, []*models.TagListDataItems0{aaa, bbb, ccc})
	load(userIDs[1], 1, []*models.TagListDataItems0{aaa})

	checkDeleteEntry(t, e0.ID, userIDs[0], true)
	checkDeleteEntry(t, e1.ID, userIDs[1], true)
	checkDeleteEntry(t, e2.ID, userIDs[2], true)
	checkDeleteEntry(t, e3.ID, userIDs[0], true)
}
