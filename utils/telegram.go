package utils

import (
	"database/sql"
	"github.com/leporo/sqlf"
	"go.uber.org/zap"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	cache "github.com/patrickmn/go-cache"
	"github.com/sevings/mindwell-server/models"
	"golang.org/x/net/proxy"
)

const errorText = "Что-то пошло не так…"
const unrecognisedText = "Неизвестная команда. Попробуй /help."

var tgHtmlEsc = strings.NewReplacer(
	"<", "&lt;",
	">", "&gt;",
	"&", "&amp;",
	"\"", "&quot;",
	"'", "&quot;",
	"\r", "",
)

var idRe = regexp.MustCompile(`\d+$`)
var loginRe = regexp.MustCompile(`[0-9\-_]*[a-zA-Z][a-zA-Z0-9\-_]*$`)

type TelegramBot struct {
	srv    *MindwellServer
	api    *tgbotapi.BotAPI
	url    string
	log    *zap.Logger
	admins []int64
	logins *cache.Cache
	cmts   *cache.Cache
	msgs   *cache.Cache
	send   chan func()
	stop   chan interface{}
}

type messageID struct {
	chat int64
	msg  int
}

type messageIDs []messageID

func NewTelegramBot(srv *MindwellServer) *TelegramBot {
	bot := &TelegramBot{
		srv:    srv,
		url:    srv.ConfigString("server.base_url"),
		log:    srv.LogTelegram(),
		admins: srv.ConfigInt64s("telegram.admins"),
		logins: cache.New(10*time.Minute, 10*time.Minute),
		cmts:   cache.New(12*time.Hour, 1*time.Hour),
		msgs:   cache.New(12*time.Hour, 1*time.Hour),
		send:   make(chan func(), 200),
		stop:   make(chan interface{}),
	}

	go bot.run()

	return bot
}

func (bot *TelegramBot) sendMessageNow(chat int64, text string) tgbotapi.Message {
	msg := tgbotapi.NewMessage(chat, text)
	msg.DisableWebPagePreview = true
	msg.ParseMode = "HTML"
	message, err := bot.api.Send(msg)
	if err != nil {
		bot.log.Error(err.Error())
	}

	return message
}

func (bot *TelegramBot) sendMessage(chat int64, text string) {
	bot.send <- func() { bot.sendMessageNow(chat, text) }
}

func (bot *TelegramBot) isAdmin(upd *tgbotapi.Update) bool {
	chat := upd.Message.Chat.ID

	for _, admin := range bot.admins {
		if admin == chat {
			return true
		}
	}

	return false
}

func (bot *TelegramBot) connectToProxy() *http.Client {
	auth := proxy.Auth{
		User:     bot.srv.ConfigString("proxy.user"),
		Password: bot.srv.ConfigString("proxy.password"),
	}

	if len(auth.User) == 0 {
		return http.DefaultClient
	}

	url := bot.srv.ConfigString("proxy.url")
	dialer, err := proxy.SOCKS5("tcp", url, &auth, proxy.Direct)
	if err != nil {
		bot.log.Error(err.Error())
		return nil
	}

	tr := &http.Transport{Dial: dialer.Dial}
	return &http.Client{
		Transport: tr,
	}
}

func (bot *TelegramBot) run() {
	token := bot.srv.ConfigString("telegram.token")
	if len(token) == 0 {
		return
	}

	client := bot.connectToProxy()
	if client == nil {
		return
	}

	api, err := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, client)
	if err != nil {
		bot.log.Error(err.Error())
		return
	}

	bot.api = api
	// api.Debug = true

	bot.log.Sugar().Infof("Running Telegram bot %s", api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.api.GetUpdatesChan(u)
	if err != nil {
		bot.log.Error(err.Error())
	}

	for {
		select {
		case send := <-bot.send:
			send()
		case <-bot.stop:
			return
		case upd := <-updates:
			if upd.Message == nil || !upd.Message.IsCommand() {
				continue
			}

			cmd := upd.Message.Command()
			bot.log.Info("update",
				zap.String("cmd", cmd),
				zap.String("from", upd.Message.From.UserName),
			)

			var reply string
			switch cmd {
			case "start":
				reply = bot.start(&upd)
			case "login":
				reply = bot.login(&upd)
			case "logout":
				reply = bot.logout(&upd)
			case "help":
				reply = bot.help(&upd)
			case "hide":
				reply = bot.hide(&upd)
			case "ban":
				reply = bot.ban(&upd)
			case "unban":
				reply = bot.unban(&upd)
			case "info":
				reply = bot.info(&upd)
			case "stat":
				reply = bot.stat(&upd)
			default:
				reply = unrecognisedText
			}

			bot.sendMessageNow(upd.Message.Chat.ID, reply)
		}
	}
}

