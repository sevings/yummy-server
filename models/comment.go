// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// Comment comment
//
// swagger:model Comment
type Comment struct {

	// author
	Author *User `json:"author,omitempty"`

	// content
	Content string `json:"content,omitempty"`

	// created at
	CreatedAt float64 `json:"createdAt,omitempty"`

	// edit content
	EditContent string `json:"editContent,omitempty"`

	// entry Id
	// Minimum: 1
	EntryID int64 `json:"entryId,omitempty"`

	// id
	// Minimum: 1
	ID int64 `json:"id,omitempty"`

	// rating
	Rating *Rating `json:"rating,omitempty"`

	// rights
	Rights *CommentRights `json:"rights,omitempty"`
}

// Validate validates this comment
func (m *Comment) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAuthor(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateEntryID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRating(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRights(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Comment) validateAuthor(formats strfmt.Registry) error {

	if swag.IsZero(m.Author) { // not required
		return nil
	}

	if m.Author != nil {
		if err := m.Author.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("author")
			}
			return err
		}
	}

	return nil
}

func (m *Comment) validateEntryID(formats strfmt.Registry) error {

	if swag.IsZero(m.EntryID) { // not required
		return nil
	}

	if err := validate.MinimumInt("entryId", "body", int64(m.EntryID), 1, false); err != nil {
		return err
	}

	return nil
}

func (m *Comment) validateID(formats strfmt.Registry) error {

	if swag.IsZero(m.ID) { // not required
		return nil
	}

	if err := validate.MinimumInt("id", "body", int64(m.ID), 1, false); err != nil {
		return err
	}

	return nil
}

func (m *Comment) validateRating(formats strfmt.Registry) error {

	if swag.IsZero(m.Rating) { // not required
		return nil
	}

	if m.Rating != nil {
		if err := m.Rating.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("rating")
			}
			return err
		}
	}

	return nil
}

func (m *Comment) validateRights(formats strfmt.Registry) error {

	if swag.IsZero(m.Rights) { // not required
		return nil
	}

	if m.Rights != nil {
		if err := m.Rights.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("rights")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Comment) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Comment) UnmarshalBinary(b []byte) error {
	var res Comment
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// CommentRights comment rights
//
// swagger:model CommentRights
type CommentRights struct {

	// complain
	Complain bool `json:"complain,omitempty"`

	// delete
	Delete bool `json:"delete,omitempty"`

	// edit
	Edit bool `json:"edit,omitempty"`

	// vote
	Vote bool `json:"vote,omitempty"`
}

// Validate validates this comment rights
func (m *CommentRights) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *CommentRights) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *CommentRights) UnmarshalBinary(b []byte) error {
	var res CommentRights
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
