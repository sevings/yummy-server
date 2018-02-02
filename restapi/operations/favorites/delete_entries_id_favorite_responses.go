// Code generated by go-swagger; DO NOT EDIT.

package favorites

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/yummy-server/models"
)

// DeleteEntriesIDFavoriteOKCode is the HTTP code returned for type DeleteEntriesIDFavoriteOK
const DeleteEntriesIDFavoriteOKCode int = 200

/*DeleteEntriesIDFavoriteOK favorite status

swagger:response deleteEntriesIdFavoriteOK
*/
type DeleteEntriesIDFavoriteOK struct {

	/*
	  In: Body
	*/
	Payload *models.FavoriteStatus `json:"body,omitempty"`
}

// NewDeleteEntriesIDFavoriteOK creates DeleteEntriesIDFavoriteOK with default headers values
func NewDeleteEntriesIDFavoriteOK() *DeleteEntriesIDFavoriteOK {
	return &DeleteEntriesIDFavoriteOK{}
}

// WithPayload adds the payload to the delete entries Id favorite o k response
func (o *DeleteEntriesIDFavoriteOK) WithPayload(payload *models.FavoriteStatus) *DeleteEntriesIDFavoriteOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete entries Id favorite o k response
func (o *DeleteEntriesIDFavoriteOK) SetPayload(payload *models.FavoriteStatus) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteEntriesIDFavoriteOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeleteEntriesIDFavoriteForbiddenCode is the HTTP code returned for type DeleteEntriesIDFavoriteForbidden
const DeleteEntriesIDFavoriteForbiddenCode int = 403

/*DeleteEntriesIDFavoriteForbidden access denied

swagger:response deleteEntriesIdFavoriteForbidden
*/
type DeleteEntriesIDFavoriteForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteEntriesIDFavoriteForbidden creates DeleteEntriesIDFavoriteForbidden with default headers values
func NewDeleteEntriesIDFavoriteForbidden() *DeleteEntriesIDFavoriteForbidden {
	return &DeleteEntriesIDFavoriteForbidden{}
}

// WithPayload adds the payload to the delete entries Id favorite forbidden response
func (o *DeleteEntriesIDFavoriteForbidden) WithPayload(payload *models.Error) *DeleteEntriesIDFavoriteForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete entries Id favorite forbidden response
func (o *DeleteEntriesIDFavoriteForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteEntriesIDFavoriteForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeleteEntriesIDFavoriteNotFoundCode is the HTTP code returned for type DeleteEntriesIDFavoriteNotFound
const DeleteEntriesIDFavoriteNotFoundCode int = 404

/*DeleteEntriesIDFavoriteNotFound Entry not found

swagger:response deleteEntriesIdFavoriteNotFound
*/
type DeleteEntriesIDFavoriteNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteEntriesIDFavoriteNotFound creates DeleteEntriesIDFavoriteNotFound with default headers values
func NewDeleteEntriesIDFavoriteNotFound() *DeleteEntriesIDFavoriteNotFound {
	return &DeleteEntriesIDFavoriteNotFound{}
}

// WithPayload adds the payload to the delete entries Id favorite not found response
func (o *DeleteEntriesIDFavoriteNotFound) WithPayload(payload *models.Error) *DeleteEntriesIDFavoriteNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete entries Id favorite not found response
func (o *DeleteEntriesIDFavoriteNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteEntriesIDFavoriteNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}