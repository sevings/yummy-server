// Code generated by go-swagger; DO NOT EDIT.

package relations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/yummy-server/models"
)

// DeleteRelationsToIDOKCode is the HTTP code returned for type DeleteRelationsToIDOK
const DeleteRelationsToIDOKCode int = 200

/*DeleteRelationsToIDOK your relationship with the user

swagger:response deleteRelationsToIdOK
*/
type DeleteRelationsToIDOK struct {

	/*
	  In: Body
	*/
	Payload *models.Relationship `json:"body,omitempty"`
}

// NewDeleteRelationsToIDOK creates DeleteRelationsToIDOK with default headers values
func NewDeleteRelationsToIDOK() *DeleteRelationsToIDOK {
	return &DeleteRelationsToIDOK{}
}

// WithPayload adds the payload to the delete relations to Id o k response
func (o *DeleteRelationsToIDOK) WithPayload(payload *models.Relationship) *DeleteRelationsToIDOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete relations to Id o k response
func (o *DeleteRelationsToIDOK) SetPayload(payload *models.Relationship) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteRelationsToIDOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeleteRelationsToIDNotFoundCode is the HTTP code returned for type DeleteRelationsToIDNotFound
const DeleteRelationsToIDNotFoundCode int = 404

/*DeleteRelationsToIDNotFound User not found

swagger:response deleteRelationsToIdNotFound
*/
type DeleteRelationsToIDNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteRelationsToIDNotFound creates DeleteRelationsToIDNotFound with default headers values
func NewDeleteRelationsToIDNotFound() *DeleteRelationsToIDNotFound {
	return &DeleteRelationsToIDNotFound{}
}

// WithPayload adds the payload to the delete relations to Id not found response
func (o *DeleteRelationsToIDNotFound) WithPayload(payload *models.Error) *DeleteRelationsToIDNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete relations to Id not found response
func (o *DeleteRelationsToIDNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteRelationsToIDNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
