// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/mindwell-server/models"
)

// GetUsersNameTlogHandlerFunc turns a function with the right signature into a get users name tlog handler
type GetUsersNameTlogHandlerFunc func(GetUsersNameTlogParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetUsersNameTlogHandlerFunc) Handle(params GetUsersNameTlogParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetUsersNameTlogHandler interface for that can handle valid get users name tlog params
type GetUsersNameTlogHandler interface {
	Handle(GetUsersNameTlogParams, *models.UserID) middleware.Responder
}

// NewGetUsersNameTlog creates a new http.Handler for the get users name tlog operation
func NewGetUsersNameTlog(ctx *middleware.Context, handler GetUsersNameTlogHandler) *GetUsersNameTlog {
	return &GetUsersNameTlog{Context: ctx, Handler: handler}
}

/*GetUsersNameTlog swagger:route GET /users/{name}/tlog users getUsersNameTlog

GetUsersNameTlog get users name tlog API

*/
type GetUsersNameTlog struct {
	Context *middleware.Context
	Handler GetUsersNameTlogHandler
}

func (o *GetUsersNameTlog) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetUsersNameTlogParams()

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
