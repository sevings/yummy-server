// Code generated by go-swagger; DO NOT EDIT.

package images

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// DeleteImagesIDNoContentCode is the HTTP code returned for type DeleteImagesIDNoContent
const DeleteImagesIDNoContentCode int = 204

/*DeleteImagesIDNoContent OK

swagger:response deleteImagesIdNoContent
*/
type DeleteImagesIDNoContent struct {
}

// NewDeleteImagesIDNoContent creates DeleteImagesIDNoContent with default headers values
func NewDeleteImagesIDNoContent() *DeleteImagesIDNoContent {

	return &DeleteImagesIDNoContent{}
}

// WriteResponse to the client
func (o *DeleteImagesIDNoContent) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(204)
}

// DeleteImagesIDForbiddenCode is the HTTP code returned for type DeleteImagesIDForbidden
const DeleteImagesIDForbiddenCode int = 403

/*DeleteImagesIDForbidden access denied

swagger:response deleteImagesIdForbidden
*/
type DeleteImagesIDForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteImagesIDForbidden creates DeleteImagesIDForbidden with default headers values
func NewDeleteImagesIDForbidden() *DeleteImagesIDForbidden {

	return &DeleteImagesIDForbidden{}
}

// WithPayload adds the payload to the delete images Id forbidden response
func (o *DeleteImagesIDForbidden) WithPayload(payload *models.Error) *DeleteImagesIDForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete images Id forbidden response
func (o *DeleteImagesIDForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteImagesIDForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeleteImagesIDNotFoundCode is the HTTP code returned for type DeleteImagesIDNotFound
const DeleteImagesIDNotFoundCode int = 404

/*DeleteImagesIDNotFound Image not found

swagger:response deleteImagesIdNotFound
*/
type DeleteImagesIDNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteImagesIDNotFound creates DeleteImagesIDNotFound with default headers values
func NewDeleteImagesIDNotFound() *DeleteImagesIDNotFound {

	return &DeleteImagesIDNotFound{}
}

// WithPayload adds the payload to the delete images Id not found response
func (o *DeleteImagesIDNotFound) WithPayload(payload *models.Error) *DeleteImagesIDNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete images Id not found response
func (o *DeleteImagesIDNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteImagesIDNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
