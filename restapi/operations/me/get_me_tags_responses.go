// Code generated by go-swagger; DO NOT EDIT.

package me

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// GetMeTagsOKCode is the HTTP code returned for type GetMeTagsOK
const GetMeTagsOKCode int = 200

/*GetMeTagsOK my tags

swagger:response getMeTagsOK
*/
type GetMeTagsOK struct {

	/*
	  In: Body
	*/
	Payload *models.TagList `json:"body,omitempty"`
}

// NewGetMeTagsOK creates GetMeTagsOK with default headers values
func NewGetMeTagsOK() *GetMeTagsOK {

	return &GetMeTagsOK{}
}

// WithPayload adds the payload to the get me tags o k response
func (o *GetMeTagsOK) WithPayload(payload *models.TagList) *GetMeTagsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get me tags o k response
func (o *GetMeTagsOK) SetPayload(payload *models.TagList) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetMeTagsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
