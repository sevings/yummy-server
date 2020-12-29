// Code generated by go-swagger; DO NOT EDIT.

package account

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// PostAccountRegisterHandlerFunc turns a function with the right signature into a post account register handler
type PostAccountRegisterHandlerFunc func(PostAccountRegisterParams) middleware.Responder

// Handle executing the request and returning a response
func (fn PostAccountRegisterHandlerFunc) Handle(params PostAccountRegisterParams) middleware.Responder {
	return fn(params)
}

// PostAccountRegisterHandler interface for that can handle valid post account register params
type PostAccountRegisterHandler interface {
	Handle(PostAccountRegisterParams) middleware.Responder
}

// NewPostAccountRegister creates a new http.Handler for the post account register operation
func NewPostAccountRegister(ctx *middleware.Context, handler PostAccountRegisterHandler) *PostAccountRegister {
	return &PostAccountRegister{Context: ctx, Handler: handler}
}

/*PostAccountRegister swagger:route POST /account/register account postAccountRegister

register new account

*/
type PostAccountRegister struct {
	Context *middleware.Context
	Handler PostAccountRegisterHandler
}

func (o *PostAccountRegister) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewPostAccountRegisterParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
