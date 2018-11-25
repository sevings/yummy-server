package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/sevings/mindwell-server/utils"
)

func takeRandom(s []string) (string, []string) {
	var result string
	i := rand.Intn(len(s))
	result, s[i] = s[i], s[len(s)-1]
	return result, s[:len(s)-1]
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

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}

	log.Println("Completed.")
}
