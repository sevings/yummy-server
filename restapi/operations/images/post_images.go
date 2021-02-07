// Code generated by go-swagger; DO NOT EDIT.

package images

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/mindwell-server/models"
)

// PostImagesHandlerFunc turns a function with the right signature into a post images handler
type PostImagesHandlerFunc func(PostImagesParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn PostImagesHandlerFunc) Handle(params PostImagesParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// PostImagesHandler interface for that can handle valid post images params
type PostImagesHandler interface {
	Handle(PostImagesParams, *models.UserID) middleware.Responder
}

// NewPostImages creates a new http.Handler for the post images operation
func NewPostImages(ctx *middleware.Context, handler PostImagesHandler) *PostImages {
	return &PostImages{Context: ctx, Handler: handler}
}

/* PostImages swagger:route POST /images images postImages

PostImages post images API

*/
type PostImages struct {
	Context *middleware.Context
	Handler PostImagesHandler
}

func (o *PostImages) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewPostImagesParams()
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
