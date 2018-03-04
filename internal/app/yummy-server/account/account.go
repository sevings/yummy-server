package account

import (
	"crypto/sha256"
	"database/sql"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/o1egl/govatar"

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
func isEmailFree(tx *utils.AutoTx, email string) bool {
	const q = `
        select id 
        from users 
		where lower(email) = lower($1)`

	var id int64
	tx.Query(q, email).Scan(&id)

	return tx.Error() == sql.ErrNoRows
}

func newEmailChecker(db *sql.DB) func(account.GetAccountEmailEmailParams) middleware.Responder {
	return func(params account.GetAccountEmailEmailParams) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			free := isEmailFree(tx, params.Email)
			data := models.GetAccountEmailEmailOKBody{Email: &params.Email, IsFree: &free}
			return account.NewGetAccountEmailEmailOK().WithPayload(&data)
		})
	}
}

func isNameFree(tx *utils.AutoTx, name string) bool {
	const q = `
        select id 
        from users 
		where lower(name) = lower($1)`

	var id int64
	tx.Query(q, name).Scan(&id)

	return tx.Error() == sql.ErrNoRows
}

func newNameChecker(db *sql.DB) func(account.GetAccountNameNameParams) middleware.Responder {
	return func(params account.GetAccountNameNameParams) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			free := isNameFree(tx, params.Name)
			data := models.GetAccountNameNameOKBody{Name: &params.Name, IsFree: &free}
			return account.NewGetAccountNameNameOK().WithPayload(&data)
		})
	}
}

func removeInvite(tx *utils.AutoTx, ref string, invite string) (int64, bool) {
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
	tx.Query(q,
		strings.ToLower(words[0]),
		strings.ToLower(words[1]),
		strings.ToLower(words[2]),
		strings.ToLower(ref)).Scan(&inviteID, &userID)

	tx.Exec(`
		delete from invites 
		where id = $1`,
		inviteID)

	rows := tx.RowsAffected()

	return userID, rows == 1
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func generateString(length int) string {
	b := make([]byte, length)
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

func generateAvatar(name, gender string) string {
	var g govatar.Gender
	if gender == models.ProfileAllOf1GenderMale {
		g = govatar.MALE
	} else if gender == models.ProfileAllOf1GenderFemale {
		g = govatar.FEMALE
	} else if ch := name[len(name)-1]; ch == 'a' || ch == 'y' || ch == 'u' || ch == 'e' || ch == 'o' || ch == 'i' {
		g = govatar.FEMALE
	} else {
		g = govatar.MALE
	}

	err := os.MkdirAll("../avatars/"+name[:1], 0777)
	if err != nil {
		log.Print(err)
	}

	path := "/avatars/" + name[:1] + "/" + generateString(5) + ".png"
	err = govatar.GenerateFileFromUsername(g, name, ".."+path)
	if err != nil {
		log.Print(err)
	}

	return path
}

func createUser(tx *utils.AutoTx, params account.PostAccountRegisterParams, ref int64) int64 {
	hash := passwordHash(params.Password)
	apiKey := generateString(32)

	const q = `
		INSERT INTO users 
		(name, show_name, email, password_hash, invited_by, api_key,
		gender, 
		country, city, avatar)
		values($1, $1, $2, $3, $4, $5, 
			(select id from gender where type = $6), 
			$7, $8, $9)
		RETURNING id`

	if params.Gender == nil {
		gender := models.ProfileAllOf1GenderNotSet
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

	avatar := generateAvatar(params.Name, *params.Gender)

	var user int64
	tx.Query(q,
		params.Name, params.Email, hash, ref, apiKey,
		*params.Gender,
		*params.Country, *params.City, avatar).Scan(&user)

	if params.Birthday != nil {
		tx.Exec("UPDATE users SET birthday = $1 WHERE id = $2", *params.Birthday, user)
	}

	return user
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

func loadAuthProfile(tx *utils.AutoTx, query string, args ...interface{}) *models.AuthProfile {
	var profile models.AuthProfile
	profile.InvitedBy = &models.User{}
	profile.Design = &models.Design{}
	profile.Counts = &models.ProfileAllOf1Counts{}
	profile.Account = &models.AuthProfileAllOf1Account{}

	var backColor string
	var textColor string

	var age sql.NullInt64
	var bday sql.NullString

	tx.Query(query, args...)
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
		&profile.Account.Email, &profile.Account.Verified, &bday,
		&profile.Account.APIKey, &profile.Account.ValidThru,
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
		profile.AgeUpperBound = profile.AgeLowerBound + 4
	}

	return &profile
}

const authProfileQueryByID = authProfileQuery + "WHERE long_users.id = $1"

func newRegistrator(db *sql.DB) func(account.PostAccountRegisterParams) middleware.Responder {
	return func(params account.PostAccountRegisterParams) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			if ok := isEmailFree(tx, params.Email); !ok {
				return account.NewPostAccountRegisterBadRequest().WithPayload(utils.NewError("email_is_not_free"))
			}

			if ok := isNameFree(tx, params.Name); !ok {
				return account.NewPostAccountRegisterBadRequest().WithPayload(utils.NewError("name_is_not_free"))
			}

			ref, ok := removeInvite(tx, params.Referrer, params.Invite)
			if !ok {
				return account.NewPostAccountRegisterBadRequest().WithPayload(utils.NewError("invalid_invite"))
			}

			id := createUser(tx, params, ref)
			if tx.Error() != nil {
				return account.NewPostAccountRegisterBadRequest().WithPayload(utils.NewError("internal_error"))
			}

			user := loadAuthProfile(tx, authProfileQueryByID, id)
			if tx.Error() != nil {
				return account.NewPostAccountRegisterBadRequest().WithPayload(utils.NewError("internal_error"))
			}

			return account.NewPostAccountRegisterCreated().WithPayload(user)
		})
	}
}

