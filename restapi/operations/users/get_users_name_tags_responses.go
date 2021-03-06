// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// GetUsersNameTagsOKCode is the HTTP code returned for type GetUsersNameTagsOK
const GetUsersNameTagsOKCode int = 200

/*GetUsersNameTagsOK user tags

swagger:response getUsersNameTagsOK
*/
type GetUsersNameTagsOK struct {

	/*
	  In: Body
	*/
	Payload *models.TagList `json:"body,omitempty"`
}

// NewGetUsersNameTagsOK creates GetUsersNameTagsOK with default headers values
func NewGetUsersNameTagsOK() *GetUsersNameTagsOK {

	return &GetUsersNameTagsOK{}
}

// WithPayload adds the payload to the get users name tags o k response
func (o *GetUsersNameTagsOK) WithPayload(payload *models.TagList) *GetUsersNameTagsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users name tags o k response
func (o *GetUsersNameTagsOK) SetPayload(payload *models.TagList) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersNameTagsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetUsersNameTagsNotFoundCode is the HTTP code returned for type GetUsersNameTagsNotFound
const GetUsersNameTagsNotFoundCode int = 404

/*GetUsersNameTagsNotFound User not found

swagger:response getUsersNameTagsNotFound
*/
type GetUsersNameTagsNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetUsersNameTagsNotFound creates GetUsersNameTagsNotFound with default headers values
func NewGetUsersNameTagsNotFound() *GetUsersNameTagsNotFound {

	return &GetUsersNameTagsNotFound{}
}

// WithPayload adds the payload to the get users name tags not found response
func (o *GetUsersNameTagsNotFound) WithPayload(payload *models.Error) *GetUsersNameTagsNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users name tags not found response
func (o *GetUsersNameTagsNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersNameTagsNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
