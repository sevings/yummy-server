// Code generated by go-swagger; DO NOT EDIT.

package design

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
	models "github.com/sevings/mindwell-server/models"
)

// GetDesignHandlerFunc turns a function with the right signature into a get design handler
type GetDesignHandlerFunc func(GetDesignParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetDesignHandlerFunc) Handle(params GetDesignParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetDesignHandler interface for that can handle valid get design params
type GetDesignHandler interface {
	Handle(GetDesignParams, *models.UserID) middleware.Responder
}

// NewGetDesign creates a new http.Handler for the get design operation
func NewGetDesign(ctx *middleware.Context, handler GetDesignHandler) *GetDesign {
	return &GetDesign{Context: ctx, Handler: handler}
}

/*GetDesign swagger:route GET /design design getDesign

GetDesign get design API

*/
type GetDesign struct {
	Context *middleware.Context
	Handler GetDesignHandler
}

func (o *GetDesign) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetDesignParams()

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
