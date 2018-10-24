// Code generated by go-swagger; DO NOT EDIT.

package favorites

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
	models "github.com/sevings/mindwell-server/models"
)

// DeleteEntriesIDFavoriteHandlerFunc turns a function with the right signature into a delete entries ID favorite handler
type DeleteEntriesIDFavoriteHandlerFunc func(DeleteEntriesIDFavoriteParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn DeleteEntriesIDFavoriteHandlerFunc) Handle(params DeleteEntriesIDFavoriteParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// DeleteEntriesIDFavoriteHandler interface for that can handle valid delete entries ID favorite params
type DeleteEntriesIDFavoriteHandler interface {
	Handle(DeleteEntriesIDFavoriteParams, *models.UserID) middleware.Responder
}

// NewDeleteEntriesIDFavorite creates a new http.Handler for the delete entries ID favorite operation
func NewDeleteEntriesIDFavorite(ctx *middleware.Context, handler DeleteEntriesIDFavoriteHandler) *DeleteEntriesIDFavorite {
	return &DeleteEntriesIDFavorite{Context: ctx, Handler: handler}
}

/*DeleteEntriesIDFavorite swagger:route DELETE /entries/{id}/favorite favorites deleteEntriesIdFavorite

DeleteEntriesIDFavorite delete entries ID favorite API

*/
type DeleteEntriesIDFavorite struct {
	Context *middleware.Context
	Handler DeleteEntriesIDFavoriteHandler
}

func (o *DeleteEntriesIDFavorite) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewDeleteEntriesIDFavoriteParams()

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
