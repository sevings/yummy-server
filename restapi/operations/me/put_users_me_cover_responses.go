// Code generated by go-swagger; DO NOT EDIT.

package me

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// PutUsersMeCoverOKCode is the HTTP code returned for type PutUsersMeCoverOK
const PutUsersMeCoverOKCode int = 200

/*PutUsersMeCoverOK Cover

swagger:response putUsersMeCoverOK
*/
type PutUsersMeCoverOK struct {

	/*
	  In: Body
	*/
	Payload *models.PutUsersMeCoverOKBody `json:"body,omitempty"`
}

// NewPutUsersMeCoverOK creates PutUsersMeCoverOK with default headers values
func NewPutUsersMeCoverOK() *PutUsersMeCoverOK {
	return &PutUsersMeCoverOK{}
}

// WithPayload adds the payload to the put users me cover o k response
func (o *PutUsersMeCoverOK) WithPayload(payload *models.PutUsersMeCoverOKBody) *PutUsersMeCoverOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put users me cover o k response
func (o *PutUsersMeCoverOK) SetPayload(payload *models.PutUsersMeCoverOKBody) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutUsersMeCoverOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PutUsersMeCoverBadRequestCode is the HTTP code returned for type PutUsersMeCoverBadRequest
const PutUsersMeCoverBadRequestCode int = 400

/*PutUsersMeCoverBadRequest bad request

swagger:response putUsersMeCoverBadRequest
*/
type PutUsersMeCoverBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPutUsersMeCoverBadRequest creates PutUsersMeCoverBadRequest with default headers values
func NewPutUsersMeCoverBadRequest() *PutUsersMeCoverBadRequest {
	return &PutUsersMeCoverBadRequest{}
}

// WithPayload adds the payload to the put users me cover bad request response
func (o *PutUsersMeCoverBadRequest) WithPayload(payload *models.Error) *PutUsersMeCoverBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put users me cover bad request response
func (o *PutUsersMeCoverBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutUsersMeCoverBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
