// Code generated by go-swagger; DO NOT EDIT.

package entries

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/mindwell-server/models"
)

// GetEntriesTagsHandlerFunc turns a function with the right signature into a get entries tags handler
type GetEntriesTagsHandlerFunc func(GetEntriesTagsParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetEntriesTagsHandlerFunc) Handle(params GetEntriesTagsParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetEntriesTagsHandler interface for that can handle valid get entries tags params
type GetEntriesTagsHandler interface {
	Handle(GetEntriesTagsParams, *models.UserID) middleware.Responder
}

// NewGetEntriesTags creates a new http.Handler for the get entries tags operation
func NewGetEntriesTags(ctx *middleware.Context, handler GetEntriesTagsHandler) *GetEntriesTags {
	return &GetEntriesTags{Context: ctx, Handler: handler}
}

/*GetEntriesTags swagger:route GET /entries/tags entries getEntriesTags

GetEntriesTags get entries tags API

*/
type GetEntriesTags struct {
	Context *middleware.Context
	Handler GetEntriesTagsHandler
}

func (o *GetEntriesTags) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetEntriesTagsParams()

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
