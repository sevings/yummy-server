package main

import (
	"log"
	"os"

	"github.com/sevings/mindwell-server/helper"
	"github.com/sevings/mindwell-server/utils"
)

const admArg = "adm"
const commentsArg = "comments"
const entriesArg = "entries"
const helpArg = "help"

func printHelp() {
	log.Printf(
		`
Usage: mindwell-helper [option]

Options are:
%s	- update comments content.
%s		- update entries title and content.
%s		- set grandfathers in adm and sent emails to them.
%s		- print this help message.
`, commentsArg, entriesArg, admArg, helpArg)
}

func main() {
	if len(os.Args) == 1 || os.Args[1] == helpArg {
		printHelp()
		return
	}

	cfg := utils.LoadConfig("configs/server")

	domain, _ := cfg.String("mailgun.domain")
	apiKey, _ := cfg.String("mailgun.api_key")
	pubKey, _ := cfg.String("mailgun.pub_key")
	baseURL, _ := cfg.String("server.base_url")

	if len(domain) == 0 || len(apiKey) == 0 || len(pubKey) == 0 || len(baseURL) == 0 {
		log.Println("Check config consistency")
	}

	mail := utils.NewPostman(domain, apiKey, pubKey, baseURL)

	db := utils.OpenDatabase(cfg)
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Println(err)
	}

	args := os.Args[1:]
	for _, arg := range args {
		switch arg {
		case commentsArg:
			helper.UpdateComments(tx)
		case entriesArg:
			helper.UpdateEntries(tx)
		case admArg:
			helper.UpdateAdm(tx, mail)
		case helpArg:
			printHelp()
		default:
			log.Printf("Unknown argument: %s\n", arg)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
	}
}
