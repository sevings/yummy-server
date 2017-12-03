package users

import (
	"database/sql"
	"log"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/gen/models"
	"github.com/sevings/yummy-server/gen/restapi/operations/me"
)

func loadMyProfile(db *sql.DB, userID *models.UserID) (*models.AuthProfile, error) {
	const q = `
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
	birthday,
	invited_by_id, 
	invited_by_name, invited_by_show_name,
	invited_by_is_online, 
	invited_by_name_color, invited_by_avatar_color,
	invited_by_avatar
	FROM long_users 
	WHERE id = $1`

	row := db.QueryRow(q, *userID)

	var profile models.AuthProfile
	profile.InvitedBy = &models.User{}
	profile.Design = &models.Design{}
	profile.Counts = &models.ProfileAllOf1Counts{}

	var backColor string
	var textColor string

	var age sql.NullInt64
	var bday sql.NullString

	err := row.Scan(&profile.ID, &profile.Name, &profile.ShowName,
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
		&bday,
		&profile.InvitedBy.ID,
		&profile.InvitedBy.Name, &profile.InvitedBy.ShowName,
		&profile.InvitedBy.IsOnline,
		&profile.InvitedBy.Avatar)

	if err != nil {
		return &profile, err
	}

	profile.Design.BackgroundColor = models.Color(backColor)
	profile.Design.TextColor = models.Color(textColor)

	if bday.Valid {
		profile.Birthday = bday.String
	}

	if age.Valid {
		profile.AgeLowerBound = age.Int64 - age.Int64%5
		profile.AgeUpperBound = profile.AgeLowerBound + 5
	}

	return &profile, nil
}

func newMeLoader(db *sql.DB) func(me.GetUsersMeParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeParams, userID *models.UserID) middleware.Responder {
		user, err := loadMyProfile(db, userID)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Print(err)
			}

			return me.NewGetUsersMeForbidden()
		}

		return me.NewGetUsersMeOK().WithPayload(user)
	}
}

func loadRelatedToMeUsers(db *sql.DB, userID *models.UserID, query, relation string, limit, offset int64) middleware.Responder {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Commit()

	list, err := loadRelatedUsers(tx, query, userID, relation, limit, offset)
	if err != nil {
		log.Print(err)
		return me.NewGetUsersMeFollowersForbidden()
	}

	return me.NewGetUsersMeFollowersOK().WithPayload(list)
}

func newMyFollowersLoader(db *sql.DB) func(me.GetUsersMeFollowersParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeFollowersParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(db, userID, usersQueryToID,
			"followed", *params.Limit, *params.Skip)
	}
}
