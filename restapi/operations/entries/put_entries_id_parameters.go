// Code generated by go-swagger; DO NOT EDIT.

package entries

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"

	strfmt "github.com/go-openapi/strfmt"
)

// NewPutEntriesIDParams creates a new PutEntriesIDParams object
// with the default values initialized.
func NewPutEntriesIDParams() PutEntriesIDParams {

	var (
		// initialize parameters with default values

		anonymousCommentsDefault = bool(false)

		inLiveDefault    = bool(false)
		isVotableDefault = bool(false)

		titleDefault = string("")
	)

	return PutEntriesIDParams{
		AnonymousComments: &anonymousCommentsDefault,

		InLive: &inLiveDefault,

		IsVotable: &isVotableDefault,

		Title: &titleDefault,
	}
}

// PutEntriesIDParams contains all the bound params for the put entries ID operation
// typically these are obtained from a http.Request
//
// swagger:parameters PutEntriesID
type PutEntriesIDParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  In: formData
	  Default: false
	*/
	AnonymousComments *bool
	/*
	  Required: true
	  Max Length: 30000
	  Min Length: 1
	  Pattern: \s*\S+.*
	  In: formData
	*/
	Content string
	/*
	  Required: true
	  Minimum: 1
	  In: path
	*/
	ID int64
	/*
	  In: formData
	  Default: false
	*/
	InLive *bool
	/*
	  In: formData
	  Default: false
	*/
	IsVotable *bool
	/*
	  Required: true
	  In: formData
	*/
	Privacy string
	/*
	  Max Length: 500
	  In: formData
	  Default: ""
	*/
	Title *string
	/*
	  In: formData
	*/
	VisibleFor []int64
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewPutEntriesIDParams() beforehand.
func (o *PutEntriesIDParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
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

	fdAnonymousComments, fdhkAnonymousComments, _ := fds.GetOK("anonymous_comments")
	if err := o.bindAnonymousComments(fdAnonymousComments, fdhkAnonymousComments, route.Formats); err != nil {
		res = append(res, err)
	}

	fdContent, fdhkContent, _ := fds.GetOK("content")
	if err := o.bindContent(fdContent, fdhkContent, route.Formats); err != nil {
		res = append(res, err)
	}

	rID, rhkID, _ := route.Params.GetOK("id")
	if err := o.bindID(rID, rhkID, route.Formats); err != nil {
		res = append(res, err)
	}

	fdInLive, fdhkInLive, _ := fds.GetOK("inLive")
	if err := o.bindInLive(fdInLive, fdhkInLive, route.Formats); err != nil {
		res = append(res, err)
	}

	fdIsVotable, fdhkIsVotable, _ := fds.GetOK("isVotable")
	if err := o.bindIsVotable(fdIsVotable, fdhkIsVotable, route.Formats); err != nil {
		res = append(res, err)
	}

	fdPrivacy, fdhkPrivacy, _ := fds.GetOK("privacy")
	if err := o.bindPrivacy(fdPrivacy, fdhkPrivacy, route.Formats); err != nil {
		res = append(res, err)
	}

	fdTitle, fdhkTitle, _ := fds.GetOK("title")
	if err := o.bindTitle(fdTitle, fdhkTitle, route.Formats); err != nil {
		res = append(res, err)
	}

	fdVisibleFor, fdhkVisibleFor, _ := fds.GetOK("visibleFor")
	if err := o.bindVisibleFor(fdVisibleFor, fdhkVisibleFor, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindAnonymousComments binds and validates parameter AnonymousComments from formData.
func (o *PutEntriesIDParams) bindAnonymousComments(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewPutEntriesIDParams()
		return nil
	}

	value, err := swag.ConvertBool(raw)
	if err != nil {
		return errors.InvalidType("anonymous_comments", "formData", "bool", raw)
	}
	o.AnonymousComments = &value

	return nil
}

// bindContent binds and validates parameter Content from formData.
func (o *PutEntriesIDParams) bindContent(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("content", "formData")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true

	if err := validate.RequiredString("content", "formData", raw); err != nil {
		return err
	}

	o.Content = raw

	if err := o.validateContent(formats); err != nil {
		return err
	}

	return nil
}

// validateContent carries on validations for parameter Content
func (o *PutEntriesIDParams) validateContent(formats strfmt.Registry) error {

	if err := validate.MinLength("content", "formData", o.Content, 1); err != nil {
		return err
	}

	if err := validate.MaxLength("content", "formData", o.Content, 30000); err != nil {
		return err
	}

	if err := validate.Pattern("content", "formData", o.Content, `\s*\S+.*`); err != nil {
		return err
	}

	return nil
}

// bindID binds and validates parameter ID from path.
func (o *PutEntriesIDParams) bindID(rawData []string, hasKey bool, formats strfmt.Registry) error {
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
func (o *PutEntriesIDParams) validateID(formats strfmt.Registry) error {

	if err := validate.MinimumInt("id", "path", int64(o.ID), 1, false); err != nil {
		return err
	}

	return nil
}

// bindInLive binds and validates parameter InLive from formData.
func (o *PutEntriesIDParams) bindInLive(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewPutEntriesIDParams()
		return nil
	}

	value, err := swag.ConvertBool(raw)
	if err != nil {
		return errors.InvalidType("inLive", "formData", "bool", raw)
	}
	o.InLive = &value

	return nil
}

// bindIsVotable binds and validates parameter IsVotable from formData.
func (o *PutEntriesIDParams) bindIsVotable(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewPutEntriesIDParams()
		return nil
	}

	value, err := swag.ConvertBool(raw)
	if err != nil {
		return errors.InvalidType("isVotable", "formData", "bool", raw)
	}
	o.IsVotable = &value

	return nil
}

