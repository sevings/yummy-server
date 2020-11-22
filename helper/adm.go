package helper

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/sevings/mindwell-server/utils"
)

type user struct {
	email  string
	name   string
	gender string
}

func genderNames(tx *utils.AutoTx) [][]string {
	var names [][]string

	tx.Query("SELECT users.name, users.gender FROM users, adm " +
		"WHERE lower(adm.name) = lower(users.name)" +
		"ORDER BY gender")

	for {
		var name string
		var gender int64
		ok := tx.Scan(&name, &gender)
		if !ok {
			break
		}

		for int(gender) >= len(names) {
			names = append(names, []string{})
		}

		names[gender] = append(names[gender], name)
	}

	return names
}

func mixNames(names [][]string) []string {
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

	return adm
}

func setAdm(adm []string, tx *utils.AutoTx) {
	set := func(gs, gf string) {
		tx.Exec("UPDATE adm SET grandfather = $2 WHERE lower(name) = lower($1)", gf, gs)

		rows := tx.RowsAffected()
		if rows != 1 {
			log.Printf("Couldn't set grandfather for %s\n", gs)
		}
	}

	for i := 0; i < len(adm)-1; i++ {
		set(adm[i], adm[i+1])
	}

	set(adm[len(adm)-1], adm[0])
}

func loadUsers(adm []string, tx *utils.AutoTx) []user {
	var users []user

	for _, name := range adm {
		tx.Query("SELECT show_name, gender.type, email, verified FROM users, gender "+
			"WHERE lower(name) = lower($1) AND users.gender = gender.id", name)
		for {
			var verified bool
			var usr user
			ok := tx.Scan(&usr.name, &usr.gender, &usr.email, &verified)
			if !ok {
				break
			}
			if verified {
				users = append(users, usr)
			}
		}
	}

	return users
}

func takeRandom(s []string) (string, []string) {
	var result string
	i := rand.Intn(len(s))
	result, s[i] = s[i], s[len(s)-1]
	return result, s[:len(s)-1]
}

func UpdateAdm(tx *utils.AutoTx, mail *utils.Postman) {
	rand.Seed(time.Now().UTC().UnixNano())

	names := genderNames(tx)
	if tx.Error() != nil {
		log.Println(tx.Error())
		return
	}

	adm := mixNames(names)
	if tx.Error() != nil {
		log.Println(tx.Error())
		return
	}

	setAdm(adm, tx)
	if tx.Error() != nil {
		log.Println(tx.Error())
		return
	}

	users := loadUsers(adm, tx)
	if tx.Error() != nil {
		log.Println(tx.Error())
		return
	}

	for _, usr := range users {
		mail.SendAdm(usr.email, usr.name, usr.gender)
	}

	log.Println("Completed. Sending emails... (press Enter to exit)")

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}
