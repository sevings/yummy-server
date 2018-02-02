// Code generated by go-swagger; DO NOT EDIT.

package me

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/yummy-server/models"
)

// GetUsersMeRequestedOKCode is the HTTP code returned for type GetUsersMeRequestedOK
const GetUsersMeRequestedOKCode int = 200

/*GetUsersMeRequestedOK User list

swagger:response getUsersMeRequestedOK
*/
type GetUsersMeRequestedOK struct {

	/*
	  In: Body
	*/
	Payload *models.UserList `json:"body,omitempty"`
}

// NewGetUsersMeRequestedOK creates GetUsersMeRequestedOK with default headers values
func NewGetUsersMeRequestedOK() *GetUsersMeRequestedOK {
	return &GetUsersMeRequestedOK{}
}

// WithPayload adds the payload to the get users me requested o k response
func (o *GetUsersMeRequestedOK) WithPayload(payload *models.UserList) *GetUsersMeRequestedOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users me requested o k response
func (o *GetUsersMeRequestedOK) SetPayload(payload *models.UserList) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersMeRequestedOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetUsersMeRequestedForbiddenCode is the HTTP code returned for type GetUsersMeRequestedForbidden
const GetUsersMeRequestedForbiddenCode int = 403

/*GetUsersMeRequestedForbidden access denied

swagger:response getUsersMeRequestedForbidden
*/
type GetUsersMeRequestedForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetUsersMeRequestedForbidden creates GetUsersMeRequestedForbidden with default headers values
func NewGetUsersMeRequestedForbidden() *GetUsersMeRequestedForbidden {
	return &GetUsersMeRequestedForbidden{}
}

// WithPayload adds the payload to the get users me requested forbidden response
func (o *GetUsersMeRequestedForbidden) WithPayload(payload *models.Error) *GetUsersMeRequestedForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users me requested forbidden response
func (o *GetUsersMeRequestedForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersMeRequestedForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}