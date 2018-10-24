// Code generated by go-swagger; DO NOT EDIT.

package entries

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
	models "github.com/sevings/mindwell-server/models"
)

// PutEntriesIDOKCode is the HTTP code returned for type PutEntriesIDOK
const PutEntriesIDOKCode int = 200

/*PutEntriesIDOK Entry data

swagger:response putEntriesIdOK
*/
type PutEntriesIDOK struct {

	/*
	  In: Body
	*/
	Payload *models.Entry `json:"body,omitempty"`
}

// NewPutEntriesIDOK creates PutEntriesIDOK with default headers values
func NewPutEntriesIDOK() *PutEntriesIDOK {

	return &PutEntriesIDOK{}
}

// WithPayload adds the payload to the put entries Id o k response
func (o *PutEntriesIDOK) WithPayload(payload *models.Entry) *PutEntriesIDOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put entries Id o k response
func (o *PutEntriesIDOK) SetPayload(payload *models.Entry) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutEntriesIDOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PutEntriesIDForbiddenCode is the HTTP code returned for type PutEntriesIDForbidden
const PutEntriesIDForbiddenCode int = 403

/*PutEntriesIDForbidden access denied or post in live restriction

swagger:response putEntriesIdForbidden
*/
type PutEntriesIDForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPutEntriesIDForbidden creates PutEntriesIDForbidden with default headers values
func NewPutEntriesIDForbidden() *PutEntriesIDForbidden {

	return &PutEntriesIDForbidden{}
}

// WithPayload adds the payload to the put entries Id forbidden response
func (o *PutEntriesIDForbidden) WithPayload(payload *models.Error) *PutEntriesIDForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put entries Id forbidden response
func (o *PutEntriesIDForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutEntriesIDForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PutEntriesIDNotFoundCode is the HTTP code returned for type PutEntriesIDNotFound
const PutEntriesIDNotFoundCode int = 404

/*PutEntriesIDNotFound Entry not found

swagger:response putEntriesIdNotFound
*/
type PutEntriesIDNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPutEntriesIDNotFound creates PutEntriesIDNotFound with default headers values
func NewPutEntriesIDNotFound() *PutEntriesIDNotFound {

	return &PutEntriesIDNotFound{}
}

// WithPayload adds the payload to the put entries Id not found response
func (o *PutEntriesIDNotFound) WithPayload(payload *models.Error) *PutEntriesIDNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put entries Id not found response
func (o *PutEntriesIDNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutEntriesIDNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
