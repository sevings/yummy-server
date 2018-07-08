// Code generated by go-swagger; DO NOT EDIT.

package me

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// GetMeFavoritesOKCode is the HTTP code returned for type GetMeFavoritesOK
const GetMeFavoritesOKCode int = 200

/*GetMeFavoritesOK Entry list

swagger:response getMeFavoritesOK
*/
type GetMeFavoritesOK struct {

	/*
	  In: Body
	*/
	Payload *models.Feed `json:"body,omitempty"`
}

// NewGetMeFavoritesOK creates GetMeFavoritesOK with default headers values
func NewGetMeFavoritesOK() *GetMeFavoritesOK {
	return &GetMeFavoritesOK{}
}

// WithPayload adds the payload to the get me favorites o k response
func (o *GetMeFavoritesOK) WithPayload(payload *models.Feed) *GetMeFavoritesOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get me favorites o k response
func (o *GetMeFavoritesOK) SetPayload(payload *models.Feed) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetMeFavoritesOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}