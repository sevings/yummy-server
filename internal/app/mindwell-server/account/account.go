package account

import (
	"database/sql"
	"fmt"
	"github.com/sevings/mindwell-server/internal/app/mindwell-server/chats"
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

	srv.API.AccountGetAccountSettingsTelegramHandler = account.GetAccountSettingsTelegramHandlerFunc(newTelegramSettingsLoader(srv))
	srv.API.AccountPutAccountSettingsTelegramHandler = account.PutAccountSettingsTelegramHandlerFunc(newTelegramSettingsEditor(srv))

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
			data := account.GetAccountEmailEmailOKBody{Email: params.Email, IsFree: free}
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
			data := account.GetAccountNameNameOKBody{Name: params.Name, IsFree: free}
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
		srv.LogApi().Error(err.Error())
	}

	w := img.Bounds().Dx()
	if size < w {
		img = imaging.Resize(img, size, size, imaging.CatmullRom)
	}

	err = imaging.Save(img, path+name, imaging.JPEGQuality(85))
	if err != nil {
		srv.LogApi().Error(err.Error())
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

	img, err := govatar.GenerateForUsername(g, name)
	if err != nil {
		srv.LogApi().Error(err.Error())
	}

	folder := name[:1] + "/"
	fileName := utils.GenerateString(5) + ".jpg"

	saveAvatar(srv, img, 124, folder, fileName)
	saveAvatar(srv, img, 92, folder, fileName)
	saveAvatar(srv, img, 42, folder, fileName)

	return folder + fileName
}

