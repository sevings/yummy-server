// Code generated by go-swagger; DO NOT EDIT.

package watchings

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"

	strfmt "github.com/go-openapi/strfmt"
)

// NewGetEntriesIDWatchingParams creates a new GetEntriesIDWatchingParams object
// no default values defined in spec.
func NewGetEntriesIDWatchingParams() GetEntriesIDWatchingParams {

	return GetEntriesIDWatchingParams{}
}

// GetEntriesIDWatchingParams contains all the bound params for the get entries ID watching operation
// typically these are obtained from a http.Request
//
// swagger:parameters GetEntriesIDWatching
type GetEntriesIDWatchingParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  Minimum: 1
	  In: path
	*/
	ID int64
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetEntriesIDWatchingParams() beforehand.
func (o *GetEntriesIDWatchingParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rID, rhkID, _ := route.Params.GetOK("id")
	if err := o.bindID(rID, rhkID, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindID binds and validates parameter ID from path.
func (o *GetEntriesIDWatchingParams) bindID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	value, err := swag.ConvertInt64(raw)
	if err != nil {
		return errors.InvalidType("id", "path", "int64", raw)
	}
	o.ID = value

	if err := o.validateID(formats); err != nil {
		return err
	}

	return nil
}

// validateID carries on validations for parameter ID
func (o *GetEntriesIDWatchingParams) validateID(formats strfmt.Registry) error {

	if err := validate.MinimumInt("id", "path", int64(o.ID), 1, false); err != nil {
		return err
	}

	return nil
}
