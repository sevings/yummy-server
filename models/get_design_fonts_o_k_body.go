// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// GetDesignFontsOKBody get design fonts o k body
// swagger:model getDesignFontsOKBody
type GetDesignFontsOKBody struct {

	// fonts
	Fonts []string `json:"fonts"`
}

// Validate validates this get design fonts o k body
func (m *GetDesignFontsOKBody) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *GetDesignFontsOKBody) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *GetDesignFontsOKBody) UnmarshalBinary(b []byte) error {
	var res GetDesignFontsOKBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}