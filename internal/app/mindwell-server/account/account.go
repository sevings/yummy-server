package account

import (
	"database/sql"
	"fmt"
	"image"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/disintegration/imaging"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/o1egl/govatar"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/account"
	"github.com/sevings/mindwell-server/utils"
)

var centSecret []byte
var apiSecret []byte

// ConfigureAPI creates operations handlers
func ConfigureAPI(srv *utils.MindwellServer) {
	centSecret = []byte(srv.ConfigString("centrifugo.secret"))
	apiSecret = []byte(srv.ConfigString("server.api_secret"))

	srv.API.APIKeyHeaderAuth = utils.NewKeyAuth(srv.DB, apiSecret)

	srv.API.AccountGetAccountEmailEmailHandler = account.GetAccountEmailEmailHandlerFunc(newEmailChecker(srv))
	srv.API.AccountGetAccountNameNameHandler = account.GetAccountNameNameHandlerFunc(newNameChecker(srv))

	srv.API.AccountPostAccountRegisterHandler = account.PostAccountRegisterHandlerFunc(newRegistrator(srv))
	srv.API.AccountPostAccountLoginHandler = account.PostAccountLoginHandlerFunc(newLoginer(srv))
	srv.API.AccountPostAccountPasswordHandler = account.PostAccountPasswordHandlerFunc(newPasswordUpdater(srv))
	srv.API.AccountPostAccountEmailHandler = account.PostAccountEmailHandlerFunc(newEmailUpdater(srv))
	srv.API.AccountGetAccountInvitesHandler = account.GetAccountInvitesHandlerFunc(newInvitesLoader(srv))

	srv.API.AccountPostAccountVerificationHandler = account.PostAccountVerificationHandlerFunc(newVerificationSender(srv))
	srv.API.AccountGetAccountVerificationEmailHandler = account.GetAccountVerificationEmailHandlerFunc(newEmailVerifier(srv))

	srv.API.AccountPostAccountRecoverHandler = account.PostAccountRecoverHandlerFunc(newResetPasswordSender(srv))
	srv.API.AccountPostAccountRecoverPasswordHandler = account.PostAccountRecoverPasswordHandlerFunc(newPasswordResetter(srv))

	srv.API.AccountGetAccountSettingsEmailHandler = account.GetAccountSettingsEmailHandlerFunc(newEmailSettingsLoader(srv))
	srv.API.AccountPutAccountSettingsEmailHandler = account.PutAccountSettingsEmailHandlerFunc(newEmailSettingsEditor(srv))

	srv.API.AccountGetAccountSubscribeTokenHandler = account.GetAccountSubscribeTokenHandlerFunc(newConnectionTokenGenerator(srv))
	srv.API.AccountGetAccountSubscribeTelegramHandler = account.GetAccountSubscribeTelegramHandlerFunc(newTelegramTokenGenerator(srv))
	srv.API.AccountDeleteAccountSubscribeTelegramHandler = account.DeleteAccountSubscribeTelegramHandlerFunc(newTelegramDeleter(srv))
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

func newEmailChecker(srv *utils.MindwellServer) func(account.GetAccountEmailEmailParams) middleware.Responder {
	return func(params account.GetAccountEmailEmailParams) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			free := isEmailFree(tx, params.Email)
			data := account.GetAccountEmailEmailOKBody{Email: &params.Email, IsFree: &free}
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

func newNameChecker(srv *utils.MindwellServer) func(account.GetAccountNameNameParams) middleware.Responder {
	return func(params account.GetAccountNameNameParams) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			free := isNameFree(tx, params.Name)
			data := account.GetAccountNameNameOKBody{Name: &params.Name, IsFree: &free}
			return account.NewGetAccountNameNameOK().WithPayload(&data)
		})
	}
}