func createUser(srv *utils.MindwellServer, tx *utils.AutoTx, params account.PostAccountRegisterParams) int64 {
	hash := srv.PasswordHash(params.Password)
	apiKey := utils.GenerateString(32)

	const q = `
		INSERT INTO users 
		(name, show_name, email, password_hash, api_key,
		gender, 
		country, city, avatar)
		values($1, $1, $2, $3, $4,
			(select id from gender where type = $5), 
			$6, $7, $8)
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
		params.Name, params.Email, hash, apiKey,
		*params.Gender,
		*params.Country, *params.City, avatar).Scan(&user)

	if params.Birthday != nil {
		tx.Exec("UPDATE users SET birthday = $1 WHERE id = $2", *params.Birthday, user)
	}

	return user
}

const authProfileQuery = `
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
users.email, users.verified, users.birthday,
users.api_key, extract(epoch from users.valid_thru),
extract(epoch from users.invite_ban), extract(epoch from users.vote_ban),
extract(epoch from users.comment_ban), extract(epoch from users.live_ban),
invited_by.id, 
invited_by.name, invited_by.show_name,
is_online(invited_by.last_seen_at), 
invited_by.avatar
FROM users 
INNER JOIN gender ON gender.id = users.gender
INNER JOIN user_privacy ON users.privacy = user_privacy.id
INNER JOIN font_family ON users.font_family = font_family.id
INNER JOIN alignment ON users.text_alignment = alignment.id
LEFT JOIN users AS invited_by ON users.invited_by = invited_by.id
`

func loadAuthProfile(srv *utils.MindwellServer, tx *utils.AutoTx, query string, args ...interface{}) *models.AuthProfile {
	var profile models.AuthProfile
	profile.Design = &models.Design{}
	profile.Counts = &models.FriendAO1Counts{}
	profile.Account = &models.AuthProfileAO1Account{}
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
			ID:       invitedByID.Int64,
			Name:     invitedByName.String,
			ShowName: invitedByShowName.String,
			IsOnline: invitedByIsOnline.Bool,
			Avatar:   srv.NewAvatar(invitedByAvatar.String),
		}
	}

	//token, thru := utils.BuildApiToken(apiSecret, profile.ID)
	//profile.Account.APIKey = token
	//profile.Account.ValidThru = float64(thru)

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

	if profile.Ban.Comment <= now {
		profile.Ban.Comment = 0
	}

	if profile.Ban.Live <= now {
		profile.Ban.Live = 0
	}

	profile.Cover = srv.NewCover(profile.ID, cover)

	return &profile
}

const authProfileQueryByID = authProfileQuery + "WHERE users.id = $1"

func newRegistrator(srv *utils.MindwellServer) func(account.PostAccountRegisterParams) middleware.Responder {
	return func(params account.PostAccountRegisterParams) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			params.Name = strings.TrimSpace(params.Name)
			params.Password = strings.TrimSpace(params.Password)
			params.Email = strings.TrimSpace(params.Email)

			if ok := isEmailFree(tx, params.Email); !ok {
				err := srv.NewError(&i18n.Message{ID: "email_is_not_free", Other: "Email is not free."})
				return account.NewPostAccountRegisterBadRequest().WithPayload(err)
			}

			if ok := isNameFree(tx, params.Name); !ok {
				err := srv.NewError(&i18n.Message{ID: "name_is_not_free", Other: "Name is not free."})
				return account.NewPostAccountRegisterBadRequest().WithPayload(err)
			}

			id := createUser(srv, tx, params)
			if tx.Error() != nil {
				err := srv.NewError(nil)
				return account.NewPostAccountRegisterBadRequest().WithPayload(err)
			}

			user := loadAuthProfile(srv, tx, authProfileQueryByID, id)
			if tx.Error() != nil {
				err := srv.NewError(nil)
				return account.NewPostAccountRegisterBadRequest().WithPayload(err)
			}

			srv.Ntf.SendGreeting(user.Account.Email, user.ShowName)
			time.AfterFunc(2*time.Minute, func() { chats.SendWelcomeMessage(srv, user) })

			user.Account.Email = utils.HideEmail(user.Account.Email)

			return account.NewPostAccountRegisterCreated().WithPayload(user)
		})
	}
}

const authProfileQueryByPassword = authProfileQuery + `
	WHERE users.password_hash = $2
		AND (lower(users.name) = lower($1) OR lower(users.email) = lower($1))
`

func newLoginer(srv *utils.MindwellServer) func(account.PostAccountLoginParams) middleware.Responder {
	return func(params account.PostAccountLoginParams) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			params.Name = strings.TrimSpace(params.Name)
			params.Password = strings.TrimSpace(params.Password)

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

func newPasswordUpdater(srv *utils.MindwellServer) func(account.PostAccountPasswordParams, *models.UserID) middleware.Responder {
	return func(params account.PostAccountPasswordParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			params.NewPassword = strings.TrimSpace(params.NewPassword)
			params.OldPassword = strings.TrimSpace(params.OldPassword)

			ok := setPassword(srv, tx, params, userID)
			if tx.Error() != nil {
				err := srv.NewError(nil)
				return account.NewPostAccountPasswordForbidden().WithPayload(err)
			}

			if !ok {
				err := srv.NewError(&i18n.Message{ID: "invalid_password_or_api_key", Other: "Password or ApiKey is invalid."})
				return account.NewPostAccountPasswordForbidden().WithPayload(err)
			}

			srv.Ntf.SendPasswordChanged(tx, userID)

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

func newEmailUpdater(srv *utils.MindwellServer) func(account.PostAccountEmailParams, *models.UserID) middleware.Responder {
	return func(params account.PostAccountEmailParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			params.Email = strings.TrimSpace(params.Email)
			params.Password = strings.TrimSpace(params.Password)

			err, oldEmail := setEmail(srv, tx, params, userID)

			if tx.Error() != nil {
				err = srv.NewError(nil)
				return account.NewPostAccountEmailForbidden().WithPayload(err)
			}

			if err != nil {
				return account.NewPostAccountEmailForbidden().WithPayload(err)
			}

			srv.Ntf.SendEmailChanged(tx, userID, oldEmail, params.Email)

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

			srv.Ntf.SendGreeting(email, name)

			return account.NewPostAccountVerificationOK()
		})
	}
}

func newEmailVerifier(srv *utils.MindwellServer) func(account.GetAccountVerificationEmailParams) middleware.Responder {
	return func(params account.GetAccountVerificationEmailParams) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			params.Email = strings.TrimSpace(params.Email)
			params.Code = strings.TrimSpace(params.Code)

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
			params.Email = strings.TrimSpace(params.Email)

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

			srv.Ntf.SendResetPassword(params.Email, name, gender)

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
			params.Email = strings.TrimSpace(params.Email)
			params.Password = strings.TrimSpace(params.Password)
			params.Code = strings.TrimSpace(params.Code)

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

			const q = "SELECT email_comments, email_followers, email_invites from users where id = $1"
			tx.Query(q, userID.ID).Scan(&settings.Comments, &settings.Followers, &settings.Invites)

			return account.NewGetAccountSettingsEmailOK().WithPayload(&settings)
		})
	}
}

func newEmailSettingsEditor(srv *utils.MindwellServer) func(account.PutAccountSettingsEmailParams, *models.UserID) middleware.Responder {
	return func(params account.PutAccountSettingsEmailParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			const q = "UPDATE users SET email_comments = $2, email_followers = $3, email_invites = $4 where id = $1"
			tx.Exec(q, userID.ID, *params.Comments, *params.Followers, *params.Invites)

			return account.NewPutAccountSettingsEmailOK()
		})
	}
}

func newTelegramSettingsLoader(srv *utils.MindwellServer) func(account.GetAccountSettingsTelegramParams, *models.UserID) middleware.Responder {
	return func(params account.GetAccountSettingsTelegramParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			const q = `
				SELECT telegram_comments, telegram_followers, 
					telegram_invites, telegram_messages 
				FROM users 
				WHERE id = $1`

			settings := account.GetAccountSettingsTelegramOKBody{}
			tx.Query(q, userID.ID).Scan(&settings.Comments, &settings.Followers, &settings.Invites, &settings.Messages)

			return account.NewGetAccountSettingsTelegramOK().WithPayload(&settings)
		})
	}
}

func newTelegramSettingsEditor(srv *utils.MindwellServer) func(account.PutAccountSettingsTelegramParams, *models.UserID) middleware.Responder {
	return func(params account.PutAccountSettingsTelegramParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			const q = `
				UPDATE users 
				SET telegram_comments = $2, telegram_followers = $3, 
					telegram_invites = $4, telegram_messages = $5 
				WHERE id = $1`

			tx.Exec(q, userID.ID, *params.Comments, *params.Followers, *params.Invites, *params.Messages)

			return account.NewPutAccountSettingsTelegramOK()
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
		tok := srv.Ntf.Tg.BuildToken(userID.ID)
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
