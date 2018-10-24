// Code generated by go-swagger; DO NOT EDIT.

package adm

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
)

// NewGetAdmStatParams creates a new GetAdmStatParams object
// no default values defined in spec.
func NewGetAdmStatParams() GetAdmStatParams {

	return GetAdmStatParams{}
}

// GetAdmStatParams contains all the bound params for the get adm stat operation
// typically these are obtained from a http.Request
//
// swagger:parameters GetAdmStat
type GetAdmStatParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetAdmStatParams() beforehand.
func (o *GetAdmStatParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}