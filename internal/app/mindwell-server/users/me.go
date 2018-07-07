package users

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"log"
	"net/http"
	"os"

	"github.com/sevings/mindwell-server/utils"

	"github.com/go-openapi/runtime"
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
	birthday, email, verified,
	invited_by_id, 
	invited_by_name, invited_by_show_name,
	invited_by_is_online, 
	invited_by_avatar
	FROM long_users 
	WHERE id = $1`

	var profile models.AuthProfile
	profile.InvitedBy = &models.User{}
	profile.Design = &models.Design{}
	profile.Account = &models.AuthProfileAllOf1Account{}
	profile.Counts = &models.FriendAllOf1Counts{}

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
		&bday, &profile.Account.Email, &profile.Account.Verified,
		&profile.InvitedBy.ID,
		&profile.InvitedBy.Name, &profile.InvitedBy.ShowName,
		&profile.InvitedBy.IsOnline,
		&invitedAvatar)

	profile.Design.BackgroundColor = models.Color(backColor)
	profile.Design.TextColor = models.Color(textColor)

	profile.Avatar = srv.NewAvatar(avatar)
	profile.InvitedBy.Avatar = srv.NewAvatar(invitedAvatar)

	if bday.Valid {
		profile.Birthday = bday.String
	}

	if age.Valid {
		profile.AgeLowerBound = age.Int64 - age.Int64%5
		profile.AgeUpperBound = profile.AgeLowerBound + 5
	}

	profile.Cover = srv.NewCover(profile.ID, cover)

	return &profile
}

func newMeLoader(srv *utils.MindwellServer) func(me.GetUsersMeParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			user := loadMyProfile(srv, tx, userID)

			if tx.Error() != nil {
				err := srv.NewError(nil)
				return me.NewGetUsersMeForbidden().WithPayload(err)
			}

			return me.NewGetUsersMeOK().WithPayload(user)
		})
	}
}

func storeAvatar(tx *utils.AutoTx, userID int64, avatar *runtime.File) error {
	var oldPath string
	tx.Query("select avatar_path from users where id = $1", userID).Scan(&oldPath)

	if err := os.Remove(oldPath); err != nil {
		log.Print(err)
	}

	// imaginary -a 127.0.0.1 -p 7000 -mount $HOME/images -http-cache-ttl 31556926 -enable-url-signature -url-signature-key 4f46feebafc4b5e988f131c4ff8b5997 -url-signature-salt 88f131c4ff8b59974f46feebafc4b5e9
	signKey := "4f46feebafc4b5e988f131c4ff8b5997"
	signSalt := "88f131c4ff8b59974f46feebafc4b5e9"
	urlPath := "/thumbnail"
	urlQuery := "width=1280&height=1280&type=jpeg"

	h := hmac.New(sha256.New, []byte(signKey))
	h.Write([]byte(urlPath))
	h.Write([]byte(urlQuery))
	h.Write([]byte(signSalt))
	buf := h.Sum(nil)
	sign := base64.RawURLEncoding.EncodeToString(buf)

	url := "http://127.0.0.1:7000" + urlPath + "?" + urlQuery + "&sign=" + sign
	req, err := http.NewRequest("POST", url, avatar.Data)
	if err != nil {
		return err
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return err
	}

	image := make([]byte, 0)
	_, err = resp.Body.Read(image)
	resp.Body.Close()
	if err != nil {
		return err
	}

	path := "a/aaaaa.jpeg"
	file, err := os.Create("../avatars/" + path)
	if err != nil {
		return err
	}

	_, err = file.Write(image)
	if err != nil {
		return err
	}

	tx.Exec("update users set avatar = $1, avatar_path = $2 where id = $3", "/avatars/"+path, path, userID)
	return nil
}

func editMyProfile(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, params me.PutUsersMeParams) *models.Profile {
	id := userID.ID

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
	return loadUserProfile(srv, tx, loadQuery, userID, id)
}

func newMeEditor(srv *utils.MindwellServer) func(me.PutUsersMeParams, *models.UserID) middleware.Responder {
	return func(params me.PutUsersMeParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			user := editMyProfile(srv, tx, userID, params)

			if tx.Error() != nil {
				err := srv.NewError(nil)
				return me.NewPutUsersMeForbidden().WithPayload(err)
			}

			return me.NewPutUsersMeOK().WithPayload(user)
		})
	}
}

func loadRelatedToMeUsers(srv *utils.MindwellServer, userID *models.UserID, query, relation string, args ...interface{}) middleware.Responder {
	return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
		id := userID.ID
		list := loadRelatedUsers(srv, tx, query, loadUserQueryID, relation, append([]interface{}{id}, args...)...)
		if tx.Error() != nil && tx.Error() != sql.ErrNoRows {
			err := srv.NewError(nil)
			return me.NewGetUsersMeFollowersForbidden().WithPayload(err)
		}

		return me.NewGetUsersMeFollowersOK().WithPayload(list)
	})
}

func newMyFollowersLoader(srv *utils.MindwellServer) func(me.GetUsersMeFollowersParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeFollowersParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(srv, userID, usersQueryToID, models.FriendListRelationFollowers,
			models.RelationshipRelationFollowed, *params.Limit, *params.Skip)
	}
}

func newMyFollowingsLoader(srv *utils.MindwellServer) func(me.GetUsersMeFollowingsParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeFollowingsParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(srv, userID, usersQueryFromID, models.FriendListRelationFollowings,
			models.RelationshipRelationFollowed, *params.Limit, *params.Skip)
	}
}

func newMyInvitedLoader(srv *utils.MindwellServer) func(me.GetUsersMeInvitedParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeInvitedParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(srv, userID, invitedUsersQuery, models.FriendListRelationInvited,
			*params.Limit, *params.Skip)
	}
}

func newMyIgnoredLoader(srv *utils.MindwellServer) func(me.GetUsersMeIgnoredParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeIgnoredParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(srv, userID, usersQueryFromID, models.FriendListRelationIgnored,
			models.RelationshipRelationIgnored, *params.Limit, *params.Skip)
	}
}

func newMyRequestedLoader(srv *utils.MindwellServer) func(me.GetUsersMeRequestedParams, *models.UserID) middleware.Responder {
	return func(params me.GetUsersMeRequestedParams, userID *models.UserID) middleware.Responder {
		return loadRelatedToMeUsers(srv, userID, usersQueryFromID, models.FriendListRelationRequested,
			models.RelationshipRelationRequested, *params.Limit, *params.Skip)
	}
}

func newMyOnlineSetter(srv *utils.MindwellServer) func(me.PutUsersMeOnlineParams, *models.UserID) middleware.Responder {
	return func(params me.PutUsersMeOnlineParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			id := userID.ID
			const q = `UPDATE users SET last_seen_at = DEFAULT WHERE id = $1`
			tx.Exec(q, id)
			return me.NewPutUsersMeOnlineOK()
		})
	}
}
