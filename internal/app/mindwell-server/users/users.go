package users

import (
	"database/sql"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/me"
	"github.com/sevings/mindwell-server/restapi/operations/users"
	"github.com/sevings/mindwell-server/utils"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	srv.API.APIKeyHeaderAuth = utils.NewKeyAuth(srv.DB)

	srv.API.MeGetUsersMeHandler = me.GetUsersMeHandlerFunc(newMeLoader(srv))
	srv.API.MePutUsersMeHandler = me.PutUsersMeHandlerFunc(newMeEditor(srv))

	srv.API.UsersGetUsersIDHandler = users.GetUsersIDHandlerFunc(newUserLoader(srv))
	srv.API.UsersGetUsersByNameNameHandler = users.GetUsersByNameNameHandlerFunc(newUserLoaderByName(srv))

	srv.API.MeGetUsersMeFollowersHandler = me.GetUsersMeFollowersHandlerFunc(newMyFollowersLoader(srv))
	srv.API.UsersGetUsersIDFollowersHandler = users.GetUsersIDFollowersHandlerFunc(newFollowersLoader(srv))
	srv.API.UsersGetUsersByNameNameFollowersHandler = users.GetUsersByNameNameFollowersHandlerFunc(newFollowersLoaderByName(srv))

	srv.API.MeGetUsersMeFollowingsHandler = me.GetUsersMeFollowingsHandlerFunc(newMyFollowingsLoader(srv))
	srv.API.UsersGetUsersIDFollowingsHandler = users.GetUsersIDFollowingsHandlerFunc(newFollowingsLoader(srv))
	srv.API.UsersGetUsersByNameNameFollowingsHandler = users.GetUsersByNameNameFollowingsHandlerFunc(newFollowingsLoaderByName(srv))

	srv.API.MeGetUsersMeInvitedHandler = me.GetUsersMeInvitedHandlerFunc(newMyInvitedLoader(srv))
	srv.API.UsersGetUsersIDInvitedHandler = users.GetUsersIDInvitedHandlerFunc(newInvitedLoader(srv))
	srv.API.UsersGetUsersByNameNameInvitedHandler = users.GetUsersByNameNameInvitedHandlerFunc(newInvitedLoaderByName(srv))

	srv.API.MeGetUsersMeIgnoredHandler = me.GetUsersMeIgnoredHandlerFunc(newMyIgnoredLoader(srv))
	srv.API.MeGetUsersMeRequestedHandler = me.GetUsersMeRequestedHandlerFunc(newMyRequestedLoader(srv))

	srv.API.MePutUsersMeOnlineHandler = me.PutUsersMeOnlineHandlerFunc(newMyOnlineSetter(srv))
}

const profileQuery = `
SELECT id, name, show_name,
avatar,
gender, is_daylog,
privacy,
title, karma, 
extract(epoch from created_at), extract(epoch from last_seen_at), is_online,
age,
entries_count, followings_count, followers_count, 
ignored_count, invited_count, comments_count, 
favorites_count, tags_count,
country, city,
cover,
css, background_color, text_color, 
font_family, font_size, text_alignment, 
invited_by_id, 
invited_by_name, invited_by_show_name,
invited_by_is_online, 
invited_by_avatar
FROM long_users `

