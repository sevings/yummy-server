// Code generated by go-swagger; DO NOT EDIT.

package adm

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
	models "github.com/sevings/mindwell-server/models"
)

// PostAdmGrandsonHandlerFunc turns a function with the right signature into a post adm grandson handler
type PostAdmGrandsonHandlerFunc func(PostAdmGrandsonParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn PostAdmGrandsonHandlerFunc) Handle(params PostAdmGrandsonParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// PostAdmGrandsonHandler interface for that can handle valid post adm grandson params
type PostAdmGrandsonHandler interface {
	Handle(PostAdmGrandsonParams, *models.UserID) middleware.Responder
}

// NewPostAdmGrandson creates a new http.Handler for the post adm grandson operation
func NewPostAdmGrandson(ctx *middleware.Context, handler PostAdmGrandsonHandler) *PostAdmGrandson {
	return &PostAdmGrandson{Context: ctx, Handler: handler}
}

/*PostAdmGrandson swagger:route POST /adm/grandson adm postAdmGrandson

PostAdmGrandson post adm grandson API

*/
type PostAdmGrandson struct {
	Context *middleware.Context
	Handler PostAdmGrandsonHandler
}

func (o *PostAdmGrandson) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewPostAdmGrandsonParams()

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