const authProfileQueryByPassword = authProfileQuery + "WHERE lower(long_users.name) = lower($1) and long_users.password_hash = $2"

func newLoginer(db *sql.DB) func(account.PostAccountLoginParams) middleware.Responder {
	return func(params account.PostAccountLoginParams) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			hash := passwordHash(params.Password)
			user := loadAuthProfile(tx, authProfileQueryByPassword, params.Name, hash)
			if tx.Error() != nil {
				return account.NewPostAccountLoginBadRequest().WithPayload(utils.NewError("invalid_name_or_password"))
			}

			return account.NewPostAccountLoginOK().WithPayload(user)
		})
	}
}

func setPassword(tx *utils.AutoTx, params account.PostAccountPasswordParams, userID *models.UserID) bool {
	const q = `
        update users
        set password_hash = $1
        where password_hash = $2 and id = $3`

	oldHash := passwordHash(params.OldPassword)
	newHash := passwordHash(params.NewPassword)

	tx.Exec(q, newHash, oldHash, int64(*userID))

	rows := tx.RowsAffected()

	return rows == 1
}

func newPasswordUpdater(db *sql.DB) func(account.PostAccountPasswordParams, *models.UserID) middleware.Responder {
	return func(params account.PostAccountPasswordParams, userID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			ok := setPassword(tx, params, userID)
			if tx.Error() != nil {
				return account.NewPostAccountPasswordForbidden().WithPayload(utils.NewError("internal_error"))
			}

			if !ok {
				return account.NewPostAccountPasswordForbidden().WithPayload(utils.NewError("invalid_password_or_api_key"))
			}

			return account.NewPostAccountPasswordOK()
		})
	}
}

func loadInvites(tx *utils.AutoTx, userID *models.UserID) []string {
	const q = `
        select word1 || ' ' || word2 || ' ' || word3 
        from unwrapped_invites
        where user_id = $1`

	tx.Query(q, int64(*userID))

	var invites []string
	for {
		var invite string
		if !tx.Scan(&invite) {
			break
		}

		invites = append(invites, invite)
	}

	return invites
}

func newInvitesLoader(db *sql.DB) func(account.GetAccountInvitesParams, *models.UserID) middleware.Responder {
	return func(params account.GetAccountInvitesParams, userID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			invites := loadInvites(tx, userID)
			if tx.Error() != nil && tx.Error() != sql.ErrNoRows {
				return account.NewGetAccountInvitesForbidden().WithPayload(utils.NewError("invalid_api_key"))
			}

			res := models.GetAccountInvitesOKBody{Invites: invites}
			return account.NewGetAccountInvitesOK().WithPayload(&res)
		})
	}
}