func (bot *TelegramBot) Stop() {
	if bot.api == nil {
		return
	}

	bot.api.StopReceivingUpdates()
	close(bot.stop)
}

func (bot *TelegramBot) BuildToken(userID int64) string {
	token := GenerateString(8)
	bot.logins.SetDefault(token, userID)
	return token
}

func (bot *TelegramBot) VerifyToken(token string) int64 {
	userID, found := bot.logins.Get(token)
	if !found {
		return 0
	}

	bot.logins.Delete(token)
	return userID.(int64)
}

func (bot *TelegramBot) start(upd *tgbotapi.Update) string {
	token := upd.Message.CommandArguments()
	if len(token) == 0 {
		return `Привет! Я могу отправлять тебе уведомления из Mindwell.
Чтобы начать, скопируй ключ со <a href="` + bot.url + `account/notifications">страницы настроек</a>.
Отправь его мне, используя команду <code>/login &lt;ключ&gt;</code>. Так ты подтвердишь свой аккаунт.
Чтобы я забыл твой номер в Телеграме, введи /logout.`
	}

	return bot.login(upd)
}

func (bot *TelegramBot) login(upd *tgbotapi.Update) string {
	if upd.Message.From == nil {
		return errorText
	}

	token := upd.Message.CommandArguments()
	userID := bot.VerifyToken(token)

	if userID == 0 {
		return `Скопируй верный ключ со <a href="` + bot.url + `account/notifications">своей страницы настроек</a>.`
	}

	const q = `
		UPDATE users
		SET telegram = $2
		WHERE id = $1
		RETURNING show_name
	`

	var name string
	err := bot.srv.DB.QueryRow(q, userID, upd.Message.Chat.ID).Scan(&name)
	if err != nil {
		bot.log.Error(err.Error())
		return errorText
	}

	return "Привет, " + name + "! Теперь я буду отправлять тебе уведомления из Mindwell. " +
		"Используй команду /logout, если захочешь прекратить."
}

func (bot *TelegramBot) logout(upd *tgbotapi.Update) string {
	if upd.Message.From == nil {
		return errorText
	}

	from := upd.Message.From.ID

	const q = `
		UPDATE users
		SET telegram = NULL
		WHERE telegram = $1
		RETURNING show_name
	`

	var name string
	err := bot.srv.DB.QueryRow(q, from).Scan(&name)
	if err == nil {
		return "Я больше не буду беспокоить тебя, " + name + "."
	} else if err == sql.ErrNoRows {
		return "К этому номеру не привязан аккаунт в Mindwell."
	} else {
		bot.log.Error(err.Error())
		return errorText
	}
}

func (bot *TelegramBot) help(upd *tgbotapi.Update) string {
	text := `Я бот для получения уведомлений из Mindwell. Доступные команды:
<code>/login &lt;ключ&gt;</code> — авторизоваться с использованием автоматически сгенерированного пароля. Его можно получить на <a href="` + bot.url + `account/notifications">странице настроек</a>.
/logout — не получать больше уведомления на этот аккаунт.
/help — вывести данную краткую справку.`

	if bot.isAdmin(upd) {
		text += `

Команды администрирования:
<code>/hide {id или ссылка}</code> — скрыть запись.
<code>/ban {live | vote | comment | invite | adm} {N} {login или ссылка}</code> — запретить пользователю выбранные действия на N дней, в случае adm — навсегда.
<code>/ban user {login или ссылка}</code> — заблокировать пользователя.
<code>/unban {login или ссылка}</code> — разблокировать пользователя.
<code>/info {email, login или ссылка}</code> — информация о пользователе.
<code>/delete {email}</code> — удалить пользователя.
/stat — статистика сервера.
`
	}

	return text
}

