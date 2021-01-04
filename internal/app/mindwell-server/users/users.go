package users

import (
	"database/sql"
	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/me"
	"github.com/sevings/mindwell-server/restapi/operations/users"
	"github.com/sevings/mindwell-server/utils"
	"strconv"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	srv.API.MeGetMeHandler = me.GetMeHandlerFunc(newMeLoader(srv))
	srv.API.MePutMeHandler = me.PutMeHandlerFunc(newMeEditor(srv))

	srv.API.UsersGetUsersNameHandler = users.GetUsersNameHandlerFunc(newUserLoader(srv))

	srv.API.MeGetMeFollowersHandler = me.GetMeFollowersHandlerFunc(newMyFollowersLoader(srv))
	srv.API.UsersGetUsersNameFollowersHandler = users.GetUsersNameFollowersHandlerFunc(newFollowersLoader(srv))

	srv.API.MeGetMeFollowingsHandler = me.GetMeFollowingsHandlerFunc(newMyFollowingsLoader(srv))
	srv.API.UsersGetUsersNameFollowingsHandler = users.GetUsersNameFollowingsHandlerFunc(newFollowingsLoader(srv))

	srv.API.MeGetMeInvitedHandler = me.GetMeInvitedHandlerFunc(newMyInvitedLoader(srv))
	srv.API.UsersGetUsersNameInvitedHandler = users.GetUsersNameInvitedHandlerFunc(newInvitedLoader(srv))

	srv.API.MeGetMeIgnoredHandler = me.GetMeIgnoredHandlerFunc(newMyIgnoredLoader(srv))
	srv.API.MeGetMeHiddenHandler = me.GetMeHiddenHandlerFunc(newMyHiddenLoader(srv))
	srv.API.MeGetMeRequestedHandler = me.GetMeRequestedHandlerFunc(newMyRequestedLoader(srv))

	srv.API.MePutMeOnlineHandler = me.PutMeOnlineHandlerFunc(newMyOnlineSetter(srv))

	srv.API.UsersGetUsersHandler = users.GetUsersHandlerFunc(newUsersLoader(srv))
}

const profileQuery = `
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
users.invited_by, 
invited_by.name, invited_by.show_name,
is_online(invited_by.last_seen_at), 
invited_by.avatar
FROM users 
INNER JOIN gender ON gender.id = users.gender
INNER JOIN user_privacy ON users.privacy = user_privacy.id
INNER JOIN font_family ON users.font_family = font_family.id
INNER JOIN alignment ON users.text_alignment = alignment.id
LEFT JOIN users AS invited_by ON users.invited_by = invited_by.id `

func loadUserProfile(srv *utils.MindwellServer, tx *utils.AutoTx, query string, userID *models.UserID, arg interface{}) *models.Profile {
	var profile models.Profile
	profile.Design = &models.Design{}
	profile.Counts = &models.FriendAO1Counts{}

	var backColor string
	var textColor string

	var age sql.NullInt64
	var avatar, cover string

	var invitedByID sql.NullInt64
	var invitedByName, invitedByShowName sql.NullString
	var invitedByIsOnline sql.NullBool
	var invitedByAvatar sql.NullString

	tx.Query(query, arg)
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
		&invitedByID,
		&invitedByName, &invitedByShowName,
		&invitedByIsOnline,
		&invitedByAvatar)

	if userID.ID == 0 && profile.Privacy != "all" {
		return &models.Profile{}
	}

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

	profile.Cover = srv.NewCover(profile.ID, cover)

	if age.Valid {
		profile.AgeLowerBound = age.Int64 - age.Int64%5
		profile.AgeUpperBound = profile.AgeLowerBound + 5
	}

	profile.Relations = &models.ProfileAO1Relations{}

	if profile.ID == userID.ID {
		profile.Relations.IsOpenForMe = true
		return &profile
	}

	const relationQuery = `
			SELECT relation.type
			FROM relations, relation
			WHERE relations.from_id = $1
				AND relations.to_id = $2
				AND relations.type = relation.id`

	profile.Relations.FromMe = relationship(tx, relationQuery, userID.ID, profile.ID)
	profile.Relations.ToMe = relationship(tx, relationQuery, profile.ID, userID.ID)
	profile.Relations.IsOpenForMe = isOpenForMe(&profile, userID)

	return &profile
}

