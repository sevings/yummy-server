// Code generated by go-swagger; DO NOT EDIT.

package votes

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// NewPutEntriesIDVoteParams creates a new PutEntriesIDVoteParams object
// with the default values initialized.
func NewPutEntriesIDVoteParams() PutEntriesIDVoteParams {

	var (
		// initialize parameters with default values

		positiveDefault = bool(true)
	)

	return PutEntriesIDVoteParams{
		Positive: &positiveDefault,
	}
}

// PutEntriesIDVoteParams contains all the bound params for the put entries ID vote operation
// typically these are obtained from a http.Request
//
// swagger:parameters PutEntriesIDVote
type PutEntriesIDVoteParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  Minimum: 1
	  In: path
	*/
	ID int64
	/*
	  In: query
	  Default: true
	*/
	Positive *bool
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewPutEntriesIDVoteParams() beforehand.
func (o *PutEntriesIDVoteParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	rID, rhkID, _ := route.Params.GetOK("id")
	if err := o.bindID(rID, rhkID, route.Formats); err != nil {
		res = append(res, err)
	}

	qPositive, qhkPositive, _ := qs.GetOK("positive")
	if err := o.bindPositive(qPositive, qhkPositive, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindID binds and validates parameter ID from path.
func (o *PutEntriesIDVoteParams) bindID(rawData []string, hasKey bool, formats strfmt.Registry) error {
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
func (o *PutEntriesIDVoteParams) validateID(formats strfmt.Registry) error {

	if err := validate.MinimumInt("id", "path", int64(o.ID), 1, false); err != nil {
		return err
	}

	return nil
}

// bindPositive binds and validates parameter Positive from query.
func (o *PutEntriesIDVoteParams) bindPositive(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false
	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewPutEntriesIDVoteParams()
		return nil
	}

	value, err := swag.ConvertBool(raw)
	if err != nil {
		return errors.InvalidType("positive", "query", "bool", raw)
	}
	o.Positive = &value

	return nil
}