// bindPrivacy binds and validates parameter Privacy from formData.
func (o *PutEntriesIDParams) bindPrivacy(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("privacy", "formData")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true

	if err := validate.RequiredString("privacy", "formData", raw); err != nil {
		return err
	}

	o.Privacy = raw

	if err := o.validatePrivacy(formats); err != nil {
		return err
	}

	return nil
}

// validatePrivacy carries on validations for parameter Privacy
func (o *PutEntriesIDParams) validatePrivacy(formats strfmt.Registry) error {

	if err := validate.Enum("privacy", "formData", o.Privacy, []interface{}{"all", "followers", "some", "me", "anonymous"}); err != nil {
		return err
	}

	return nil
}

// bindTitle binds and validates parameter Title from formData.
func (o *PutEntriesIDParams) bindTitle(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		// Default values have been previously initialized by NewPutEntriesIDParams()
		return nil
	}

	o.Title = &raw

	if err := o.validateTitle(formats); err != nil {
		return err
	}

	return nil
}

// validateTitle carries on validations for parameter Title
func (o *PutEntriesIDParams) validateTitle(formats strfmt.Registry) error {

	if err := validate.MaxLength("title", "formData", (*o.Title), 500); err != nil {
		return err
	}

	return nil
}

// bindVisibleFor binds and validates array parameter VisibleFor from formData.
//
// Arrays are parsed according to CollectionFormat: "" (defaults to "csv" when empty).
func (o *PutEntriesIDParams) bindVisibleFor(rawData []string, hasKey bool, formats strfmt.Registry) error {

	var qvVisibleFor string
	if len(rawData) > 0 {
		qvVisibleFor = rawData[len(rawData)-1]
	}

	// CollectionFormat:
	visibleForIC := swag.SplitByFormat(qvVisibleFor, "")
	if len(visibleForIC) == 0 {
		return nil
	}

	var visibleForIR []int64
	for i, visibleForIV := range visibleForIC {
		// items.Format: "int64"
		visibleForI, err := swag.ConvertInt64(visibleForIV)
		if err != nil {
			return errors.InvalidType(fmt.Sprintf("%s.%v", "visibleFor", i), "formData", "int64", visibleForI)
		}

		if err := validate.MinimumInt(fmt.Sprintf("%s.%v", "visibleFor", i), "formData", int64(visibleForI), 1, false); err != nil {
			return err
		}

		visibleForIR = append(visibleForIR, visibleForI)
	}

	o.VisibleFor = visibleForIR

	return nil
}
