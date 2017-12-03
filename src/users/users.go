package users

import (
	"database/sql"
	"log"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations"
	"github.com/sevings/yummy-server/gen/restapi/operations/me"
	"github.com/sevings/yummy-server/gen/restapi/operations/users"
	yummy "github.com/sevings/yummy-server/src"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.APIKeyHeaderAuth = newKeyAuth(db)

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
invited_by_avatar
FROM long_users `

func loadProfile(db *sql.DB, query string, userID *models.UserID, arg interface{}) middleware.Responder {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Commit()

	row := tx.QueryRow(query, arg)

	var profile models.Profile
	profile.InvitedBy = &models.User{}
	profile.Design = &models.Design{}
	profile.Counts = &models.ProfileAllOf1Counts{}

	var backColor string
	var textColor string

	var age sql.NullInt64

	err = row.Scan(&profile.ID, &profile.Name, &profile.ShowName,
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

	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}

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
	userID *models.UserID, arg interface{}) (bool, error) {
	var privacy string
	err := tx.QueryRow(privacyQuery, arg).Scan(&privacy)
	if err != nil {
		return false, err
	}

	if privacy == "all" {
		return true, nil
	}

	if privacy == "registered" {
		return true, nil
	}

	relation := relationship(tx, relationQuery, int64(*userID), arg)
	return relation == models.RelationshipRelationFollowed, nil
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
			&user.Avatar)
		list.Users = append(list.Users, &user)
	}

	return &list, rows.Err()
}

func loadUsers(db *sql.DB, usersQuery, privacyQuery, relationQuery string,
	userID *models.UserID,
	arg interface{}, relation string, limit, offset int64) middleware.Responder {

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Commit()

	open, err := isOpenForMe(tx, privacyQuery, relationQuery, userID, arg)
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
avatar
FROM long_users
WHERE `

func loadUser(tx yummy.AutoTx, query string, arg interface{}) (*models.User, bool) {
	var user models.User

	err := tx.QueryRow(query, arg).Scan(&user.ID, &user.Name, &user.ShowName,
		&user.IsOnline,
		&user.Avatar)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}

		return &user, false
	}

	return &user, true
}

// LoadUserByID returns short user profile by its ID.
func LoadUserByID(tx yummy.AutoTx, id int64) (*models.User, bool) {
	const q = loadUserQuery + "id = $1"
	return loadUser(tx, q, id)
}
