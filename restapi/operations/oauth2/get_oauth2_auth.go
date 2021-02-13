// Code generated by go-swagger; DO NOT EDIT.

package oauth2

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"context"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/sevings/mindwell-server/models"
)

// GetOauth2AuthHandlerFunc turns a function with the right signature into a get oauth2 auth handler
type GetOauth2AuthHandlerFunc func(GetOauth2AuthParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetOauth2AuthHandlerFunc) Handle(params GetOauth2AuthParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetOauth2AuthHandler interface for that can handle valid get oauth2 auth params
type GetOauth2AuthHandler interface {
	Handle(GetOauth2AuthParams, *models.UserID) middleware.Responder
}

// NewGetOauth2Auth creates a new http.Handler for the get oauth2 auth operation
func NewGetOauth2Auth(ctx *middleware.Context, handler GetOauth2AuthHandler) *GetOauth2Auth {
	return &GetOauth2Auth{Context: ctx, Handler: handler}
}

/* GetOauth2Auth swagger:route GET /oauth2/auth oauth2 getOauth2Auth

GetOauth2Auth get oauth2 auth API

*/
type GetOauth2Auth struct {
	Context *middleware.Context
	Handler GetOauth2AuthHandler
}

func (o *GetOauth2Auth) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetOauth2AuthParams()
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

// GetOauth2AuthOKBody get oauth2 auth o k body
//
// swagger:model GetOauth2AuthOKBody
type GetOauth2AuthOKBody struct {

	// code
	Code string `json:"code,omitempty"`

	// state
	State string `json:"state,omitempty"`
}

// Validate validates this get oauth2 auth o k body
func (o *GetOauth2AuthOKBody) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this get oauth2 auth o k body based on context it is used
func (o *GetOauth2AuthOKBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *GetOauth2AuthOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *GetOauth2AuthOKBody) UnmarshalBinary(b []byte) error {
	var res GetOauth2AuthOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}
