// Code generated by go-swagger; DO NOT EDIT.

package account

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// PostAccountRecoverPasswordHandlerFunc turns a function with the right signature into a post account recover password handler
type PostAccountRecoverPasswordHandlerFunc func(PostAccountRecoverPasswordParams) middleware.Responder

// Handle executing the request and returning a response
func (fn PostAccountRecoverPasswordHandlerFunc) Handle(params PostAccountRecoverPasswordParams) middleware.Responder {
	return fn(params)
}

// PostAccountRecoverPasswordHandler interface for that can handle valid post account recover password params
type PostAccountRecoverPasswordHandler interface {
	Handle(PostAccountRecoverPasswordParams) middleware.Responder
}

// NewPostAccountRecoverPassword creates a new http.Handler for the post account recover password operation
func NewPostAccountRecoverPassword(ctx *middleware.Context, handler PostAccountRecoverPasswordHandler) *PostAccountRecoverPassword {
	return &PostAccountRecoverPassword{Context: ctx, Handler: handler}
}

/* PostAccountRecoverPassword swagger:route POST /account/recover/password account postAccountRecoverPassword

reset password

*/
type PostAccountRecoverPassword struct {
	Context *middleware.Context
	Handler PostAccountRecoverPasswordHandler
}

func (o *PostAccountRecoverPassword) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewPostAccountRecoverPasswordParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
