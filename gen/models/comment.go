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

// Comment comment
// swagger:model Comment

type Comment struct {

	// author
	Author *User `json:"author,omitempty"`

	// content
	// Min Length: 1
	Content string `json:"content,omitempty"`

	// created at
	CreatedAt strfmt.DateTime `json:"createdAt,omitempty"`

	// entry Id
	// Minimum: 1
	EntryID int64 `json:"entryId,omitempty"`

	// id
	// Minimum: 1
	ID int64 `json:"id,omitempty"`

	// rating
	Rating int64 `json:"rating,omitempty"`

	// vote
	Vote string `json:"vote,omitempty"`
}

/* polymorph Comment author false */

/* polymorph Comment content false */

/* polymorph Comment createdAt false */

/* polymorph Comment entryId false */

/* polymorph Comment id false */

/* polymorph Comment rating false */

/* polymorph Comment vote false */

// Validate validates this comment
func (m *Comment) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAuthor(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateContent(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateEntryID(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateID(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateVote(formats); err != nil {
		// prop
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

func (m *Comment) validateContent(formats strfmt.Registry) error {

	if swag.IsZero(m.Content) { // not required
		return nil
	}

	if err := validate.MinLength("content", "body", string(m.Content), 1); err != nil {
		return err
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

var commentTypeVotePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["not","pos","neg"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		commentTypeVotePropEnum = append(commentTypeVotePropEnum, v)
	}
}

const (
	// CommentVoteNot captures enum value "not"
	CommentVoteNot string = "not"
	// CommentVotePos captures enum value "pos"
	CommentVotePos string = "pos"
	// CommentVoteNeg captures enum value "neg"
	CommentVoteNeg string = "neg"
)

// prop value enum
func (m *Comment) validateVoteEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, commentTypeVotePropEnum); err != nil {
		return err
	}
	return nil
}

func (m *Comment) validateVote(formats strfmt.Registry) error {

	if swag.IsZero(m.Vote) { // not required
		return nil
	}

	// value enum
	if err := m.validateVoteEnum("vote", "body", m.Vote); err != nil {
		return err
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
