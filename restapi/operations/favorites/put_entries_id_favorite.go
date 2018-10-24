// Code generated by go-swagger; DO NOT EDIT.

package favorites

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
	models "github.com/sevings/mindwell-server/models"
)

// PutEntriesIDFavoriteHandlerFunc turns a function with the right signature into a put entries ID favorite handler
type PutEntriesIDFavoriteHandlerFunc func(PutEntriesIDFavoriteParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn PutEntriesIDFavoriteHandlerFunc) Handle(params PutEntriesIDFavoriteParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// PutEntriesIDFavoriteHandler interface for that can handle valid put entries ID favorite params
type PutEntriesIDFavoriteHandler interface {
	Handle(PutEntriesIDFavoriteParams, *models.UserID) middleware.Responder
}

// NewPutEntriesIDFavorite creates a new http.Handler for the put entries ID favorite operation
func NewPutEntriesIDFavorite(ctx *middleware.Context, handler PutEntriesIDFavoriteHandler) *PutEntriesIDFavorite {
	return &PutEntriesIDFavorite{Context: ctx, Handler: handler}
}

/*PutEntriesIDFavorite swagger:route PUT /entries/{id}/favorite favorites putEntriesIdFavorite

PutEntriesIDFavorite put entries ID favorite API

*/
type PutEntriesIDFavorite struct {
	Context *middleware.Context
	Handler PutEntriesIDFavoriteHandler
}

func (o *PutEntriesIDFavorite) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewPutEntriesIDFavoriteParams()

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
