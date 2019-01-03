// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"

	models "github.com/sevings/mindwell-server/models"
)

// GetUsersNameHandlerFunc turns a function with the right signature into a get users name handler
type GetUsersNameHandlerFunc func(GetUsersNameParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetUsersNameHandlerFunc) Handle(params GetUsersNameParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetUsersNameHandler interface for that can handle valid get users name params
type GetUsersNameHandler interface {
	Handle(GetUsersNameParams, *models.UserID) middleware.Responder
}

// NewGetUsersName creates a new http.Handler for the get users name operation
func NewGetUsersName(ctx *middleware.Context, handler GetUsersNameHandler) *GetUsersName {
	return &GetUsersName{Context: ctx, Handler: handler}
}

/*GetUsersName swagger:route GET /users/{name} users getUsersName

GetUsersName get users name API

*/
type GetUsersName struct {
	Context *middleware.Context
	Handler GetUsersNameHandler
}

func (o *GetUsersName) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetUsersNameParams()

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
