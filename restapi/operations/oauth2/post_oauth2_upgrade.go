// Code generated by go-swagger; DO NOT EDIT.

package oauth2

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/mindwell-server/models"
)

// PostOauth2UpgradeHandlerFunc turns a function with the right signature into a post oauth2 upgrade handler
type PostOauth2UpgradeHandlerFunc func(PostOauth2UpgradeParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn PostOauth2UpgradeHandlerFunc) Handle(params PostOauth2UpgradeParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// PostOauth2UpgradeHandler interface for that can handle valid post oauth2 upgrade params
type PostOauth2UpgradeHandler interface {
	Handle(PostOauth2UpgradeParams, *models.UserID) middleware.Responder
}

// NewPostOauth2Upgrade creates a new http.Handler for the post oauth2 upgrade operation
func NewPostOauth2Upgrade(ctx *middleware.Context, handler PostOauth2UpgradeHandler) *PostOauth2Upgrade {
	return &PostOauth2Upgrade{Context: ctx, Handler: handler}
}

/* PostOauth2Upgrade swagger:route POST /oauth2/upgrade oauth2 postOauth2Upgrade

only for internal usage

*/
type PostOauth2Upgrade struct {
	Context *middleware.Context
	Handler PostOauth2UpgradeHandler
}

func (o *PostOauth2Upgrade) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewPostOauth2UpgradeParams()
	uprinc, aCtx, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	if aCtx != nil {
		r = aCtx
	}
	var principal *models.UserID
	if uprinc != nil {
		principal = uprinc.(*models.UserID) // this is really a models.UserID, I promise
	}

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params, principal) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}