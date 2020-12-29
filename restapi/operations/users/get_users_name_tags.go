// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/mindwell-server/models"
)

// GetUsersNameTagsHandlerFunc turns a function with the right signature into a get users name tags handler
type GetUsersNameTagsHandlerFunc func(GetUsersNameTagsParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetUsersNameTagsHandlerFunc) Handle(params GetUsersNameTagsParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetUsersNameTagsHandler interface for that can handle valid get users name tags params
type GetUsersNameTagsHandler interface {
	Handle(GetUsersNameTagsParams, *models.UserID) middleware.Responder
}

// NewGetUsersNameTags creates a new http.Handler for the get users name tags operation
func NewGetUsersNameTags(ctx *middleware.Context, handler GetUsersNameTagsHandler) *GetUsersNameTags {
	return &GetUsersNameTags{Context: ctx, Handler: handler}
}

/*GetUsersNameTags swagger:route GET /users/{name}/tags users getUsersNameTags

GetUsersNameTags get users name tags API

*/
type GetUsersNameTags struct {
	Context *middleware.Context
	Handler GetUsersNameTagsHandler
}

func (o *GetUsersNameTags) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetUsersNameTagsParams()

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
