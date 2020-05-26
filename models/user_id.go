// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// UserID user ID
// swagger:model UserID
type UserID struct {

	// ban
	Ban *UserIDBan `json:"ban,omitempty"`

	// followers count
	FollowersCount int64 `json:"followersCount,omitempty"`

	// id
	ID int64 `json:"id,omitempty"`

	// is invited
	IsInvited bool `json:"isInvited,omitempty"`

	// name
	// Max Length: 20
	// Min Length: 1
	Name string `json:"name,omitempty"`

	// neg karma
	NegKarma bool `json:"negKarma,omitempty"`

	// verified
	Verified bool `json:"verified,omitempty"`
}

// Validate validates this user ID
func (m *UserID) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateBan(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateName(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *UserID) validateBan(formats strfmt.Registry) error {

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

func (m *UserID) validateName(formats strfmt.Registry) error {

	if swag.IsZero(m.Name) { // not required
		return nil
	}

	if err := validate.MinLength("name", "body", string(m.Name), 1); err != nil {
		return err
	}

	if err := validate.MaxLength("name", "body", string(m.Name), 20); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *UserID) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *UserID) UnmarshalBinary(b []byte) error {
	var res UserID
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// UserIDBan user ID ban
// swagger:model UserIDBan
type UserIDBan struct {

	// comment
	Comment bool `json:"comment,omitempty"`

	// invite
	Invite bool `json:"invite,omitempty"`

	// live
	Live bool `json:"live,omitempty"`

	// vote
	Vote bool `json:"vote,omitempty"`
}

// Validate validates this user ID ban
func (m *UserIDBan) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *UserIDBan) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *UserIDBan) UnmarshalBinary(b []byte) error {
	var res UserIDBan
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
