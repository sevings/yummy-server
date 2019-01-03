// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// Profile profile
// swagger:model Profile
type Profile struct {
	Friend

	// age lower bound
	AgeLowerBound int64 `json:"ageLowerBound,omitempty"`

	// age upper bound
	AgeUpperBound int64 `json:"ageUpperBound,omitempty"`

	// city
	// Max Length: 50
	City string `json:"city,omitempty"`

	// country
	// Max Length: 50
	Country string `json:"country,omitempty"`

	// created at
	CreatedAt float64 `json:"createdAt,omitempty"`

	// design
	Design *Design `json:"design,omitempty"`

	// invited by
	InvitedBy *User `json:"invitedBy,omitempty"`

	// is daylog
	IsDaylog bool `json:"isDaylog,omitempty"`

	// relations
	Relations *ProfileAO1Relations `json:"relations,omitempty"`
}

// UnmarshalJSON unmarshals this object from a JSON structure
func (m *Profile) UnmarshalJSON(raw []byte) error {
	// AO0
	var aO0 Friend
	if err := swag.ReadJSON(raw, &aO0); err != nil {
		return err
	}
	m.Friend = aO0

	// AO1
	var dataAO1 struct {
		AgeLowerBound int64 `json:"ageLowerBound,omitempty"`

		AgeUpperBound int64 `json:"ageUpperBound,omitempty"`

		City string `json:"city,omitempty"`

		Country string `json:"country,omitempty"`

		CreatedAt float64 `json:"createdAt,omitempty"`

		Design *Design `json:"design,omitempty"`

		InvitedBy *User `json:"invitedBy,omitempty"`

		IsDaylog bool `json:"isDaylog,omitempty"`

		Relations *ProfileAO1Relations `json:"relations,omitempty"`
	}
	if err := swag.ReadJSON(raw, &dataAO1); err != nil {
		return err
	}

	m.AgeLowerBound = dataAO1.AgeLowerBound

	m.AgeUpperBound = dataAO1.AgeUpperBound

	m.City = dataAO1.City

	m.Country = dataAO1.Country

	m.CreatedAt = dataAO1.CreatedAt

	m.Design = dataAO1.Design

	m.InvitedBy = dataAO1.InvitedBy

	m.IsDaylog = dataAO1.IsDaylog

	m.Relations = dataAO1.Relations

	return nil
}

// MarshalJSON marshals this object to a JSON structure
func (m Profile) MarshalJSON() ([]byte, error) {
	_parts := make([][]byte, 0, 2)

	aO0, err := swag.WriteJSON(m.Friend)
	if err != nil {
		return nil, err
	}
	_parts = append(_parts, aO0)

	var dataAO1 struct {
		AgeLowerBound int64 `json:"ageLowerBound,omitempty"`

		AgeUpperBound int64 `json:"ageUpperBound,omitempty"`

		City string `json:"city,omitempty"`

		Country string `json:"country,omitempty"`

		CreatedAt float64 `json:"createdAt,omitempty"`

		Design *Design `json:"design,omitempty"`

		InvitedBy *User `json:"invitedBy,omitempty"`

		IsDaylog bool `json:"isDaylog,omitempty"`

		Relations *ProfileAO1Relations `json:"relations,omitempty"`
	}

	dataAO1.AgeLowerBound = m.AgeLowerBound

	dataAO1.AgeUpperBound = m.AgeUpperBound

	dataAO1.City = m.City

	dataAO1.Country = m.Country

	dataAO1.CreatedAt = m.CreatedAt

	dataAO1.Design = m.Design

	dataAO1.InvitedBy = m.InvitedBy

	dataAO1.IsDaylog = m.IsDaylog

	dataAO1.Relations = m.Relations

	jsonDataAO1, errAO1 := swag.WriteJSON(dataAO1)
	if errAO1 != nil {
		return nil, errAO1
	}
	_parts = append(_parts, jsonDataAO1)

	return swag.ConcatJSON(_parts...), nil
}

