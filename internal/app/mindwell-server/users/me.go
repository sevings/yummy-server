package users

import (
	"database/sql"
	"strings"
	"time"

	"github.com/sevings/mindwell-server/utils"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/me"
)

func loadMyProfile(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID) *models.AuthProfile {
	const q = `
	SELECT users.id, users.name, users.show_name,
	users.avatar,
	gender.type, users.is_daylog,
	user_privacy.type,
	users.title, users.rank, 
	extract(epoch from users.created_at), extract(epoch from users.last_seen_at), is_online(users.last_seen_at),
	user_age(users.birthday),
	users.entries_count, users.followings_count, users.followers_count, 
	users.ignored_count, users.invited_count, users.comments_count, 
	users.favorites_count, users.tags_count, CURRENT_DATE - users.created_at::date,
	users.country, users.city,
	users.cover,
	users.css, users.background_color, users.text_color, 
	font_family.type, users.font_size, alignment.type, 
	users.birthday, users.email, users.verified,
	extract(epoch from users.invite_ban), extract(epoch from users.vote_ban),
	extract(epoch from users.comment_ban), extract(epoch from users.live_ban),
	users.invited_by, 
	invited_by.name, invited_by.show_name,
	is_online(invited_by.last_seen_at), 
	invited_by.avatar
	FROM users 
	INNER JOIN gender ON gender.id = users.gender
	INNER JOIN user_privacy ON users.privacy = user_privacy.id
	INNER JOIN font_family ON users.font_family = font_family.id
	INNER JOIN alignment ON users.text_alignment = alignment.id
	LEFT JOIN users AS invited_by ON users.invited_by = invited_by.id
	WHERE users.id = $1`

	var profile models.AuthProfile
	profile.Design = &models.Design{}
	profile.Account = &models.AuthProfileAO1Account{}
	profile.Counts = &models.FriendAO1Counts{}
	profile.Ban = &models.AuthProfileAO1Ban{}

	var backColor string
	var textColor string

	var age sql.NullInt64
	var bday sql.NullString
	var avatar, cover string

	var invitedByID sql.NullInt64
	var invitedByName, invitedByShowName sql.NullString
	var invitedByIsOnline sql.NullBool
	var invitedByAvatar sql.NullString

	tx.Query(q, userID.ID)
	tx.Scan(&profile.ID, &profile.Name, &profile.ShowName,
		&avatar,
		&profile.Gender, &profile.IsDaylog,
		&profile.Privacy,
		&profile.Title, &profile.Rank,
		&profile.CreatedAt, &profile.LastSeenAt, &profile.IsOnline,
		&age,
		&profile.Counts.Entries, &profile.Counts.Followings, &profile.Counts.Followers,
		&profile.Counts.Ignored, &profile.Counts.Invited, &profile.Counts.Comments,
		&profile.Counts.Favorites, &profile.Counts.Tags, &profile.Counts.Days,
		&profile.Country, &profile.City,
		&cover,
		&profile.Design.CSS, &backColor, &textColor,
		&profile.Design.FontFamily, &profile.Design.FontSize, &profile.Design.TextAlignment,
		&bday, &profile.Account.Email, &profile.Account.Verified,
		&profile.Ban.Invite, &profile.Ban.Vote,
		&profile.Ban.Comment, &profile.Ban.Live,
		&invitedByID,
		&invitedByName, &invitedByShowName,
		&invitedByIsOnline,
		&invitedByAvatar)

	profile.Design.BackgroundColor = models.Color(backColor)
	profile.Design.TextColor = models.Color(textColor)

	profile.Avatar = srv.NewAvatar(avatar)

	if invitedByID.Valid {
		profile.InvitedBy = &models.User{
			ID: invitedByID.Int64,
		}

		if invitedByName.Valid {
			profile.InvitedBy.Name = invitedByName.String
			profile.InvitedBy.ShowName = invitedByShowName.String
			profile.InvitedBy.IsOnline = invitedByIsOnline.Bool
			profile.InvitedBy.Avatar = srv.NewAvatar(invitedByAvatar.String)
		}
	}

	profile.Account.Email = utils.HideEmail(profile.Account.Email)

	if bday.Valid {
		profile.Birthday = bday.String
	}

	if age.Valid {
		profile.AgeLowerBound = age.Int64 - age.Int64%5
		profile.AgeUpperBound = profile.AgeLowerBound + 5
	}

	now := float64(time.Now().Unix())

	if profile.Ban.Invite <= now {
		profile.Ban.Invite = 0
	}

	if profile.Ban.Vote <= now {
		profile.Ban.Vote = 0
	}

	if profile.Ban.Comment <= now {
		profile.Ban.Comment = 0
	}

	if profile.Ban.Live <= now {
		profile.Ban.Live = 0
	}

	profile.Cover = srv.NewCover(profile.ID, cover)

	profile.Relations = &models.ProfileAO1Relations{
		IsOpenForMe: true,
	}

	return &profile
}

