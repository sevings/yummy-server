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
	SELECT id, name, show_name,
	avatar,
	gender, is_daylog,
	privacy,
	title, rank, 
	extract(epoch from created_at), extract(epoch from last_seen_at), is_online,
	age,
	entries_count, followings_count, followers_count, 
	ignored_count, invited_count, comments_count, 
	favorites_count, tags_count,
	country, city,
	cover,
	css, background_color, text_color, 
	font_family, font_size, text_alignment, 
	birthday, email, verified,
	extract(epoch from invite_ban), extract(epoch from vote_ban),
	invited_by_id, 
	invited_by_name, invited_by_show_name,
	invited_by_is_online, 
	invited_by_avatar
	FROM long_users 
	WHERE id = $1`

	var profile models.AuthProfile
	profile.InvitedBy = &models.User{}
	profile.Design = &models.Design{}
	profile.Account = &models.AuthProfileAO1Account{}
	profile.Counts = &models.FriendAO1Counts{}
	profile.Ban = &models.AuthProfileAO1Ban{}

	var backColor string
	var textColor string

	var age sql.NullInt64
	var bday sql.NullString
	var avatar, cover string
	var invitedAvatar string

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
		&profile.Counts.Favorites, &profile.Counts.Tags,
		&profile.Country, &profile.City,
		&cover,
		&profile.Design.CSS, &backColor, &textColor,
		&profile.Design.FontFamily, &profile.Design.FontSize, &profile.Design.TextAlignment,
		&bday, &profile.Account.Email, &profile.Account.Verified,
		&profile.Ban.Invite, &profile.Ban.Vote,
		&profile.InvitedBy.ID,
		&profile.InvitedBy.Name, &profile.InvitedBy.ShowName,
		&profile.InvitedBy.IsOnline,
		&invitedAvatar)

	profile.Design.BackgroundColor = models.Color(backColor)
	profile.Design.TextColor = models.Color(textColor)

	profile.Avatar = srv.NewAvatar(avatar)
	profile.InvitedBy.Avatar = srv.NewAvatar(invitedAvatar)

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

	profile.Cover = srv.NewCover(profile.ID, cover)

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

	const loadQuery = profileQuery + "WHERE long_users.id = $1"
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

func loadRelatedToMeUsers(srv *utils.MindwellServer, userID *models.UserID, query, relation string, args ...interface{}) middleware.Responder {
	return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
		list := loadRelatedUsers(srv, tx, query, loadUserQueryName, relation, append([]interface{}{userID.Name}, args...)...)
		if tx.Error() != nil && tx.Error() != sql.ErrNoRows {
			err := srv.NewError(nil)
			return me.NewPutMeCoverBadRequest().WithPayload(err)
		}

		return me.NewGetMeFollowersOK().WithPayload(list)
	})
}

func newMyFollowersLoader(srv *utils.MindwellServer) func(me.GetMeFollowersParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeFollowersParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(srv, userID, usersQueryToName, models.FriendListRelationFollowers,
			models.RelationshipRelationFollowed, *params.Limit, *params.Skip)
	}
}

func newMyFollowingsLoader(srv *utils.MindwellServer) func(me.GetMeFollowingsParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeFollowingsParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(srv, userID, usersQueryFromName, models.FriendListRelationFollowings,
			models.RelationshipRelationFollowed, *params.Limit, *params.Skip)
	}
}

func newMyInvitedLoader(srv *utils.MindwellServer) func(me.GetMeInvitedParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeInvitedParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(srv, userID, invitedUsersQuery, models.FriendListRelationInvited,
			*params.Limit, *params.Skip)
	}
}

func newMyIgnoredLoader(srv *utils.MindwellServer) func(me.GetMeIgnoredParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeIgnoredParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(srv, userID, usersQueryFromName, models.FriendListRelationIgnored,
			models.RelationshipRelationIgnored, *params.Limit, *params.Skip)
	}
}

func newMyRequestedLoader(srv *utils.MindwellServer) func(me.GetMeRequestedParams, *models.UserID) middleware.Responder {
	return func(params me.GetMeRequestedParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(srv, userID, usersQueryFromName, models.FriendListRelationRequested,
			models.RelationshipRelationRequested, *params.Limit, *params.Skip)
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
