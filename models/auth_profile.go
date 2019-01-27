// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// AuthProfile auth profile
// swagger:model AuthProfile
type AuthProfile struct {
	Profile

	// account
	Account *AuthProfileAO1Account `json:"account,omitempty"`

	// ban
	Ban *AuthProfileAO1Ban `json:"ban,omitempty"`

	// birthday
	Birthday string `json:"birthday,omitempty"`

	// show in tops
	ShowInTops bool `json:"showInTops,omitempty"`
}

// UnmarshalJSON unmarshals this object from a JSON structure
func (m *AuthProfile) UnmarshalJSON(raw []byte) error {
	// AO0
	var aO0 Profile
	if err := swag.ReadJSON(raw, &aO0); err != nil {
		return err
	}
	m.Profile = aO0

	// AO1
	var dataAO1 struct {
		Account *AuthProfileAO1Account `json:"account,omitempty"`

		Ban *AuthProfileAO1Ban `json:"ban,omitempty"`

		Birthday string `json:"birthday,omitempty"`

		ShowInTops bool `json:"showInTops,omitempty"`
	}
	if err := swag.ReadJSON(raw, &dataAO1); err != nil {
		return err
	}

	m.Account = dataAO1.Account

	m.Ban = dataAO1.Ban

	m.Birthday = dataAO1.Birthday

	m.ShowInTops = dataAO1.ShowInTops

	return nil
}

// MarshalJSON marshals this object to a JSON structure
func (m AuthProfile) MarshalJSON() ([]byte, error) {
	_parts := make([][]byte, 0, 2)

	aO0, err := swag.WriteJSON(m.Profile)
	if err != nil {
		return nil, err
	}
	_parts = append(_parts, aO0)

	var dataAO1 struct {
		Account *AuthProfileAO1Account `json:"account,omitempty"`

		Ban *AuthProfileAO1Ban `json:"ban,omitempty"`

		Birthday string `json:"birthday,omitempty"`

		ShowInTops bool `json:"showInTops,omitempty"`
	}

	dataAO1.Account = m.Account

	dataAO1.Ban = m.Ban

	dataAO1.Birthday = m.Birthday

	dataAO1.ShowInTops = m.ShowInTops

	jsonDataAO1, errAO1 := swag.WriteJSON(dataAO1)
	if errAO1 != nil {
		return nil, errAO1
	}
	_parts = append(_parts, jsonDataAO1)

	return swag.ConcatJSON(_parts...), nil
}

// Validate validates this auth profile
func (m *AuthProfile) Validate(formats strfmt.Registry) error {
	var res []error

	// validation for a type composition with Profile
	if err := m.Profile.Validate(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateAccount(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateBan(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *AuthProfile) validateAccount(formats strfmt.Registry) error {

	if swag.IsZero(m.Account) { // not required
		return nil
	}

	if m.Account != nil {
		if err := m.Account.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("account")
			}
			return err
		}
	}

	return nil
}

func (m *AuthProfile) validateBan(formats strfmt.Registry) error {

	if swag.IsZero(m.Ban) { // not required
		return nil
	}

	if m.Ban != nil {
		if err := m.Ban.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("ban")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *AuthProfile) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *AuthProfile) UnmarshalBinary(b []byte) error {
	var res AuthProfile
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// AuthProfileAO1Account auth profile a o1 account
// swagger:model AuthProfileAO1Account
type AuthProfileAO1Account struct {

	// api key
	APIKey string `json:"apiKey,omitempty"`

	// email
	Email string `json:"email,omitempty"`

	// valid thru
	ValidThru float64 `json:"validThru,omitempty"`

	// verified
	Verified bool `json:"verified,omitempty"`
}

// Validate validates this auth profile a o1 account
func (m *AuthProfileAO1Account) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *AuthProfileAO1Account) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *AuthProfileAO1Account) UnmarshalBinary(b []byte) error {
	var res AuthProfileAO1Account
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// AuthProfileAO1Ban auth profile a o1 ban
// swagger:model AuthProfileAO1Ban
type AuthProfileAO1Ban struct {

	// invite
	Invite float64 `json:"invite,omitempty"`

	// vote
	Vote float64 `json:"vote,omitempty"`
}

// Validate validates this auth profile a o1 ban
func (m *AuthProfileAO1Ban) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *AuthProfileAO1Ban) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *AuthProfileAO1Ban) UnmarshalBinary(b []byte) error {
	var res AuthProfileAO1Ban
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
