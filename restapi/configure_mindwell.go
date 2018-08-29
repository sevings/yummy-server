// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"math/rand"
	"net/http"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	accountImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/account"
	commentsImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/comments"
	designImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/design"
	entriesImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/entries"
	favoritesImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/favorites"
	relationsImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/relations"
	usersImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/users"
	votesImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/votes"
	watchingsImpl "github.com/sevings/mindwell-server/internal/app/mindwell-server/watchings"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/restapi/operations"
	"github.com/sevings/mindwell-server/utils"
)

//go:generate swagger generate server --target .. --name  --spec ../web/swagger.yaml --principal models.UserID

func configureFlags(api *operations.MindwellAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.MindwellAPI) http.Handler {
	rand.Seed(time.Now().UTC().UnixNano())

	srv := utils.NewMindwellServer(api, "configs/server")

	domain := srv.ConfigString("mailgun.domain")
	apiKey := srv.ConfigString("mailgun.api_key")
	pubKey := srv.ConfigString("mailgun.pub_key")

	srv.Mail = utils.NewPostman(domain, apiKey, pubKey)

	accountImpl.ConfigureAPI(srv)
	usersImpl.ConfigureAPI(srv)
	entriesImpl.ConfigureAPI(srv)
	votesImpl.ConfigureAPI(srv)
	favoritesImpl.ConfigureAPI(srv)
	watchingsImpl.ConfigureAPI(srv)
	commentsImpl.ConfigureAPI(srv)
	designImpl.ConfigureAPI(srv)
	relationsImpl.ConfigureAPI(srv)

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
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	lmt := tollbooth.NewLimiter(5, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	lmt.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"})
	// lmt.SetHeader("X-Api-Token", []string{})
	// lmt.SetHeaderEntryExpirationTTL(time.Hour)
	lmt.SetMessage(`{"message":"You have reached maximum request limit."}`)
	lmt.SetMessageContentType("application/json")

	return tollbooth.LimitHandler(lmt, handler)
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
