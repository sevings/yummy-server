// Code generated by go-swagger; DO NOT EDIT.

package chats

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/mindwell-server/models"
)

// GetChatsHandlerFunc turns a function with the right signature into a get chats handler
type GetChatsHandlerFunc func(GetChatsParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetChatsHandlerFunc) Handle(params GetChatsParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetChatsHandler interface for that can handle valid get chats params
type GetChatsHandler interface {
	Handle(GetChatsParams, *models.UserID) middleware.Responder
}

// NewGetChats creates a new http.Handler for the get chats operation
func NewGetChats(ctx *middleware.Context, handler GetChatsHandler) *GetChats {
	return &GetChats{Context: ctx, Handler: handler}
}

/* GetChats swagger:route GET /chats chats getChats

GetChats get chats API

*/
type GetChats struct {
	Context *middleware.Context
	Handler GetChatsHandler
}

func (o *GetChats) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetChatsParams()
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
