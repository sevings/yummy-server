// Code generated by go-swagger; DO NOT EDIT.

package account

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// NewGetAccountNameNameParams creates a new GetAccountNameNameParams object
// no default values defined in spec.
func NewGetAccountNameNameParams() GetAccountNameNameParams {

	return GetAccountNameNameParams{}
}

// GetAccountNameNameParams contains all the bound params for the get account name name operation
// typically these are obtained from a http.Request
//
// swagger:parameters GetAccountNameName
type GetAccountNameNameParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  Max Length: 20
	  Min Length: 1
	  In: path
	*/
	Name string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetAccountNameNameParams() beforehand.
func (o *GetAccountNameNameParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rName, rhkName, _ := route.Params.GetOK("name")
	if err := o.bindName(rName, rhkName, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindName binds and validates parameter Name from path.
func (o *GetAccountNameNameParams) bindName(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	o.Name = raw

	if err := o.validateName(formats); err != nil {
		return err
	}

	return nil
}

// validateName carries on validations for parameter Name
func (o *GetAccountNameNameParams) validateName(formats strfmt.Registry) error {

	if err := validate.MinLength("name", "path", o.Name, 1); err != nil {
		return err
	}

	if err := validate.MaxLength("name", "path", o.Name, 20); err != nil {
		return err
	}

	return nil
}
