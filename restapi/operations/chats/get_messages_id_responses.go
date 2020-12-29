// Code generated by go-swagger; DO NOT EDIT.

package chats

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// GetMessagesIDOKCode is the HTTP code returned for type GetMessagesIDOK
const GetMessagesIDOKCode int = 200

/*GetMessagesIDOK Message data

swagger:response getMessagesIdOK
*/
type GetMessagesIDOK struct {

	/*
	  In: Body
	*/
	Payload *models.Message `json:"body,omitempty"`
}

// NewGetMessagesIDOK creates GetMessagesIDOK with default headers values
func NewGetMessagesIDOK() *GetMessagesIDOK {

	return &GetMessagesIDOK{}
}

// WithPayload adds the payload to the get messages Id o k response
func (o *GetMessagesIDOK) WithPayload(payload *models.Message) *GetMessagesIDOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get messages Id o k response
func (o *GetMessagesIDOK) SetPayload(payload *models.Message) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetMessagesIDOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetMessagesIDForbiddenCode is the HTTP code returned for type GetMessagesIDForbidden
const GetMessagesIDForbiddenCode int = 403

/*GetMessagesIDForbidden access denied

swagger:response getMessagesIdForbidden
*/
type GetMessagesIDForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetMessagesIDForbidden creates GetMessagesIDForbidden with default headers values
func NewGetMessagesIDForbidden() *GetMessagesIDForbidden {

	return &GetMessagesIDForbidden{}
}

// WithPayload adds the payload to the get messages Id forbidden response
func (o *GetMessagesIDForbidden) WithPayload(payload *models.Error) *GetMessagesIDForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get messages Id forbidden response
func (o *GetMessagesIDForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetMessagesIDForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetMessagesIDNotFoundCode is the HTTP code returned for type GetMessagesIDNotFound
const GetMessagesIDNotFoundCode int = 404

/*GetMessagesIDNotFound Message not found

swagger:response getMessagesIdNotFound
*/
type GetMessagesIDNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetMessagesIDNotFound creates GetMessagesIDNotFound with default headers values
func NewGetMessagesIDNotFound() *GetMessagesIDNotFound {

	return &GetMessagesIDNotFound{}
}

// WithPayload adds the payload to the get messages Id not found response
func (o *GetMessagesIDNotFound) WithPayload(payload *models.Error) *GetMessagesIDNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get messages Id not found response
func (o *GetMessagesIDNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetMessagesIDNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
