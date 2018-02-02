// Code generated by go-swagger; DO NOT EDIT.

package entries

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/yummy-server/models"
)

// DeleteEntriesIDOKCode is the HTTP code returned for type DeleteEntriesIDOK
const DeleteEntriesIDOKCode int = 200

/*DeleteEntriesIDOK OK

swagger:response deleteEntriesIdOK
*/
type DeleteEntriesIDOK struct {
}

// NewDeleteEntriesIDOK creates DeleteEntriesIDOK with default headers values
func NewDeleteEntriesIDOK() *DeleteEntriesIDOK {
	return &DeleteEntriesIDOK{}
}

// WriteResponse to the client
func (o *DeleteEntriesIDOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

// DeleteEntriesIDForbiddenCode is the HTTP code returned for type DeleteEntriesIDForbidden
const DeleteEntriesIDForbiddenCode int = 403

/*DeleteEntriesIDForbidden access denied

swagger:response deleteEntriesIdForbidden
*/
type DeleteEntriesIDForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteEntriesIDForbidden creates DeleteEntriesIDForbidden with default headers values
func NewDeleteEntriesIDForbidden() *DeleteEntriesIDForbidden {
	return &DeleteEntriesIDForbidden{}
}

// WithPayload adds the payload to the delete entries Id forbidden response
func (o *DeleteEntriesIDForbidden) WithPayload(payload *models.Error) *DeleteEntriesIDForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete entries Id forbidden response
func (o *DeleteEntriesIDForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteEntriesIDForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeleteEntriesIDNotFoundCode is the HTTP code returned for type DeleteEntriesIDNotFound
const DeleteEntriesIDNotFoundCode int = 404

/*DeleteEntriesIDNotFound Entry not found

swagger:response deleteEntriesIdNotFound
*/
type DeleteEntriesIDNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteEntriesIDNotFound creates DeleteEntriesIDNotFound with default headers values
func NewDeleteEntriesIDNotFound() *DeleteEntriesIDNotFound {
	return &DeleteEntriesIDNotFound{}
}

// WithPayload adds the payload to the delete entries Id not found response
func (o *DeleteEntriesIDNotFound) WithPayload(payload *models.Error) *DeleteEntriesIDNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete entries Id not found response
func (o *DeleteEntriesIDNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteEntriesIDNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}