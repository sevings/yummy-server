// Code generated by go-swagger; DO NOT EDIT.

package account

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
	strfmt "github.com/go-openapi/strfmt"
	swag "github.com/go-openapi/swag"

	models "github.com/sevings/mindwell-server/models"
)

// GetAccountSettingsEmailHandlerFunc turns a function with the right signature into a get account settings email handler
type GetAccountSettingsEmailHandlerFunc func(GetAccountSettingsEmailParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetAccountSettingsEmailHandlerFunc) Handle(params GetAccountSettingsEmailParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetAccountSettingsEmailHandler interface for that can handle valid get account settings email params
type GetAccountSettingsEmailHandler interface {
	Handle(GetAccountSettingsEmailParams, *models.UserID) middleware.Responder
}

// NewGetAccountSettingsEmail creates a new http.Handler for the get account settings email operation
func NewGetAccountSettingsEmail(ctx *middleware.Context, handler GetAccountSettingsEmailHandler) *GetAccountSettingsEmail {
	return &GetAccountSettingsEmail{Context: ctx, Handler: handler}
}

/*GetAccountSettingsEmail swagger:route GET /account/settings/email account getAccountSettingsEmail

GetAccountSettingsEmail get account settings email API

*/
type GetAccountSettingsEmail struct {
	Context *middleware.Context
	Handler GetAccountSettingsEmailHandler
}

func (o *GetAccountSettingsEmail) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetAccountSettingsEmailParams()

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

// GetAccountSettingsEmailOKBody get account settings email o k body
// swagger:model GetAccountSettingsEmailOKBody
type GetAccountSettingsEmailOKBody struct {

	// comments
	Comments bool `json:"comments,omitempty"`

	// followers
	Followers bool `json:"followers,omitempty"`
}

// Validate validates this get account settings email o k body
func (o *GetAccountSettingsEmailOKBody) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *GetAccountSettingsEmailOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *GetAccountSettingsEmailOKBody) UnmarshalBinary(b []byte) error {
	var res GetAccountSettingsEmailOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}
