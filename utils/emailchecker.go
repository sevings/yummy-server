package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const checkUrl = "http://open.kickbox.com/v1/disposable/"

type reply struct {
	Disposable bool
}

type EmailChecker struct {
	srv     *MindwellServer
	client  *http.Client
	trusted []string
	banned  []string
}

func NewEmailChecker(srv *MindwellServer) *EmailChecker {
	return &EmailChecker{
		srv: srv,
		client: &http.Client{
			Timeout: time.Second,
		},
		trusted: srv.ConfigStrings("server.trust_email"),
		banned:  srv.ConfigStrings("server.ban_email"),
	}
}

func (ec *EmailChecker) IsAllowed(email string) bool {
	loginAtService := strings.Split(email, "@")
	if len(loginAtService) < 2 {
		return false
	}

	service := loginAtService[1]

	for _, s := range ec.trusted {
		if s == service {
			return true
		}
	}

	for _, s := range ec.banned {
		if s == service {
			return false
		}
	}

	resp, err := ec.client.Get(checkUrl + service)
	if err != nil {
		ec.logError(err)
		return true
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ec.logError(err)
		return true
	}

	if resp.StatusCode != 200 {
		ec.srv.LogSystem().Error(string(body))
		return true
	}

	var result reply
	err = json.Unmarshal(body, &result)
	if err != nil {
		ec.logError(err)
		return true
	}

	if result.Disposable {
		ec.banned = append(ec.banned, service)
	}

	return !result.Disposable
}

func (ec *EmailChecker) logError(err error) {
	ec.srv.LogSystem().Error(err.Error())
}
