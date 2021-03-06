// Code generated by go-swagger; DO NOT EDIT.

package watchings

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// DeleteEntriesIDWatchingOKCode is the HTTP code returned for type DeleteEntriesIDWatchingOK
const DeleteEntriesIDWatchingOKCode int = 200

/*DeleteEntriesIDWatchingOK watching status

swagger:response deleteEntriesIdWatchingOK
*/
type DeleteEntriesIDWatchingOK struct {

	/*
	  In: Body
	*/
	Payload *models.WatchingStatus `json:"body,omitempty"`
}

// NewDeleteEntriesIDWatchingOK creates DeleteEntriesIDWatchingOK with default headers values
func NewDeleteEntriesIDWatchingOK() *DeleteEntriesIDWatchingOK {

	return &DeleteEntriesIDWatchingOK{}
}

// WithPayload adds the payload to the delete entries Id watching o k response
func (o *DeleteEntriesIDWatchingOK) WithPayload(payload *models.WatchingStatus) *DeleteEntriesIDWatchingOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete entries Id watching o k response
func (o *DeleteEntriesIDWatchingOK) SetPayload(payload *models.WatchingStatus) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteEntriesIDWatchingOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeleteEntriesIDWatchingNotFoundCode is the HTTP code returned for type DeleteEntriesIDWatchingNotFound
const DeleteEntriesIDWatchingNotFoundCode int = 404

/*DeleteEntriesIDWatchingNotFound Entry not found

swagger:response deleteEntriesIdWatchingNotFound
*/
type DeleteEntriesIDWatchingNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteEntriesIDWatchingNotFound creates DeleteEntriesIDWatchingNotFound with default headers values
func NewDeleteEntriesIDWatchingNotFound() *DeleteEntriesIDWatchingNotFound {

	return &DeleteEntriesIDWatchingNotFound{}
}

// WithPayload adds the payload to the delete entries Id watching not found response
func (o *DeleteEntriesIDWatchingNotFound) WithPayload(payload *models.Error) *DeleteEntriesIDWatchingNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete entries Id watching not found response
func (o *DeleteEntriesIDWatchingNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteEntriesIDWatchingNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
