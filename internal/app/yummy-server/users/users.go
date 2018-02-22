package users

import (
	"database/sql"
	"log"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"
	"github.com/sevings/yummy-server/restapi/operations/me"
	"github.com/sevings/yummy-server/restapi/operations/users"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.APIKeyHeaderAuth = newKeyAuth(db)

	api.MeGetUsersMeHandler = me.GetUsersMeHandlerFunc(newMeLoader(db))
	api.UsersGetUsersIDHandler = users.GetUsersIDHandlerFunc(newUserLoader(db))
	api.UsersGetUsersByNameNameHandler = users.GetUsersByNameNameHandlerFunc(newUserLoaderByName(db))

	api.MeGetUsersMeFollowersHandler = me.GetUsersMeFollowersHandlerFunc(newMyFollowersLoader(db))
	api.UsersGetUsersIDFollowersHandler = users.GetUsersIDFollowersHandlerFunc(newFollowersLoader(db))
	api.UsersGetUsersByNameNameFollowersHandler = users.GetUsersByNameNameFollowersHandlerFunc(newFollowersLoaderByName(db))

	api.MeGetUsersMeFollowingsHandler = me.GetUsersMeFollowingsHandlerFunc(newMyFollowingsLoader(db))
	api.UsersGetUsersIDFollowingsHandler = users.GetUsersIDFollowingsHandlerFunc(newFollowingsLoader(db))
	api.UsersGetUsersByNameNameFollowingsHandler = users.GetUsersByNameNameFollowingsHandlerFunc(newFollowingsLoaderByName(db))

	api.MeGetUsersMeInvitedHandler = me.GetUsersMeInvitedHandlerFunc(newMyInvitedLoader(db))
	api.UsersGetUsersIDInvitedHandler = users.GetUsersIDInvitedHandlerFunc(newInvitedLoader(db))
	api.UsersGetUsersByNameNameInvitedHandler = users.GetUsersByNameNameInvitedHandlerFunc(newInvitedLoaderByName(db))

	api.MeGetUsersMeIgnoredHandler = me.GetUsersMeIgnoredHandlerFunc(newMyIgnoredLoader(db))
	api.MeGetUsersMeRequestedHandler = me.GetUsersMeRequestedHandlerFunc(newMyRequestedLoader(db))

	api.MePutUsersMeOnlineHandler = me.PutUsersMeOnlineHandlerFunc(newMyOnlineSetter(db))
}

const profileQuery = `
SELECT id, name, show_name,
avatar,
gender, is_daylog,
privacy,
title, karma, 
created_at, last_seen_at, is_online,
age,
entries_count, followings_count, followers_count, 
ignored_count, invited_count, comments_count, 
favorites_count, tags_count,
country, city,
css, background_color, text_color, 
font_family, font_size, text_alignment, 
invited_by_id, 
invited_by_name, invited_by_show_name,
invited_by_is_online, 
invited_by_avatar
FROM long_users `

func loadProfile(db *sql.DB, query string, userID *models.UserID, arg interface{}) middleware.Responder {
	return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
		var profile models.Profile
		profile.InvitedBy = &models.User{}
		profile.Design = &models.Design{}
		profile.Counts = &models.ProfileAllOf1Counts{}

		var backColor string
		var textColor string

		var age sql.NullInt64

		tx.Query(query, arg)
		tx.Scan(&profile.ID, &profile.Name, &profile.ShowName,
			&profile.Avatar,
			&profile.Gender, &profile.IsDaylog,
			&profile.Privacy,
			&profile.Title, &profile.Karma,
			&profile.CreatedAt, &profile.LastSeenAt, &profile.IsOnline,
			&age,
			&profile.Counts.Entries, &profile.Counts.Followings, &profile.Counts.Followers,
			&profile.Counts.Ignored, &profile.Counts.Invited, &profile.Counts.Comments,
			&profile.Counts.Favorites, &profile.Counts.Tags,
			&profile.Country, &profile.City,
			&profile.Design.CSS, &backColor, &textColor,
			&profile.Design.FontFamily, &profile.Design.FontSize, &profile.Design.TextAlignment,
			&profile.InvitedBy.ID,
			&profile.InvitedBy.Name, &profile.InvitedBy.ShowName,
			&profile.InvitedBy.IsOnline,
			&profile.InvitedBy.Avatar)

		if tx.Error() != nil {
			return users.NewGetUsersIDNotFound()
		}

		profile.Design.BackgroundColor = models.Color(backColor)
		profile.Design.TextColor = models.Color(textColor)

		if age.Valid {
			profile.AgeLowerBound = age.Int64 - age.Int64%5
			profile.AgeUpperBound = profile.AgeLowerBound + 5
		}

		result := users.NewGetUsersIDOK().WithPayload(&profile)
		if int64(*userID) == profile.ID {
			return result
		}

		profile.Relations = &models.ProfileAllOf1Relations{}
		profile.Relations.FromMe = relationship(tx, relationToIDQuery, int64(*userID), profile.ID)
		profile.Relations.ToMe = relationship(tx, relationToIDQuery, profile.ID, int64(*userID))

		return result
	})
}