func isOpenForMe(profile *models.Profile, me *models.UserID) bool {
	if profile.ID == me.ID {
		return true
	}

	if profile.Relations.ToMe == models.RelationshipRelationIgnored {
		return false
	}

	switch profile.Privacy {
	case "all":
		return true
	case "registered":
		return me.ID > 0
	case "followers":
		return profile.Relations.FromMe == models.RelationshipRelationFollowed
	case "invited":
		return me.IsInvited
	}

	return false
}

func relationship(tx *utils.AutoTx, query string, from, to int64) string {
	if from == 0 || to == 0 {
		return models.RelationshipRelationNone
	}

	var relation string
	tx.Query(query, from, to).Scan(&relation)
	if tx.Error() == sql.ErrNoRows {
		return models.RelationshipRelationNone
	}

	return relation
}

const usersQuerySelect = `
SELECT users.id, users.name, users.show_name, gender.type,
is_online(users.last_seen_at), extract(epoch from users.last_seen_at), users.title, users.rank,
user_privacy.type, users.avatar, users.cover,
users.entries_count, users.followings_count, users.followers_count, 
users.ignored_count, users.invited_count, users.comments_count, 
users.favorites_count, users.tags_count, CURRENT_DATE - users.created_at::date
`

const usersQueryStart = usersQuerySelect + `, extract(epoch from relations.changed_at)
FROM users, relation, relations, gender, user_privacy
WHERE `

const usersQueryJoins = ` AND users.gender = gender.id AND users.privacy = user_privacy.id `

const usersQueryEnd = `
 AND relations.type = relation.id AND relation.type = $2` + usersQueryJoins

func loadUserList(srv *utils.MindwellServer, tx *utils.AutoTx, reverse bool) ([]*models.Friend, interface{}, interface{}) {
	list := make([]*models.Friend, 0, 50)

	var nextBefore, nextAfter interface{}

	for {
		var user models.Friend
		user.Counts = &models.FriendAO1Counts{}
		var avatar, cover string

		ok := tx.Scan(&user.ID, &user.Name, &user.ShowName, &user.Gender,
			&user.IsOnline, &user.LastSeenAt, &user.Title, &user.Rank,
			&user.Privacy, &avatar, &cover,
			&user.Counts.Entries, &user.Counts.Followings, &user.Counts.Followers,
			&user.Counts.Ignored, &user.Counts.Invited, &user.Counts.Comments,
			&user.Counts.Favorites, &user.Counts.Tags, &user.Counts.Days,
			&nextBefore)
		if !ok {
			break
		}

		user.Avatar = srv.NewAvatar(avatar)
		user.Cover = srv.NewCover(user.ID, cover)
		list = append(list, &user)

		if nextAfter == nil {
			nextAfter = nextBefore
		}
	}

	if reverse {
		for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
			list[i], list[j] = list[j], list[i]
		}
		nextAfter, nextBefore = nextBefore, nextAfter
	}

	return list, nextBefore, nextAfter
}

