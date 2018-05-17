// Code generated by go-swagger; DO NOT EDIT.

package me

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"

	strfmt "github.com/go-openapi/strfmt"
)

// NewPutUsersMeParams creates a new PutUsersMeParams object
// with the default values initialized.
func NewPutUsersMeParams() PutUsersMeParams {
	var (
		birthdayDefault   = string("")
		cityDefault       = string("")
		countryDefault    = string("")
		genderDefault     = string("not set")
		isDaylogDefault   = bool(false)
		showInTopsDefault = bool(false)
		titleDefault      = string("")
	)
	return PutUsersMeParams{
		Birthday: &birthdayDefault,

		City: &cityDefault,

		Country: &countryDefault,

		Gender: &genderDefault,

		IsDaylog: &isDaylogDefault,

		ShowInTops: &showInTopsDefault,

		Title: &titleDefault,
	}
}

// PutUsersMeParams contains all the bound params for the put users me operation
// typically these are obtained from a http.Request
//
// swagger:parameters PutUsersMe
type PutUsersMeParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  In: formData
	  Default: ""
	*/
	Birthday *string
	/*
	  Max Length: 50
	  In: formData
	  Default: ""
	*/
	City *string
	/*
	  Max Length: 50
	  In: formData
	  Default: ""
	*/
	Country *string
	/*
	  In: formData
	  Default: "not set"
	*/
	Gender *string
	/*
	  In: formData
	  Default: false
	*/
	IsDaylog *bool
	/*
	  Required: true
	  In: formData
	*/
	Privacy string
	/*
	  In: formData
	  Default: false
	*/
	ShowInTops *bool
	/*
	  Required: true
	  Max Length: 20
	  Min Length: 1
	  In: formData
	*/
	ShowName string
	/*
	  Max Length: 500
	  In: formData
	  Default: ""
	*/
	Title *string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls
func (o *PutUsersMeParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error
	o.HTTPRequest = r

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if err != http.ErrNotMultipart {
			return err
		} else if err := r.ParseForm(); err != nil {
			return err
		}
	}
	fds := runtime.Values(r.Form)

	fdBirthday, fdhkBirthday, _ := fds.GetOK("birthday")
	if err := o.bindBirthday(fdBirthday, fdhkBirthday, route.Formats); err != nil {
		res = append(res, err)
	}

	fdCity, fdhkCity, _ := fds.GetOK("city")
	if err := o.bindCity(fdCity, fdhkCity, route.Formats); err != nil {
		res = append(res, err)
	}

	fdCountry, fdhkCountry, _ := fds.GetOK("country")
	if err := o.bindCountry(fdCountry, fdhkCountry, route.Formats); err != nil {
		res = append(res, err)
	}

	fdGender, fdhkGender, _ := fds.GetOK("gender")
	if err := o.bindGender(fdGender, fdhkGender, route.Formats); err != nil {
		res = append(res, err)
	}

	fdIsDaylog, fdhkIsDaylog, _ := fds.GetOK("isDaylog")
	if err := o.bindIsDaylog(fdIsDaylog, fdhkIsDaylog, route.Formats); err != nil {
		res = append(res, err)
	}

	fdPrivacy, fdhkPrivacy, _ := fds.GetOK("privacy")
	if err := o.bindPrivacy(fdPrivacy, fdhkPrivacy, route.Formats); err != nil {
		res = append(res, err)
	}

	fdShowInTops, fdhkShowInTops, _ := fds.GetOK("showInTops")
	if err := o.bindShowInTops(fdShowInTops, fdhkShowInTops, route.Formats); err != nil {
		res = append(res, err)
	}

	fdShowName, fdhkShowName, _ := fds.GetOK("showName")
	if err := o.bindShowName(fdShowName, fdhkShowName, route.Formats); err != nil {
		res = append(res, err)
	}

	fdTitle, fdhkTitle, _ := fds.GetOK("title")
	if err := o.bindTitle(fdTitle, fdhkTitle, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *PutUsersMeParams) bindBirthday(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}
	if raw == "" { // empty values pass all other validations
		var birthdayDefault string = string("")
		o.Birthday = &birthdayDefault
		return nil
	}

	o.Birthday = &raw

	return nil
}

