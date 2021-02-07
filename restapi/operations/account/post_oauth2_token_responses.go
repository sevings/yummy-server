// Code generated by go-swagger; DO NOT EDIT.

package account

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
)

// PostOauth2TokenOKCode is the HTTP code returned for type PostOauth2TokenOK
const PostOauth2TokenOKCode int = 200

/*PostOauth2TokenOK authorized

swagger:response postOauth2TokenOK
*/
type PostOauth2TokenOK struct {

	/*
	  In: Body
	*/
	Payload *PostOauth2TokenOKBody `json:"body,omitempty"`
}

// NewPostOauth2TokenOK creates PostOauth2TokenOK with default headers values
func NewPostOauth2TokenOK() *PostOauth2TokenOK {

	return &PostOauth2TokenOK{}
}

// WithPayload adds the payload to the post oauth2 token o k response
func (o *PostOauth2TokenOK) WithPayload(payload *PostOauth2TokenOKBody) *PostOauth2TokenOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post oauth2 token o k response
func (o *PostOauth2TokenOK) SetPayload(payload *PostOauth2TokenOKBody) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostOauth2TokenOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PostOauth2TokenBadRequestCode is the HTTP code returned for type PostOauth2TokenBadRequest
const PostOauth2TokenBadRequestCode int = 400

/*PostOauth2TokenBadRequest some error happened

swagger:response postOauth2TokenBadRequest
*/
type PostOauth2TokenBadRequest struct {

	/*
	  In: Body
	*/
	Payload *PostOauth2TokenBadRequestBody `json:"body,omitempty"`
}

// NewPostOauth2TokenBadRequest creates PostOauth2TokenBadRequest with default headers values
func NewPostOauth2TokenBadRequest() *PostOauth2TokenBadRequest {

	return &PostOauth2TokenBadRequest{}
}

// WithPayload adds the payload to the post oauth2 token bad request response
func (o *PostOauth2TokenBadRequest) WithPayload(payload *PostOauth2TokenBadRequestBody) *PostOauth2TokenBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post oauth2 token bad request response
func (o *PostOauth2TokenBadRequest) SetPayload(payload *PostOauth2TokenBadRequestBody) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostOauth2TokenBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PostOauth2TokenUnauthorizedCode is the HTTP code returned for type PostOauth2TokenUnauthorized
const PostOauth2TokenUnauthorizedCode int = 401

/*PostOauth2TokenUnauthorized invalid client

swagger:response postOauth2TokenUnauthorized
*/
type PostOauth2TokenUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *PostOauth2TokenUnauthorizedBody `json:"body,omitempty"`
}

// NewPostOauth2TokenUnauthorized creates PostOauth2TokenUnauthorized with default headers values
func NewPostOauth2TokenUnauthorized() *PostOauth2TokenUnauthorized {

	return &PostOauth2TokenUnauthorized{}
}

// WithPayload adds the payload to the post oauth2 token unauthorized response
func (o *PostOauth2TokenUnauthorized) WithPayload(payload *PostOauth2TokenUnauthorizedBody) *PostOauth2TokenUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post oauth2 token unauthorized response
func (o *PostOauth2TokenUnauthorized) SetPayload(payload *PostOauth2TokenUnauthorizedBody) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostOauth2TokenUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
