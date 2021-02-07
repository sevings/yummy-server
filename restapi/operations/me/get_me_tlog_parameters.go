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

// NewGetMeTlogParams creates a new GetMeTlogParams object
// with the default values initialized.
func NewGetMeTlogParams() GetMeTlogParams {

	var (
		// initialize parameters with default values

		afterDefault  = string("")
		beforeDefault = string("")
		limitDefault  = int64(30)
		queryDefault  = string("")
		sortDefault   = string("new")
		tagDefault    = string("")
	)

	return GetMeTlogParams{
		After: &afterDefault,

		Before: &beforeDefault,

		Limit: &limitDefault,

		Query: &queryDefault,

		Sort: &sortDefault,

		Tag: &tagDefault,
	}
}

// GetMeTlogParams contains all the bound params for the get me tlog operation
// typically these are obtained from a http.Request
//
// swagger:parameters GetMeTlog
type GetMeTlogParams struct {

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
	/*
	  Max Length: 100
	  In: query
	  Default: ""
	*/
	Query *string
	/*
	  In: query
	  Default: "new"
	*/
	Sort *string
	/*
	  Max Length: 50
	  In: query
	  Default: ""
	*/
	Tag *string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetMeTlogParams() beforehand.
func (o *GetMeTlogParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
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

	qQuery, qhkQuery, _ := qs.GetOK("query")
	if err := o.bindQuery(qQuery, qhkQuery, route.Formats); err != nil {
		res = append(res, err)
	}

	qSort, qhkSort, _ := qs.GetOK("sort")
	if err := o.bindSort(qSort, qhkSort, route.Formats); err != nil {
		res = append(res, err)
	}

	qTag, qhkTag, _ := qs.GetOK("tag")
	if err := o.bindTag(qTag, qhkTag, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindAfter binds and validates parameter After from query.
func (o *GetMeTlogParams) bindAfter(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewGetMeTlogParams()
		return nil
	}
	o.After = &raw

	return nil
}

// bindBefore binds and validates parameter Before from query.
func (o *GetMeTlogParams) bindBefore(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewGetMeTlogParams()
		return nil
	}
	o.Before = &raw

	return nil
}

// bindLimit binds and validates parameter Limit from query.
func (o *GetMeTlogParams) bindLimit(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewGetMeTlogParams()
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
func (o *GetMeTlogParams) validateLimit(formats strfmt.Registry) error {

	if err := validate.MinimumInt("limit", "query", *o.Limit, 1, false); err != nil {
		return err
	}

	if err := validate.MaximumInt("limit", "query", *o.Limit, 100, false); err != nil {
		return err
	}

	return nil
}

// bindQuery binds and validates parameter Query from query.
func (o *GetMeTlogParams) bindQuery(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewGetMeTlogParams()
		return nil
	}
	o.Query = &raw

	if err := o.validateQuery(formats); err != nil {
		return err
	}

	return nil
}

// validateQuery carries on validations for parameter Query
func (o *GetMeTlogParams) validateQuery(formats strfmt.Registry) error {

	if err := validate.MaxLength("query", "query", *o.Query, 100); err != nil {
		return err
	}

	return nil
}

// bindSort binds and validates parameter Sort from query.
func (o *GetMeTlogParams) bindSort(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewGetMeTlogParams()
		return nil
	}
	o.Sort = &raw

	if err := o.validateSort(formats); err != nil {
		return err
	}

	return nil
}

// validateSort carries on validations for parameter Sort
func (o *GetMeTlogParams) validateSort(formats strfmt.Registry) error {

	if err := validate.EnumCase("sort", "query", *o.Sort, []interface{}{"new", "old", "best"}, true); err != nil {
		return err
	}

	return nil
}

// bindTag binds and validates parameter Tag from query.
func (o *GetMeTlogParams) bindTag(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewGetMeTlogParams()
		return nil
	}
	o.Tag = &raw

	if err := o.validateTag(formats); err != nil {
		return err
	}

	return nil
}

// validateTag carries on validations for parameter Tag
func (o *GetMeTlogParams) validateTag(formats strfmt.Registry) error {

	if err := validate.MaxLength("tag", "query", *o.Tag, 50); err != nil {
		return err
	}

	return nil
}
