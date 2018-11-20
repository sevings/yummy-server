// Code generated by go-swagger; DO NOT EDIT.

package account

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// NewPostAccountLoginParams creates a new PostAccountLoginParams object
// no default values defined in spec.
func NewPostAccountLoginParams() PostAccountLoginParams {

	return PostAccountLoginParams{}
}

// PostAccountLoginParams contains all the bound params for the post account login operation
// typically these are obtained from a http.Request
//
// swagger:parameters PostAccountLogin
type PostAccountLoginParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  Max Length: 500
	  Min Length: 1
	  In: formData
	*/
	Name string
	/*
	  Required: true
	  Max Length: 500
	  Min Length: 6
	  In: formData
	*/
	Password string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewPostAccountLoginParams() beforehand.
func (o *PostAccountLoginParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
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

	fdName, fdhkName, _ := fds.GetOK("name")
	if err := o.bindName(fdName, fdhkName, route.Formats); err != nil {
		res = append(res, err)
	}

	fdPassword, fdhkPassword, _ := fds.GetOK("password")
	if err := o.bindPassword(fdPassword, fdhkPassword, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindName binds and validates parameter Name from formData.
func (o *PostAccountLoginParams) bindName(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("name", "formData")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true

	if err := validate.RequiredString("name", "formData", raw); err != nil {
		return err
	}

	o.Name = raw

	if err := o.validateName(formats); err != nil {
		return err
	}

	return nil
}

// validateName carries on validations for parameter Name
func (o *PostAccountLoginParams) validateName(formats strfmt.Registry) error {

	if err := validate.MinLength("name", "formData", o.Name, 1); err != nil {
		return err
	}

	if err := validate.MaxLength("name", "formData", o.Name, 500); err != nil {
		return err
	}

	return nil
}

// bindPassword binds and validates parameter Password from formData.
func (o *PostAccountLoginParams) bindPassword(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("password", "formData")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true

	if err := validate.RequiredString("password", "formData", raw); err != nil {
		return err
	}

	o.Password = raw

	if err := o.validatePassword(formats); err != nil {
		return err
	}

	return nil
}

// validatePassword carries on validations for parameter Password
func (o *PostAccountLoginParams) validatePassword(formats strfmt.Registry) error {

	if err := validate.MinLength("password", "formData", o.Password, 6); err != nil {
		return err
	}

	if err := validate.MaxLength("password", "formData", o.Password, 500); err != nil {
		return err
	}

	return nil
}
