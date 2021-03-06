// Code generated by go-swagger; DO NOT EDIT.

package design

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

// GetDesignFontsHandlerFunc turns a function with the right signature into a get design fonts handler
type GetDesignFontsHandlerFunc func(GetDesignFontsParams, *models.UserID) middleware.Responder

// Handle executing the request and returning a response
func (fn GetDesignFontsHandlerFunc) Handle(params GetDesignFontsParams, principal *models.UserID) middleware.Responder {
	return fn(params, principal)
}

// GetDesignFontsHandler interface for that can handle valid get design fonts params
type GetDesignFontsHandler interface {
	Handle(GetDesignFontsParams, *models.UserID) middleware.Responder
}

// NewGetDesignFonts creates a new http.Handler for the get design fonts operation
func NewGetDesignFonts(ctx *middleware.Context, handler GetDesignFontsHandler) *GetDesignFonts {
	return &GetDesignFonts{Context: ctx, Handler: handler}
}

/* GetDesignFonts swagger:route GET /design/fonts design getDesignFonts

GetDesignFonts get design fonts API

*/
type GetDesignFonts struct {
	Context *middleware.Context
	Handler GetDesignFontsHandler
}

func (o *GetDesignFonts) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetDesignFontsParams()
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

// GetDesignFontsOKBody get design fonts o k body
//
// swagger:model GetDesignFontsOKBody
type GetDesignFontsOKBody struct {

	// fonts
	Fonts []string `json:"fonts"`
}

// Validate validates this get design fonts o k body
func (o *GetDesignFontsOKBody) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this get design fonts o k body based on context it is used
func (o *GetDesignFontsOKBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *GetDesignFontsOKBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *GetDesignFontsOKBody) UnmarshalBinary(b []byte) error {
	var res GetDesignFontsOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}
