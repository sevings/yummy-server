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
	"github.com/go-openapi/swag"
)

// PutAccountSettingsTelegramMaxParseMemory sets the maximum size in bytes for
// the multipart form parser for this operation.
//
// The default value is 32 MB.
// The multipart parser stores up to this + 10MB.
var PutAccountSettingsTelegramMaxParseMemory int64 = 32 << 20

// NewPutAccountSettingsTelegramParams creates a new PutAccountSettingsTelegramParams object
// with the default values initialized.
func NewPutAccountSettingsTelegramParams() PutAccountSettingsTelegramParams {

	var (
		// initialize parameters with default values

		commentsDefault  = bool(false)
		followersDefault = bool(false)
		invitesDefault   = bool(false)
		messagesDefault  = bool(false)
	)

	return PutAccountSettingsTelegramParams{
		Comments: &commentsDefault,

		Followers: &followersDefault,

		Invites: &invitesDefault,

		Messages: &messagesDefault,
	}
}

// PutAccountSettingsTelegramParams contains all the bound params for the put account settings telegram operation
// typically these are obtained from a http.Request
//
// swagger:parameters PutAccountSettingsTelegram
type PutAccountSettingsTelegramParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  In: formData
	  Default: false
	*/
	Comments *bool
	/*
	  In: formData
	  Default: false
	*/
	Followers *bool
	/*
	  In: formData
	  Default: false
	*/
	Invites *bool
	/*
	  In: formData
	  Default: false
	*/
	Messages *bool
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewPutAccountSettingsTelegramParams() beforehand.
func (o *PutAccountSettingsTelegramParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if err := r.ParseMultipartForm(PutAccountSettingsTelegramMaxParseMemory); err != nil {
		if err != http.ErrNotMultipart {
			return errors.New(400, "%v", err)
		} else if err := r.ParseForm(); err != nil {
			return errors.New(400, "%v", err)
		}
	}
	fds := runtime.Values(r.Form)

	fdComments, fdhkComments, _ := fds.GetOK("comments")
	if err := o.bindComments(fdComments, fdhkComments, route.Formats); err != nil {
		res = append(res, err)
	}

	fdFollowers, fdhkFollowers, _ := fds.GetOK("followers")
	if err := o.bindFollowers(fdFollowers, fdhkFollowers, route.Formats); err != nil {
		res = append(res, err)
	}

	fdInvites, fdhkInvites, _ := fds.GetOK("invites")
	if err := o.bindInvites(fdInvites, fdhkInvites, route.Formats); err != nil {
		res = append(res, err)
	}

	fdMessages, fdhkMessages, _ := fds.GetOK("messages")
	if err := o.bindMessages(fdMessages, fdhkMessages, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindComments binds and validates parameter Comments from formData.
func (o *PutAccountSettingsTelegramParams) bindComments(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewPutAccountSettingsTelegramParams()
		return nil
	}

	value, err := swag.ConvertBool(raw)
	if err != nil {
		return errors.InvalidType("comments", "formData", "bool", raw)
	}
	o.Comments = &value

	return nil
}

// bindFollowers binds and validates parameter Followers from formData.
func (o *PutAccountSettingsTelegramParams) bindFollowers(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewPutAccountSettingsTelegramParams()
		return nil
	}

	value, err := swag.ConvertBool(raw)
	if err != nil {
		return errors.InvalidType("followers", "formData", "bool", raw)
	}
	o.Followers = &value

	return nil
}

// bindInvites binds and validates parameter Invites from formData.
func (o *PutAccountSettingsTelegramParams) bindInvites(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewPutAccountSettingsTelegramParams()
		return nil
	}

	value, err := swag.ConvertBool(raw)
	if err != nil {
		return errors.InvalidType("invites", "formData", "bool", raw)
	}
	o.Invites = &value

	return nil
}

// bindMessages binds and validates parameter Messages from formData.
func (o *PutAccountSettingsTelegramParams) bindMessages(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewPutAccountSettingsTelegramParams()
		return nil
	}

	value, err := swag.ConvertBool(raw)
	if err != nil {
		return errors.InvalidType("messages", "formData", "bool", raw)
	}
	o.Messages = &value

	return nil
}