func (bot *TelegramBot) hide(upd *tgbotapi.Update) string {
	if !bot.isAdmin(upd) {
		return unrecognisedText
	}

	if upd.Message.From == nil {
		return errorText
	}

	url := upd.Message.CommandArguments()
	strID := idRe.FindString(url)
	if strID == "" {
		return "Укажи ID записи."
	}

	id, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		return err.Error()
	}

	atx := NewAutoTx(bot.srv.DB)
	defer atx.Finish()

	const q = "UPDATE entries SET visible_for = (SELECT id FROM entry_privacy WHERE type = 'me') WHERE id = $1"
	atx.Exec(q, id)
	if atx.RowsAffected() < 1 {
		return "Запись " + strID + " не найдена."
	}

	return "Запись " + strID + " скрыта."
}

func (bot *TelegramBot) ban(upd *tgbotapi.Update) string {
	if !bot.isAdmin(upd) {
		return unrecognisedText
	}

	if upd.Message.From == nil {
		return errorText
	}

	args := strings.Split(upd.Message.CommandArguments(), " ")
	if len(args) < 2 {
		return "Укажи необходимые аргументы."
	}

	url := args[len(args)-1]
	login := loginRe.FindString(url)
	if login == "" {
		return "Укажи логин пользователя."
	}

	if args[0] == "user" {
		return bot.banUser(login)
	} else {
		return bot.restrictUser(args[:len(args)-1], login)
	}
}

func (bot *TelegramBot) restrictUser(args []string, login string) string {
	dayCount := args[len(args)-1]
	banUntil := "CURRENT_DATE + interval '" + dayCount + " days'"
	banTypes := args[:len(args)-1]
	if len(banTypes) == 0 {
		return "Укажи необходимые ограничения."
	}

	q := sqlf.Update("users").
		Where("lower(name) = lower(?)", login)
	for _, ban := range banTypes {
		switch ban {
		case "live":
			q.SetExpr("live_ban", banUntil)
		case "vote":
			q.SetExpr("vote_ban", banUntil)
		case "comment":
			q.SetExpr("comment_ban", banUntil)
		case "invite":
			q.SetExpr("invite_ban", banUntil)
		case "adm":
			q.Set("adm_ban", true)
		default:
			q.Close()
			return "Неизвестный аргумент: " + ban + "."
		}
	}

	atx := NewAutoTx(bot.srv.DB)
	defer atx.Finish()

	atx.ExecStmt(q)
	if atx.RowsAffected() < 1 {
		return "Пользователь " + login + " не найден."
	}

	return "Пользователь " + login + " ограничен в правах на " + dayCount + " дней."
}

func (bot *TelegramBot) banUser(login string) string {
	atx := NewAutoTx(bot.srv.DB)
	defer atx.Finish()

	const q = "SELECT ban_user($1)"
	invitedBy := atx.QueryString(q, login)

	if invitedBy == "" {
		return "Пользователь не найден."
	}

	const emailQ = "SELECT email FROM users WHERE lower(name) = lower($1)"
	email := atx.QueryString(emailQ, login)

	link := bot.url + "users/" + invitedBy
	return "Пользователь " + login +
		` заблокирован. Приглашен <a href="` + link + `">` + invitedBy + `</a>. Почта ` + email + "."
}

func (bot *TelegramBot) unban(upd *tgbotapi.Update) string {
	if !bot.isAdmin(upd) {
		return unrecognisedText
	}

	if upd.Message.From == nil {
		return errorText
	}

	url := upd.Message.CommandArguments()
	login := loginRe.FindString(url)
	if login == "" {
		return "Укажи логин пользователя."
	}

	atx := NewAutoTx(bot.srv.DB)
	defer atx.Finish()

	const q = "UPDATE users SET verified = true WHERE lower(name) = lower($1) RETURNING email"
	email := atx.QueryString(q, login)
	if email == "" {
		return "Пользователь " + login + " не найден."
	}

	return "Пользователь " + login +
		" разблокирован. Теперь можно запросить сброс пароля на почту " + email + "."
}