// Validate validates this profile
func (m *Profile) Validate(formats strfmt.Registry) error {
	var res []error

	// validation for a type composition with Friend
	if err := m.Friend.Validate(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCity(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCountry(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateDesign(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateInvitedBy(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRelations(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Profile) validateCity(formats strfmt.Registry) error {

	if swag.IsZero(m.City) { // not required
		return nil
	}

	if err := validate.MaxLength("city", "body", string(m.City), 50); err != nil {
		return err
	}

	return nil
}

func (m *Profile) validateCountry(formats strfmt.Registry) error {

	if swag.IsZero(m.Country) { // not required
		return nil
	}

	if err := validate.MaxLength("country", "body", string(m.Country), 50); err != nil {
		return err
	}

	return nil
}

func (m *Profile) validateDesign(formats strfmt.Registry) error {

	if swag.IsZero(m.Design) { // not required
		return nil
	}

	if m.Design != nil {
		if err := m.Design.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("design")
			}
			return err
		}
	}

	return nil
}

func (m *Profile) validateInvitedBy(formats strfmt.Registry) error {

	if swag.IsZero(m.InvitedBy) { // not required
		return nil
	}

	if m.InvitedBy != nil {
		if err := m.InvitedBy.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("invitedBy")
			}
			return err
		}
	}

	return nil
}

func (m *Profile) validateRelations(formats strfmt.Registry) error {

	if swag.IsZero(m.Relations) { // not required
		return nil
	}

	if m.Relations != nil {
		if err := m.Relations.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("relations")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Profile) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Profile) UnmarshalBinary(b []byte) error {
	var res Profile
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// ProfileAO1Relations profile a o1 relations
// swagger:model ProfileAO1Relations
type ProfileAO1Relations struct {

	// from me
	// Enum: [followed requested ignored none]
	FromMe string `json:"fromMe,omitempty"`

	// to me
	// Enum: [followed requested ignored none]
	ToMe string `json:"toMe,omitempty"`
}

// Validate validates this profile a o1 relations
func (m *ProfileAO1Relations) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateFromMe(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateToMe(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

var profileAO1RelationsTypeFromMePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["followed","requested","ignored","none"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		profileAO1RelationsTypeFromMePropEnum = append(profileAO1RelationsTypeFromMePropEnum, v)
	}
}

const (

	// ProfileAO1RelationsFromMeFollowed captures enum value "followed"
	ProfileAO1RelationsFromMeFollowed string = "followed"

	// ProfileAO1RelationsFromMeRequested captures enum value "requested"
	ProfileAO1RelationsFromMeRequested string = "requested"

	// ProfileAO1RelationsFromMeIgnored captures enum value "ignored"
	ProfileAO1RelationsFromMeIgnored string = "ignored"

	// ProfileAO1RelationsFromMeNone captures enum value "none"
	ProfileAO1RelationsFromMeNone string = "none"
)

// prop value enum
func (m *ProfileAO1Relations) validateFromMeEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, profileAO1RelationsTypeFromMePropEnum); err != nil {
		return err
	}
	return nil
}

func (m *ProfileAO1Relations) validateFromMe(formats strfmt.Registry) error {

	if swag.IsZero(m.FromMe) { // not required
		return nil
	}

	// value enum
	if err := m.validateFromMeEnum("relations"+"."+"fromMe", "body", m.FromMe); err != nil {
		return err
	}

	return nil
}

var profileAO1RelationsTypeToMePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["followed","requested","ignored","none"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		profileAO1RelationsTypeToMePropEnum = append(profileAO1RelationsTypeToMePropEnum, v)
	}
}

const (

	// ProfileAO1RelationsToMeFollowed captures enum value "followed"
	ProfileAO1RelationsToMeFollowed string = "followed"

	// ProfileAO1RelationsToMeRequested captures enum value "requested"
	ProfileAO1RelationsToMeRequested string = "requested"

	// ProfileAO1RelationsToMeIgnored captures enum value "ignored"
	ProfileAO1RelationsToMeIgnored string = "ignored"

	// ProfileAO1RelationsToMeNone captures enum value "none"
	ProfileAO1RelationsToMeNone string = "none"
)

// prop value enum
func (m *ProfileAO1Relations) validateToMeEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, profileAO1RelationsTypeToMePropEnum); err != nil {
		return err
	}
	return nil
}

func (m *ProfileAO1Relations) validateToMe(formats strfmt.Registry) error {

	if swag.IsZero(m.ToMe) { // not required
		return nil
	}

	// value enum
	if err := m.validateToMeEnum("relations"+"."+"toMe", "body", m.ToMe); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *ProfileAO1Relations) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ProfileAO1Relations) UnmarshalBinary(b []byte) error {
	var res ProfileAO1Relations
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
