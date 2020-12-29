// Code generated by go-swagger; DO NOT EDIT.

package relations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// PutRelationsFromNameOKCode is the HTTP code returned for type PutRelationsFromNameOK
const PutRelationsFromNameOKCode int = 200

/*PutRelationsFromNameOK the user relationship with you

swagger:response putRelationsFromNameOK
*/
type PutRelationsFromNameOK struct {

	/*
	  In: Body
	*/
	Payload *models.Relationship `json:"body,omitempty"`
}

// NewPutRelationsFromNameOK creates PutRelationsFromNameOK with default headers values
func NewPutRelationsFromNameOK() *PutRelationsFromNameOK {

	return &PutRelationsFromNameOK{}
}

// WithPayload adds the payload to the put relations from name o k response
func (o *PutRelationsFromNameOK) WithPayload(payload *models.Relationship) *PutRelationsFromNameOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put relations from name o k response
func (o *PutRelationsFromNameOK) SetPayload(payload *models.Relationship) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutRelationsFromNameOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PutRelationsFromNameForbiddenCode is the HTTP code returned for type PutRelationsFromNameForbidden
const PutRelationsFromNameForbiddenCode int = 403

/*PutRelationsFromNameForbidden access denied

swagger:response putRelationsFromNameForbidden
*/
type PutRelationsFromNameForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPutRelationsFromNameForbidden creates PutRelationsFromNameForbidden with default headers values
func NewPutRelationsFromNameForbidden() *PutRelationsFromNameForbidden {

	return &PutRelationsFromNameForbidden{}
}

// WithPayload adds the payload to the put relations from name forbidden response
func (o *PutRelationsFromNameForbidden) WithPayload(payload *models.Error) *PutRelationsFromNameForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put relations from name forbidden response
func (o *PutRelationsFromNameForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutRelationsFromNameForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PutRelationsFromNameNotFoundCode is the HTTP code returned for type PutRelationsFromNameNotFound
const PutRelationsFromNameNotFoundCode int = 404

/*PutRelationsFromNameNotFound User not found

swagger:response putRelationsFromNameNotFound
*/
type PutRelationsFromNameNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPutRelationsFromNameNotFound creates PutRelationsFromNameNotFound with default headers values
func NewPutRelationsFromNameNotFound() *PutRelationsFromNameNotFound {

	return &PutRelationsFromNameNotFound{}
}

// WithPayload adds the payload to the put relations from name not found response
func (o *PutRelationsFromNameNotFound) WithPayload(payload *models.Error) *PutRelationsFromNameNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put relations from name not found response
func (o *PutRelationsFromNameNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutRelationsFromNameNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
