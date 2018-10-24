// Code generated by go-swagger; DO NOT EDIT.

package account

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
	models "github.com/sevings/mindwell-server/models"
)

// PostAccountRecoverPasswordOKCode is the HTTP code returned for type PostAccountRecoverPasswordOK
const PostAccountRecoverPasswordOKCode int = 200

/*PostAccountRecoverPasswordOK Password changed

swagger:response postAccountRecoverPasswordOK
*/
type PostAccountRecoverPasswordOK struct {
}

// NewPostAccountRecoverPasswordOK creates PostAccountRecoverPasswordOK with default headers values
func NewPostAccountRecoverPasswordOK() *PostAccountRecoverPasswordOK {

	return &PostAccountRecoverPasswordOK{}
}

// WriteResponse to the client
func (o *PostAccountRecoverPasswordOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

// PostAccountRecoverPasswordBadRequestCode is the HTTP code returned for type PostAccountRecoverPasswordBadRequest
const PostAccountRecoverPasswordBadRequestCode int = 400

/*PostAccountRecoverPasswordBadRequest Email not found

swagger:response postAccountRecoverPasswordBadRequest
*/
type PostAccountRecoverPasswordBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPostAccountRecoverPasswordBadRequest creates PostAccountRecoverPasswordBadRequest with default headers values
func NewPostAccountRecoverPasswordBadRequest() *PostAccountRecoverPasswordBadRequest {

	return &PostAccountRecoverPasswordBadRequest{}
}

// WithPayload adds the payload to the post account recover password bad request response
func (o *PostAccountRecoverPasswordBadRequest) WithPayload(payload *models.Error) *PostAccountRecoverPasswordBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post account recover password bad request response
func (o *PostAccountRecoverPasswordBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostAccountRecoverPasswordBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