func (bot *TelegramBot) info(upd *tgbotapi.Update) string {
	if !bot.isAdmin(upd) {
		return unrecognisedText
	}

	if upd.Message.From == nil {
		return errorText
	}

	arg := upd.Message.CommandArguments()
	if strings.Contains(arg, "/") {
		arg = loginRe.FindString(arg)
	}
	if arg == "" {
		return "Укажи логин или адрес почты."
	}

	atx := NewAutoTx(bot.srv.DB)
	defer atx.Finish()

	const q = `
SELECT users.id, users.name, users.show_name, created_at, 
	email, verified, valid_thru, rank, karma,
	invited.name, invited.show_name,
	entries_count, followers_count, followings_count, comments_count, invited_count,
	invite_ban, vote_ban, comment_ban, live_ban, adm_ban
FROM users
JOIN (SELECT id, name, show_name FROM users) AS invited ON users.invited_by = invited.id
WHERE users.email = lower($1) OR lower(users.name) = lower($1)`

	atx.Query(q, arg)

	var id int64
	var name, showName, email, invitedByName, invitedByShowName string
	var verified bool
	var createdAt, validThru time.Time
	var rank int64
	var karma float64
	var entries, followers, followings, comments, invited int64
	var inviteBan, voteBan, commentBan, liveBan time.Time
	var admBan bool
	atx.Scan(&id, &name, &showName, &createdAt,
		&email, &verified, &validThru, &rank, &karma,
		&invitedByName, &invitedByShowName,
		&entries, &followers, &followings, &comments, &invited,
		&inviteBan, &voteBan, &commentBan, &liveBan, &admBan)

	if id == 0 {
		return "Пользователь с логином или адресом почты " + arg + " не найден."
	}

	today := time.Now()

	var text string
	text += "\n<b>id</b>: " + strconv.FormatInt(id, 10)
	text += "\n<b>url</b>: " + `<a href="` + bot.url + "users/" + name + `">` + showName + `</a>`
	text += "\n<b>email</b>: " + email
	text += "\n<b>verified</b>: " + strconv.FormatBool(verified)
	text += "\n<b>created at</b>: " + createdAt.Format("15:04:05 02 Jan 2006 MST")
	text += "\n<b>valid thru</b>: " + validThru.Format("15:04:05 02 Jan 2006 MST")
	text += "\n<b>rank</b>: " + strconv.FormatInt(rank, 10)
	text += "\n<b>karma</b>: " + strconv.FormatFloat(karma, 'f', 2, 64)
	text += "\n<b>invited by</b>: " + `<a href="` + bot.url + "users/" + invitedByName + `">` + invitedByShowName + `</a>`
	text += "\n<b>entries</b>: " + strconv.FormatInt(entries, 10)
	text += "\n<b>followers</b>: " + strconv.FormatInt(followers, 10)
	text += "\n<b>followings</b>: " + strconv.FormatInt(followings, 10)
	text += "\n<b>comments</b>: " + strconv.FormatInt(comments, 10)
	text += "\n<b>invited</b>: " + strconv.FormatInt(invited, 10)
	text += "\n<b>invite ban</b>: " + strconv.FormatBool(inviteBan.After(today)) + ", " + inviteBan.Format("02 Jan 2006")
	text += "\n<b>vote ban</b>: " + strconv.FormatBool(voteBan.After(today)) + ", " + voteBan.Format("02 Jan 2006")
	text += "\n<b>comment ban</b>: " + strconv.FormatBool(commentBan.After(today)) + ", " + commentBan.Format("02 Jan 2006")
	text += "\n<b>live ban</b>: " + strconv.FormatBool(liveBan.After(today)) + ", " + liveBan.Format("02 Jan 2006")
	text += "\n<b>adm ban</b>: " + strconv.FormatBool(admBan)

	return text
}

