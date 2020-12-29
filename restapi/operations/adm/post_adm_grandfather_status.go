// Code generated by go-swagger; DO NOT EDIT.

package adm

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/mindwell-server/models"
)

// PostAdmGrandfatherStatusHandlerFunc turns a function with the right signature into a post adm grandfather status handler
type PostAdmGrandfatherStatusHandlerFunc func(PostAdmGrandfatherStatusParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn PostAdmGrandfatherStatusHandlerFunc) Handle(params PostAdmGrandfatherStatusParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// PostAdmGrandfatherStatusHandler interface for that can handle valid post adm grandfather status params
type PostAdmGrandfatherStatusHandler interface {
	Handle(PostAdmGrandfatherStatusParams, *models.UserID) middleware.Responder
}

// NewPostAdmGrandfatherStatus creates a new http.Handler for the post adm grandfather status operation
func NewPostAdmGrandfatherStatus(ctx *middleware.Context, handler PostAdmGrandfatherStatusHandler) *PostAdmGrandfatherStatus {
	return &PostAdmGrandfatherStatus{Context: ctx, Handler: handler}
}

/*PostAdmGrandfatherStatus swagger:route POST /adm/grandfather/status adm postAdmGrandfatherStatus

PostAdmGrandfatherStatus post adm grandfather status API

*/
type PostAdmGrandfatherStatus struct {
	Context *middleware.Context
	Handler PostAdmGrandfatherStatusHandler
}

func (o *PostAdmGrandfatherStatus) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewPostAdmGrandfatherStatusParams()

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
