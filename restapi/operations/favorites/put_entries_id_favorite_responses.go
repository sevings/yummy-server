// Code generated by go-swagger; DO NOT EDIT.

package favorites

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/yummy-server/models"
)

// PutEntriesIDFavoriteOKCode is the HTTP code returned for type PutEntriesIDFavoriteOK
const PutEntriesIDFavoriteOKCode int = 200

/*PutEntriesIDFavoriteOK favorite status

swagger:response putEntriesIdFavoriteOK
*/
type PutEntriesIDFavoriteOK struct {

	/*
	  In: Body
	*/
	Payload *models.FavoriteStatus `json:"body,omitempty"`
}

// NewPutEntriesIDFavoriteOK creates PutEntriesIDFavoriteOK with default headers values
func NewPutEntriesIDFavoriteOK() *PutEntriesIDFavoriteOK {
	return &PutEntriesIDFavoriteOK{}
}

// WithPayload adds the payload to the put entries Id favorite o k response
func (o *PutEntriesIDFavoriteOK) WithPayload(payload *models.FavoriteStatus) *PutEntriesIDFavoriteOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put entries Id favorite o k response
func (o *PutEntriesIDFavoriteOK) SetPayload(payload *models.FavoriteStatus) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutEntriesIDFavoriteOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PutEntriesIDFavoriteNotFoundCode is the HTTP code returned for type PutEntriesIDFavoriteNotFound
const PutEntriesIDFavoriteNotFoundCode int = 404

/*PutEntriesIDFavoriteNotFound Entry not found

swagger:response putEntriesIdFavoriteNotFound
*/
type PutEntriesIDFavoriteNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPutEntriesIDFavoriteNotFound creates PutEntriesIDFavoriteNotFound with default headers values
func NewPutEntriesIDFavoriteNotFound() *PutEntriesIDFavoriteNotFound {
	return &PutEntriesIDFavoriteNotFound{}
}

// WithPayload adds the payload to the put entries Id favorite not found response
func (o *PutEntriesIDFavoriteNotFound) WithPayload(payload *models.Error) *PutEntriesIDFavoriteNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put entries Id favorite not found response
func (o *PutEntriesIDFavoriteNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutEntriesIDFavoriteNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
