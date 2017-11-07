package users

import (
	"database/sql"
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations"
	"github.com/sevings/yummy-server/gen/restapi/operations/me"
	"github.com/sevings/yummy-server/gen/restapi/operations/users"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.UsersGetUsersIDHandler = users.GetUsersIDHandlerFunc(newUserLoader(db))
	api.UsersGetUsersByNameNameHandler = users.GetUsersByNameNameHandlerFunc(newUserLoaderByName(db))

	api.UsersGetUsersIDFollowersHandler = users.GetUsersIDFollowersHandlerFunc(newFollowersLoader(db))
	api.UsersGetUsersByNameNameFollowersHandler = users.GetUsersByNameNameFollowersHandlerFunc(newFollowersLoaderByName(db))

	api.UsersGetUsersIDFollowingsHandler = users.GetUsersIDFollowingsHandlerFunc(newFollowingsLoader(db))

	api.MeGetUsersMeHandler = me.GetUsersMeHandlerFunc(newMeLoader(db))
	api.MeGetUsersMeFollowersHandler = me.GetUsersMeFollowersHandlerFunc(newMyFollowersLoader(db))
}

const profileQuery = `
SELECT id, name, show_name,
name_color, avatar_color, avatar,
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
invited_by_name_color, invited_by_avatar_color,
invited_by_avatar
FROM long_users `

func loadProfile(db *sql.DB, query string, apiKey *string, arg interface{}) middleware.Responder {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Commit()

	row := tx.QueryRow(query, arg)

	var profile models.Profile
	profile.InvitedBy = &models.User{}
	profile.Design = &models.Design{}
	profile.Counts = &models.ProfileAO1Counts{}

	var nameColor string
	var avatarColor string
	var backColor string
	var textColor string
	var invNameColor string
	var invAvColor string

	var age sql.NullInt64

	err = row.Scan(&profile.ID, &profile.Name, &profile.ShowName,
		&nameColor, &avatarColor, &profile.Avatar,
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
		&invNameColor, &invAvColor,
		&profile.InvitedBy.Avatar)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}

		return users.NewGetUsersIDNotFound()
	}

	profile.NameColor = models.Color(nameColor)
	profile.AvatarColor = models.Color(avatarColor)
	profile.Design.BackgroundColor = models.Color(backColor)
	profile.Design.TextColor = models.Color(textColor)
	profile.InvitedBy.NameColor = models.Color(invNameColor)
	profile.InvitedBy.AvatarColor = models.Color(invAvColor)

	if age.Valid {
		profile.AgeLowerBound = age.Int64 - age.Int64%5
		profile.AgeUpperBound = profile.AgeLowerBound + 5
	}

	result := users.NewGetUsersIDOK().WithPayload(&profile)
	if apiKey == nil {
		return result
	}

	userID, found := FindAuthUser(tx, apiKey)
	if !found || userID == profile.ID {
		return result
	}

	profile.Relations = &models.ProfileAO1Relations{}
	profile.Relations.FromMe = relationship(tx, relationToIDQuery, userID, profile.ID)
	profile.Relations.ToMe = relationship(tx, relationToIDQuery, profile.ID, userID)

	return result
}

const authUserQuery = `
SELECT id
FROM users
WHERE api_key = $1 AND valid_thru > CURRENT_TIMESTAMP`

// FindAuthUser returns ID of the authorized user or false if the key is invalid or expired.
func FindAuthUser(tx *sql.Tx, apiKey *string) (int64, bool) {
	if apiKey == nil {
		return 0, false
	}

	var id int64
	err := tx.QueryRow(authUserQuery, *apiKey).Scan(&id)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}

		return id, false
	}

	return id, true
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
WHERE users.name = $2
	relations.from_id = $1
	AND relations.to_id = users.id
	AND relations.type = relation.id`

func relationship(tx *sql.Tx, query string, from int64, to interface{}) string {
	var relation string
	err := tx.QueryRow(query, from, to).Scan(&relation)
	switch {
	case err == sql.ErrNoRows:
		return "none"
	case err != nil:
		log.Print(err)
		return ""
	}

	return relation
}

const privacyQueryStart = `
SELECT user_privacy.type
FROM users, user_privacy
WHERE users.privacy = user_privacy.id AND `

func isOpenForMe(tx *sql.Tx, privacyQuery, relationQuery string,
	apiKey *string, arg interface{}) (bool, error) {
	var privacy string
	err := tx.QueryRow(privacyQuery, arg).Scan(&privacy)
	if err != nil {
		return false, err
	}

	if privacy == "all" {
		return true, nil
	}

	userID, found := FindAuthUser(tx, apiKey)
	if !found {
		return false, nil
	}

	if privacy == "registered" {
		return true, nil
	}

	relation := relationship(tx, relationQuery, userID, arg)
	if relation == models.RelationshipRelationFollowed {
		return true, nil
	}

	return false, nil
}

const usersQueryStart = `
SELECT short_users.id, name, show_name,
is_online, 
name_color, avatar_color, avatar
FROM short_users, relation, relations
WHERE `

const usersQueryEnd = `
 and relations.type = relation.id and relation.type = $2
ORDER BY relations.changed_at DESC
LIMIT $3 OFFSET $4`

func loadRelatedUsers(tx *sql.Tx, usersQuery string,
	arg interface{}, relation string, limit, offset int64) (*models.UserList, error) {
	var list models.UserList
	rows, err := tx.Query(usersQuery, arg, relation, limit, offset)
	if err != nil {
		return &list, err
	}

	for rows.Next() {
		var user models.User
		rows.Scan(&user.ID, &user.Name, &user.ShowName,
			&user.IsOnline,
			&user.NameColor, &user.AvatarColor, &user.Avatar)
		list.Users = append(list.Users, &user)
	}

	return &list, rows.Err()
}

func loadUsers(db *sql.DB, usersQuery, privacyQuery, relationQuery string,
	apiKey *string,
	arg interface{}, relation string, limit, offset int64) middleware.Responder {

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Commit()

	open, err := isOpenForMe(tx, privacyQuery, relationQuery, apiKey, arg)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}

		return users.NewGetUsersIDFollowersNotFound()
	}

	if !open {
		return users.NewGetUsersIDFollowersForbidden()
	}

	list, err := loadRelatedUsers(tx, usersQuery, arg, relation, limit, offset)
	if err != nil {
		return users.NewGetUsersIDFollowersNotFound()
	}

	return users.NewGetUsersIDFollowersOK().WithPayload(list)
}

const loadUserQuery = `
SELECT id, name, show_name,
is_online, 
name_color, avatar_color, avatar
FROM long_users
WHERE `

func loadUser(tx *sql.Tx, query string, arg interface{}) (*models.User, bool) {
	var user models.User
	err := tx.QueryRow(query, arg).Scan(&user.ID, &user.Name, &user.ShowName,
		&user.IsOnline,
		&user.NameColor, &user.AvatarColor, &user.Avatar)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}

		return &user, false
	}

	return &user, true
}

// LoadAuthUser returns short profile of the authorized user or false if the key is invalid or expired.
func LoadAuthUser(tx *sql.Tx, apiKey *string) (*models.User, bool) {
	if apiKey == nil {
		return nil, false
	}

	const q = loadUserQuery + "api_key = $1 AND valid_thru > CURRENT_TIMESTAMP"
	return loadUser(tx, q, *apiKey)
}

// LoadUserByID returns short user profile by its ID.
func LoadUserByID(tx *sql.Tx, id int64) (*models.User, bool) {
	const q = loadUserQuery + "id = $1"
	return loadUser(tx, q, id)
}
