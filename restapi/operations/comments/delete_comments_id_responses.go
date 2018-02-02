// Code generated by go-swagger; DO NOT EDIT.

package comments

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/yummy-server/models"
)

// DeleteCommentsIDOKCode is the HTTP code returned for type DeleteCommentsIDOK
const DeleteCommentsIDOKCode int = 200

/*DeleteCommentsIDOK OK

swagger:response deleteCommentsIdOK
*/
type DeleteCommentsIDOK struct {
}

// NewDeleteCommentsIDOK creates DeleteCommentsIDOK with default headers values
func NewDeleteCommentsIDOK() *DeleteCommentsIDOK {
	return &DeleteCommentsIDOK{}
}

// WriteResponse to the client
func (o *DeleteCommentsIDOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

// DeleteCommentsIDForbiddenCode is the HTTP code returned for type DeleteCommentsIDForbidden
const DeleteCommentsIDForbiddenCode int = 403

/*DeleteCommentsIDForbidden access denied

swagger:response deleteCommentsIdForbidden
*/
type DeleteCommentsIDForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteCommentsIDForbidden creates DeleteCommentsIDForbidden with default headers values
func NewDeleteCommentsIDForbidden() *DeleteCommentsIDForbidden {
	return &DeleteCommentsIDForbidden{}
}

// WithPayload adds the payload to the delete comments Id forbidden response
func (o *DeleteCommentsIDForbidden) WithPayload(payload *models.Error) *DeleteCommentsIDForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete comments Id forbidden response
func (o *DeleteCommentsIDForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteCommentsIDForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeleteCommentsIDNotFoundCode is the HTTP code returned for type DeleteCommentsIDNotFound
const DeleteCommentsIDNotFoundCode int = 404

/*DeleteCommentsIDNotFound Comment not found

swagger:response deleteCommentsIdNotFound
*/
type DeleteCommentsIDNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteCommentsIDNotFound creates DeleteCommentsIDNotFound with default headers values
func NewDeleteCommentsIDNotFound() *DeleteCommentsIDNotFound {
	return &DeleteCommentsIDNotFound{}
}

// WithPayload adds the payload to the delete comments Id not found response
func (o *DeleteCommentsIDNotFound) WithPayload(payload *models.Error) *DeleteCommentsIDNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete comments Id not found response
func (o *DeleteCommentsIDNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteCommentsIDNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}