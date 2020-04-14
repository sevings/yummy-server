package chats

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/utils"
	"time"
)

var userIDs *cache.Cache
var chatIDs *cache.Cache
var messages *cache.Cache
var partners *cache.Cache

func init() {
	userIDs = cache.New(48*time.Hour, 24*time.Hour)
	chatIDs = cache.New(48*time.Hour, 24*time.Hour)
	messages = cache.New(48*time.Hour, 24*time.Hour)
	partners = cache.New(48*time.Hour, 24*time.Hour)
}

func findUserID(tx *utils.AutoTx, name string) int64 {
	id, found := userIDs.Get(name)
	if found {
		return id.(int64)
	}

	const findUserQuery = "SELECT id FROM users WHERE lower(name) = lower($1)"
	userID := tx.QueryInt64(findUserQuery, name)
	if userID != 0 {
		userIDs.SetDefault(name, userID)
	}

	return userID
}

func findDialogID(tx *utils.AutoTx, userID, otherID int64) int64 {
	var creatorID, partnerID int64
	if userID < otherID {
		creatorID = userID
		partnerID = otherID
	} else {
		creatorID = otherID
		partnerID = userID
	}

	key := fmt.Sprintf("%x_%x", creatorID, partnerID)
	id, found := chatIDs.Get(key)
	if found {
		return id.(int64)
	}

	const findDialogQuery = "SELECT id FROM chats WHERE creator_id = $1 AND partner_id = $2"
	chatID := tx.QueryInt64(findDialogQuery, creatorID, partnerID)
	if chatID != 0 {
		chatIDs.SetDefault(key, chatID)
	}

	return chatID
}

func findDialog(tx *utils.AutoTx, userID int64, partnerName string) (chatID, otherID int64) {
	otherID = findUserID(tx, partnerName)
	if otherID != 0 {
		chatID = findDialogID(tx, userID, otherID)
	}

	return
}

func getCachedMessage(userID, uid int64, partnerName string) *models.Message {
	key := fmt.Sprintf("%x_%x_%s", userID, uid, partnerName)
	msg, found := messages.Get(key)
	if !found {
		return nil
	}

	return msg.(*models.Message)
}

func setCachedMessage(userID, uid int64, partnerName string, msg *models.Message) {
	key := fmt.Sprintf("%x_%x_%s", userID, uid, partnerName)
	messages.SetDefault(key, msg)
}

func findPartner(tx *utils.AutoTx, chatID, userID int64) string {
	key := fmt.Sprintf("%d_%d", chatID, userID)
	name, found := partners.Get(key)
	if found {
		return name.(string)
	}

	const q = `
		SELECT name 
		FROM users
		JOIN talkers ON user_id = users.id
		WHERE chat_id = $1 and user_id <> $2
	`

	partner := tx.QueryString(q, chatID, userID)
	if partner != "" {
		partners.SetDefault(key, partner)
	}

	return partner
}