func (o *PutUsersMeParams) bindCity(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}
	if raw == "" { // empty values pass all other validations
		var cityDefault string = string("")
		o.City = &cityDefault
		return nil
	}

	o.City = &raw

	if err := o.validateCity(formats); err != nil {
		return err
	}

	return nil
}

func (o *PutUsersMeParams) validateCity(formats strfmt.Registry) error {

	if err := validate.MaxLength("city", "formData", (*o.City), 50); err != nil {
		return err
	}

	return nil
}

func (o *PutUsersMeParams) bindCountry(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}
	if raw == "" { // empty values pass all other validations
		var countryDefault string = string("")
		o.Country = &countryDefault
		return nil
	}

	o.Country = &raw

	if err := o.validateCountry(formats); err != nil {
		return err
	}

	return nil
}

func (o *PutUsersMeParams) validateCountry(formats strfmt.Registry) error {

	if err := validate.MaxLength("country", "formData", (*o.Country), 50); err != nil {
		return err
	}

	return nil
}

func (o *PutUsersMeParams) bindGender(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}
	if raw == "" { // empty values pass all other validations
		var genderDefault string = string("not set")
		o.Gender = &genderDefault
		return nil
	}

	o.Gender = &raw

	if err := o.validateGender(formats); err != nil {
		return err
	}

	return nil
}

func (o *PutUsersMeParams) validateGender(formats strfmt.Registry) error {

	if err := validate.Enum("gender", "formData", *o.Gender, []interface{}{"male", "female", "not set"}); err != nil {
		return err
	}

	return nil
}

func (o *PutUsersMeParams) bindIsDaylog(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}
	if raw == "" { // empty values pass all other validations
		var isDaylogDefault bool = bool(false)
		o.IsDaylog = &isDaylogDefault
		return nil
	}

	value, err := swag.ConvertBool(raw)
	if err != nil {
		return errors.InvalidType("isDaylog", "formData", "bool", raw)
	}
	o.IsDaylog = &value

	return nil
}

func (o *PutUsersMeParams) bindPrivacy(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("privacy", "formData")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}
	if err := validate.RequiredString("privacy", "formData", raw); err != nil {
		return err
	}

	o.Privacy = raw

	if err := o.validatePrivacy(formats); err != nil {
		return err
	}

	return nil
}

func (o *PutUsersMeParams) validatePrivacy(formats strfmt.Registry) error {

	if err := validate.Enum("privacy", "formData", o.Privacy, []interface{}{"all", "followers"}); err != nil {
		return err
	}

	return nil
}

func (o *PutUsersMeParams) bindShowInTops(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}
	if raw == "" { // empty values pass all other validations
		var showInTopsDefault bool = bool(false)
		o.ShowInTops = &showInTopsDefault
		return nil
	}

	value, err := swag.ConvertBool(raw)
	if err != nil {
		return errors.InvalidType("showInTops", "formData", "bool", raw)
	}
	o.ShowInTops = &value

	return nil
}

func (o *PutUsersMeParams) bindShowName(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("showName", "formData")
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}
	if err := validate.RequiredString("showName", "formData", raw); err != nil {
		return err
	}

	o.ShowName = raw

	if err := o.validateShowName(formats); err != nil {
		return err
	}

	return nil
}

func (o *PutUsersMeParams) validateShowName(formats strfmt.Registry) error {

	if err := validate.MinLength("showName", "formData", o.ShowName, 1); err != nil {
		return err
	}

	if err := validate.MaxLength("showName", "formData", o.ShowName, 20); err != nil {
		return err
	}

	return nil
}

func (o *PutUsersMeParams) bindTitle(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}
	if raw == "" { // empty values pass all other validations
		var titleDefault string = string("")
		o.Title = &titleDefault
		return nil
	}

	o.Title = &raw

	if err := o.validateTitle(formats); err != nil {
		return err
	}

	return nil
}

func (o *PutUsersMeParams) validateTitle(formats strfmt.Registry) error {

	if err := validate.MaxLength("title", "formData", (*o.Title), 500); err != nil {
		return err
	}

	return nil
}
