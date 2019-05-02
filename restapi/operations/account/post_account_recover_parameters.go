// Code generated by go-swagger; DO NOT EDIT.

package account

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/validate"

	strfmt "github.com/go-openapi/strfmt"
)

// NewPostAccountRecoverParams creates a new PostAccountRecoverParams object
// no default values defined in spec.
func NewPostAccountRecoverParams() PostAccountRecoverParams {

	return PostAccountRecoverParams{}
}

// PostAccountRecoverParams contains all the bound params for the post account recover operation
// typically these are obtained from a http.Request
//
// swagger:parameters PostAccountRecover
type PostAccountRecoverParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  Max Length: 500
	  Pattern: .+@.+
	  In: formData
	*/
	Email string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewPostAccountRecoverParams() beforehand.
func (o *PostAccountRecoverParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if err != http.ErrNotMultipart {
			return errors.New(400, "%v", err)
		} else if err := r.ParseForm(); err != nil {
			return errors.New(400, "%v", err)
		}
	}
	fds := runtime.Values(r.Form)

	fdEmail, fdhkEmail, _ := fds.GetOK("email")
	if err := o.bindEmail(fdEmail, fdhkEmail, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindEmail binds and validates parameter Email from formData.
func (o *PostAccountRecoverParams) bindEmail(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("email", "formData")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true

	if err := validate.RequiredString("email", "formData", raw); err != nil {
		return err
	}

	o.Email = raw

	if err := o.validateEmail(formats); err != nil {
		return err
	}

	return nil
}

// validateEmail carries on validations for parameter Email
func (o *PostAccountRecoverParams) validateEmail(formats strfmt.Registry) error {

	if err := validate.MaxLength("email", "formData", o.Email, 500); err != nil {
		return err
	}

	if err := validate.Pattern("email", "formData", o.Email, `.+@.+`); err != nil {
		return err
	}

	return nil
}