func removeInvite(tx *utils.AutoTx, invite string) (int64, bool) {
	words := strings.Fields(invite)
	if len(words) != 3 {
		return 0, false
	}

	const q = `
		DELETE FROM invites
		WHERE word1 = (SELECT id FROM invite_words WHERE word = $1)
		  AND word2 = (SELECT id FROM invite_words WHERE word = $2)
		  AND word3 = (SELECT id FROM invite_words WHERE word = $3)
		RETURNING referrer_id`

	var userID int64
	tx.Query(q,
		strings.ToLower(words[0]),
		strings.ToLower(words[1]),
		strings.ToLower(words[2])).Scan(&userID)

	return userID, userID != 0
}

func saveAvatar(srv *utils.MindwellServer, img image.Image, size int, folder, name string) {
	path := srv.ImagesFolder() + "avatars/" + strconv.Itoa(size) + "/" + folder
	err := os.MkdirAll(path, 0777)
	if err != nil {
		log.Print(err)
	}

	w := img.Bounds().Dx()
	if size < w {
		img = imaging.Resize(img, size, size, imaging.CatmullRom)
	}

	err = imaging.Save(img, path+name, imaging.JPEGQuality(85))
	if err != nil {
		log.Print(err)
	}
}

func generateAvatar(srv *utils.MindwellServer, name, gender string) string {
	var g govatar.Gender
	if gender == "male" {
		g = govatar.MALE
	} else if gender == "female" {
		g = govatar.FEMALE
	} else if ch := name[len(name)-1]; ch == 'a' || ch == 'y' || ch == 'u' || ch == 'e' || ch == 'o' || ch == 'i' {
		g = govatar.FEMALE
	} else {
		g = govatar.MALE
	}

	img, err := govatar.GenerateFromUsername(g, name)
	if err != nil {
		log.Print(err)
	}

	folder := name[:1] + "/"
	fileName := utils.GenerateString(5) + ".jpg"

	saveAvatar(srv, img, 124, folder, fileName)
	saveAvatar(srv, img, 92, folder, fileName)
	saveAvatar(srv, img, 42, folder, fileName)

	return folder + fileName
}

