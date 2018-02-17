package users

import (
	"database/sql"

	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations/me"
)

func loadMyProfile(tx *utils.AutoTx, userID *models.UserID) *models.AuthProfile {
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
	invited_by_avatar
	FROM long_users 
	WHERE id = $1`

	var profile models.AuthProfile
	profile.InvitedBy = &models.User{}
	profile.Design = &models.Design{}
	profile.Counts = &models.ProfileAllOf1Counts{}

	var backColor string
	var textColor string

	var age sql.NullInt64
	var bday sql.NullString

	tx.Query(q, *userID)
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
		&bday,
		&profile.InvitedBy.ID,
		&profile.InvitedBy.Name, &profile.InvitedBy.ShowName,
		&profile.InvitedBy.IsOnline,
		&profile.InvitedBy.Avatar)

	profile.Design.BackgroundColor = models.Color(backColor)
	profile.Design.TextColor = models.Color(textColor)

	if bday.Valid {
		profile.Birthday = bday.String
	}

	if age.Valid {
		profile.AgeLowerBound = age.Int64 - age.Int64%5
		profile.AgeUpperBound = profile.AgeLowerBound + 5
	}

	return &profile
}

func newMeLoader(db *sql.DB) func(me.GetUsersMeParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeParams, userID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			user := loadMyProfile(tx, userID)

			if tx.Error() != nil {
				return me.NewGetUsersMeForbidden()
			}

			return me.NewGetUsersMeOK().WithPayload(user)
		})
	}
}

func loadRelatedToMeUsers(db *sql.DB, userID *models.UserID, query string, args ...interface{}) middleware.Responder {
	return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
		id := int64(*userID)
		list := loadRelatedUsers(tx, query, append([]interface{}{id}, args...)...)
		if tx.Error() != nil {
			return me.NewGetUsersMeFollowersForbidden()
		}

		return me.NewGetUsersMeFollowersOK().WithPayload(list)
	})
}

func newMyFollowersLoader(db *sql.DB) func(me.GetUsersMeFollowersParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeFollowersParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(db, userID, usersQueryToID,
			models.RelationshipRelationFollowed, *params.Limit, *params.Skip)
	}
}

func newMyFollowingsLoader(db *sql.DB) func(me.GetUsersMeFollowingsParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeFollowingsParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(db, userID, usersQueryFromID,
			models.RelationshipRelationFollowed, *params.Limit, *params.Skip)
	}
}

func newMyInvitedLoader(db *sql.DB) func(me.GetUsersMeInvitedParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeInvitedParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(db, userID, invitedUsersQuery,
			*params.Limit, *params.Skip)
	}
}

func newMyIgnoredLoader(db *sql.DB) func(me.GetUsersMeIgnoredParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeIgnoredParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(db, userID, usersQueryFromID,
			models.RelationshipRelationIgnored, *params.Limit, *params.Skip)
	}
}

func newMyRequestedLoader(db *sql.DB) func(me.GetUsersMeRequestedParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeRequestedParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(db, userID, usersQueryFromID,
			models.RelationshipRelationRequested, *params.Limit, *params.Skip)
	}
}
