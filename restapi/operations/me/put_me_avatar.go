// Code generated by go-swagger; DO NOT EDIT.

package me

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/mindwell-server/models"
)

// PutMeAvatarHandlerFunc turns a function with the right signature into a put me avatar handler
type PutMeAvatarHandlerFunc func(PutMeAvatarParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn PutMeAvatarHandlerFunc) Handle(params PutMeAvatarParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// PutMeAvatarHandler interface for that can handle valid put me avatar params
type PutMeAvatarHandler interface {
	Handle(PutMeAvatarParams, *models.UserID) middleware.Responder
}

// NewPutMeAvatar creates a new http.Handler for the put me avatar operation
func NewPutMeAvatar(ctx *middleware.Context, handler PutMeAvatarHandler) *PutMeAvatar {
	return &PutMeAvatar{Context: ctx, Handler: handler}
}

/*PutMeAvatar swagger:route PUT /me/avatar me putMeAvatar

PutMeAvatar put me avatar API

*/
type PutMeAvatar struct {
	Context *middleware.Context
	Handler PutMeAvatarHandler
}

func (o *PutMeAvatar) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewPutMeAvatarParams()

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
