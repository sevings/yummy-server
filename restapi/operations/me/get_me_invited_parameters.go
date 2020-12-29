// Code generated by go-swagger; DO NOT EDIT.

package me

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// NewGetMeInvitedParams creates a new GetMeInvitedParams object
// with the default values initialized.
func NewGetMeInvitedParams() GetMeInvitedParams {

	var (
		// initialize parameters with default values

		afterDefault  = string("")
		beforeDefault = string("")
		limitDefault  = int64(30)
	)

	return GetMeInvitedParams{
		After: &afterDefault,

		Before: &beforeDefault,

		Limit: &limitDefault,
	}
}

// GetMeInvitedParams contains all the bound params for the get me invited operation
// typically these are obtained from a http.Request
//
// swagger:parameters GetMeInvited
type GetMeInvitedParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  In: query
	  Default: ""
	*/
	After *string
	/*
	  In: query
	  Default: ""
	*/
	Before *string
	/*
	  Maximum: 100
	  Minimum: 1
	  In: query
	  Default: 30
	*/
	Limit *int64
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetMeInvitedParams() beforehand.
func (o *GetMeInvitedParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	qAfter, qhkAfter, _ := qs.GetOK("after")
	if err := o.bindAfter(qAfter, qhkAfter, route.Formats); err != nil {
		res = append(res, err)
	}

	qBefore, qhkBefore, _ := qs.GetOK("before")
	if err := o.bindBefore(qBefore, qhkBefore, route.Formats); err != nil {
		res = append(res, err)
	}

	qLimit, qhkLimit, _ := qs.GetOK("limit")
	if err := o.bindLimit(qLimit, qhkLimit, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindAfter binds and validates parameter After from query.
func (o *GetMeInvitedParams) bindAfter(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false
	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewGetMeInvitedParams()
		return nil
	}

	o.After = &raw

	return nil
}

// bindBefore binds and validates parameter Before from query.
func (o *GetMeInvitedParams) bindBefore(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false
	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewGetMeInvitedParams()
		return nil
	}

	o.Before = &raw

	return nil
}

// bindLimit binds and validates parameter Limit from query.
func (o *GetMeInvitedParams) bindLimit(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false
	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewGetMeInvitedParams()
		return nil
	}

	value, err := swag.ConvertInt64(raw)
	if err != nil {
		return errors.InvalidType("limit", "query", "int64", raw)
	}
	o.Limit = &value

	if err := o.validateLimit(formats); err != nil {
		return err
	}

	return nil
}

// validateLimit carries on validations for parameter Limit
func (o *GetMeInvitedParams) validateLimit(formats strfmt.Registry) error {

	if err := validate.MinimumInt("limit", "query", int64(*o.Limit), 1, false); err != nil {
		return err
	}

	if err := validate.MaximumInt("limit", "query", int64(*o.Limit), 100, false); err != nil {
		return err
	}

	return nil
}
