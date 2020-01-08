package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sevings/mindwell-server/utils"
	"github.com/zpatrick/go-config"
	"io/ioutil"
	"log"
	"net/http"
)

var conf *config.Config
var secret []byte
var linked map[int64]string

func main() {
	conf = utils.LoadConfig("configs/im")

	sec, err := conf.String("server.im_secret")
	if err != nil {
		log.Println(err)
	}
	secret = []byte(sec)

	linked = make(map[int64]string)

	router := gin.Default()
	router.POST("/add", unsupportedHandler)
	router.POST("/auth", authHandler)
	router.POST("/checkunique", unsupportedHandler)
	router.POST("/del", unsupportedHandler)
	router.POST("/gen", unsupportedHandler)
	router.POST("/link", linkHandler)
	router.POST("/upd", unsupportedHandler)
	router.POST("/rtagns", rtagnsHandler)
	router.NoRoute(notFoundHandler)

	//gin.SetMode(gin.ReleaseMode)

	srv := &http.Server{
		Addr:    ":5000",
		Handler: router,
	}

	// service connections
	if err := srv.ListenAndServe(); err != nil {
		log.Printf("listen: %s\n", err)
	}
}

type record struct {
	Uid      string   `json:"uid,omitempty"`
	Authlvl  string   `json:"authlvl,omitempty"`
	Lifetime string   `json:"lifetime,omitempty"`
	Features string   `json:"features,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

type request struct {
	Endpoint string `json:"endpoint,omitempty"`
	Secret   string `json:"secret,omitempty"`
	Rec      record `json:"rec,omitempty"`
}

type public struct {
	Id   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type private struct {
	Comment  string `json:"comment,omitempty"`
	Arch     bool   `json:"arch,omitempty"`
	Accepted string `json:"accepted,omitempty"`
}

type account struct {
	Auth    string  `json:"auth,omitempty"`
	Anon    string  `json:"anon,omitempty"`
	Public  public  `json:"public,omitempty"`
	Private private `json:"private,omitempty"`
}

type response struct {
	Err     string   `json:"err,omitempty"`
	Rec     record   `json:"rec,omitempty"`
	Byteval string   `json:"byteval,omitempty"`
	Ts      string   `json:"ts,omitempty"`
	Boolval bool     `json:"boolval,omitempty"`
	Strarr  []string `json:"strarr,omitempty"`
	Newacc  account  `json:"newacc,omitempty"`
}

func replyJson(ctx *gin.Context, resp response) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = ctx.Writer.Write(data)
	if err != nil {
		log.Println(err)
	}

	log.Println("reply:", string(data))
}

func replyError(ctx *gin.Context, err string) {
	resp := response{Err: err}
	replyJson(ctx, resp)
}

func readRequest(ctx *gin.Context, endpoint string) (req request, err error) {
	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Println(err)
		replyError(ctx, "malformed")
		return
	}

	log.Println("request:", string(data))

	err = json.Unmarshal(data, &req)
	if err != nil {
		log.Println(err)
		replyError(ctx, "malformed")
		return
	}

	if req.Endpoint != endpoint {
		err = errors.Errorf("Expected endpoint '%s', but got '%s'", endpoint, req.Endpoint)
		log.Println(err)
		replyError(ctx, "malformed")
		return
	}

	return
}

func unsupportedHandler(ctx *gin.Context) {
	replyError(ctx, "unsupported")
}

func notFoundHandler(ctx *gin.Context) {
	ctx.Writer.WriteHeader(404)
	replyError(ctx, "not found")
}

func parseSecret(tokenString string) (id int64, name string, err error) {
	tokenB, err := base64.StdEncoding.DecodeString(tokenString)
	if err != nil {
		return
	}
	tokenString = string(tokenB)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})

	if err != nil {
		return
	}

	if !token.Valid {
		err = errors.Errorf("Invalid token: %s\n", tokenString)
		return

	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = errors.Errorf("Error get claims: %s\n", tokenString)
		return
	}

	if claims.Valid() != nil {
		err = errors.Errorf("Claims is invalid: %s\n", tokenString)
		return
	}

	idF, ok := claims["uid"].(float64)
	if !ok {
		err = errors.Errorf("UserID not found: %s\n", tokenString)
		return
	}

	id = int64(idF)

	name, ok = claims["sub"].(string)
	if !ok {
		err = errors.Errorf("Name not found: %s\n", tokenString)
		return
	}

	return
}

func authHandler(ctx *gin.Context) {
	req, err := readRequest(ctx, "auth")
	if err != nil {
		return
	}

	userID, name, err := parseSecret(req.Secret)
	if err != nil {
		log.Println(err)
		replyError(ctx, "failed")
		return
	}

	uid, found := linked[userID]
	if found {
		resp := response{
			Rec: record{
				Uid:     uid,
				Authlvl: "auth",
			},
		}
		replyJson(ctx, resp)
		return
	}

	resp := response{
		Rec: record{
			Authlvl: "auth",
			Tags:    []string{"name:" + name},
		},
		Newacc: account{
			Auth: "JRWP",
			Anon: "N",
			Public: public{
				Id:   userID,
				Name: name,
			},
			Private: private{},
		},
	}
	replyJson(ctx, resp)
}

func linkHandler(ctx *gin.Context) {
	req, err := readRequest(ctx, "link")
	if err != nil {
		return
	}

	if req.Secret == "" || req.Rec.Uid == "" {
		replyError(ctx, "malformed")
		return
	}

	userID, _, err := parseSecret(req.Secret)
	if err != nil {
		log.Println(err)
		replyError(ctx, "failed")
		return
	}

	_, found := linked[userID]
	if found {
		replyError(ctx, "duplicate value")
		return
	}

	linked[userID] = req.Rec.Uid
	replyJson(ctx, response{})
}

func rtagnsHandler(ctx *gin.Context) {
	_, err := readRequest(ctx, "rtagns")
	if err != nil {
		return
	}

	resp := response{
		Strarr: []string{"name"},
	}
	replyJson(ctx, resp)
}
