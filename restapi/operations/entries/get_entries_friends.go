// Code generated by go-swagger; DO NOT EDIT.

package entries

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"

	models "github.com/sevings/mindwell-server/models"
)

// GetEntriesFriendsHandlerFunc turns a function with the right signature into a get entries friends handler
type GetEntriesFriendsHandlerFunc func(GetEntriesFriendsParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetEntriesFriendsHandlerFunc) Handle(params GetEntriesFriendsParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetEntriesFriendsHandler interface for that can handle valid get entries friends params
type GetEntriesFriendsHandler interface {
	Handle(GetEntriesFriendsParams, *models.UserID) middleware.Responder
}

// NewGetEntriesFriends creates a new http.Handler for the get entries friends operation
func NewGetEntriesFriends(ctx *middleware.Context, handler GetEntriesFriendsHandler) *GetEntriesFriends {
	return &GetEntriesFriends{Context: ctx, Handler: handler}
}

/*GetEntriesFriends swagger:route GET /entries/friends entries getEntriesFriends

GetEntriesFriends get entries friends API

*/
type GetEntriesFriends struct {
	Context *middleware.Context
	Handler GetEntriesFriendsHandler
}

func (o *GetEntriesFriends) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetEntriesFriendsParams()

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