func loadUserProfile(srv *utils.MindwellServer, tx *utils.AutoTx, query string, userID *models.UserID, arg interface{}) *models.Profile {
	var profile models.Profile
	profile.InvitedBy = &models.User{}
	profile.Design = &models.Design{}
	profile.Counts = &models.FriendAllOf1Counts{}

	var backColor string
	var textColor string

	var age sql.NullInt64
	var avatar, cover string
	var invitedAvatar string

	tx.Query(query, arg)
	tx.Scan(&profile.ID, &profile.Name, &profile.ShowName,
		&avatar,
		&profile.Gender, &profile.IsDaylog,
		&profile.Privacy,
		&profile.Title, &profile.Karma,
		&profile.CreatedAt, &profile.LastSeenAt, &profile.IsOnline,
		&age,
		&profile.Counts.Entries, &profile.Counts.Followings, &profile.Counts.Followers,
		&profile.Counts.Ignored, &profile.Counts.Invited, &profile.Counts.Comments,
		&profile.Counts.Favorites, &profile.Counts.Tags,
		&profile.Country, &profile.City,
		&cover,
		&profile.Design.CSS, &backColor, &textColor,
		&profile.Design.FontFamily, &profile.Design.FontSize, &profile.Design.TextAlignment,
		&profile.InvitedBy.ID,
		&profile.InvitedBy.Name, &profile.InvitedBy.ShowName,
		&profile.InvitedBy.IsOnline,
		&invitedAvatar)

	profile.Design.BackgroundColor = models.Color(backColor)
	profile.Design.TextColor = models.Color(textColor)

	profile.Avatar = srv.NewAvatar(avatar)
	profile.InvitedBy.Avatar = srv.NewAvatar(invitedAvatar)

	profile.Cover = srv.NewCover(profile.ID, cover)

	if age.Valid {
		profile.AgeLowerBound = age.Int64 - age.Int64%5
		profile.AgeUpperBound = profile.AgeLowerBound + 5
	}

	return &profile
}

func loadProfile(srv *utils.MindwellServer, query string, userID *models.UserID, arg interface{}) middleware.Responder {
	return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
		profile := loadUserProfile(srv, tx, query, userID, arg)
		if tx.Error() != nil {
			return users.NewGetUsersIDNotFound()
		}

		result := users.NewGetUsersIDOK().WithPayload(profile)
		if int64(*userID) == profile.ID {
			return result
		}

		profile.Relations = &models.ProfileAllOf1Relations{}
		profile.Relations.FromMe = relationship(tx, relationToIDQuery, int64(*userID), profile.ID)
		profile.Relations.ToMe = relationship(tx, relationToIDQuery, profile.ID, int64(*userID))

		return result
	})
}

const relationToIDQuery = `
SELECT relation.type
FROM relations, relation
WHERE relations.from_id = $1
	AND relations.to_id = $2
	AND relations.type = relation.id`

const relationToNameQuery = `
SELECT relation.type
FROM users, relations, relation
WHERE lower(users.name) = lower($2)
	AND relations.from_id = $1
	AND relations.to_id = users.id
	AND relations.type = relation.id`

func relationship(tx *utils.AutoTx, query string, from int64, to interface{}) string {
	var relation string
	tx.Query(query, from, to).Scan(&relation)
	if tx.Error() == sql.ErrNoRows {
		return models.RelationshipRelationNone
	}

	return relation
}

const privacyQueryStart = `
SELECT users.id, user_privacy.type
FROM users, user_privacy
WHERE users.privacy = user_privacy.id AND `

func isOpenForMe(tx *utils.AutoTx, privacyQuery, relationQuery string,
	userID *models.UserID, arg interface{}) bool {
	var subjectID int64
	var privacy string
	tx.Query(privacyQuery, arg).Scan(&subjectID, &privacy)
	if tx.Error() != nil {
		return false
	}

	if subjectID == int64(*userID) {
		return true
	}

	if privacy == models.ProfileAllOf1PrivacyAll {
		return true
	}

	relation := relationship(tx, relationQuery, int64(*userID), arg)
	return relation == models.RelationshipRelationFollowed
}

const usersQueryStart = `
SELECT long_users.id, name, show_name,
is_online, extract(epoch from last_seen_at), title, karma,
avatar, cover,
entries_count, followings_count, followers_count, 
ignored_count, invited_count, comments_count, 
favorites_count, tags_count
FROM long_users, relation, relations
WHERE `

const usersQueryEnd = `
 AND relations.type = relation.id AND relation.type = $2
ORDER BY relations.changed_at DESC
LIMIT $3 OFFSET $4`