func loadRelatedUsers(srv *utils.MindwellServer, userID *models.UserID, query, queryWhere, related, name,
	relation, afterS, beforeS string, limit int64) middleware.Responder {
	return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
		open := utils.IsOpenForMe(tx, userID, name)
		if tx.Error() != nil {
			err := srv.NewError(nil)
			return users.NewGetUsersNameFollowersNotFound().WithPayload(err)
		}

		if !open {
			err := srv.StandardError("no_tlog")
			return users.NewGetUsersNameFollowersForbidden().WithPayload(err)
		}

		before := utils.ParseFloat(beforeS)
		after := utils.ParseFloat(afterS)

		if after > 0 {
			q := query + " AND relations.changed_at > to_timestamp($3) ORDER BY relations.changed_at ASC LIMIT $4"
			tx.Query(q, name, related, after, limit)
		} else if before > 0 {
			q := query + " AND relations.changed_at < to_timestamp($3) ORDER BY relations.changed_at DESC LIMIT $4"
			tx.Query(q, name, related, before, limit)
		} else {
			q := query + " ORDER BY relations.changed_at DESC LIMIT $3"
			tx.Query(q, name, related, limit)
		}

		var list models.FriendList
		var nextBefore, nextAfter interface{}
		list.Users, nextBefore, nextAfter = loadUserList(srv, tx, after > 0)

		list.Subject = loadUser(srv, tx, loadUserQueryName, name)
		list.Relation = relation

		if tx.Error() != nil && tx.Error() != sql.ErrNoRows {
			err := srv.StandardError("no_tlog")
			return users.NewGetUsersNameFollowersNotFound().WithPayload(err)
		}

		if len(list.Users) == 0 {
			return users.NewGetUsersNameFollowersOK().WithPayload(&list)
		}

		beforeQuery := `SELECT EXISTS(
		SELECT 1 
		FROM users, relation, relations, gender, user_privacy
		WHERE ` + queryWhere + " AND relations.changed_at < to_timestamp($3))"

		list.NextBefore = utils.FormatFloat(nextBefore.(float64))
		list.HasBefore = tx.QueryBool(beforeQuery, name, related, nextBefore)

		afterQuery := `SELECT EXISTS(
		SELECT 1 
		FROM users, relation, relations, gender, user_privacy
		WHERE ` + queryWhere + " AND relations.changed_at > to_timestamp($3))"

		list.NextAfter = utils.FormatFloat(nextAfter.(float64))
		list.HasAfter = tx.QueryBool(afterQuery, name, related, nextAfter)

		return users.NewGetUsersNameFollowersOK().WithPayload(&list)
	})
}

func loadInvitedUsers(srv *utils.MindwellServer, userID *models.UserID, query, queryWhere, name,
	afterS, beforeS string, limit int64) middleware.Responder {
	return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
		open := utils.IsOpenForMe(tx, userID, name)
		if tx.Error() != nil {
			err := srv.NewError(nil)
			return users.NewGetUsersNameFollowersNotFound().WithPayload(err)
		}

		if !open {
			err := srv.StandardError("no_tlog")
			return users.NewGetUsersNameFollowersForbidden().WithPayload(err)
		}

		before, err := strconv.ParseInt(beforeS, 10, 64)
		if err != nil {
			before = 0
		}

		after, err := strconv.ParseInt(afterS, 10, 64)
		if err != nil {
			after = 0
		}

		if after > 0 {
			q := query + " AND users.id > $2 ORDER BY users.id ASC LIMIT $3"
			tx.Query(q, name, after, limit)
		} else if before > 0 {
			q := query + " AND users.id < $2 ORDER BY users.id DESC LIMIT $3"
			tx.Query(q, name, before, limit)
		} else {
			q := query + " ORDER BY users.id DESC LIMIT $2"
			tx.Query(q, name, limit)
		}

		var list models.FriendList
		var nextBefore, nextAfter interface{}
		list.Users, nextBefore, nextAfter = loadUserList(srv, tx, after > 0)

		list.Subject = loadUser(srv, tx, loadUserQueryName, name)
		list.Relation = models.FriendListRelationInvited

		if tx.Error() != nil && tx.Error() != sql.ErrNoRows {
			err := srv.StandardError("no_tlog")
			return users.NewGetUsersNameFollowersNotFound().WithPayload(err)
		}

		if len(list.Users) == 0 {
			return users.NewGetUsersNameFollowersOK().WithPayload(&list)
		}

		beforeQuery := `SELECT EXISTS(
		SELECT 1 
		FROM users, gender, user_privacy
		WHERE ` + queryWhere + " AND users.id < $2)"

		list.NextBefore = strconv.FormatInt(nextBefore.(int64), 10)
		list.HasBefore = tx.QueryBool(beforeQuery, name, nextBefore)

		afterQuery := `SELECT EXISTS(
		SELECT 1 
		FROM users, gender, user_privacy
		WHERE ` + queryWhere + " AND users.id > $2)"

		list.NextAfter = strconv.FormatInt(nextAfter.(int64), 10)
		list.HasAfter = tx.QueryBool(afterQuery, name, nextAfter)

		return users.NewGetUsersNameFollowersOK().WithPayload(&list)
	})
}

const loadUserQuery = `
SELECT id, name, show_name,
is_online(last_seen_at), avatar
FROM users
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

// LoadUserByName returns short user profile by its name.
func LoadUserByName(srv *utils.MindwellServer, tx *utils.AutoTx, name string) *models.User {
	return loadUser(srv, tx, loadUserQueryName, name)
}
