package helper

import (
	"bufio"
	"database/sql"
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

func genderNames(tx *sql.Tx) ([][]string, error) {
	var names [][]string
	for i := 0; i < 3; i++ {
		names = append(names, []string{})
	}

	rows, err := tx.Query("SELECT users.name, users.gender FROM users, adm " +
		"WHERE lower(adm.name) = lower(users.name)" +
		"ORDER BY gender")
	if err != nil {
		return names, err
	}

	for rows.Next() {
		var name string
		var gender int64
		rows.Scan(&name, &gender)
		names[gender] = append(names[gender], name)
	}

	return names, nil
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

func setAdm(adm []string, tx *sql.Tx) error {
	set := func(gs, gf string) error {
		res, err := tx.Exec("UPDATE adm SET grandfather = $2 WHERE lower(name) = lower($1)", gf, gs)
		if err != nil {
			return err
		}

		rows, err := res.RowsAffected()
		if rows != 1 {
			log.Printf("Couldn't set grandfather for %s\n", gs)
		}

		return nil
	}

	for i := 0; i < len(adm)-1; i++ {
		if err := set(adm[i], adm[i+1]); err != nil {
			return err
		}
	}

	return set(adm[len(adm)-1], adm[0])
}

func loadUsers(adm []string, tx *sql.Tx) ([]user, error) {
	var users []user

	for _, name := range adm {
		rows, err := tx.Query("SELECT show_name, gender.type, email, verified FROM users, gender "+
			"WHERE lower(name) = lower($1) AND users.gender = gender.id", name)
		if err != nil {
			return users, err
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

	return users, nil
}

func takeRandom(s []string) (string, []string) {
	var result string
	i := rand.Intn(len(s))
	result, s[i] = s[i], s[len(s)-1]
	return result, s[:len(s)-1]
}

func UpdateAdm(tx *sql.Tx, mail *utils.Postman) {
	rand.Seed(time.Now().UTC().UnixNano())

	names, err := genderNames(tx)
	if err != nil {
		log.Println(err)
		return
	}

	adm := mixNames(names)

	err = setAdm(adm, tx)
	if err != nil {
		log.Println(err)
		return
	}

	users, err := loadUsers(adm, tx)
	if err != nil {
		log.Println(err)
		return
	}

	for _, usr := range users {
		mail.SendAdm(usr.email, usr.name, usr.gender)
	}

	log.Println("Completed. Sending emails... (press Enter to exit)")

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}
