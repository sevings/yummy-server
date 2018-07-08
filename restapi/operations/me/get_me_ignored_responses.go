// Code generated by go-swagger; DO NOT EDIT.

package me

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// GetMeIgnoredOKCode is the HTTP code returned for type GetMeIgnoredOK
const GetMeIgnoredOKCode int = 200

/*GetMeIgnoredOK User list

swagger:response getMeIgnoredOK
*/
type GetMeIgnoredOK struct {

	/*
	  In: Body
	*/
	Payload *models.FriendList `json:"body,omitempty"`
}

// NewGetMeIgnoredOK creates GetMeIgnoredOK with default headers values
func NewGetMeIgnoredOK() *GetMeIgnoredOK {
	return &GetMeIgnoredOK{}
}

// WithPayload adds the payload to the get me ignored o k response
func (o *GetMeIgnoredOK) WithPayload(payload *models.FriendList) *GetMeIgnoredOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get me ignored o k response
func (o *GetMeIgnoredOK) SetPayload(payload *models.FriendList) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetMeIgnoredOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}