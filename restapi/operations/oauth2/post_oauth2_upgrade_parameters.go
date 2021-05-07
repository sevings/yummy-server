// Code generated by go-swagger; DO NOT EDIT.

package oauth2

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

// PostOauth2UpgradeMaxParseMemory sets the maximum size in bytes for
// the multipart form parser for this operation.
//
// The default value is 32 MB.
// The multipart parser stores up to this + 10MB.
var PostOauth2UpgradeMaxParseMemory int64 = 32 << 20

// NewPostOauth2UpgradeParams creates a new PostOauth2UpgradeParams object
//
// There are no default values defined in the spec.
func NewPostOauth2UpgradeParams() PostOauth2UpgradeParams {

	return PostOauth2UpgradeParams{}
}

// PostOauth2UpgradeParams contains all the bound params for the post oauth2 upgrade operation
// typically these are obtained from a http.Request
//
// swagger:parameters PostOauth2Upgrade
type PostOauth2UpgradeParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  In: formData
	*/
	ClientID int64
	/*
	  Required: true
	  Max Length: 64
	  In: formData
	*/
	ClientSecret string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewPostOauth2UpgradeParams() beforehand.
func (o *PostOauth2UpgradeParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if err := r.ParseMultipartForm(PostOauth2UpgradeMaxParseMemory); err != nil {
		if err != http.ErrNotMultipart {
			return errors.New(400, "%v", err)
		} else if err := r.ParseForm(); err != nil {
			return errors.New(400, "%v", err)
		}
	}
	fds := runtime.Values(r.Form)

	fdClientID, fdhkClientID, _ := fds.GetOK("client_id")
	if err := o.bindClientID(fdClientID, fdhkClientID, route.Formats); err != nil {
		res = append(res, err)
	}

	fdClientSecret, fdhkClientSecret, _ := fds.GetOK("client_secret")
	if err := o.bindClientSecret(fdClientSecret, fdhkClientSecret, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindClientID binds and validates parameter ClientID from formData.
func (o *PostOauth2UpgradeParams) bindClientID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("client_id", "formData", rawData)
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true

	if err := validate.RequiredString("client_id", "formData", raw); err != nil {
		return err
	}

	value, err := swag.ConvertInt64(raw)
	if err != nil {
		return errors.InvalidType("client_id", "formData", "int64", raw)
	}
	o.ClientID = value

	return nil
}

// bindClientSecret binds and validates parameter ClientSecret from formData.
func (o *PostOauth2UpgradeParams) bindClientSecret(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("client_secret", "formData", rawData)
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true

	if err := validate.RequiredString("client_secret", "formData", raw); err != nil {
		return err
	}
	o.ClientSecret = raw

	if err := o.validateClientSecret(formats); err != nil {
		return err
	}

	return nil
}

// validateClientSecret carries on validations for parameter ClientSecret
func (o *PostOauth2UpgradeParams) validateClientSecret(formats strfmt.Registry) error {

	if err := validate.MaxLength("client_secret", "formData", o.ClientSecret, 64); err != nil {
		return err
	}

	return nil
}