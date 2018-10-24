// Code generated by go-swagger; DO NOT EDIT.

package votes

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
	models "github.com/sevings/mindwell-server/models"
)

// DeleteEntriesIDVoteOKCode is the HTTP code returned for type DeleteEntriesIDVoteOK
const DeleteEntriesIDVoteOKCode int = 200

/*DeleteEntriesIDVoteOK vote status

swagger:response deleteEntriesIdVoteOK
*/
type DeleteEntriesIDVoteOK struct {

	/*
	  In: Body
	*/
	Payload *models.Rating `json:"body,omitempty"`
}

// NewDeleteEntriesIDVoteOK creates DeleteEntriesIDVoteOK with default headers values
func NewDeleteEntriesIDVoteOK() *DeleteEntriesIDVoteOK {

	return &DeleteEntriesIDVoteOK{}
}

// WithPayload adds the payload to the delete entries Id vote o k response
func (o *DeleteEntriesIDVoteOK) WithPayload(payload *models.Rating) *DeleteEntriesIDVoteOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete entries Id vote o k response
func (o *DeleteEntriesIDVoteOK) SetPayload(payload *models.Rating) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteEntriesIDVoteOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeleteEntriesIDVoteForbiddenCode is the HTTP code returned for type DeleteEntriesIDVoteForbidden
const DeleteEntriesIDVoteForbiddenCode int = 403

/*DeleteEntriesIDVoteForbidden access denied

swagger:response deleteEntriesIdVoteForbidden
*/
type DeleteEntriesIDVoteForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteEntriesIDVoteForbidden creates DeleteEntriesIDVoteForbidden with default headers values
func NewDeleteEntriesIDVoteForbidden() *DeleteEntriesIDVoteForbidden {

	return &DeleteEntriesIDVoteForbidden{}
}

// WithPayload adds the payload to the delete entries Id vote forbidden response
func (o *DeleteEntriesIDVoteForbidden) WithPayload(payload *models.Error) *DeleteEntriesIDVoteForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete entries Id vote forbidden response
func (o *DeleteEntriesIDVoteForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteEntriesIDVoteForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeleteEntriesIDVoteNotFoundCode is the HTTP code returned for type DeleteEntriesIDVoteNotFound
const DeleteEntriesIDVoteNotFoundCode int = 404

/*DeleteEntriesIDVoteNotFound Entry not found

swagger:response deleteEntriesIdVoteNotFound
*/
type DeleteEntriesIDVoteNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteEntriesIDVoteNotFound creates DeleteEntriesIDVoteNotFound with default headers values
func NewDeleteEntriesIDVoteNotFound() *DeleteEntriesIDVoteNotFound {

	return &DeleteEntriesIDVoteNotFound{}
}

// WithPayload adds the payload to the delete entries Id vote not found response
func (o *DeleteEntriesIDVoteNotFound) WithPayload(payload *models.Error) *DeleteEntriesIDVoteNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete entries Id vote not found response
func (o *DeleteEntriesIDVoteNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteEntriesIDVoteNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
