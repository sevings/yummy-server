// Code generated by go-swagger; DO NOT EDIT.

package account

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/mindwell-server/models"
)

// PutAccountSettingsEmailHandlerFunc turns a function with the right signature into a put account settings email handler
type PutAccountSettingsEmailHandlerFunc func(PutAccountSettingsEmailParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn PutAccountSettingsEmailHandlerFunc) Handle(params PutAccountSettingsEmailParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// PutAccountSettingsEmailHandler interface for that can handle valid put account settings email params
type PutAccountSettingsEmailHandler interface {
	Handle(PutAccountSettingsEmailParams, *models.UserID) middleware.Responder
}

// NewPutAccountSettingsEmail creates a new http.Handler for the put account settings email operation
func NewPutAccountSettingsEmail(ctx *middleware.Context, handler PutAccountSettingsEmailHandler) *PutAccountSettingsEmail {
	return &PutAccountSettingsEmail{Context: ctx, Handler: handler}
}

/*PutAccountSettingsEmail swagger:route PUT /account/settings/email account putAccountSettingsEmail

PutAccountSettingsEmail put account settings email API

*/
type PutAccountSettingsEmail struct {
	Context *middleware.Context
	Handler PutAccountSettingsEmailHandler
}

func (o *PutAccountSettingsEmail) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewPutAccountSettingsEmailParams()

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
