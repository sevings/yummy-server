// Code generated by go-swagger; DO NOT EDIT.

package relations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
	models "github.com/sevings/mindwell-server/models"
)

// DeleteRelationsFromNameHandlerFunc turns a function with the right signature into a delete relations from name handler
type DeleteRelationsFromNameHandlerFunc func(DeleteRelationsFromNameParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn DeleteRelationsFromNameHandlerFunc) Handle(params DeleteRelationsFromNameParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// DeleteRelationsFromNameHandler interface for that can handle valid delete relations from name params
type DeleteRelationsFromNameHandler interface {
	Handle(DeleteRelationsFromNameParams, *models.UserID) middleware.Responder
}

// NewDeleteRelationsFromName creates a new http.Handler for the delete relations from name operation
func NewDeleteRelationsFromName(ctx *middleware.Context, handler DeleteRelationsFromNameHandler) *DeleteRelationsFromName {
	return &DeleteRelationsFromName{Context: ctx, Handler: handler}
}

/*DeleteRelationsFromName swagger:route DELETE /relations/from/{name} relations deleteRelationsFromName

cancel following request or unsubscribe the user

*/
type DeleteRelationsFromName struct {
	Context *middleware.Context
	Handler DeleteRelationsFromNameHandler
}

func (o *DeleteRelationsFromName) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewDeleteRelationsFromNameParams()

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
