// Code generated by go-swagger; DO NOT EDIT.

package relations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
	models "github.com/sevings/mindwell-server/models"
)

// GetRelationsToNameHandlerFunc turns a function with the right signature into a get relations to name handler
type GetRelationsToNameHandlerFunc func(GetRelationsToNameParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetRelationsToNameHandlerFunc) Handle(params GetRelationsToNameParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetRelationsToNameHandler interface for that can handle valid get relations to name params
type GetRelationsToNameHandler interface {
	Handle(GetRelationsToNameParams, *models.UserID) middleware.Responder
}

// NewGetRelationsToName creates a new http.Handler for the get relations to name operation
func NewGetRelationsToName(ctx *middleware.Context, handler GetRelationsToNameHandler) *GetRelationsToName {
	return &GetRelationsToName{Context: ctx, Handler: handler}
}

/*GetRelationsToName swagger:route GET /relations/to/{name} relations getRelationsToName

GetRelationsToName get relations to name API

*/
type GetRelationsToName struct {
	Context *middleware.Context
	Handler GetRelationsToNameHandler
}

func (o *GetRelationsToName) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetRelationsToNameParams()

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