func (bot *TelegramBot) stat(upd *tgbotapi.Update) string {
	if !bot.isAdmin(upd) {
		return unrecognisedText
	}

	if upd.Message.From == nil {
		return errorText
	}

	atx := NewAutoTx(bot.srv.DB)
	defer atx.Finish()

	var text string

	addInt64 := func(key string, value int64) {
		text += "\n<b>" + key + "</b>: " + strconv.FormatInt(value, 10)
	}

	addFloat64 := func(key string, value float64) {
		text += "\n<b>" + key + "</b>: " + strconv.FormatFloat(value, 'f', 2, 64)
	}

	const usersQuery = `SELECT count(*) FROM users`
	users := atx.QueryInt64(usersQuery)
	addInt64("users", users)

	const invitedUsersQuery = `SELECT count(*) FROM users WHERE invited_by IS NOT NULL`
	invitedUsers := atx.QueryInt64(invitedUsersQuery)
	addInt64("invited users", invitedUsers)

	const negKarmaUsersQuery = `SELECT count(*) FROM users WHERE karma < -1`
	negKarmaUsers := atx.QueryInt64(negKarmaUsersQuery)
	addInt64("users with karma &lt; -1", negKarmaUsers)

	const posKarmaUsersQuery = `SELECT count(*) FROM users WHERE karma > 1`
	posKarmaUsers := atx.QueryInt64(posKarmaUsersQuery)
	addInt64("users with karma &gt; 1", posKarmaUsers)

	const genderUsersQuery = `
SELECT gender.type AS sex, count(*)
FROM users
JOIN gender ON users.gender = gender.id
GROUP BY sex
ORDER BY sex`
	atx.Query(genderUsersQuery)
	for {
		var gender string
		var count int64
		if !atx.Scan(&gender, &count) {
			break
		}

		addInt64(gender+" gender users", count)
	}

	const newUsersMonthQuery = `SELECT count(*) FROM users WHERE now() - created_at < interval '1 month'`
	newUsersMonth := atx.QueryInt64(newUsersMonthQuery)
	addInt64("last month new users", newUsersMonth)

	const onlineUsersNowQuery = `SELECT count(*) FROM users WHERE is_online(last_seen_at)`
	onlineUsersNow := atx.QueryInt64(onlineUsersNowQuery)
	addInt64("online users", onlineUsersNow)

	const onlineUsersWeekQuery = `SELECT count(*) FROM users WHERE now() - last_seen_at < interval '7 days'`
	onlineUsersWeek := atx.QueryInt64(onlineUsersWeekQuery)
	addInt64("last week online users", onlineUsersWeek)

	const onlineUsersMonthQuery = `SELECT count(*) FROM users WHERE now() - last_seen_at < interval '1 month'`
	onlineUsersMonth := atx.QueryInt64(onlineUsersMonthQuery)
	addInt64("last month online users", onlineUsersMonth)

	const postingUsersMonthQuery = `
SELECT count(distinct author_id)
FROM entries
WHERE now() - created_at < interval '1 month'`
	postingUsersMonth := atx.QueryInt64(postingUsersMonthQuery)
	addInt64("last month posting users", postingUsersMonth)

	const chatsQuery = `
SELECT count(*)
FROM chats
JOIN messages ON last_message = messages.id
WHERE last_message > 0 AND messages.author_id <> 1`
	chats := atx.QueryInt64(chatsQuery)
	addInt64("user chats", chats)

	const avgEntriesQuery = `
SELECT count(*) / 7.0
FROM entries
WHERE created_at::date < current_date
	AND created_at::date >= current_date - interval '7 days'`
	avgEntries := atx.QueryFloat64(avgEntriesQuery)
	addFloat64("avg entries", avgEntries)

	const avgCommentsQuery = `
SELECT count(*) / 7.0
FROM comments
WHERE created_at::date < current_date
	AND created_at::date >= current_date - interval '7 days'`
	avgComments := atx.QueryFloat64(avgCommentsQuery)
	addFloat64("avg comments", avgComments)

	const avgMessagesQuery = `
SELECT count(*) / 7.0
FROM messages
WHERE created_at::date < current_date
	AND created_at::date >= current_date - interval '7 days'
	AND author_id <> 1`
	avgMessages := atx.QueryFloat64(avgMessagesQuery)
	addFloat64("avg user messages", avgMessages)

	if atx.Error() != nil {
		return errorText
	}

	return text
}

