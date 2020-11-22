package main

import (
	"go.uber.org/zap"
	"log"
	"os"

	"github.com/sevings/mindwell-server/helper"
	"github.com/sevings/mindwell-server/utils"
)

const admArg = "adm"
const helpArg = "help"

func printHelp() {
	log.Printf(
		`
Usage: mindwell-helper [option]

Options are:
%s		- set grandfathers in adm and sent emails to them.
%s		- print this help message.
`, admArg, helpArg)
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
	support, _ := cfg.String("server.support")

	if len(domain) == 0 || len(apiKey) == 0 || len(pubKey) == 0 || len(baseURL) == 0 {
		log.Println("Check config consistency")
	}

	zapLog, err := zap.NewProduction(zap.WithCaller(false))
	if err != nil {
		log.Println(err)
	}

	emailLog := zapLog.With(zap.String("type", "email"))
	mail := utils.NewPostman(domain, apiKey, pubKey, baseURL, support, emailLog)

	db := utils.OpenDatabase(cfg)
	tx := utils.NewAutoTx(db)
	defer tx.Finish()

	args := os.Args[1:]
	for _, arg := range args {
		switch arg {
		case admArg:
			helper.UpdateAdm(tx, mail)
		case helpArg:
			printHelp()
		default:
			log.Printf("Unknown argument: %s\n", arg)
		}
	}
}
