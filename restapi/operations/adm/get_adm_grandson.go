// Code generated by go-swagger; DO NOT EDIT.

package adm

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
	strfmt "github.com/go-openapi/strfmt"
	swag "github.com/go-openapi/swag"
	models "github.com/sevings/mindwell-server/models"
)

// GetAdmGrandsonHandlerFunc turns a function with the right signature into a get adm grandson handler
type GetAdmGrandsonHandlerFunc func(GetAdmGrandsonParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetAdmGrandsonHandlerFunc) Handle(params GetAdmGrandsonParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetAdmGrandsonHandler interface for that can handle valid get adm grandson params
type GetAdmGrandsonHandler interface {
	Handle(GetAdmGrandsonParams, *models.UserID) middleware.Responder
}

// NewGetAdmGrandson creates a new http.Handler for the get adm grandson operation
func NewGetAdmGrandson(ctx *middleware.Context, handler GetAdmGrandsonHandler) *GetAdmGrandson {
	return &GetAdmGrandson{Context: ctx, Handler: handler}
}

/*GetAdmGrandson swagger:route GET /adm/grandson adm getAdmGrandson

GetAdmGrandson get adm grandson API

*/
type GetAdmGrandson struct {
	Context *middleware.Context
	Handler GetAdmGrandsonHandler
}

func (o *GetAdmGrandson) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetAdmGrandsonParams()

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

// GetAdmGrandsonOKBody get adm grandson o k body
// swagger:model GetAdmGrandsonOKBody
type GetAdmGrandsonOKBody struct {

	// address
	Address string `json:"address,omitempty"`

	// anonymous
	Anonymous bool `json:"anonymous,omitempty"`

	// comment
	Comment string `json:"comment,omitempty"`

	// country
	Country string `json:"country,omitempty"`

	// name
	Name string `json:"name,omitempty"`

	// postcode
	Postcode string `json:"postcode,omitempty"`
}

// Validate validates this get adm grandson o k body
func (o *GetAdmGrandsonOKBody) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *GetAdmGrandsonOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *GetAdmGrandsonOKBody) UnmarshalBinary(b []byte) error {
	var res GetAdmGrandsonOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}