func idToString(id int64) string {
	return strconv.FormatInt(id, 32)
}

func (bot *TelegramBot) comment(entryTitle string, cmt *models.Comment) (cmtID, text string) {
	cmtID = idToString(cmt.ID)

	link := bot.url + "entries/" + strconv.FormatInt(cmt.EntryID, 10) + "#comments"

	text = tgHtmlEsc.Replace(cmt.Author.ShowName) + " пишет: \n" +
		"«" + tgHtmlEsc.Replace(cmt.EditContent) + "»\n"

	if len(entryTitle) == 0 {
		text += `К <a href="` + link + `">записи</a>`
	} else {
		text += `<a href="` + link + `">` + entryTitle + `</a>`
	}

	return
}

func (bot *TelegramBot) SendNewComment(chat int64, entryTitle string, cmt *models.Comment) {
	if bot.api == nil {
		return
	}

	cmtID, text := bot.comment(entryTitle, cmt)

	bot.send <- func() {
		msg := bot.sendMessageNow(chat, text)
		if msg.MessageID <= 0 {
			return
		}

		var msgIDs messageIDs
		msgIDsVar, found := bot.cmts.Get(cmtID)
		if found {
			msgIDs = msgIDsVar.(messageIDs)
		}
		msgIDs = append(msgIDs, messageID{chat, msg.MessageID})

		bot.cmts.SetDefault(cmtID, msgIDs)
	}
}

func (bot *TelegramBot) SendUpdateComment(entryTitle string, cmt *models.Comment) {
	if bot.api == nil {
		return
	}

	cmtID, text := bot.comment(entryTitle, cmt)
	msgIDsVar, found := bot.cmts.Get(cmtID)
	if !found {
		return
	}

	msgIDs := msgIDsVar.(messageIDs)

	bot.send <- func() {
		for _, msgID := range msgIDs {
			msg := tgbotapi.NewEditMessageText(msgID.chat, msgID.msg, text)
			msg.DisableWebPagePreview = true
			msg.ParseMode = "HTML"
			_, err := bot.api.Send(msg)
			if err != nil {
				bot.log.Error(err.Error())
			}
		}
	}
}

func (bot *TelegramBot) SendRemoveComment(commentID int64) {
	if bot.api == nil {
		return
	}

	cmtID := idToString(commentID)
	msgIDsVar, found := bot.cmts.Get(cmtID)
	if !found {
		return
	}

	msgIDs := msgIDsVar.(messageIDs)

	bot.send <- func() {
		for _, msgID := range msgIDs {
			msg := tgbotapi.NewDeleteMessage(msgID.chat, msgID.msg)
			_, err := bot.api.DeleteMessage(msg)
			if err != nil {
				bot.log.Error(err.Error())
			}
		}
	}
}

func (bot *TelegramBot) SendPasswordChanged(chat int64) {
	if bot.api == nil {
		return
	}

	const text = "Пароль к твоему тлогу был изменен."
	bot.sendMessage(chat, text)
}

func (bot *TelegramBot) SendEmailChanged(chat int64) {
	if bot.api == nil {
		return
	}

	const text = "Твой адрес почты был изменен."
	bot.sendMessage(chat, text)
}

func (bot *TelegramBot) SendNewFollower(chat int64, fromName, fromShowName, fromGender string, toPrivate bool) {
	if bot.api == nil {
		return
	}

	var ending string
	if fromGender == "female" {
		ending = "ась"
	} else {
		ending = "ся"
	}

	link := `<a href="` + bot.url + `users/` + fromName + `">` + fromShowName + `</a>`

	var text string
	if toPrivate {
		text = link + " запрашивает доступ к чтению твоего тлога."
	} else {
		text = link + " подписал" + ending + " на твой тлог."
	}

	bot.sendMessage(chat, text)
}

