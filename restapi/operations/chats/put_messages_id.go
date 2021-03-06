// Code generated by go-swagger; DO NOT EDIT.

package chats

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/mindwell-server/models"
)

// PutMessagesIDHandlerFunc turns a function with the right signature into a put messages ID handler
type PutMessagesIDHandlerFunc func(PutMessagesIDParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn PutMessagesIDHandlerFunc) Handle(params PutMessagesIDParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// PutMessagesIDHandler interface for that can handle valid put messages ID params
type PutMessagesIDHandler interface {
	Handle(PutMessagesIDParams, *models.UserID) middleware.Responder
}

// NewPutMessagesID creates a new http.Handler for the put messages ID operation
func NewPutMessagesID(ctx *middleware.Context, handler PutMessagesIDHandler) *PutMessagesID {
	return &PutMessagesID{Context: ctx, Handler: handler}
}

/* PutMessagesID swagger:route PUT /messages/{id} chats putMessagesId

PutMessagesID put messages ID API

*/
type PutMessagesID struct {
	Context *middleware.Context
	Handler PutMessagesIDHandler
}

func (o *PutMessagesID) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewPutMessagesIDParams()
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
