package comments

import (
	"github.com/sevings/mindwell-server/restapi/operations/comments"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sevings/mindwell-server/models"
)

var prevComments *cache.Cache

func init() {
	prevComments = cache.New(time.Hour, time.Hour)
}

func checkPrev(params comments.PostEntriesIDCommentsParams, userID *models.UserID) (prev *models.Comment, found bool) {
	c, found := prevComments.Get(userID.Name)
	if !found {
		return
	}

	prev = c.(*models.Comment)
	if prev.EntryID != params.ID || prev.EditContent != params.Content {
		found = false
		return
	}

	return
}

func setPrev(cmt *models.Comment, userID *models.UserID) {
	prevComments.SetDefault(userID.Name, cmt)
}

func removePrev(cmtID int64, userID *models.UserID) bool {
	c, found := prevComments.Get(userID.Name)
	if !found {
		return false
	}

	prev := c.(*models.Comment)
	if prev.ID != cmtID {
		return false
	}

	prevComments.Delete(userID.Name)
	return true
}

func updatePrev(cmt *models.Comment, userID *models.UserID) {
	same := removePrev(cmt.ID, userID)
	if !same {
		return
	}

	setPrev(cmt, userID)
}
