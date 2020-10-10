package utils

import (
	"database/sql"
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

var tgHtmlEsc *strings.Replacer = strings.NewReplacer(
	"<", "&lt;",
	">", "&gt;",
	"&", "&amp;",
	"\"", "&quot;",
	"'", "&quot;",
	"\r", "",
)

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
<code>/ban user {login или ссылка}</code> — забанить пользователя.
<code>/unban user {login или ссылка}</code> — разбанить пользователя.
<code>/info {email}</code> — информация о пользователе.
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
	idRe := regexp.MustCompile(`\d+$`)
	strID := idRe.FindString(url)
	if strID == "" {
		return "Укажите ID записи."
	}

	id, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		return err.Error()
	}

	const q = "UPDATE entries SET visible_for = (SELECT id FROM entry_privacy WHERE type = 'me') WHERE id = $1"
	res, err := bot.srv.DB.Exec(q, id)
	if err != nil {
		return err.Error()
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err.Error()
	}

	if rows < 1 {
		return "Запись " + strID + " не найдена."
	}

	return "Запись " + strID + " скрыта."
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
