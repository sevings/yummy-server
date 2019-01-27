package utils

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sevings/mindwell-server/models"
	"golang.org/x/net/proxy"
)

const errorText = "Что-то пошло не так…"

type tgMsg struct {
	chat int64
	text string
}

type TelegramBot struct {
	srv    *MindwellServer
	api    *tgbotapi.BotAPI
	secret []byte
	url    string
	send   chan *tgMsg
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
		secret: []byte(srv.ConfigString("telegram.secret")),
		url:    srv.ConfigString("server.base_url"),
		send:   make(chan *tgMsg, 200),
	}

	proxy := connectToProxy(srv)
	if proxy == nil {
		return bot
	}

	token := srv.ConfigString("telegram.token")
	if len(token) == 0 {
		return bot
	}

	api, err := tgbotapi.NewBotAPIWithClient(token, proxy)
	if err != nil {
		log.Print(err)
		return bot
	}

	bot.api = api
	// api.Debug = true

	log.Printf("Running Telegram bot %s", api.Self.UserName)

	go bot.run()

	return bot
}

func (bot *TelegramBot) run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.api.GetUpdatesChan(u)
	if err != nil {
		log.Print(err)
	}

	for {
		select {
		case upd := <-updates:
			if upd.Message == nil || !upd.Message.IsCommand() {
				continue
			}

			var reply string
			switch upd.Message.Command() {
			case "login":
				reply = bot.login(&upd)
			case "logout":
				reply = bot.logout(&upd)
			case "help":
				reply = bot.help(&upd)
			default:
				reply = bot.help(&upd)
			}

			msg := tgbotapi.NewMessage(upd.Message.Chat.ID, reply)
			msg.DisableWebPagePreview = true
			bot.api.Send(msg)
		case ntf := <-bot.send:
			msg := tgbotapi.NewMessage(ntf.chat, ntf.text)
			msg.DisableWebPagePreview = true
			bot.api.Send(msg)
		}
	}
}

func (bot *TelegramBot) Stop() {
	if bot.api == nil {
		return
	}

	bot.api.StopReceivingUpdates()
}

func (bot *TelegramBot) BuildToken(userID int64) string {
	if len(bot.secret) == 0 {
		return ""
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Unix() + 60*60,
		"id":  userID,
	})

	tokenString, err := token.SignedString(bot.secret)
	if err != nil {
		log.Print(err)
	}

	return tokenString
}

func (bot *TelegramBot) VerifyToken(tokenString string) int64 {
	if len(bot.secret) == 0 {
		return 0
	}

	if len(tokenString) == 0 {
		return 0
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return bot.secret, nil
	})

	if err != nil {
		log.Println(err)
		return 0
	}

	if !token.Valid {
		log.Printf("Invalid token: %s\n", tokenString)
		return 0

	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("Error get claims: %s\n", tokenString)
		return 0
	}

	now := time.Now().Unix()
	exp := int64(claims["exp"].(float64))
	if exp < now {
		return 0
	}

	id := claims["id"].(float64)

	return int64(id)
}

func (bot *TelegramBot) login(upd *tgbotapi.Update) string {
	if upd.Message.From == nil {
		return errorText
	}

	token := upd.Message.CommandArguments()
	userID := bot.VerifyToken(token)

	if userID == 0 {
		return "Скопируй верный ключ со своей страницы настроек https://mindwell.win/account/notifications"
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
	return "Привет! Я могу отправлять тебе уведомления из Mindwell.\n" +
		"Чтобы начать, скопируй ключ со страницы настроек https://mindwell.win/account/notifications. " +
		"Отправь его мне, используя команду '/login <ключ>'. Так ты подтвердишь свой аккаунт.\n" +
		"Чтобы я забыл твой номер в телеграме, введи '/logout'."
}

func (bot *TelegramBot) SendNewComment(chat int64, cmt *models.Comment) {
	if bot.api == nil {
		return
	}

	text := cmt.Author.ShowName + ": \n" +
		"«" + cmt.EditContent + "».\n" +
		bot.url + "/entries/" + strconv.FormatInt(cmt.EntryID, 10) + "#comments"

	bot.send <- &tgMsg{chat: chat, text: text}
}
