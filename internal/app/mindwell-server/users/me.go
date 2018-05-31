package users

import (
	"database/sql"

	"github.com/sevings/mindwell-server/utils"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/me"
)

func loadMyProfile(tx *utils.AutoTx, userID *models.UserID) *models.AuthProfile {
	const q = `
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
	var avatar string
	var invitedAvatar string

	tx.Query(q, *userID)
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
		&profile.Cover,
		&profile.Design.CSS, &backColor, &textColor,
		&profile.Design.FontFamily, &profile.Design.FontSize, &profile.Design.TextAlignment,
		&bday,
		&profile.InvitedBy.ID,
		&profile.InvitedBy.Name, &profile.InvitedBy.ShowName,
		&profile.InvitedBy.IsOnline,
		&invitedAvatar)

	profile.Design.BackgroundColor = models.Color(backColor)
	profile.Design.TextColor = models.Color(textColor)

	profile.Avatar = utils.NewAvatar(avatar)
	profile.InvitedBy.Avatar = utils.NewAvatar(invitedAvatar)

	if bday.Valid {
		profile.Birthday = bday.String
	}

	if age.Valid {
		profile.AgeLowerBound = age.Int64 - age.Int64%5
		profile.AgeUpperBound = profile.AgeLowerBound + 5
	}

	profile.Cover = utils.CoverUrl(profile.Cover)

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

func editMyProfile(tx *utils.AutoTx, userID *models.UserID, params me.PutUsersMeParams) *models.Profile {
	id := int64(*userID)

	if params.Birthday != nil && len(*params.Birthday) > 0 {
		const q = "update users set birthday = $2 where id = $1"
		tx.Exec(q, id, *params.Birthday)
	}

	if params.City != nil {
		const q = "update users set city = $2 where id = $1"
		tx.Exec(q, id, *params.City)
	}

	if params.Country != nil {
		const q = "update users set country = $2 where id = $1"
		tx.Exec(q, id, *params.Country)
	}

	if params.Gender != nil {
		const q = "update users set gender = (select id from gender where type = $2) where id = $1"
		tx.Exec(q, id, *params.Gender)
	}

	if params.IsDaylog != nil {
		const q = "update users set is_daylog = $2 where id = $1"
		tx.Exec(q, id, *params.IsDaylog)
	}

	const q = "update users set privacy = (select id from user_privacy where type = $2), show_name = $3 where id = $1"
	tx.Exec(q, id, params.Privacy, params.ShowName)

	if params.ShowInTops != nil {
		const q = "update users set show_in_tops = $2 where id = $1"
		tx.Exec(q, id, *params.ShowInTops)
	}

	if params.Title != nil {
		const q = "update users set title = $2 where id = $1"
		tx.Exec(q, id, *params.Title)
	}

	const loadQuery = profileQuery + "WHERE long_users.id = $1"
	return loadUserProfile(tx, loadQuery, userID, id)
}

func newMeEditor(db *sql.DB) func(me.PutUsersMeParams, *models.UserID) middleware.Responder {
	return func(params me.PutUsersMeParams, userID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			user := editMyProfile(tx, userID, params)

			if tx.Error() != nil {
				return me.NewPutUsersMeForbidden()
			}

			return me.NewPutUsersMeOK().WithPayload(user)
		})
	}
}

func loadRelatedToMeUsers(db *sql.DB, userID *models.UserID, query, relation string, args ...interface{}) middleware.Responder {
	return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
		id := int64(*userID)
		list := loadRelatedUsers(tx, query, loadUserQueryID, relation, append([]interface{}{id}, args...)...)
		if tx.Error() != nil && tx.Error() != sql.ErrNoRows {
			return me.NewGetUsersMeFollowersForbidden()
		}

		return me.NewGetUsersMeFollowersOK().WithPayload(list)
	})
}

func newMyFollowersLoader(db *sql.DB) func(me.GetUsersMeFollowersParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeFollowersParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(db, userID, usersQueryToID, models.UserListRelationFollowers,
			models.RelationshipRelationFollowed, *params.Limit, *params.Skip)
	}
}

func newMyFollowingsLoader(db *sql.DB) func(me.GetUsersMeFollowingsParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeFollowingsParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(db, userID, usersQueryFromID, models.UserListRelationFollowings,
			models.RelationshipRelationFollowed, *params.Limit, *params.Skip)
	}
}

func newMyInvitedLoader(db *sql.DB) func(me.GetUsersMeInvitedParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeInvitedParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(db, userID, invitedUsersQuery, models.UserListRelationInvited,
			*params.Limit, *params.Skip)
	}
}

func newMyIgnoredLoader(db *sql.DB) func(me.GetUsersMeIgnoredParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeIgnoredParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(db, userID, usersQueryFromID, models.UserListRelationIgnored,
			models.RelationshipRelationIgnored, *params.Limit, *params.Skip)
	}
}

func newMyRequestedLoader(db *sql.DB) func(me.GetUsersMeRequestedParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeRequestedParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(db, userID, usersQueryFromID, models.UserListRelationRequested,
			models.RelationshipRelationRequested, *params.Limit, *params.Skip)
	}
}

func newMyOnlineSetter(db *sql.DB) func(me.PutUsersMeOnlineParams, *models.UserID) middleware.Responder {
	return func(params me.PutUsersMeOnlineParams, userID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			id := int64(*userID)
			const q = `UPDATE users SET last_seen_at = DEFAULT WHERE id = $1`
			tx.Exec(q, id)
			return me.NewPutUsersMeOnlineOK()
		})
	}
}