func createUser(srv *utils.MindwellServer, tx *utils.AutoTx, params account.PostAccountRegisterParams, ref int64) int64 {
	hash := srv.PasswordHash(params.Password)
	apiKey := utils.GenerateString(32)

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

	avatar := generateAvatar(srv, params.Name, *params.Gender)

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
title, rank, 
extract(epoch from created_at), extract(epoch from last_seen_at), is_online,
age,
entries_count, followings_count, followers_count, 
ignored_count, invited_count, comments_count, 
favorites_count, tags_count, days_count,
country, city,
cover,
css, background_color, text_color, 
font_family, font_size, text_alignment, 
email, verified, birthday,
api_key, extract(epoch from valid_thru),
extract(epoch from invite_ban), extract(epoch from vote_ban),
invited_by_id, 
invited_by_name, invited_by_show_name,
invited_by_is_online, 
invited_by_avatar
FROM long_users `

func loadAuthProfile(srv *utils.MindwellServer, tx *utils.AutoTx, query string, args ...interface{}) *models.AuthProfile {
	var profile models.AuthProfile
	profile.InvitedBy = &models.User{}
	profile.Design = &models.Design{}
	profile.Counts = &models.FriendAO1Counts{}
	profile.Account = &models.AuthProfileAO1Account{}
	profile.Ban = &models.AuthProfileAO1Ban{}

	var backColor string
	var textColor string

	var age sql.NullInt64
	var bday sql.NullString
	var avatar, cover string
	var invitedAvatar string

	tx.Query(query, args...)
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
		&profile.Account.Email, &profile.Account.Verified, &bday,
		&profile.Account.APIKey, &profile.Account.ValidThru,
		&profile.Ban.Invite, &profile.Ban.Vote,
		&profile.InvitedBy.ID,
		&profile.InvitedBy.Name, &profile.InvitedBy.ShowName,
		&profile.InvitedBy.IsOnline,
		&invitedAvatar)

	profile.Design.BackgroundColor = models.Color(backColor)
	profile.Design.TextColor = models.Color(textColor)
	profile.Avatar = srv.NewAvatar(avatar)
	profile.InvitedBy.Avatar = srv.NewAvatar(invitedAvatar)

	// token, thru := utils.BuildApiToken(apiSecret, &models.UserID{
	// 	ID:   profile.ID,
	// 	Name: profile.Name,
	// })
	// profile.Account.APIKey = token
	// profile.Account.ValidThru = float64(thru)

	if bday.Valid {
		profile.Birthday = bday.String
	}

	if age.Valid {
		profile.AgeLowerBound = age.Int64 - age.Int64%5
		profile.AgeUpperBound = profile.AgeLowerBound + 4
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

const authProfileQueryByID = authProfileQuery + "WHERE long_users.id = $1"

func newRegistrator(srv *utils.MindwellServer) func(account.PostAccountRegisterParams) middleware.Responder {
	return func(params account.PostAccountRegisterParams) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			if ok := isEmailFree(tx, params.Email); !ok {
				err := srv.NewError(&i18n.Message{ID: "email_is_not_free", Other: "Email is not free."})
				return account.NewPostAccountRegisterBadRequest().WithPayload(err)
			}

			if ok := isNameFree(tx, params.Name); !ok {
				err := srv.NewError(&i18n.Message{ID: "name_is_not_free", Other: "Name is not free."})
				return account.NewPostAccountRegisterBadRequest().WithPayload(err)
			}

			ref, ok := removeInvite(tx, params.Invite)
			if !ok {
				err := srv.NewError(&i18n.Message{ID: "invalid_invite", Other: "Invite is invalid."})
				return account.NewPostAccountRegisterBadRequest().WithPayload(err)
			}

			id := createUser(srv, tx, params, ref)
			if tx.Error() != nil {
				err := srv.NewError(nil)
				return account.NewPostAccountRegisterBadRequest().WithPayload(err)
			}

			user := loadAuthProfile(srv, tx, authProfileQueryByID, id)
			if tx.Error() != nil {
				err := srv.NewError(nil)
				return account.NewPostAccountRegisterBadRequest().WithPayload(err)
			}

			code := srv.VerificationCode(user.Account.Email)
			srv.Mail.SendGreeting(user.Account.Email, user.ShowName, code)

			user.Account.Email = utils.HideEmail(user.Account.Email)

			return account.NewPostAccountRegisterCreated().WithPayload(user)
		})
	}
}

const authProfileQueryByPassword = authProfileQuery + `
	WHERE password_hash = $2
		AND (lower(name) = lower($1) OR lower(email) = lower($1))
`

func newLoginer(srv *utils.MindwellServer) func(account.PostAccountLoginParams) middleware.Responder {
	return func(params account.PostAccountLoginParams) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			hash := srv.PasswordHash(params.Password)
			user := loadAuthProfile(srv, tx, authProfileQueryByPassword, params.Name, hash)
			if tx.Error() != nil {
				err := srv.NewError(&i18n.Message{ID: "invalid_name_or_password", Other: "Name or password is invalid."})
				return account.NewPostAccountLoginBadRequest().WithPayload(err)
			}

			user.Account.Email = utils.HideEmail(user.Account.Email)

			return account.NewPostAccountLoginOK().WithPayload(user)
		})
	}
}

func setPassword(srv *utils.MindwellServer, tx *utils.AutoTx, params account.PostAccountPasswordParams, userID *models.UserID) bool {
	const q = `
        update users
        set password_hash = $1
        where password_hash = $2 and id = $3`

	oldHash := srv.PasswordHash(params.OldPassword)
	newHash := srv.PasswordHash(params.NewPassword)

	tx.Exec(q, newHash, oldHash, userID.ID)

	rows := tx.RowsAffected()

	return rows == 1
}

func sendPasswordChanged(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID) {
	const q = `
		SELECT email, verified, show_name, telegram
		FROM users
		WHERE id = $1
	`

	var email, name string
	var verified bool
	var tg sql.NullInt64
	tx.Query(q, userID.ID).Scan(&email, &verified, &name, &tg)

	if tx.Error() != nil {
		if tx.Error() != sql.ErrNoRows {
			log.Print(tx.Error())
		}
		return
	}

	if verified {
		srv.Mail.SendPasswordChanged(email, name)
	}

	if tg.Valid {
		srv.Tg.SendPasswordChanged(tg.Int64)
	}
}

func newPasswordUpdater(srv *utils.MindwellServer) func(account.PostAccountPasswordParams, *models.UserID) middleware.Responder {
	return func(params account.PostAccountPasswordParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			ok := setPassword(srv, tx, params, userID)
			if tx.Error() != nil {
				err := srv.NewError(nil)
				return account.NewPostAccountPasswordForbidden().WithPayload(err)
			}

			if !ok {
				err := srv.NewError(&i18n.Message{ID: "invalid_password_or_api_key", Other: "Password or ApiKey is invalid."})
				return account.NewPostAccountPasswordForbidden().WithPayload(err)
			}

			sendPasswordChanged(srv, tx, userID)

			return account.NewPostAccountPasswordOK()
		})
	}
}

func setEmail(srv *utils.MindwellServer, tx *utils.AutoTx, params account.PostAccountEmailParams, userID *models.UserID) (*models.Error, string) {
	var oldEmail string
	var verified bool
	tx.Query("SELECT email, verified FROM users WHERE id = $1", userID.ID).Scan(&oldEmail, &verified)

	newEmail := strings.TrimSpace(params.Email)
	if strings.ToLower(oldEmail) == strings.ToLower(newEmail) {
		return srv.NewError(&i18n.Message{ID: "email_is_the_same", Other: "Email is the same as old one."}), oldEmail
	}

	if !verified {
		oldEmail = ""
	}

	if !isEmailFree(tx, newEmail) {
		return srv.NewError(&i18n.Message{ID: "email_is_used", Other: "Email is already used."}), oldEmail
	}

	const q = `
        update users
        set email = $1, verified = false
        where password_hash = $2 and id = $3`

	hash := srv.PasswordHash(params.Password)
	tx.Exec(q, newEmail, hash, userID.ID)

	if tx.RowsAffected() < 1 {
		return srv.NewError(&i18n.Message{ID: "invalid_password_or_api_key", Other: "Password or ApiKey is invalid."}), oldEmail
	}

	return nil, oldEmail
}

func sendEmailChanged(srv *utils.MindwellServer, tx *utils.AutoTx, userID *models.UserID, oldEmail, newEmail string) {
	const q = `
		SELECT show_name, telegram
		FROM users
		WHERE id = $1
	`

	var name string
	var tg sql.NullInt64
	tx.Query(q, userID.ID).Scan(&name, &tg)

	if tx.Error() != nil {
		if tx.Error() != sql.ErrNoRows {
			log.Print(tx.Error())
		}
		return
	}

	if len(oldEmail) > 0 {
		srv.Mail.SendEmailChanged(oldEmail, name)
	}

	code := srv.VerificationCode(newEmail)
	srv.Mail.SendGreeting(newEmail, name, code)

	if tg.Valid {
		srv.Tg.SendEmailChanged(tg.Int64)
	}
}

func newEmailUpdater(srv *utils.MindwellServer) func(account.PostAccountEmailParams, *models.UserID) middleware.Responder {
	return func(params account.PostAccountEmailParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			err, oldEmail := setEmail(srv, tx, params, userID)

			if tx.Error() != nil {
				err = srv.NewError(nil)
				return account.NewPostAccountEmailForbidden().WithPayload(err)
			}

			if err != nil {
				return account.NewPostAccountEmailForbidden().WithPayload(err)
			}

			sendEmailChanged(srv, tx, userID, oldEmail, params.Email)

			return account.NewPostAccountEmailOK()
		})
	}
}

func loadInvites(tx *utils.AutoTx, userID *models.UserID) []string {
	const q = `
		SELECT one.word || ' ' || two.word || ' ' || three.word
		FROM mindwell.invites,
			mindwell.invite_words AS one,
			mindwell.invite_words AS two,
			mindwell.invite_words AS three
		WHERE invites.referrer_id = $1
			AND invites.word1 = one.id 
			AND invites.word2 = two.id 
			AND invites.word3 = three.id
		ORDER BY created_at ASC`

	tx.Query(q, userID.ID)

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

func newInvitesLoader(srv *utils.MindwellServer) func(account.GetAccountInvitesParams, *models.UserID) middleware.Responder {
	return func(params account.GetAccountInvitesParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			invites := loadInvites(tx, userID)
			if tx.Error() != nil && tx.Error() != sql.ErrNoRows {
				err := srv.NewError(nil)
				return account.NewPostAccountLoginBadRequest().WithPayload(err)
			}

			res := account.GetAccountInvitesOKBody{Invites: invites}
			return account.NewGetAccountInvitesOK().WithPayload(&res)
		})
	}
}

func newVerificationSender(srv *utils.MindwellServer) func(account.PostAccountVerificationParams, *models.UserID) middleware.Responder {
	return func(params account.PostAccountVerificationParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			const q = "SELECT verified, email, show_name from users where id = $1"

			var verified bool
			var email string
			var name string
			tx.Query(q, userID.ID).Scan(&verified, &email, &name)
			if tx.Error() != nil {
				err := srv.NewError(nil)
				return account.NewPostAccountLoginBadRequest().WithPayload(err)
			}

			if verified {
				err := srv.NewError(&i18n.Message{ID: "already_verified", Other: "Your email had been verified earlier."})
				return account.NewPostAccountVerificationForbidden().WithPayload(err)
			}

			code := srv.VerificationCode(email)
			srv.Mail.SendGreeting(email, name, code)

			return account.NewPostAccountVerificationOK()
		})
	}
}

func newEmailVerifier(srv *utils.MindwellServer) func(account.GetAccountVerificationEmailParams) middleware.Responder {
	return func(params account.GetAccountVerificationEmailParams) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			code := srv.VerificationCode(params.Email)
			if params.Code != code {
				return account.NewGetAccountVerificationEmailBadRequest()
			}

			const q = "UPDATE users SET verified = true WHERE email = $1"
			tx.Exec(q, params.Email)

			return account.NewGetAccountVerificationEmailOK()
		})
	}
}

func newResetPasswordSender(srv *utils.MindwellServer) func(account.PostAccountRecoverParams) middleware.Responder {
	return func(params account.PostAccountRecoverParams) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			const q = `
				SELECT show_name, gender.type 
				FROM users, gender 
				WHERE lower(users.email) = lower($1) and verified and users.gender = gender.id`

			var gender string
			var name string
			tx.Query(q, params.Email).Scan(&name, &gender)
			if tx.Error() == sql.ErrNoRows {
				err := srv.NewError(&i18n.Message{ID: "no_email", Other: "User with this email not found or not verified."})
				return account.NewPostAccountRecoverBadRequest().WithPayload(err)
			}

			if tx.Error() != nil {
				err := srv.NewError(nil)
				return account.NewPostAccountRecoverBadRequest().WithPayload(err)
			}

			code, date := srv.ResetPasswordCode(params.Email)
			srv.Mail.SendResetPassword(params.Email, name, gender, code, date)

			return account.NewPostAccountRecoverOK()
		})
	}
}

func resetPassword(srv *utils.MindwellServer, tx *utils.AutoTx, email, password string) bool {
	const q = `
        update users
        set password_hash = $2
        where lower(email) = lower($1)`

	hash := srv.PasswordHash(password)
	tx.Exec(q, email, hash)

	return tx.RowsAffected() == 1
}

func newPasswordResetter(srv *utils.MindwellServer) func(account.PostAccountRecoverPasswordParams) middleware.Responder {
	return func(params account.PostAccountRecoverPasswordParams) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			if !srv.CheckResetPasswordCode(params.Email, params.Code, params.Date) {
				err := srv.NewError(&i18n.Message{ID: "invalid_code", Other: "Invalid reset link."})
				return account.NewPostAccountRecoverPasswordBadRequest().WithPayload(err)
			}

			if !resetPassword(srv, tx, params.Email, params.Password) {
				err := srv.NewError(nil)
				return account.NewPostAccountRecoverPasswordBadRequest().WithPayload(err)
			}

			return account.NewPostAccountRecoverPasswordOK()
		})
	}
}

func newEmailSettingsLoader(srv *utils.MindwellServer) func(account.GetAccountSettingsEmailParams, *models.UserID) middleware.Responder {
	return func(params account.GetAccountSettingsEmailParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			settings := account.GetAccountSettingsEmailOKBody{}

			const q = "SELECT email_comments, email_followers from users where id = $1"
			tx.Query(q, userID.ID).Scan(&settings.Comments, &settings.Followers)

			return account.NewGetAccountSettingsEmailOK().WithPayload(&settings)
		})
	}
}

func newEmailSettingsEditor(srv *utils.MindwellServer) func(account.PutAccountSettingsEmailParams, *models.UserID) middleware.Responder {
	return func(params account.PutAccountSettingsEmailParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			const q = "UPDATE users SET email_comments = $2, email_followers = $3 where id = $1"
			tx.Exec(q, userID.ID, *params.Comments, *params.Followers)

			return account.NewPutAccountSettingsEmailOK()
		})
	}
}

func generateToken(claims jwt.MapClaims) string {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := tok.SignedString(centSecret)
	if err != nil {
		log.Println(err)
	}

	return str
}

func connectionToken(userID *models.UserID) string {
	return generateToken(jwt.MapClaims{
		"sub":  userID.Name,
		"info": fmt.Sprintf(`{"id":%d,"name":"%s"}`, userID.ID, userID.Name),
	})
}

func privateChannelToken(userID *models.UserID, channel string) string {
	return generateToken(jwt.MapClaims{
		"client":  userID.Name,
		"channel": channel,
	})
}

func newConnectionTokenGenerator(srv *utils.MindwellServer) func(account.GetAccountSubscribeTokenParams, *models.UserID) middleware.Responder {
	return func(params account.GetAccountSubscribeTokenParams, userID *models.UserID) middleware.Responder {
		tok := connectionToken(userID)
		res := account.GetAccountSubscribeTokenOKBody{Token: tok}
		return account.NewGetAccountSubscribeTokenOK().WithPayload(&res)
	}
}

func newTelegramTokenGenerator(srv *utils.MindwellServer) func(account.GetAccountSubscribeTelegramParams, *models.UserID) middleware.Responder {
	return func(params account.GetAccountSubscribeTelegramParams, userID *models.UserID) middleware.Responder {
		tok := srv.Tg.BuildToken(userID.ID)
		res := account.GetAccountSubscribeTelegramOKBody{Token: tok}
		return account.NewGetAccountSubscribeTelegramOK().WithPayload(&res)
	}
}

func newTelegramDeleter(srv *utils.MindwellServer) func(account.DeleteAccountSubscribeTelegramParams, *models.UserID) middleware.Responder {
	return func(params account.DeleteAccountSubscribeTelegramParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			const q = "UPDATE users SET telegram = NULL WHERE id = $1"
			tx.Exec(q, userID.ID)

			return account.NewDeleteAccountSubscribeTelegramNoContent()
		})
	}
}
