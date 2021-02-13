// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// App app
//
// swagger:model App
type App struct {

	// id
	ID int64 `json:"id,omitempty"`

	// info
	Info string `json:"info,omitempty"`

	// name
	Name string `json:"name,omitempty"`

	// platform
	Platform string `json:"platform,omitempty"`

	// show name
	ShowName string `json:"showName,omitempty"`
}

// Validate validates this app
func (m *App) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this app based on context it is used
func (m *App) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *App) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *App) UnmarshalBinary(b []byte) error {
	var res App
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
