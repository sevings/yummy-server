// Code generated by go-swagger; DO NOT EDIT.

package comments

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/yummy-server/models"
)

// PutCommentsIDOKCode is the HTTP code returned for type PutCommentsIDOK
const PutCommentsIDOKCode int = 200

/*PutCommentsIDOK Comment data

swagger:response putCommentsIdOK
*/
type PutCommentsIDOK struct {

	/*
	  In: Body
	*/
	Payload *models.Comment `json:"body,omitempty"`
}

// NewPutCommentsIDOK creates PutCommentsIDOK with default headers values
func NewPutCommentsIDOK() *PutCommentsIDOK {
	return &PutCommentsIDOK{}
}

// WithPayload adds the payload to the put comments Id o k response
func (o *PutCommentsIDOK) WithPayload(payload *models.Comment) *PutCommentsIDOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put comments Id o k response
func (o *PutCommentsIDOK) SetPayload(payload *models.Comment) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutCommentsIDOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PutCommentsIDForbiddenCode is the HTTP code returned for type PutCommentsIDForbidden
const PutCommentsIDForbiddenCode int = 403

/*PutCommentsIDForbidden access denied

swagger:response putCommentsIdForbidden
*/
type PutCommentsIDForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPutCommentsIDForbidden creates PutCommentsIDForbidden with default headers values
func NewPutCommentsIDForbidden() *PutCommentsIDForbidden {
	return &PutCommentsIDForbidden{}
}

// WithPayload adds the payload to the put comments Id forbidden response
func (o *PutCommentsIDForbidden) WithPayload(payload *models.Error) *PutCommentsIDForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put comments Id forbidden response
func (o *PutCommentsIDForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutCommentsIDForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PutCommentsIDNotFoundCode is the HTTP code returned for type PutCommentsIDNotFound
const PutCommentsIDNotFoundCode int = 404

/*PutCommentsIDNotFound Comment not found

swagger:response putCommentsIdNotFound
*/
type PutCommentsIDNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPutCommentsIDNotFound creates PutCommentsIDNotFound with default headers values
func NewPutCommentsIDNotFound() *PutCommentsIDNotFound {
	return &PutCommentsIDNotFound{}
}

// WithPayload adds the payload to the put comments Id not found response
func (o *PutCommentsIDNotFound) WithPayload(payload *models.Error) *PutCommentsIDNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put comments Id not found response
func (o *PutCommentsIDNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutCommentsIDNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}