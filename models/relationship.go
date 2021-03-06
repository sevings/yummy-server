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

// Relationship relationship
//
// swagger:model Relationship
type Relationship struct {

	// from
	From string `json:"from,omitempty"`

	// relation
	// Enum: [followed requested ignored hidden none]
	Relation string `json:"relation,omitempty"`

	// to
	To string `json:"to,omitempty"`
}

// Validate validates this relationship
func (m *Relationship) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateRelation(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

var relationshipTypeRelationPropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["followed","requested","ignored","hidden","none"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		relationshipTypeRelationPropEnum = append(relationshipTypeRelationPropEnum, v)
	}
}

const (

	// RelationshipRelationFollowed captures enum value "followed"
	RelationshipRelationFollowed string = "followed"

	// RelationshipRelationRequested captures enum value "requested"
	RelationshipRelationRequested string = "requested"

	// RelationshipRelationIgnored captures enum value "ignored"
	RelationshipRelationIgnored string = "ignored"

	// RelationshipRelationHidden captures enum value "hidden"
	RelationshipRelationHidden string = "hidden"

	// RelationshipRelationNone captures enum value "none"
	RelationshipRelationNone string = "none"
)

// prop value enum
func (m *Relationship) validateRelationEnum(path, location string, value string) error {
	if err := validate.EnumCase(path, location, value, relationshipTypeRelationPropEnum, true); err != nil {
		return err
	}
	return nil
}

func (m *Relationship) validateRelation(formats strfmt.Registry) error {
	if swag.IsZero(m.Relation) { // not required
		return nil
	}

	// value enum
	if err := m.validateRelationEnum("relation", "body", m.Relation); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this relationship based on context it is used
func (m *Relationship) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *Relationship) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Relationship) UnmarshalBinary(b []byte) error {
	var res Relationship
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
