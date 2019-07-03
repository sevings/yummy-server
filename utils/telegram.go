package utils

import (
	"database/sql"
	"log"
	"net/http"
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
	logins *cache.Cache
	send   chan *tgbotapi.MessageConfig
	stop   chan interface{}
}

func connectToProxy(srv *MindwellServer) *http.Client {
	auth := proxy.Auth{
		User:     srv.ConfigString("proxy.user"),
		Password: srv.ConfigString("proxy.password"),
	}

	if len(auth.User) == 0 {
		return nil
	}

	url := srv.ConfigString("proxy.url")
	dialer, err := proxy.SOCKS5("tcp", url, &auth, proxy.Direct)
	if err != nil {
		log.Println(err)
		return nil
	}

	tr := &http.Transport{Dial: dialer.Dial}
	return &http.Client{
		Transport: tr,
	}
}

func NewTelegramBot(srv *MindwellServer) *TelegramBot {
	bot := &TelegramBot{
		srv:    srv,
		url:    srv.ConfigString("server.base_url"),
		logins: cache.New(10*time.Minute, 10*time.Minute),
		send:   make(chan *tgbotapi.MessageConfig, 200),
		stop:   make(chan interface{}),
	}

	go bot.run(srv)

	return bot
}

func (bot *TelegramBot) sendMessage(chat int64, text string) {
	msg := tgbotapi.NewMessage(chat, text)
	msg.DisableWebPagePreview = true
	msg.ParseMode = "HTML"
	bot.send <- &msg
}

func (bot *TelegramBot) run(srv *MindwellServer) {
	token := srv.ConfigString("telegram.token")
	if len(token) == 0 {
		return
	}

	proxy := connectToProxy(srv)
	if proxy == nil {
		return
	}

	api, err := tgbotapi.NewBotAPIWithClient(token, proxy)
	if err != nil {
		log.Print(err)
		return
	}

	bot.api = api
	// api.Debug = true

	log.Printf("Running Telegram bot %s", api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.api.GetUpdatesChan(u)
	if err != nil {
		log.Print(err)
	}

	for {
		select {
		case msg := <-bot.send:
			_, err := bot.api.Send(msg)
			if err != nil {
				log.Println("Telegram: ", err)
			}
		case <-bot.stop:
			return
		case upd := <-updates:
			if upd.Message == nil || !upd.Message.IsCommand() {
				continue
			}

			cmd := upd.Message.Command()
			log.Printf("Telegram: [%s] %s", upd.Message.From.UserName, cmd)

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
			default:
				reply = unrecognisedText
			}

			bot.sendMessage(upd.Message.Chat.ID, reply)
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
	bot.logins.Set(token, userID, cache.DefaultExpiration)
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
		log.Print(err)
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
		log.Print(err)
		return errorText
	}
}

func (bot *TelegramBot) help(upd *tgbotapi.Update) string {
	return `Я бот для получения уведомлений из Mindwell. Доступные команды:
<code>/login &lt;ключ&gt;</code> - авторизоваться с использованием автоматически сгенерированного пароля. Его можно получить на <a href="` + bot.url + `account/notifications">странице настроек</a>.
/logout - не получать больше уведомления на этот аккаунт.
/help - вывести данную краткую справку.`
}

func (bot *TelegramBot) SendNewComment(chat int64, entryTitle string, cmt *models.Comment) {
	if bot.api == nil {
		return
	}

	link := bot.url + "entries/" + strconv.FormatInt(cmt.EntryID, 10) + "#comments"

	text := cmt.Author.ShowName + " пишет: \n" +
		"«" + tgHtmlEsc.Replace(cmt.EditContent) + "»\n"

	if len(entryTitle) == 0 {
		text += `К <a href="` + link + `">записи</a>`
	} else {
		text += `<a href="` + link + `">` + entryTitle + `</a>`
	}

	bot.sendMessage(chat, text)
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

	text := `У тебя в настройках <a href="` + bot.url + `account/invites">появилось новое приглашение</a>! Используй его с умом.`
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
