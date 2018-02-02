package account

import (
	"crypto/sha256"
	"database/sql"
	"log"
	"math/rand"
	"strings"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"
	"github.com/sevings/yummy-server/restapi/operations/account"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.YummyAPI) {
	api.AccountGetAccountEmailEmailHandler = account.GetAccountEmailEmailHandlerFunc(newEmailChecker(db))
	api.AccountGetAccountNameNameHandler = account.GetAccountNameNameHandlerFunc(newNameChecker(db))
	api.AccountPostAccountRegisterHandler = account.PostAccountRegisterHandlerFunc(newRegistrator(db))
	api.AccountPostAccountLoginHandler = account.PostAccountLoginHandlerFunc(newLoginer(db))
	api.AccountPostAccountPasswordHandler = account.PostAccountPasswordHandlerFunc(newPasswordUpdater(db))
	api.AccountGetAccountInvitesHandler = account.GetAccountInvitesHandlerFunc(newInvitesLoader(db))
}

// IsEmailFree returns true if there is no account with such an email
func isEmailFree(tx utils.AutoTx, email string) bool {
	const q = `
        select id 
        from users 
		where lower(email) = $1`

	var id int64
	err := tx.QueryRow(q, strings.ToLower(email)).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return true
	case err != nil:
		log.Print(err)
	}

	return false
}

func newEmailChecker(db *sql.DB) func(account.GetAccountEmailEmailParams) middleware.Responder {
	return func(params account.GetAccountEmailEmailParams) middleware.Responder {
		return utils.Transact(db, func(tx utils.AutoTx) (middleware.Responder, bool) {
			free := isEmailFree(tx, params.Email)
			data := models.GetAccountEmailEmailOKBody{Email: &params.Email, IsFree: &free}
			return account.NewGetAccountEmailEmailOK().WithPayload(&data), true
		})
	}
}

func isNameFree(tx utils.AutoTx, name string) bool {
	const q = `
        select id 
        from users 
		where lower(name) = $1`

	var id int64
	err := tx.QueryRow(q, strings.ToLower(name)).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return true
	case err != nil:
		log.Print(err)
	}

	return false
}

func newNameChecker(db *sql.DB) func(account.GetAccountNameNameParams) middleware.Responder {
	return func(params account.GetAccountNameNameParams) middleware.Responder {
		return utils.Transact(db, func(tx utils.AutoTx) (middleware.Responder, bool) {
			free := isNameFree(tx, params.Name)
			data := models.GetAccountNameNameOKBody{Name: &params.Name, IsFree: &free}
			return account.NewGetAccountNameNameOK().WithPayload(&data), true
		})
	}
}

