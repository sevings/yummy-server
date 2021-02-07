// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"encoding/json"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// Design design
//
// swagger:model Design
type Design struct {

	// background color
	BackgroundColor Color `json:"backgroundColor,omitempty"`

	// css
	CSS string `json:"css,omitempty"`

	// font family
	FontFamily string `json:"fontFamily,omitempty"`

	// font size
	FontSize int64 `json:"fontSize,omitempty"`

	// text alignment
	// Enum: [left right center justify]
	TextAlignment string `json:"textAlignment,omitempty"`

	// text color
	TextColor Color `json:"textColor,omitempty"`
}

// Validate validates this design
func (m *Design) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateBackgroundColor(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTextAlignment(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTextColor(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Design) validateBackgroundColor(formats strfmt.Registry) error {
	if swag.IsZero(m.BackgroundColor) { // not required
		return nil
	}

	if err := m.BackgroundColor.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("backgroundColor")
		}
		return err
	}

	return nil
}

var designTypeTextAlignmentPropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["left","right","center","justify"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		designTypeTextAlignmentPropEnum = append(designTypeTextAlignmentPropEnum, v)
	}
}

const (

	// DesignTextAlignmentLeft captures enum value "left"
	DesignTextAlignmentLeft string = "left"

	// DesignTextAlignmentRight captures enum value "right"
	DesignTextAlignmentRight string = "right"

	// DesignTextAlignmentCenter captures enum value "center"
	DesignTextAlignmentCenter string = "center"

	// DesignTextAlignmentJustify captures enum value "justify"
	DesignTextAlignmentJustify string = "justify"
)

// prop value enum
func (m *Design) validateTextAlignmentEnum(path, location string, value string) error {
	if err := validate.EnumCase(path, location, value, designTypeTextAlignmentPropEnum, true); err != nil {
		return err
	}
	return nil
}

func (m *Design) validateTextAlignment(formats strfmt.Registry) error {
	if swag.IsZero(m.TextAlignment) { // not required
		return nil
	}

	// value enum
	if err := m.validateTextAlignmentEnum("textAlignment", "body", m.TextAlignment); err != nil {
		return err
	}

	return nil
}

func (m *Design) validateTextColor(formats strfmt.Registry) error {
	if swag.IsZero(m.TextColor) { // not required
		return nil
	}

	if err := m.TextColor.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("textColor")
		}
		return err
	}

	return nil
}

// ContextValidate validate this design based on the context it is used
func (m *Design) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateBackgroundColor(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateTextColor(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Design) contextValidateBackgroundColor(ctx context.Context, formats strfmt.Registry) error {

	if err := m.BackgroundColor.ContextValidate(ctx, formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("backgroundColor")
		}
		return err
	}

	return nil
}

func (m *Design) contextValidateTextColor(ctx context.Context, formats strfmt.Registry) error {

	if err := m.TextColor.ContextValidate(ctx, formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("textColor")
		}
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Design) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Design) UnmarshalBinary(b []byte) error {
	var res Design
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
