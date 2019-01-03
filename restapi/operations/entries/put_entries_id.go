// Code generated by go-swagger; DO NOT EDIT.

package entries

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"

	models "github.com/sevings/mindwell-server/models"
)

// PutEntriesIDHandlerFunc turns a function with the right signature into a put entries ID handler
type PutEntriesIDHandlerFunc func(PutEntriesIDParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn PutEntriesIDHandlerFunc) Handle(params PutEntriesIDParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// PutEntriesIDHandler interface for that can handle valid put entries ID params
type PutEntriesIDHandler interface {
	Handle(PutEntriesIDParams, *models.UserID) middleware.Responder
}

// NewPutEntriesID creates a new http.Handler for the put entries ID operation
func NewPutEntriesID(ctx *middleware.Context, handler PutEntriesIDHandler) *PutEntriesID {
	return &PutEntriesID{Context: ctx, Handler: handler}
}

/*PutEntriesID swagger:route PUT /entries/{id} entries putEntriesId

PutEntriesID put entries ID API

*/
type PutEntriesID struct {
	Context *middleware.Context
	Handler PutEntriesIDHandler
}

func (o *PutEntriesID) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewPutEntriesIDParams()

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
