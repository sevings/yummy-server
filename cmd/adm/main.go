package main

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/sevings/mindwell-server/utils"
	goconf "github.com/zpatrick/go-config"
)

func takeRandom(s []string) (string, []string) {
	var result string
	i := rand.Intn(len(s))
	result, s[i] = s[i], s[len(s)-1]
	return result, s[:len(s)-1]
}

func postman(cfg *goconf.Config) *utils.Postman {
	domain, _ := cfg.String("mailgun.domain")
	apiKey, _ := cfg.String("mailgun.api_key")
	pubKey, _ := cfg.String("mailgun.pub_key")
	baseURL, _ := cfg.String("server.base_url")

	if len(domain) == 0 || len(apiKey) == 0 || len(pubKey) == 0 || len(baseURL) == 0 {
		log.Println("Check config consistency")
		return nil
	}

	return utils.NewPostman(domain, apiKey, pubKey, baseURL)
}

type user struct {
	email  string
	name   string
	gender string
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	cfg := utils.LoadConfig("configs/server")
	db := utils.OpenDatabase(cfg)
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Println(err)
	}

	var names [][]string
	for i := 0; i < 3; i++ {
		names = append(names, []string{})
	}

	rows, err := tx.Query("SELECT users.name, users.gender FROM users, adm " +
		"WHERE lower(adm.name) = lower(users.name)" +
		"ORDER BY gender")
	if err != nil {
		log.Println(err)
		return
	}

	for rows.Next() {
		var name string
		var gender int64
		rows.Scan(&name, &gender)
		names[gender] = append(names[gender], name)
	}

	cnt := len(names[0]) + len(names[1]) + len(names[2])
	log.Printf("Found %d adms...\n", cnt)

	adm := make([]string, cnt)
	var i int

	fillGender := func(gender int) {
		for ; len(names[gender]) > 0; i += 2 {
			if i >= cnt {
				i = 1
			}

			adm[i], names[gender] = takeRandom(names[gender])
		}
	}

	for j := 2; j >= 0; j-- {
		fillGender(j)
	}

	setAdm := func(gs, gf string) {
		res, err := tx.Exec("UPDATE adm SET grandfather = $2 WHERE lower(name) = lower($1)", gf, gs)
		if err != nil {
			log.Println(err)
		}

		rows, _ := res.RowsAffected()
		if rows != 1 {
			log.Printf("Couldn't set grandfather for %s\n", gs)
		}
	}

	for i = 0; i < cnt-1; i++ {
		setAdm(adm[i], adm[i+1])
	}
	setAdm(adm[cnt-1], adm[0])

	var users []user

	for _, name := range adm {
		rows, err := tx.Query("SELECT show_name, gender.type, email, verified FROM users, gender "+
			"WHERE lower(name) = lower($1) AND users.gender = gender.id", name)
		if err != nil {
			log.Println(err)
		}
		for rows.Next() {
			var verified bool
			var usr user
			rows.Scan(&usr.name, &usr.gender, &usr.email, &verified)
			if verified {
				users = append(users, usr)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	mail := postman(cfg)
	for _, usr := range users {
		mail.SendAdm(usr.email, usr.name, usr.gender)
	}

	log.Println("Completed. Sending emails... (press Enter to exit)")

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}
