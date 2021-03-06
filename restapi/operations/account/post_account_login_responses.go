// Code generated by go-swagger; DO NOT EDIT.

package account

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// PostAccountLoginOKCode is the HTTP code returned for type PostAccountLoginOK
const PostAccountLoginOKCode int = 200

/*PostAccountLoginOK OK

swagger:response postAccountLoginOK
*/
type PostAccountLoginOK struct {

	/*
	  In: Body
	*/
	Payload *models.AuthProfile `json:"body,omitempty"`
}

// NewPostAccountLoginOK creates PostAccountLoginOK with default headers values
func NewPostAccountLoginOK() *PostAccountLoginOK {

	return &PostAccountLoginOK{}
}

// WithPayload adds the payload to the post account login o k response
func (o *PostAccountLoginOK) WithPayload(payload *models.AuthProfile) *PostAccountLoginOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post account login o k response
func (o *PostAccountLoginOK) SetPayload(payload *models.AuthProfile) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostAccountLoginOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PostAccountLoginBadRequestCode is the HTTP code returned for type PostAccountLoginBadRequest
const PostAccountLoginBadRequestCode int = 400

/*PostAccountLoginBadRequest User not found

swagger:response postAccountLoginBadRequest
*/
type PostAccountLoginBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPostAccountLoginBadRequest creates PostAccountLoginBadRequest with default headers values
func NewPostAccountLoginBadRequest() *PostAccountLoginBadRequest {

	return &PostAccountLoginBadRequest{}
}

// WithPayload adds the payload to the post account login bad request response
func (o *PostAccountLoginBadRequest) WithPayload(payload *models.Error) *PostAccountLoginBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post account login bad request response
func (o *PostAccountLoginBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostAccountLoginBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
