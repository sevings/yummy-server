package utils

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/matcornic/hermes"
	"github.com/sevings/mindwell-server/models"
	"gopkg.in/mailgun/mailgun-go.v1"
)

type Postman struct {
	url  string
	mg   mailgun.Mailgun
	h    hermes.Hermes
	sprt string
	ch   chan *mailgun.Message
	stop chan interface{}
}

func NewPostman(domain, apiKey, pubKey, baseURL, support string) *Postman {
	pm := &Postman{
		url: baseURL,
		mg:  mailgun.NewMailgun(domain, apiKey, pubKey),
		h: hermes.Hermes{
			Theme: &hermes.Flat{},
			Product: hermes.Product{
				Name:      "команда Mindwell",
				Link:      baseURL,
				Logo:      baseURL + "assets/olympus/img/logo-mindwell.png",
				Copyright: "© Mindwell.",
				TroubleText: "Если кнопка '{ACTION}' по какой-то причине не работает, " +
					"скопируй и вставь в адресную строку браузера следующую ссылку: ",
			},
		},
		sprt: support,
		ch:   make(chan *mailgun.Message, 200),
		stop: make(chan interface{}),
	}

	go func() {
		const limitPerInt = 1000
		const interval = time.Hour

		until := time.Now().Add(interval)
		count := 0

		resetCounter := func() {
			until = time.Now().Add(interval)
			count = 0
		}

		for msg := range pm.ch {
			timeLeft := time.Until(until)
			if timeLeft < 0 {
				resetCounter()
			}

			if count == limitPerInt {
				fmt.Printf("Exceeded the limit of emails. Sleeping for %.0f minutes...\n", timeLeft.Minutes())
				time.Sleep(timeLeft)
				resetCounter()
			}

			count++

			resp, id, err := pm.mg.Send(msg)
			if err == nil {
				fmt.Printf("Email ID: %s. Resp: %s.\n", id, resp)
			} else {
				log.Println(err)
			}
		}

		close(pm.stop)
	}()

	return pm
}

func (pm *Postman) Stop() {
	close(pm.ch)
	<-pm.stop
}

func (pm *Postman) send(email hermes.Email, address, subj, name string) {
	email.Body.Title = "Привет, " + name
	email.Body.Signature = "С наилучшими пожеланиями"
	email.Body.Outros = []string{
		// "Изменить настройки уведомлений можно в панели управления учетной записью: " +
		// 	pm.url + "account/email",
		"Появились вопросы или какая-то проблема? " +
			"Не стесняйся и просто ответь на это письмо. Мы будем рады помочь. ",
	}

	text, err := pm.h.GeneratePlainText(email)
	if err != nil {
		log.Println(err)
	}

	from := "Команда Mindwell <support@mindwell.win>"
	recp := name + " <" + address + ">"
	msg := pm.mg.NewMessage(from, subj, text, recp)

	// html, err := pm.h.GenerateHTML(email)
	// if err != nil {
	// 	log.Println(err)
	// }

	// html, err = inliner.Inline(html)
	// if err != nil {
	// 	log.Println(err)
	// }

	// msg.SetHtml(html)

	// err = ioutil.WriteFile("preview.html", []byte(html), 0644)
	// err = ioutil.WriteFile("preview.txt", []byte(text), 0644)

	pm.ch <- msg
}

func (pm *Postman) sendComplain(email hermes.Email, subj string) {
	email.Body.Title = "Привет, дорогой модератор"

	text, err := pm.h.GeneratePlainText(email)
	if err != nil {
		log.Println(err)
	}

	from := "Команда Mindwell <support@mindwell.win>"
	recp := "Команда Mindwell <" + pm.sprt + ">"
	msg := pm.mg.NewMessage(from, subj, text, recp)

	pm.ch <- msg
}

func (pm *Postman) SendGreeting(address, name, code string) {
	email := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				"добро пожаловать на борт нашего корабля!",
				"Располагайся, чувствуй себя как дома. Тебе у нас понравится. ",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Открой эту ссылку, чтобы подтвердить свой email:",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  "Начать пользоваться Mindwell",
						Link:  pm.url + "account/verification/" + address + "?code=" + code,
					},
				},
			},
		},
	}

	subj := "Приветствуем в Mindwell, " + name + "!"
	pm.send(email, address, subj, name)
}

func (pm *Postman) SendPasswordChanged(address, name string) {
	email := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				"пароль к твоему тлогу был изменен.",
			},
		},
	}

	subj := "Изменение пароля"
	pm.send(email, address, subj, name)
}

func (pm *Postman) SendEmailChanged(address, name string) {
	email := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				"твой адрес почты был изменен.",
			},
		},
	}

	subj := "Изменение адреса почты"
	pm.send(email, address, subj, name)
}

func (pm *Postman) SendResetPassword(address, name, gender, code string, date int64) {
	var ending string
	if gender == "female" {
		ending = "а"
	}

	email := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				"кто-то запросил сброс пароля для твоего аккаунта.",
				"Если это был" + ending + " не ты, можешь просто проигнорировать данное письмо.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Или открой эту ссылку и придумай хороший новый пароль. Она будет действительна в течение часа.",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  "Сбросить пароль",
						Link: pm.url + "account/recover?email=" + address +
							"&code=" + code + "&date=" + strconv.FormatInt(date, 10),
					},
				},
			},
		},
	}

	subj := "Забыл" + ending + " пароль, " + name + "?"
	pm.send(email, address, subj, name)
}