func (bot *TelegramBot) SendNewAccept(chat int64, fromName, fromShowName, fromGender string) {
	if bot.api == nil {
		return
	}

	var ending string
	if fromGender == "female" {
		ending = "а"
	} else {
		ending = ""
	}

	link := `<a href="` + bot.url + `users/` + fromName + `">` + fromShowName + `</a>`
	text := link + " разрешил" + ending + " тебе читать свой тлог."

	bot.sendMessage(chat, text)
}

func (bot *TelegramBot) SendNewInvite(chat int64) {
	if bot.api == nil {
		return
	}

	text := `У тебя появилось новое приглашение! <a href="` + bot.url + `users?top=waiting">Используй</a> его с умом.`
	bot.sendMessage(chat, text)
}

func (bot *TelegramBot) SendInvited(chat int64, fromName, fromShowName, fromGender string) {
	if bot.api == nil {
		return
	}

	var ending string
	if fromGender == "female" {
		ending = "а"
	} else {
		ending = ""
	}

	link := `<a href="` + bot.url + `users/` + fromName + `">` + fromShowName + `</a>`
	text := link + " отправил" + ending + " тебе приглашение на Mindwell. " +
		"Теперь тебе доступны все функции сайта (при отсутствии других ограничений)."

	bot.sendMessage(chat, text)
}

func (bot *TelegramBot) SendAdmSent(chat int64) {
	if bot.api == nil {
		return
	}

	text := `Твой Дед Мороз отправил тебе подарок! Когда получишь, не забудь <a href="` + bot.url +
		`adm">поставить нужный флажок</a>. И не открывай до Нового года.`
	bot.sendMessage(chat, text)
}

func (bot *TelegramBot) SendAdmReceived(chat int64) {
	if bot.api == nil {
		return
	}

	text := `Внук получил твой новогодний подарок.`
	bot.sendMessage(chat, text)
}

func (bot *TelegramBot) message(msg *models.Message) (msgID, text string) {
	msgID = idToString(msg.ID)

	link := bot.url + "chats/" + msg.Author.Name

	text = tgHtmlEsc.Replace(msg.Author.ShowName) + " пишет: \n" +
		"«" + tgHtmlEsc.Replace(msg.EditContent) + "»\n" +
		`В <a href="` + link + `">беседе</a>`

	return
}

func (bot *TelegramBot) SendNewMessage(chat int64, msg *models.Message) {
	if bot.api == nil {
		return
	}

	msgID, text := bot.message(msg)

	bot.send <- func() {
		msg := bot.sendMessageNow(chat, text)
		if msg.MessageID <= 0 {
			return
		}

		var msgIDs messageIDs
		msgIDsVar, found := bot.msgs.Get(msgID)
		if found {
			msgIDs = msgIDsVar.(messageIDs)
		}
		msgIDs = append(msgIDs, messageID{chat, msg.MessageID})

		bot.msgs.SetDefault(msgID, msgIDs)
	}
}

func (bot *TelegramBot) SendUpdateMessage(msg *models.Message) {
	if bot.api == nil {
		return
	}

	msgID, text := bot.message(msg)
	msgIDsVar, found := bot.msgs.Get(msgID)
	if !found {
		return
	}

	msgIDs := msgIDsVar.(messageIDs)

	bot.send <- func() {
		for _, msgID := range msgIDs {
			msg := tgbotapi.NewEditMessageText(msgID.chat, msgID.msg, text)
			msg.DisableWebPagePreview = true
			msg.ParseMode = "HTML"
			_, err := bot.api.Send(msg)
			if err != nil {
				bot.log.Error(err.Error())
			}
		}
	}
}

func (bot *TelegramBot) SendRemoveMessage(messageID int64) {
	if bot.api == nil {
		return
	}

	cmtID := idToString(messageID)
	msgIDsVar, found := bot.msgs.Get(cmtID)
	if !found {
		return
	}

	msgIDs := msgIDsVar.(messageIDs)

	bot.send <- func() {
		for _, msgID := range msgIDs {
			msg := tgbotapi.NewDeleteMessage(msgID.chat, msgID.msg)
			_, err := bot.api.DeleteMessage(msg)
			if err != nil {
				bot.log.Error(err.Error())
			}
		}
	}
}
