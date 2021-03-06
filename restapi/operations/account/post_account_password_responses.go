// Code generated by go-swagger; DO NOT EDIT.

package account

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// PostAccountPasswordOKCode is the HTTP code returned for type PostAccountPasswordOK
const PostAccountPasswordOKCode int = 200

/*PostAccountPasswordOK password has been set

swagger:response postAccountPasswordOK
*/
type PostAccountPasswordOK struct {
}

// NewPostAccountPasswordOK creates PostAccountPasswordOK with default headers values
func NewPostAccountPasswordOK() *PostAccountPasswordOK {

	return &PostAccountPasswordOK{}
}

// WriteResponse to the client
func (o *PostAccountPasswordOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

// PostAccountPasswordForbiddenCode is the HTTP code returned for type PostAccountPasswordForbidden
const PostAccountPasswordForbiddenCode int = 403

/*PostAccountPasswordForbidden old password is invalid

swagger:response postAccountPasswordForbidden
*/
type PostAccountPasswordForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPostAccountPasswordForbidden creates PostAccountPasswordForbidden with default headers values
func NewPostAccountPasswordForbidden() *PostAccountPasswordForbidden {

	return &PostAccountPasswordForbidden{}
}

// WithPayload adds the payload to the post account password forbidden response
func (o *PostAccountPasswordForbidden) WithPayload(payload *models.Error) *PostAccountPasswordForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post account password forbidden response
func (o *PostAccountPasswordForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostAccountPasswordForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
