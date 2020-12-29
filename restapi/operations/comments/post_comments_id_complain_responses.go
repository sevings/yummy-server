// Code generated by go-swagger; DO NOT EDIT.

package comments

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// PostCommentsIDComplainNoContentCode is the HTTP code returned for type PostCommentsIDComplainNoContent
const PostCommentsIDComplainNoContentCode int = 204

/*PostCommentsIDComplainNoContent OK

swagger:response postCommentsIdComplainNoContent
*/
type PostCommentsIDComplainNoContent struct {
}

// NewPostCommentsIDComplainNoContent creates PostCommentsIDComplainNoContent with default headers values
func NewPostCommentsIDComplainNoContent() *PostCommentsIDComplainNoContent {

	return &PostCommentsIDComplainNoContent{}
}

// WriteResponse to the client
func (o *PostCommentsIDComplainNoContent) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(204)
}

// PostCommentsIDComplainForbiddenCode is the HTTP code returned for type PostCommentsIDComplainForbidden
const PostCommentsIDComplainForbiddenCode int = 403

/*PostCommentsIDComplainForbidden access denied

swagger:response postCommentsIdComplainForbidden
*/
type PostCommentsIDComplainForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPostCommentsIDComplainForbidden creates PostCommentsIDComplainForbidden with default headers values
func NewPostCommentsIDComplainForbidden() *PostCommentsIDComplainForbidden {

	return &PostCommentsIDComplainForbidden{}
}

// WithPayload adds the payload to the post comments Id complain forbidden response
func (o *PostCommentsIDComplainForbidden) WithPayload(payload *models.Error) *PostCommentsIDComplainForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post comments Id complain forbidden response
func (o *PostCommentsIDComplainForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostCommentsIDComplainForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PostCommentsIDComplainNotFoundCode is the HTTP code returned for type PostCommentsIDComplainNotFound
const PostCommentsIDComplainNotFoundCode int = 404

/*PostCommentsIDComplainNotFound Comment not found

swagger:response postCommentsIdComplainNotFound
*/
type PostCommentsIDComplainNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPostCommentsIDComplainNotFound creates PostCommentsIDComplainNotFound with default headers values
func NewPostCommentsIDComplainNotFound() *PostCommentsIDComplainNotFound {

	return &PostCommentsIDComplainNotFound{}
}

// WithPayload adds the payload to the post comments Id complain not found response
func (o *PostCommentsIDComplainNotFound) WithPayload(payload *models.Error) *PostCommentsIDComplainNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post comments Id complain not found response
func (o *PostCommentsIDComplainNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostCommentsIDComplainNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
