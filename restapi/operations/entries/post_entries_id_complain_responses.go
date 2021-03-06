// Code generated by go-swagger; DO NOT EDIT.

package entries

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// PostEntriesIDComplainNoContentCode is the HTTP code returned for type PostEntriesIDComplainNoContent
const PostEntriesIDComplainNoContentCode int = 204

/*PostEntriesIDComplainNoContent OK

swagger:response postEntriesIdComplainNoContent
*/
type PostEntriesIDComplainNoContent struct {
}

// NewPostEntriesIDComplainNoContent creates PostEntriesIDComplainNoContent with default headers values
func NewPostEntriesIDComplainNoContent() *PostEntriesIDComplainNoContent {

	return &PostEntriesIDComplainNoContent{}
}

// WriteResponse to the client
func (o *PostEntriesIDComplainNoContent) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(204)
}

// PostEntriesIDComplainForbiddenCode is the HTTP code returned for type PostEntriesIDComplainForbidden
const PostEntriesIDComplainForbiddenCode int = 403

/*PostEntriesIDComplainForbidden access denied

swagger:response postEntriesIdComplainForbidden
*/
type PostEntriesIDComplainForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPostEntriesIDComplainForbidden creates PostEntriesIDComplainForbidden with default headers values
func NewPostEntriesIDComplainForbidden() *PostEntriesIDComplainForbidden {

	return &PostEntriesIDComplainForbidden{}
}

// WithPayload adds the payload to the post entries Id complain forbidden response
func (o *PostEntriesIDComplainForbidden) WithPayload(payload *models.Error) *PostEntriesIDComplainForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post entries Id complain forbidden response
func (o *PostEntriesIDComplainForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostEntriesIDComplainForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PostEntriesIDComplainNotFoundCode is the HTTP code returned for type PostEntriesIDComplainNotFound
const PostEntriesIDComplainNotFoundCode int = 404

/*PostEntriesIDComplainNotFound Entry not found

swagger:response postEntriesIdComplainNotFound
*/
type PostEntriesIDComplainNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPostEntriesIDComplainNotFound creates PostEntriesIDComplainNotFound with default headers values
func NewPostEntriesIDComplainNotFound() *PostEntriesIDComplainNotFound {

	return &PostEntriesIDComplainNotFound{}
}

// WithPayload adds the payload to the post entries Id complain not found response
func (o *PostEntriesIDComplainNotFound) WithPayload(payload *models.Error) *PostEntriesIDComplainNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post entries Id complain not found response
func (o *PostEntriesIDComplainNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostEntriesIDComplainNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
