// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// Color color in rgb
// Example: #373737
//
// swagger:model Color
type Color string

// Validate validates this color
func (m Color) Validate(formats strfmt.Registry) error {
	var res []error

	if err := validate.Pattern("", "body", string(m), `^#[0-9a-fA-F]{6}$`); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// ContextValidate validates this color based on context it is used
func (m Color) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}
