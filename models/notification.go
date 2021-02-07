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

// Notification notification
//
// swagger:model Notification
type Notification struct {

	// comment
	Comment *Comment `json:"comment,omitempty"`

	// created at
	CreatedAt float64 `json:"createdAt,omitempty"`

	// entry
	Entry *Entry `json:"entry,omitempty"`

	// id
	// Minimum: 1
	ID int64 `json:"id,omitempty"`

	// info
	Info *NotificationInfo `json:"info,omitempty"`

	// read
	Read bool `json:"read,omitempty"`

	// type
	// Enum: [comment follower request accept invite welcome invited adm_sent adm_received info]
	Type string `json:"type,omitempty"`

	// user
	User *User `json:"user,omitempty"`
}

// Validate validates this notification
func (m *Notification) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateComment(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateEntry(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateInfo(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateType(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateUser(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Notification) validateComment(formats strfmt.Registry) error {
	if swag.IsZero(m.Comment) { // not required
		return nil
	}

	if m.Comment != nil {
		if err := m.Comment.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("comment")
			}
			return err
		}
	}

	return nil
}

func (m *Notification) validateEntry(formats strfmt.Registry) error {
	if swag.IsZero(m.Entry) { // not required
		return nil
	}

	if m.Entry != nil {
		if err := m.Entry.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("entry")
			}
			return err
		}
	}

	return nil
}

func (m *Notification) validateID(formats strfmt.Registry) error {
	if swag.IsZero(m.ID) { // not required
		return nil
	}

	if err := validate.MinimumInt("id", "body", m.ID, 1, false); err != nil {
		return err
	}

	return nil
}

func (m *Notification) validateInfo(formats strfmt.Registry) error {
	if swag.IsZero(m.Info) { // not required
		return nil
	}

	if m.Info != nil {
		if err := m.Info.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("info")
			}
			return err
		}
	}

	return nil
}

var notificationTypeTypePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["comment","follower","request","accept","invite","welcome","invited","adm_sent","adm_received","info"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		notificationTypeTypePropEnum = append(notificationTypeTypePropEnum, v)
	}
}

const (

	// NotificationTypeComment captures enum value "comment"
	NotificationTypeComment string = "comment"

	// NotificationTypeFollower captures enum value "follower"
	NotificationTypeFollower string = "follower"

	// NotificationTypeRequest captures enum value "request"
	NotificationTypeRequest string = "request"

	// NotificationTypeAccept captures enum value "accept"
	NotificationTypeAccept string = "accept"

	// NotificationTypeInvite captures enum value "invite"
	NotificationTypeInvite string = "invite"

	// NotificationTypeWelcome captures enum value "welcome"
	NotificationTypeWelcome string = "welcome"

	// NotificationTypeInvited captures enum value "invited"
	NotificationTypeInvited string = "invited"

	// NotificationTypeAdmSent captures enum value "adm_sent"
	NotificationTypeAdmSent string = "adm_sent"

	// NotificationTypeAdmReceived captures enum value "adm_received"
	NotificationTypeAdmReceived string = "adm_received"

	// NotificationTypeInfo captures enum value "info"
	NotificationTypeInfo string = "info"
)

// prop value enum
func (m *Notification) validateTypeEnum(path, location string, value string) error {
	if err := validate.EnumCase(path, location, value, notificationTypeTypePropEnum, true); err != nil {
		return err
	}
	return nil
}

func (m *Notification) validateType(formats strfmt.Registry) error {
	if swag.IsZero(m.Type) { // not required
		return nil
	}

	// value enum
	if err := m.validateTypeEnum("type", "body", m.Type); err != nil {
		return err
	}

	return nil
}

func (m *Notification) validateUser(formats strfmt.Registry) error {
	if swag.IsZero(m.User) { // not required
		return nil
	}

	if m.User != nil {
		if err := m.User.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("user")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this notification based on the context it is used
func (m *Notification) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateComment(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateEntry(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateInfo(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateUser(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Notification) contextValidateComment(ctx context.Context, formats strfmt.Registry) error {

	if m.Comment != nil {
		if err := m.Comment.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("comment")
			}
			return err
		}
	}

	return nil
}

func (m *Notification) contextValidateEntry(ctx context.Context, formats strfmt.Registry) error {

	if m.Entry != nil {
		if err := m.Entry.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("entry")
			}
			return err
		}
	}

	return nil
}

func (m *Notification) contextValidateInfo(ctx context.Context, formats strfmt.Registry) error {

	if m.Info != nil {
		if err := m.Info.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("info")
			}
			return err
		}
	}

	return nil
}

func (m *Notification) contextValidateUser(ctx context.Context, formats strfmt.Registry) error {

	if m.User != nil {
		if err := m.User.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("user")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Notification) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Notification) UnmarshalBinary(b []byte) error {
	var res Notification
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// NotificationInfo notification info
//
// swagger:model NotificationInfo
type NotificationInfo struct {

	// content
	Content string `json:"content,omitempty"`

	// link
	Link string `json:"link,omitempty"`
}

// Validate validates this notification info
func (m *NotificationInfo) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this notification info based on context it is used
func (m *NotificationInfo) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *NotificationInfo) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *NotificationInfo) UnmarshalBinary(b []byte) error {
	var res NotificationInfo
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