func newMeLoader(srv *utils.MindwellServer) func(me.GetMeParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			user := loadMyProfile(srv, tx, userID)

			if tx.Error() != nil {
				err := srv.NewError(nil)
				return me.NewPutMeCoverBadRequest().WithPayload(err)
			}

			return me.NewGetMeOK().WithPayload(user)
		})
	}
}

func editMyProfile(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, params me.PutMeParams) *models.Profile {
	id := userID.ID

	if params.Birthday != nil && len(*params.Birthday) > 0 {
		const q = "update users set birthday = $2 where id = $1"
		tx.Exec(q, id, *params.Birthday)
	}

	if params.City != nil {
		const q = "update users set city = $2 where id = $1"
		city := strings.TrimSpace(*params.City)
		tx.Exec(q, id, city)
	}

	if params.Country != nil {
		const q = "update users set country = $2 where id = $1"
		country := strings.TrimSpace(*params.Country)
		tx.Exec(q, id, country)
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
	showName := strings.TrimSpace(params.ShowName)
	tx.Exec(q, id, params.Privacy, showName)

	if params.ShowInTops != nil {
		const q = "update users set show_in_tops = $2 where id = $1"
		tx.Exec(q, id, *params.ShowInTops)
	}

	if params.Title != nil {
		const q = "update users set title = $2 where id = $1"
		title := strings.TrimSpace(*params.Title)
		tx.Exec(q, id, title)
	}

	const loadQuery = profileQuery + "WHERE users.id = $1"
	return loadUserProfile(srv, tx, loadQuery, userID, id)
}

func newMeEditor(srv *utils.MindwellServer) func(me.PutMeParams, *models.UserID) middleware.Responder {
	return func(params me.PutMeParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			user := editMyProfile(srv, tx, userID, params)

			if tx.Error() != nil {
				err := srv.NewError(nil)
				return me.NewPutMeCoverBadRequest().WithPayload(err)
			}

			return me.NewPutMeOK().WithPayload(user)
		})
	}
}

func newMyFollowersLoader(srv *utils.MindwellServer) func(me.GetMeFollowersParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeFollowersParams, userID *models.UserID) middleware.Responder {
		return loadRelatedUsers(srv, userID, usersQueryToName, usersToNameQueryWhere,
			models.RelationshipRelationFollowed, userID.Name, models.FriendListRelationFollowers,
			*params.After, *params.Before, *params.Limit)
	}
}

func newMyFollowingsLoader(srv *utils.MindwellServer) func(me.GetMeFollowingsParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeFollowingsParams, userID *models.UserID) middleware.Responder {
		return loadRelatedUsers(srv, userID, usersQueryFromName, usersFromNameQueryWhere,
			models.RelationshipRelationFollowed, userID.Name, models.FriendListRelationFollowings,
			*params.After, *params.Before, *params.Limit)
	}
}

func newMyInvitedLoader(srv *utils.MindwellServer) func(me.GetMeInvitedParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeInvitedParams, userID *models.UserID) middleware.Responder {
		return loadInvitedUsers(srv, userID, invitedUsersQuery, invitedByQueryWhere,
			userID.Name, *params.After, *params.Before, *params.Limit)
	}
}

func newMyIgnoredLoader(srv *utils.MindwellServer) func(me.GetMeIgnoredParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeIgnoredParams, userID *models.UserID) middleware.Responder {
		return loadRelatedUsers(srv, userID, usersQueryFromName, usersFromNameQueryWhere,
			models.RelationshipRelationIgnored, userID.Name, models.FriendListRelationIgnored,
			*params.After, *params.Before, *params.Limit)
	}
}

func newMyHiddenLoader(srv *utils.MindwellServer) func(me.GetMeHiddenParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeHiddenParams, userID *models.UserID) middleware.Responder {
		return loadRelatedUsers(srv, userID, usersQueryFromName, usersFromNameQueryWhere,
			models.RelationshipRelationHidden, userID.Name, models.FriendListRelationHidden,
			*params.After, *params.Before, *params.Limit)
	}
}

func newMyRequestedLoader(srv *utils.MindwellServer) func(me.GetMeRequestedParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeRequestedParams, userID *models.UserID) middleware.Responder {
		return loadRelatedUsers(srv, userID, usersQueryFromName, usersFromNameQueryWhere,
			models.RelationshipRelationRequested, userID.Name, models.FriendListRelationRequested,
			*params.After, *params.Before, *params.Limit)
	}
}

func newMyOnlineSetter(srv *utils.MindwellServer) func(me.PutMeOnlineParams, *models.UserID) middleware.Responder {
	return func(params me.PutMeOnlineParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			id := userID.ID
			const q = `UPDATE users SET last_seen_at = DEFAULT WHERE id = $1`
			tx.Exec(q, id)
			return me.NewPutMeOnlineNoContent()
		})
	}
}
