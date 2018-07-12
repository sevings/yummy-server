// Code generated by go-swagger; DO NOT EDIT.

package me

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"

	models "github.com/sevings/mindwell-server/models"
)

// GetMeTlogHandlerFunc turns a function with the right signature into a get me tlog handler
type GetMeTlogHandlerFunc func(GetMeTlogParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetMeTlogHandlerFunc) Handle(params GetMeTlogParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetMeTlogHandler interface for that can handle valid get me tlog params
type GetMeTlogHandler interface {
	Handle(GetMeTlogParams, *models.UserID) middleware.Responder
}

// NewGetMeTlog creates a new http.Handler for the get me tlog operation
func NewGetMeTlog(ctx *middleware.Context, handler GetMeTlogHandler) *GetMeTlog {
	return &GetMeTlog{Context: ctx, Handler: handler}
}

/*GetMeTlog swagger:route GET /me/tlog me getMeTlog

GetMeTlog get me tlog API

*/
type GetMeTlog struct {
	Context *middleware.Context
	Handler GetMeTlogHandler
}

func (o *GetMeTlog) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetMeTlogParams()

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
