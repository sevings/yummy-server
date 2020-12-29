// Code generated by go-swagger; DO NOT EDIT.

package entries

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/mindwell-server/models"
)

// PostEntriesIDComplainHandlerFunc turns a function with the right signature into a post entries ID complain handler
type PostEntriesIDComplainHandlerFunc func(PostEntriesIDComplainParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn PostEntriesIDComplainHandlerFunc) Handle(params PostEntriesIDComplainParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// PostEntriesIDComplainHandler interface for that can handle valid post entries ID complain params
type PostEntriesIDComplainHandler interface {
	Handle(PostEntriesIDComplainParams, *models.UserID) middleware.Responder
}

// NewPostEntriesIDComplain creates a new http.Handler for the post entries ID complain operation
func NewPostEntriesIDComplain(ctx *middleware.Context, handler PostEntriesIDComplainHandler) *PostEntriesIDComplain {
	return &PostEntriesIDComplain{Context: ctx, Handler: handler}
}

/*PostEntriesIDComplain swagger:route POST /entries/{id}/complain entries postEntriesIdComplain

PostEntriesIDComplain post entries ID complain API

*/
type PostEntriesIDComplain struct {
	Context *middleware.Context
	Handler PostEntriesIDComplainHandler
}

func (o *PostEntriesIDComplain) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewPostEntriesIDComplainParams()

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
