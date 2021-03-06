// Code generated by go-swagger; DO NOT EDIT.

package watchings

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/mindwell-server/models"
)

// GetEntriesIDWatchingHandlerFunc turns a function with the right signature into a get entries ID watching handler
type GetEntriesIDWatchingHandlerFunc func(GetEntriesIDWatchingParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetEntriesIDWatchingHandlerFunc) Handle(params GetEntriesIDWatchingParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetEntriesIDWatchingHandler interface for that can handle valid get entries ID watching params
type GetEntriesIDWatchingHandler interface {
	Handle(GetEntriesIDWatchingParams, *models.UserID) middleware.Responder
}

// NewGetEntriesIDWatching creates a new http.Handler for the get entries ID watching operation
func NewGetEntriesIDWatching(ctx *middleware.Context, handler GetEntriesIDWatchingHandler) *GetEntriesIDWatching {
	return &GetEntriesIDWatching{Context: ctx, Handler: handler}
}

/* GetEntriesIDWatching swagger:route GET /entries/{id}/watching watchings getEntriesIdWatching

GetEntriesIDWatching get entries ID watching API

*/
type GetEntriesIDWatching struct {
	Context *middleware.Context
	Handler GetEntriesIDWatchingHandler
}

func (o *GetEntriesIDWatching) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetEntriesIDWatchingParams()
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
