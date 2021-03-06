// Code generated by go-swagger; DO NOT EDIT.

package me

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/mindwell-server/models"
)

// GetMeInvitedHandlerFunc turns a function with the right signature into a get me invited handler
type GetMeInvitedHandlerFunc func(GetMeInvitedParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetMeInvitedHandlerFunc) Handle(params GetMeInvitedParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetMeInvitedHandler interface for that can handle valid get me invited params
type GetMeInvitedHandler interface {
	Handle(GetMeInvitedParams, *models.UserID) middleware.Responder
}

// NewGetMeInvited creates a new http.Handler for the get me invited operation
func NewGetMeInvited(ctx *middleware.Context, handler GetMeInvitedHandler) *GetMeInvited {
	return &GetMeInvited{Context: ctx, Handler: handler}
}

/* GetMeInvited swagger:route GET /me/invited me getMeInvited

GetMeInvited get me invited API

*/
type GetMeInvited struct {
	Context *middleware.Context
	Handler GetMeInvitedHandler
}

func (o *GetMeInvited) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetMeInvitedParams()
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
