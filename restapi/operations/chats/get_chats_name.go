// Code generated by go-swagger; DO NOT EDIT.

package chats

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"

	models "github.com/sevings/mindwell-server/models"
)

// GetChatsNameHandlerFunc turns a function with the right signature into a get chats name handler
type GetChatsNameHandlerFunc func(GetChatsNameParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetChatsNameHandlerFunc) Handle(params GetChatsNameParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetChatsNameHandler interface for that can handle valid get chats name params
type GetChatsNameHandler interface {
	Handle(GetChatsNameParams, *models.UserID) middleware.Responder
}

// NewGetChatsName creates a new http.Handler for the get chats name operation
func NewGetChatsName(ctx *middleware.Context, handler GetChatsNameHandler) *GetChatsName {
	return &GetChatsName{Context: ctx, Handler: handler}
}

/*GetChatsName swagger:route GET /chats/{name} chats getChatsName

GetChatsName get chats name API

*/
type GetChatsName struct {
	Context *middleware.Context
	Handler GetChatsNameHandler
}

func (o *GetChatsName) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetChatsNameParams()

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