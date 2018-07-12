// Code generated by go-swagger; DO NOT EDIT.

package relations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/sevings/mindwell-server/models"
)

// GetRelationsToNameOKCode is the HTTP code returned for type GetRelationsToNameOK
const GetRelationsToNameOKCode int = 200

/*GetRelationsToNameOK your relationship with the user

swagger:response getRelationsToNameOK
*/
type GetRelationsToNameOK struct {

	/*
	  In: Body
	*/
	Payload *models.Relationship `json:"body,omitempty"`
}

// NewGetRelationsToNameOK creates GetRelationsToNameOK with default headers values
func NewGetRelationsToNameOK() *GetRelationsToNameOK {

	return &GetRelationsToNameOK{}
}

// WithPayload adds the payload to the get relations to name o k response
func (o *GetRelationsToNameOK) WithPayload(payload *models.Relationship) *GetRelationsToNameOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get relations to name o k response
func (o *GetRelationsToNameOK) SetPayload(payload *models.Relationship) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetRelationsToNameOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetRelationsToNameNotFoundCode is the HTTP code returned for type GetRelationsToNameNotFound
const GetRelationsToNameNotFoundCode int = 404

/*GetRelationsToNameNotFound User not found

swagger:response getRelationsToNameNotFound
*/
type GetRelationsToNameNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetRelationsToNameNotFound creates GetRelationsToNameNotFound with default headers values
func NewGetRelationsToNameNotFound() *GetRelationsToNameNotFound {

	return &GetRelationsToNameNotFound{}
}

// WithPayload adds the payload to the get relations to name not found response
func (o *GetRelationsToNameNotFound) WithPayload(payload *models.Error) *GetRelationsToNameNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get relations to name not found response
func (o *GetRelationsToNameNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetRelationsToNameNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
