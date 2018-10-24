// Code generated by go-swagger; DO NOT EDIT.

package relations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
	models "github.com/sevings/mindwell-server/models"
)

// PutRelationsFromNameHandlerFunc turns a function with the right signature into a put relations from name handler
type PutRelationsFromNameHandlerFunc func(PutRelationsFromNameParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn PutRelationsFromNameHandlerFunc) Handle(params PutRelationsFromNameParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// PutRelationsFromNameHandler interface for that can handle valid put relations from name params
type PutRelationsFromNameHandler interface {
	Handle(PutRelationsFromNameParams, *models.UserID) middleware.Responder
}

// NewPutRelationsFromName creates a new http.Handler for the put relations from name operation
func NewPutRelationsFromName(ctx *middleware.Context, handler PutRelationsFromNameHandler) *PutRelationsFromName {
	return &PutRelationsFromName{Context: ctx, Handler: handler}
}

/*PutRelationsFromName swagger:route PUT /relations/from/{name} relations putRelationsFromName

permit the user to follow you

*/
type PutRelationsFromName struct {
	Context *middleware.Context
	Handler PutRelationsFromNameHandler
}

func (o *PutRelationsFromName) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewPutRelationsFromNameParams()

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