func newKeyAuth(db *sql.DB) func(apiKey string) (*models.UserID, error) {
	const q = `
		SELECT id
		FROM users
		WHERE api_key = $1 AND valid_thru > CURRENT_TIMESTAMP`

	return func(apiKey string) (*models.UserID, error) {
		var id int64
		err := db.QueryRow(q, apiKey).Scan(&id)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Print(err)
			}

			return nil, errors.New(403, "Access denied")
		}

		userID := models.UserID(id)
		return &userID, nil
	}
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
	relations.from_id = $1
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
SELECT user_privacy.type
FROM users, user_privacy
WHERE users.privacy = user_privacy.id AND `

func isOpenForMe(tx *utils.AutoTx, privacyQuery, relationQuery string,
	userID *models.UserID, arg interface{}) bool {
	var privacy string
	tx.Query(privacyQuery, arg).Scan(&privacy)
	if tx.Error() != nil {
		return false
	}

	if privacy == "all" {
		return true
	}

	if privacy == "registered" {
		return true
	}

	relation := relationship(tx, relationQuery, int64(*userID), arg)
	return relation == models.RelationshipRelationFollowed
}

const usersQueryStart = `
SELECT short_users.id, name, show_name,
is_online, 
avatar
FROM short_users, relation, relations
WHERE `

const usersQueryEnd = `
 and relations.type = relation.id and relation.type = $2
ORDER BY relations.changed_at DESC
LIMIT $3 OFFSET $4`

const invitedUsersQuery = `
SELECT id, name, show_name,
is_online, 
avatar
FROM long_users
WHERE invited_by = $1
ORDER BY id ASC
LIMIT $2 OFFSET $3`

const invitedUsersByNameQuery = `
WITH by AS (
	SELECT id
	FROM users
	WHERE lower(name) = lower($1)
)
SELECT long_users.id, name, show_name,
is_online, 
avatar
FROM long_users
WHERE invited_by = by.id
ORDER BY long_users.id ASC
LIMIT $2 OFFSET $3`

func loadRelatedUsers(tx *utils.AutoTx, usersQuery string, args ...interface{}) *models.UserList {
	var list models.UserList
	tx.Query(usersQuery, args...)

	for {
		var user models.User
		ok := tx.Scan(&user.ID, &user.Name, &user.ShowName,
			&user.IsOnline,
			&user.Avatar)
		if !ok {
			break
		}

		list.Users = append(list.Users, &user)
	}

	return &list
}

func loadUsers(db *sql.DB, usersQuery, privacyQuery, relationQuery string,
	userID *models.UserID, args ...interface{}) middleware.Responder {
	return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
		open := isOpenForMe(tx, privacyQuery, relationQuery, userID, args[0])
		if tx.Error() != nil {
			return users.NewGetUsersIDFollowersNotFound()
		}

		if !open {
			return users.NewGetUsersIDFollowersForbidden()
		}

		list := loadRelatedUsers(tx, usersQuery, args...)
		if tx.Error() != nil {
			return users.NewGetUsersIDFollowersNotFound()
		}

		return users.NewGetUsersIDFollowersOK().WithPayload(list)
	})
}

const loadUserQuery = `
SELECT id, name, show_name,
is_online, 
avatar
FROM long_users
WHERE `

func loadUser(tx *utils.AutoTx, query string, arg interface{}) *models.User {
	var user models.User

	tx.Query(query, arg).Scan(&user.ID, &user.Name, &user.ShowName,
		&user.IsOnline,
		&user.Avatar)

	return &user
}

// LoadUserByID returns short user profile by its ID.
func LoadUserByID(tx *utils.AutoTx, id int64) *models.User {
	const q = loadUserQuery + "id = $1"
	return loadUser(tx, q, id)
}
