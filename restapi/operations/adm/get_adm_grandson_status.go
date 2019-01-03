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

// GetAdmGrandsonStatusHandlerFunc turns a function with the right signature into a get adm grandson status handler
type GetAdmGrandsonStatusHandlerFunc func(GetAdmGrandsonStatusParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetAdmGrandsonStatusHandlerFunc) Handle(params GetAdmGrandsonStatusParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetAdmGrandsonStatusHandler interface for that can handle valid get adm grandson status params
type GetAdmGrandsonStatusHandler interface {
	Handle(GetAdmGrandsonStatusParams, *models.UserID) middleware.Responder
}

// NewGetAdmGrandsonStatus creates a new http.Handler for the get adm grandson status operation
func NewGetAdmGrandsonStatus(ctx *middleware.Context, handler GetAdmGrandsonStatusHandler) *GetAdmGrandsonStatus {
	return &GetAdmGrandsonStatus{Context: ctx, Handler: handler}
}

/*GetAdmGrandsonStatus swagger:route GET /adm/grandson/status adm getAdmGrandsonStatus

GetAdmGrandsonStatus get adm grandson status API

*/
type GetAdmGrandsonStatus struct {
	Context *middleware.Context
	Handler GetAdmGrandsonStatusHandler
}

func (o *GetAdmGrandsonStatus) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetAdmGrandsonStatusParams()

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

// GetAdmGrandsonStatusOKBody get adm grandson status o k body
// swagger:model GetAdmGrandsonStatusOKBody
type GetAdmGrandsonStatusOKBody struct {

	// received
	Received bool `json:"received,omitempty"`

	// sent
	Sent bool `json:"sent,omitempty"`
}

// Validate validates this get adm grandson status o k body
func (o *GetAdmGrandsonStatusOKBody) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *GetAdmGrandsonStatusOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *GetAdmGrandsonStatusOKBody) UnmarshalBinary(b []byte) error {
	var res GetAdmGrandsonStatusOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}