func removeInvite(tx utils.AutoTx, ref string, invite string) (int64, bool) {
	words := strings.Fields(invite)
	if len(words) != 3 {
		return 0, false
	}

	const q = `
        select id, user_id 
        from unwrapped_invites 
        where word1 = $1 and word2 = $2 and word3 = $3
        and name = $4`

	var inviteID int64
	var userID int64
	err := tx.QueryRow(q,
		strings.ToLower(words[0]),
		strings.ToLower(words[1]),
		strings.ToLower(words[2]),
		strings.ToLower(ref)).Scan(&inviteID, &userID)
	switch {
	case err == sql.ErrNoRows:
		return 0, false
	case err != nil:
		log.Print(err)
		return 0, false
	}

	res, err := tx.Exec(`
		delete from invites 
		where id = $1`,
		inviteID)
	if err != nil {
		log.Print(err)
		return 0, false
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Print(err)
		return 0, false
	}

	return userID, rows == 1
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func generateAPIKey() string {
	b := make([]byte, 32)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := len(b)-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func passwordHash(password string) []byte {
	const salt = "RZZer3fSMd1K0DZpYdJe"
	sum := sha256.Sum256([]byte(password + salt))
	return sum[:]
}

func createUser(tx utils.AutoTx, params account.PostAccountRegisterParams, ref int64) (int64, error) {
	hash := passwordHash(params.Password)
	apiKey := generateAPIKey()

	const q = `
		INSERT INTO users 
		(name, show_name, email, password_hash, invited_by, api_key,
		gender, 
		country, city)
		values($1, $1, $2, $3, $4, $5, 
			(select id from gender where type = $6), 
			$7, $8)
		RETURNING id`

	if params.Gender == nil {
		gender := "not set"
		params.Gender = &gender
	}

	if params.Country == nil {
		str := ""
		params.Country = &str
	}

	if params.City == nil {
		str := ""
		params.City = &str
	}

	var user int64
	err := tx.QueryRow(q,
		params.Name, params.Email, hash, ref, apiKey,
		*params.Gender,
		*params.Country, *params.City).Scan(&user)

	if err != nil {
		return 0, err
	}

	if params.Birthday != nil {
		_, err = tx.Exec("UPDATE users SET birthday = $1 WHERE id = $2", *params.Birthday, user)
		if err != nil {
			return 0, err
		}
	}

	return user, nil
}

const authProfileQuery = `
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
email, verified, birthday,
api_key, valid_thru,
invited_by_id, 
invited_by_name, invited_by_show_name,
invited_by_is_online, 
invited_by_avatar
FROM long_users `

func loadAuthProfile(tx utils.AutoTx, query string, args ...interface{}) (*models.AuthProfile, error) {
	row := tx.QueryRow(query, args...)

	var profile models.AuthProfile
	profile.InvitedBy = &models.User{}
	profile.Design = &models.Design{}
	profile.Counts = &models.ProfileAllOf1Counts{}
	profile.Account = &models.AuthProfileAllOf1Account{}

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
		&profile.Account.Email, &profile.Account.Verified, &bday,
		&profile.Account.APIKey, &profile.Account.ValidThru,
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
		profile.AgeUpperBound = profile.AgeLowerBound + 4
	}

	return &profile, nil
}

const authProfileQueryByID = authProfileQuery + "WHERE long_users.id = $1"

func newRegistrator(db *sql.DB) func(account.PostAccountRegisterParams) middleware.Responder {
	return func(params account.PostAccountRegisterParams) middleware.Responder {
		return utils.Transact(db, func(tx utils.AutoTx) (middleware.Responder, bool) {
			if ok := isEmailFree(tx, params.Email); !ok {
				return account.NewPostAccountRegisterBadRequest().WithPayload(utils.NewError("email_is_not_free")), false
			}

			if ok := isNameFree(tx, params.Name); !ok {
				return account.NewPostAccountRegisterBadRequest().WithPayload(utils.NewError("name_is_not_free")), false
			}

			ref, ok := removeInvite(tx, params.Referrer, params.Invite)
			if !ok {
				return account.NewPostAccountRegisterBadRequest().WithPayload(utils.NewError("invalid_invite")), false
			}

			id, err := createUser(tx, params, ref)
			if err != nil {
				log.Print(err)
				return account.NewPostAccountRegisterBadRequest().WithPayload(utils.NewError("internal_error")), false
			}

			user, err := loadAuthProfile(tx, authProfileQueryByID, id)
			if err != nil {
				log.Print(err)
				return account.NewPostAccountRegisterBadRequest().WithPayload(utils.NewError("internal_error")), false
			}

			return account.NewPostAccountRegisterOK().WithPayload(user), true
		})
	}
}

const authProfileQueryByPassword = authProfileQuery + "WHERE long_users.name = $1 and long_users.password_hash = $2"

func newLoginer(db *sql.DB) func(account.PostAccountLoginParams) middleware.Responder {
	return func(params account.PostAccountLoginParams) middleware.Responder {
		return utils.Transact(db, func(tx utils.AutoTx) (middleware.Responder, bool) {
			hash := passwordHash(params.Password)
			user, err := loadAuthProfile(tx, authProfileQueryByPassword, params.Name, hash)
			if err != nil {
				if err != sql.ErrNoRows {
					log.Print(err)
				}
				return account.NewPostAccountLoginBadRequest().WithPayload(utils.NewError("invalid_name_or_password")), false
			}

			return account.NewPostAccountLoginOK().WithPayload(user), true
		})
	}
}

func setPassword(tx utils.AutoTx, params account.PostAccountPasswordParams, userID *models.UserID) (bool, error) {
	const q = `
        update users
        set password_hash = $1
        where password_hash = $2 and id = $3`

	oldHash := passwordHash(params.OldPassword)
	newHash := passwordHash(params.NewPassword)

	res, err := tx.Exec(q, newHash, oldHash, int64(*userID))
	if err != nil {
		return false, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	if rows != 1 {
		return false, nil
	}

	return true, nil
}

func newPasswordUpdater(db *sql.DB) func(account.PostAccountPasswordParams, *models.UserID) middleware.Responder {
	return func(params account.PostAccountPasswordParams, userID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx utils.AutoTx) (middleware.Responder, bool) {
			ok, err := setPassword(tx, params, userID)
			if err != nil {
				log.Print(err)
				return account.NewPostAccountPasswordForbidden().WithPayload(utils.NewError("internal_error")), false
			}

			if !ok {
				return account.NewPostAccountPasswordForbidden().WithPayload(utils.NewError("invalid_password_or_api_key")), false
			}

			return account.NewPostAccountPasswordOK(), true
		})
	}
}

func loadInvites(tx utils.AutoTx, userID *models.UserID) ([]string, error) {
	const q = `
        select word1 || ' ' || word2 || ' ' || word3 
        from unwrapped_invites
        where user_id = $1`

	rows, err := tx.Query(q, int64(*userID))
	if err != nil {
		return nil, err
	}

	var invites []string
	for rows.Next() {
		var invite string
		rows.Scan(&invite)
		invites = append(invites, invite)
	}

	return invites, rows.Err()
}

func newInvitesLoader(db *sql.DB) func(account.GetAccountInvitesParams, *models.UserID) middleware.Responder {
	return func(params account.GetAccountInvitesParams, userID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx utils.AutoTx) (middleware.Responder, bool) {
			invites, err := loadInvites(tx, userID)
			if err != nil {
				log.Print(err)
				return account.NewGetAccountInvitesForbidden().WithPayload(utils.NewError("invalid_api_key")), false
			}

			res := models.GetAccountInvitesOKBody{Invites: invites}
			return account.NewGetAccountInvitesOK().WithPayload(&res), true
		})
	}
}
