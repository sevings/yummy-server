// Code generated by go-swagger; DO NOT EDIT.

package oauth2

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// PostOauth2TokenHandlerFunc turns a function with the right signature into a post oauth2 token handler
type PostOauth2TokenHandlerFunc func(PostOauth2TokenParams) middleware.Responder

// Handle executing the request and returning a response
func (fn PostOauth2TokenHandlerFunc) Handle(params PostOauth2TokenParams) middleware.Responder {
	return fn(params)
}

// PostOauth2TokenHandler interface for that can handle valid post oauth2 token params
type PostOauth2TokenHandler interface {
	Handle(PostOauth2TokenParams) middleware.Responder
}

// NewPostOauth2Token creates a new http.Handler for the post oauth2 token operation
func NewPostOauth2Token(ctx *middleware.Context, handler PostOauth2TokenHandler) *PostOauth2Token {
	return &PostOauth2Token{Context: ctx, Handler: handler}
}

/* PostOauth2Token swagger:route POST /oauth2/token oauth2 postOauth2Token

PostOauth2Token post oauth2 token API

*/
type PostOauth2Token struct {
	Context *middleware.Context
	Handler PostOauth2TokenHandler
}

func (o *PostOauth2Token) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewPostOauth2TokenParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
