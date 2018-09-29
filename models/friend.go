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

// Friend friend
// swagger:model Friend
type Friend struct {
	User

	// counts
	Counts *FriendAO1Counts `json:"counts,omitempty"`

	// cover
	Cover *Cover `json:"cover,omitempty"`

	// gender
	// Enum: [male female not set]
	Gender string `json:"gender,omitempty"`

	// last seen at
	LastSeenAt float64 `json:"lastSeenAt,omitempty"`

	// rank
	Rank int64 `json:"rank,omitempty"`

	// title
	// Max Length: 260
	Title string `json:"title,omitempty"`
}

// UnmarshalJSON unmarshals this object from a JSON structure
func (m *Friend) UnmarshalJSON(raw []byte) error {
	// AO0
	var aO0 User
	if err := swag.ReadJSON(raw, &aO0); err != nil {
		return err
	}
	m.User = aO0

	// AO1
	var dataAO1 struct {
		Counts *FriendAO1Counts `json:"counts,omitempty"`

		Cover *Cover `json:"cover,omitempty"`

		Gender string `json:"gender,omitempty"`

		LastSeenAt float64 `json:"lastSeenAt,omitempty"`

		Rank int64 `json:"rank,omitempty"`

		Title string `json:"title,omitempty"`
	}
	if err := swag.ReadJSON(raw, &dataAO1); err != nil {
		return err
	}

	m.Counts = dataAO1.Counts

	m.Cover = dataAO1.Cover

	m.Gender = dataAO1.Gender

	m.LastSeenAt = dataAO1.LastSeenAt

	m.Rank = dataAO1.Rank

	m.Title = dataAO1.Title

	return nil
}

// MarshalJSON marshals this object to a JSON structure
func (m Friend) MarshalJSON() ([]byte, error) {
	_parts := make([][]byte, 0, 2)

	aO0, err := swag.WriteJSON(m.User)
	if err != nil {
		return nil, err
	}
	_parts = append(_parts, aO0)

	var dataAO1 struct {
		Counts *FriendAO1Counts `json:"counts,omitempty"`

		Cover *Cover `json:"cover,omitempty"`

		Gender string `json:"gender,omitempty"`

		LastSeenAt float64 `json:"lastSeenAt,omitempty"`

		Rank int64 `json:"rank,omitempty"`

		Title string `json:"title,omitempty"`
	}

	dataAO1.Counts = m.Counts

	dataAO1.Cover = m.Cover

	dataAO1.Gender = m.Gender

	dataAO1.LastSeenAt = m.LastSeenAt

	dataAO1.Rank = m.Rank

	dataAO1.Title = m.Title

	jsonDataAO1, errAO1 := swag.WriteJSON(dataAO1)
	if errAO1 != nil {
		return nil, errAO1
	}
	_parts = append(_parts, jsonDataAO1)

	return swag.ConcatJSON(_parts...), nil
}

// Validate validates this friend
func (m *Friend) Validate(formats strfmt.Registry) error {
	var res []error

	// validation for a type composition with User
	if err := m.User.Validate(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCounts(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCover(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateGender(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTitle(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Friend) validateCounts(formats strfmt.Registry) error {

	if swag.IsZero(m.Counts) { // not required
		return nil
	}

	if m.Counts != nil {
		if err := m.Counts.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("counts")
			}
			return err
		}
	}

	return nil
}

func (m *Friend) validateCover(formats strfmt.Registry) error {

	if swag.IsZero(m.Cover) { // not required
		return nil
	}

	if m.Cover != nil {
		if err := m.Cover.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("cover")
			}
			return err
		}
	}

	return nil
}

var friendTypeGenderPropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["male","female","not set"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		friendTypeGenderPropEnum = append(friendTypeGenderPropEnum, v)
	}
}

// property enum
func (m *Friend) validateGenderEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, friendTypeGenderPropEnum); err != nil {
		return err
	}
	return nil
}

func (m *Friend) validateGender(formats strfmt.Registry) error {

	if swag.IsZero(m.Gender) { // not required
		return nil
	}

	// value enum
	if err := m.validateGenderEnum("gender", "body", m.Gender); err != nil {
		return err
	}

	return nil
}

func (m *Friend) validateTitle(formats strfmt.Registry) error {

	if swag.IsZero(m.Title) { // not required
		return nil
	}

	if err := validate.MaxLength("title", "body", string(m.Title), 260); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Friend) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Friend) UnmarshalBinary(b []byte) error {
	var res Friend
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// FriendAO1Counts friend a o1 counts
// swagger:model FriendAO1Counts
type FriendAO1Counts struct {

	// comments
	Comments int64 `json:"comments,omitempty"`

	// days
	Days int64 `json:"days,omitempty"`

	// entries
	Entries int64 `json:"entries,omitempty"`

	// favorites
	Favorites int64 `json:"favorites,omitempty"`

	// followers
	Followers int64 `json:"followers,omitempty"`

	// followings
	Followings int64 `json:"followings,omitempty"`

	// ignored
	Ignored int64 `json:"ignored,omitempty"`

	// invited
	Invited int64 `json:"invited,omitempty"`

	// tags
	Tags int64 `json:"tags,omitempty"`
}

// Validate validates this friend a o1 counts
func (m *FriendAO1Counts) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *FriendAO1Counts) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *FriendAO1Counts) UnmarshalBinary(b []byte) error {
	var res FriendAO1Counts
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
