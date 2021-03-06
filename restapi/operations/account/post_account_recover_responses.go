// Code generated by go-swagger; DO NOT EDIT.

package account

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// PostAccountRecoverOKCode is the HTTP code returned for type PostAccountRecoverOK
const PostAccountRecoverOKCode int = 200

/*PostAccountRecoverOK OK

swagger:response postAccountRecoverOK
*/
type PostAccountRecoverOK struct {
}

// NewPostAccountRecoverOK creates PostAccountRecoverOK with default headers values
func NewPostAccountRecoverOK() *PostAccountRecoverOK {

	return &PostAccountRecoverOK{}
}

// WriteResponse to the client
func (o *PostAccountRecoverOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

// PostAccountRecoverBadRequestCode is the HTTP code returned for type PostAccountRecoverBadRequest
const PostAccountRecoverBadRequestCode int = 400

/*PostAccountRecoverBadRequest Email not found

swagger:response postAccountRecoverBadRequest
*/
type PostAccountRecoverBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPostAccountRecoverBadRequest creates PostAccountRecoverBadRequest with default headers values
func NewPostAccountRecoverBadRequest() *PostAccountRecoverBadRequest {

	return &PostAccountRecoverBadRequest{}
}

// WithPayload adds the payload to the post account recover bad request response
func (o *PostAccountRecoverBadRequest) WithPayload(payload *models.Error) *PostAccountRecoverBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post account recover bad request response
func (o *PostAccountRecoverBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostAccountRecoverBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