func (pm *Postman) SendNewComment(address, fromGender, toShowName, entryTitle string, cmt *models.Comment) {
	var ending string
	if fromGender == "female" {
		ending = "а"
	}

	var entry string
	if len(entryTitle) > 0 {
		entry = " «" + entryTitle + "»"
	} else {
		entry = ", за которой ты следишь"
	}

	email := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				cmt.Author.ShowName + " оставил" + ending + " новый комментарий к записи" + entry + ".",
				"Вот, что он" + ending + " пишет:",
				"«" + cmt.EditContent + "».",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Узнать подробности и ответить можно по ссылке:",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  "Открыть запись",
						Link:  pm.url + "entries/" + strconv.FormatInt(cmt.EntryID, 10) + "#comments",
					},
				},
			},
		},
	}

	subj := "Новый комментарий к записи" + entry
	pm.send(email, address, subj, toShowName)
}

func (pm *Postman) SendNewFollower(address, fromName, fromShowName, fromGender string, toPrivate bool, toShowName string) {
	var ending, pronoun string
	if fromGender == "female" {
		ending = "ась"
		pronoun = "её"
	} else {
		ending = "ся"
		pronoun = "его"
	}

	var intro, text string
	if toPrivate {
		intro = fromShowName + " запрашивает доступ к чтению твоего тлога."
		text = "Принять или отклонить запрос можно на странице " + pronoun + " профиля: "
	} else {
		intro = fromShowName + " подписал" + ending + " на твой тлог."
		text = "Ссылка на " + pronoun + " профиль: "
	}

	email := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				intro,
			},
			Actions: []hermes.Action{
				{
					Instructions: text,
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  fromShowName,
						Link:  pm.url + "users/" + fromName,
					},
				},
			},
		},
	}

	const subj = "Новый подписчик"
	pm.send(email, address, subj, toShowName)
}

func (pm *Postman) SendNewAccept(address, fromName, fromShowName, fromGender, toShowName string) {
	var ending, pronoun string
	if fromGender == "female" {
		ending = "а"
		pronoun = "её"
	} else {
		ending = ""
		pronoun = "его"
	}

	email := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				fromShowName + " разрешил" + ending + " тебе читать свой тлог.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Ссылка на " + pronoun + " профиль: ",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  fromShowName,
						Link:  pm.url + "users/" + fromName,
					},
				},
			},
		},
	}

	const subj = "Доступ открыт"
	pm.send(email, address, subj, toShowName)
}

func (pm *Postman) SendNewInvite(address, name string) {
	email := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				"У тебя появилось новое приглашение! Используй его с умом.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Ожидающие приглашений пользователи перечислены в этом разделе:",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  "Ожидающие",
						Link:  pm.url + "users?top=waiting",
					},
				},
			},
		},
	}

	subj := "Новое приглашение"
	pm.send(email, address, subj, name)
}

func (pm *Postman) SendAdm(address, name, gender string) {
	var ending string
	if gender == "female" {
		ending = "а"
	}

	email := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				"ты подавал" + ending + " заявку для участия в Клубе АДМ.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Твой дорогой внук уже ждет от тебя подарок! " +
						"Вся необходимая информация доступна по этой ссылке:",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  "Клуб АДМ",
						Link:  pm.url + "adm",
					},
				},
			},
		},
	}

	subj := "Клуб анонимных Дедов Морозов"
	pm.send(email, address, subj, name)
}

func (pm *Postman) SendInvited(address, fromShowName, fromGender, toShowName string) {
	var ending string
	if fromGender == "female" {
		ending = "а"
	} else {
		ending = ""
	}

	email := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				fromShowName + " отправил" + ending + " тебе приглашение на Mindwell. " +
					"Теперь тебе доступны все функции сайта (при отсутствии других ограничений).",
			},
		},
	}

	const subj = "Приглашение на Mindwell"
	pm.send(email, address, subj, toShowName)
}

func (pm *Postman) SendAdmSent(address, toShowName string) {
	email := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				"твой Дед Мороз отправил тебе подарок! Когда получишь, не забудь " +
					"поставить нужный флажок. И не открывай до Нового года.",
			},
		},
	}

	subj := "Клуб анонимных Дедов Морозов"
	pm.send(email, address, subj, toShowName)
}

func (pm *Postman) SendAdmReceived(address, toShowName string) {
	email := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				"внук получил твой новогодний подарок.",
			},
		},
	}

	subj := "Клуб анонимных Дедов Морозов"
	pm.send(email, address, subj, toShowName)
}

func (pm *Postman) SendCommentComplain(from, against, content, comment string, commentID, entryID int64) {
	email := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				"Пользователь " + from + " пожаловался на комментарий " +
					strconv.FormatInt(commentID, 10) + " от " + against + ".",
				"Текст комментария:",
				"«" + comment + "».",
				"Пояснение (если есть):",
				"«" + content + "».",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Ссылка на запись с комментарием:",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  "Запись",
						Link:  pm.url + "entries/" + strconv.FormatInt(entryID, 10),
					},
				},
			},
		},
	}

	subj := "Жалоба на комментарий пользователя " + against
	pm.sendComplain(email, subj)
}

func (pm *Postman) SendEntryComplain(from, against, content, entry string, entryID int64) {
	email := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				"Пользователь " + from + " пожаловался на запись " +
					strconv.FormatInt(entryID, 10) + " от " + against + ".",
				"Текст записи:",
				"«" + entry + "».",
				"Пояснение (если есть):",
				"«" + content + "».",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Ссылка на запись:",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  "Запись",
						Link:  pm.url + "entries/" + strconv.FormatInt(entryID, 10),
					},
				},
			},
		},
	}

	subj := "Жалоба на запись пользователя " + against
	pm.sendComplain(email, subj)
}
