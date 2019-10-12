package entries

import (
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/me"
)

var prevEntries *cache.Cache

func init() {
	prevEntries = cache.New(time.Hour, time.Hour)
}

func checkPrev(params me.PostMeTlogParams, userID *models.UserID) (prev *models.Entry, found, same bool) {
	e, found := prevEntries.Get(userID.Name)
	if !found {
		return
	}

	prev = e.(*models.Entry)
	if strings.TrimSpace(prev.EditContent) != strings.TrimSpace(params.Content) {
		found = false
		return
	}

	same = prev.InLive == *params.InLive &&
		prev.Rating.IsVotable == *params.IsVotable &&
		prev.Privacy == params.Privacy &&
		strings.TrimSpace(prev.Title) == strings.TrimSpace(*params.Title) &&
		len(prev.Images) == len(params.Images)
	//! \todo check visible for
	if !same {
		return
	}

	for i := range prev.Images {
		same = prev.Images[i].ID == params.Images[i]
		if !same {
			return
		}
	}

	return
}

func setPrev(entry *models.Entry, userID *models.UserID) {
	prevEntries.SetDefault(userID.Name, entry)
}

func removePrev(entryID int64, userID *models.UserID) bool {
	e, found := prevEntries.Get(userID.Name)
	if !found {
		return false
	}

	prev := e.(*models.Entry)
	if prev.ID != entryID {
		return false
	}

	prevEntries.Delete(userID.Name)
	return true
}

func updatePrev(entry *models.Entry, userID *models.UserID) {
	same := removePrev(entry.ID, userID)
	if !same {
		return
	}

	setPrev(entry, userID)
}
