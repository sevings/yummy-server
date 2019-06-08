// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
	"strconv"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// Entry entry
// swagger:model Entry
type Entry struct {

	// author
	Author *User `json:"author,omitempty"`

	// comment count
	CommentCount int64 `json:"commentCount,omitempty"`

	// comments
	Comments *CommentList `json:"comments,omitempty"`

	// content
	Content string `json:"content,omitempty"`

	// created at
	CreatedAt float64 `json:"createdAt,omitempty"`

	// cut content
	CutContent string `json:"cutContent,omitempty"`

	// cut title
	CutTitle string `json:"cutTitle,omitempty"`

	// edit content
	EditContent string `json:"editContent,omitempty"`

	// has cut
	HasCut bool `json:"hasCut,omitempty"`

	// id
	// Minimum: 1
	ID int64 `json:"id,omitempty"`

	// images
	Images []*Image `json:"images"`

	// in live
	InLive bool `json:"inLive,omitempty"`

	// is favorited
	IsFavorited bool `json:"isFavorited,omitempty"`

	// is watching
	IsWatching bool `json:"isWatching,omitempty"`

	// privacy
	// Enum: [all some me anonymous]
	Privacy string `json:"privacy,omitempty"`

	// rating
	Rating *Rating `json:"rating,omitempty"`

	// rights
	Rights *EntryRights `json:"rights,omitempty"`

	// title
	Title string `json:"title,omitempty"`

	// visible for
	VisibleFor []*User `json:"visibleFor"`

	// word count
	WordCount int64 `json:"wordCount,omitempty"`
}

// Validate validates this entry
func (m *Entry) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAuthor(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateComments(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateImages(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validatePrivacy(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRating(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRights(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateVisibleFor(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Entry) validateAuthor(formats strfmt.Registry) error {

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

func (m *Entry) validateComments(formats strfmt.Registry) error {

	if swag.IsZero(m.Comments) { // not required
		return nil
	}

	if m.Comments != nil {
		if err := m.Comments.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("comments")
			}
			return err
		}
	}

	return nil
}

func (m *Entry) validateID(formats strfmt.Registry) error {

	if swag.IsZero(m.ID) { // not required
		return nil
	}

	if err := validate.MinimumInt("id", "body", int64(m.ID), 1, false); err != nil {
		return err
	}

	return nil
}

func (m *Entry) validateImages(formats strfmt.Registry) error {

	if swag.IsZero(m.Images) { // not required
		return nil
	}

	for i := 0; i < len(m.Images); i++ {
		if swag.IsZero(m.Images[i]) { // not required
			continue
		}

		if m.Images[i] != nil {
			if err := m.Images[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("images" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

var entryTypePrivacyPropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["all","some","me","anonymous"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		entryTypePrivacyPropEnum = append(entryTypePrivacyPropEnum, v)
	}
}

const (

	// EntryPrivacyAll captures enum value "all"
	EntryPrivacyAll string = "all"

	// EntryPrivacySome captures enum value "some"
	EntryPrivacySome string = "some"

	// EntryPrivacyMe captures enum value "me"
	EntryPrivacyMe string = "me"

	// EntryPrivacyAnonymous captures enum value "anonymous"
	EntryPrivacyAnonymous string = "anonymous"
)

// prop value enum
func (m *Entry) validatePrivacyEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, entryTypePrivacyPropEnum); err != nil {
		return err
	}
	return nil
}

func (m *Entry) validatePrivacy(formats strfmt.Registry) error {

	if swag.IsZero(m.Privacy) { // not required
		return nil
	}

	// value enum
	if err := m.validatePrivacyEnum("privacy", "body", m.Privacy); err != nil {
		return err
	}

	return nil
}

func (m *Entry) validateRating(formats strfmt.Registry) error {

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

func (m *Entry) validateRights(formats strfmt.Registry) error {

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

func (m *Entry) validateVisibleFor(formats strfmt.Registry) error {

	if swag.IsZero(m.VisibleFor) { // not required
		return nil
	}

	for i := 0; i < len(m.VisibleFor); i++ {
		if swag.IsZero(m.VisibleFor[i]) { // not required
			continue
		}

		if m.VisibleFor[i] != nil {
			if err := m.VisibleFor[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("visibleFor" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *Entry) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Entry) UnmarshalBinary(b []byte) error {
	var res Entry
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// EntryRights entry rights
// swagger:model EntryRights
type EntryRights struct {

	// comment
	Comment bool `json:"comment,omitempty"`

	// delete
	Delete bool `json:"delete,omitempty"`

	// edit
	Edit bool `json:"edit,omitempty"`

	// vote
	Vote bool `json:"vote,omitempty"`
}

// Validate validates this entry rights
func (m *EntryRights) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *EntryRights) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *EntryRights) UnmarshalBinary(b []byte) error {
	var res EntryRights
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