const invitedUsersQuery = `
SELECT id, name, show_name,
is_online, extract(epoch from last_seen_at), title, karma,
avatar, cover,
entries_count, followings_count, followers_count, 
ignored_count, invited_count, comments_count, 
favorites_count, tags_count
FROM long_users
WHERE invited_by = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3`

const invitedUsersByNameQuery = `
WITH by AS (
	SELECT id
	FROM users
	WHERE lower(name) = lower($1)
)
SELECT long_users.id, name, show_name,
is_online, extract(epoch from last_seen_at), title, karma,
avatar, cover,
entries_count, followings_count, followers_count, 
ignored_count, invited_count, comments_count, 
favorites_count, tags_count
FROM long_users, by
WHERE invited_by = by.id
ORDER BY long_users.id DESC
LIMIT $2 OFFSET $3`

func loadRelatedUsers(srv *utils.MindwellServer, tx *utils.AutoTx, usersQuery, subjectQuery, relation string, args ...interface{}) *models.FriendList {
	var list models.FriendList
	tx.Query(usersQuery, args...)

	for {
		var user models.Friend
		user.Counts = &models.FriendAllOf1Counts{}
		var avatar, cover string

		ok := tx.Scan(&user.ID, &user.Name, &user.ShowName,
			&user.IsOnline, &user.LastSeenAt, &user.Title, &user.Karma,
			&avatar, &cover,
			&user.Counts.Entries, &user.Counts.Followings, &user.Counts.Followers,
			&user.Counts.Ignored, &user.Counts.Invited, &user.Counts.Comments,
			&user.Counts.Favorites, &user.Counts.Tags)
		if !ok {
			break
		}

		user.Avatar = srv.NewAvatar(avatar)
		user.Cover = srv.NewCover(user.ID, cover)
		list.Users = append(list.Users, &user)
	}

	list.Subject = loadUser(srv, tx, subjectQuery, args[0])
	list.Relation = relation

	return &list
}

func loadUsers(srv *utils.MindwellServer, usersQuery, privacyQuery, relationQuery, subjectQuery, relation string,
	userID *models.UserID, args ...interface{}) middleware.Responder {
	return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
		open := isOpenForMe(tx, privacyQuery, relationQuery, userID, args[0])
		if tx.Error() != nil {
			err := srv.NewError(nil)
			return users.NewGetUsersIDFollowersNotFound().WithPayload(err)
		}

		if !open {
			err := srv.StandardError("no_tlog")
			return users.NewGetUsersIDFollowersForbidden().WithPayload(err)
		}

		list := loadRelatedUsers(srv, tx, usersQuery, subjectQuery, relation, args...)
		if tx.Error() != nil && tx.Error() != sql.ErrNoRows {
			err := srv.StandardError("no_tlog")
			return users.NewGetUsersIDFollowersNotFound().WithPayload(err)
		}

		return users.NewGetUsersIDFollowersOK().WithPayload(list)
	})
}

const loadUserQuery = `
SELECT id, name, show_name,
is_online, avatar
FROM short_users
WHERE `

const loadUserQueryID = loadUserQuery + "id = $1"
const loadUserQueryName = loadUserQuery + "lower(name) = lower($1)"

func loadUser(srv *utils.MindwellServer, tx *utils.AutoTx, query string, arg interface{}) *models.User {
	var user models.User
	var avatar string

	tx.Query(query, arg).Scan(&user.ID, &user.Name, &user.ShowName,
		&user.IsOnline, &avatar)

	user.Avatar = srv.NewAvatar(avatar)
	return &user
}

// LoadUserByID returns short user profile by its ID.
func LoadUserByID(srv *utils.MindwellServer, tx *utils.AutoTx, id int64) *models.User {
	return loadUser(srv, tx, loadUserQueryID, id)
}

// LoadUserByName returns short user profile by its ID.
func LoadUserByName(srv *utils.MindwellServer, tx *utils.AutoTx, name string) *models.User {
	return loadUser(srv, tx, loadUserQueryName, name)
}
