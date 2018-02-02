// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/yummy-server/models"
)

// GetUsersIDInvitedOKCode is the HTTP code returned for type GetUsersIDInvitedOK
const GetUsersIDInvitedOKCode int = 200

/*GetUsersIDInvitedOK User list

swagger:response getUsersIdInvitedOK
*/
type GetUsersIDInvitedOK struct {

	/*
	  In: Body
	*/
	Payload *models.UserList `json:"body,omitempty"`
}

// NewGetUsersIDInvitedOK creates GetUsersIDInvitedOK with default headers values
func NewGetUsersIDInvitedOK() *GetUsersIDInvitedOK {
	return &GetUsersIDInvitedOK{}
}

// WithPayload adds the payload to the get users Id invited o k response
func (o *GetUsersIDInvitedOK) WithPayload(payload *models.UserList) *GetUsersIDInvitedOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users Id invited o k response
func (o *GetUsersIDInvitedOK) SetPayload(payload *models.UserList) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersIDInvitedOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetUsersIDInvitedForbiddenCode is the HTTP code returned for type GetUsersIDInvitedForbidden
const GetUsersIDInvitedForbiddenCode int = 403

/*GetUsersIDInvitedForbidden access denied

swagger:response getUsersIdInvitedForbidden
*/
type GetUsersIDInvitedForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetUsersIDInvitedForbidden creates GetUsersIDInvitedForbidden with default headers values
func NewGetUsersIDInvitedForbidden() *GetUsersIDInvitedForbidden {
	return &GetUsersIDInvitedForbidden{}
}

// WithPayload adds the payload to the get users Id invited forbidden response
func (o *GetUsersIDInvitedForbidden) WithPayload(payload *models.Error) *GetUsersIDInvitedForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users Id invited forbidden response
func (o *GetUsersIDInvitedForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersIDInvitedForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetUsersIDInvitedNotFoundCode is the HTTP code returned for type GetUsersIDInvitedNotFound
const GetUsersIDInvitedNotFoundCode int = 404

/*GetUsersIDInvitedNotFound User not found

swagger:response getUsersIdInvitedNotFound
*/
type GetUsersIDInvitedNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetUsersIDInvitedNotFound creates GetUsersIDInvitedNotFound with default headers values
func NewGetUsersIDInvitedNotFound() *GetUsersIDInvitedNotFound {
	return &GetUsersIDInvitedNotFound{}
}

// WithPayload adds the payload to the get users Id invited not found response
func (o *GetUsersIDInvitedNotFound) WithPayload(payload *models.Error) *GetUsersIDInvitedNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users Id invited not found response
func (o *GetUsersIDInvitedNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersIDInvitedNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}