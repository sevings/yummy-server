package restapi

import (
	"crypto/tls"
	"math/rand"
	"net/http"
	"strings"
	"time"

	accountImpl "github.com/sevings/yummy-server/internal/app/yummy-server/account"
	commentsImpl "github.com/sevings/yummy-server/internal/app/yummy-server/comments"
	entriesImpl "github.com/sevings/yummy-server/internal/app/yummy-server/entries"
	favoritesImpl "github.com/sevings/yummy-server/internal/app/yummy-server/favorites"
	usersImpl "github.com/sevings/yummy-server/internal/app/yummy-server/users"
	votesImpl "github.com/sevings/yummy-server/internal/app/yummy-server/votes"
	watchingsImpl "github.com/sevings/yummy-server/internal/app/yummy-server/watchings"

	"github.com/didip/tollbooth"
	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	graceful "github.com/tylerb/graceful"

	"github.com/sevings/yummy-server/internal/app/yummy-server"
	"github.com/sevings/yummy-server/models"
	"github.com/sevings/yummy-server/restapi/operations"

	_ "github.com/lib/pq"
)

// This file is safe to edit. Once it exists it will not be overwritten

//go:generate swagger generate server --target ../gen --name  --spec ../swagger-ui/swagger.yaml

func configureFlags(api *operations.YummyAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.YummyAPI) http.Handler {
	rand.Seed(time.Now().UTC().UnixNano())

	config := yummy.LoadConfig("server")
	db := yummy.OpenDatabase(config)

	accountImpl.ConfigureAPI(db, api)
	usersImpl.ConfigureAPI(db, api)
	entriesImpl.ConfigureAPI(db, api)
	votesImpl.ConfigureAPI(db, api)
	favoritesImpl.ConfigureAPI(db, api)
	watchingsImpl.ConfigureAPI(db, api)
	commentsImpl.ConfigureAPI(db, api)

	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()
	api.UrlformConsumer = runtime.DiscardConsumer
	api.MultipartformConsumer = runtime.DiscardConsumer
	api.JSONProducer = runtime.JSONProducer()
	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *graceful.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	handleUI := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Index(r.URL.Path, "/help/api/") == 0 {
			http.StripPrefix("/help/api/", http.FileServer(http.Dir("web"))).ServeHTTP(w, r)
			return
		}
		handler.ServeHTTP(w, r)
	})

	lmt := tollbooth.NewLimiter(10, nil)
	lmt.SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"})
	lmt.SetMessage("")
	lmt.SetMessageContentType("application/json")
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		err := models.Error{Message: "Too many requests"}
		data, _ := err.MarshalBinary()
		w.Write(data)
	})
	return tollbooth.LimitFuncHandler(lmt, handleUI)
}
