// Code generated by go-swagger; DO NOT EDIT.

package account

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// PostAccountPasswordMaxParseMemory sets the maximum size in bytes for
// the multipart form parser for this operation.
//
// The default value is 32 MB.
// The multipart parser stores up to this + 10MB.
var PostAccountPasswordMaxParseMemory int64 = 32 << 20

// NewPostAccountPasswordParams creates a new PostAccountPasswordParams object
//
// There are no default values defined in the spec.
func NewPostAccountPasswordParams() PostAccountPasswordParams {

	return PostAccountPasswordParams{}
}

// PostAccountPasswordParams contains all the bound params for the post account password operation
// typically these are obtained from a http.Request
//
// swagger:parameters PostAccountPassword
type PostAccountPasswordParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  Max Length: 100
	  Min Length: 6
	  In: formData
	*/
	NewPassword string
	/*
	  Required: true
	  Max Length: 100
	  Min Length: 6
	  In: formData
	*/
	OldPassword string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewPostAccountPasswordParams() beforehand.
func (o *PostAccountPasswordParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if err := r.ParseMultipartForm(PostAccountPasswordMaxParseMemory); err != nil {
		if err != http.ErrNotMultipart {
			return errors.New(400, "%v", err)
		} else if err := r.ParseForm(); err != nil {
			return errors.New(400, "%v", err)
		}
	}
	fds := runtime.Values(r.Form)

	fdNewPassword, fdhkNewPassword, _ := fds.GetOK("new_password")
	if err := o.bindNewPassword(fdNewPassword, fdhkNewPassword, route.Formats); err != nil {
		res = append(res, err)
	}

	fdOldPassword, fdhkOldPassword, _ := fds.GetOK("old_password")
	if err := o.bindOldPassword(fdOldPassword, fdhkOldPassword, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindNewPassword binds and validates parameter NewPassword from formData.
func (o *PostAccountPasswordParams) bindNewPassword(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("new_password", "formData", rawData)
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true

	if err := validate.RequiredString("new_password", "formData", raw); err != nil {
		return err
	}
	o.NewPassword = raw

	if err := o.validateNewPassword(formats); err != nil {
		return err
	}

	return nil
}

// validateNewPassword carries on validations for parameter NewPassword
func (o *PostAccountPasswordParams) validateNewPassword(formats strfmt.Registry) error {

	if err := validate.MinLength("new_password", "formData", o.NewPassword, 6); err != nil {
		return err
	}

	if err := validate.MaxLength("new_password", "formData", o.NewPassword, 100); err != nil {
		return err
	}

	return nil
}

// bindOldPassword binds and validates parameter OldPassword from formData.
func (o *PostAccountPasswordParams) bindOldPassword(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("old_password", "formData", rawData)
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true

	if err := validate.RequiredString("old_password", "formData", raw); err != nil {
		return err
	}
	o.OldPassword = raw

	if err := o.validateOldPassword(formats); err != nil {
		return err
	}

	return nil
}

// validateOldPassword carries on validations for parameter OldPassword
func (o *PostAccountPasswordParams) validateOldPassword(formats strfmt.Registry) error {

	if err := validate.MinLength("old_password", "formData", o.OldPassword, 6); err != nil {
		return err
	}

	if err := validate.MaxLength("old_password", "formData", o.OldPassword, 100); err != nil {
		return err
	}

	return nil
}
