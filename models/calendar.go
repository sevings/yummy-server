// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// Calendar calendar
//
// swagger:model Calendar
type Calendar struct {

	// end
	End int64 `json:"end,omitempty"`

	// entries
	Entries []*CalendarEntriesItems0 `json:"entries"`

	// limit
	Limit int64 `json:"limit,omitempty"`

	// start
	Start int64 `json:"start,omitempty"`
}

// Validate validates this calendar
func (m *Calendar) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateEntries(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Calendar) validateEntries(formats strfmt.Registry) error {
	if swag.IsZero(m.Entries) { // not required
		return nil
	}

	for i := 0; i < len(m.Entries); i++ {
		if swag.IsZero(m.Entries[i]) { // not required
			continue
		}

		if m.Entries[i] != nil {
			if err := m.Entries[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("entries" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this calendar based on the context it is used
func (m *Calendar) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateEntries(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Calendar) contextValidateEntries(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Entries); i++ {

		if m.Entries[i] != nil {
			if err := m.Entries[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("entries" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *Calendar) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Calendar) UnmarshalBinary(b []byte) error {
	var res Calendar
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// CalendarEntriesItems0 calendar entries items0
//
// swagger:model CalendarEntriesItems0
type CalendarEntriesItems0 struct {

	// created at
	CreatedAt float64 `json:"createdAt,omitempty"`

	// id
	ID int64 `json:"id,omitempty"`

	// title
	Title string `json:"title,omitempty"`
}

// Validate validates this calendar entries items0
func (m *CalendarEntriesItems0) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this calendar entries items0 based on context it is used
func (m *CalendarEntriesItems0) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *CalendarEntriesItems0) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *CalendarEntriesItems0) UnmarshalBinary(b []byte) error {
	var res CalendarEntriesItems0
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
