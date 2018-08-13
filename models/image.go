// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// Image image
// swagger:model Image
type Image struct {

	// id
	ID int64 `json:"id,omitempty"`

	// large
	Large *ImageSize `json:"large,omitempty"`

	// medium
	Medium *ImageSize `json:"medium,omitempty"`

	// mime type
	MimeType string `json:"mimeType,omitempty"`

	// small
	Small *ImageSize `json:"small,omitempty"`

	// user Id
	UserID int64 `json:"userId,omitempty"`
}

// Validate validates this image
func (m *Image) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateLarge(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateMedium(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateSmall(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Image) validateLarge(formats strfmt.Registry) error {

	if swag.IsZero(m.Large) { // not required
		return nil
	}

	if m.Large != nil {
		if err := m.Large.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("large")
			}
			return err
		}
	}

	return nil
}

func (m *Image) validateMedium(formats strfmt.Registry) error {

	if swag.IsZero(m.Medium) { // not required
		return nil
	}

	if m.Medium != nil {
		if err := m.Medium.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("medium")
			}
			return err
		}
	}

	return nil
}

func (m *Image) validateSmall(formats strfmt.Registry) error {

	if swag.IsZero(m.Small) { // not required
		return nil
	}

	if m.Small != nil {
		if err := m.Small.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("small")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Image) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Image) UnmarshalBinary(b []byte) error {
	var res Image
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}