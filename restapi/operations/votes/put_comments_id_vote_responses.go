// Code generated by go-swagger; DO NOT EDIT.

package votes

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// PutCommentsIDVoteOKCode is the HTTP code returned for type PutCommentsIDVoteOK
const PutCommentsIDVoteOKCode int = 200

/*PutCommentsIDVoteOK vote status

swagger:response putCommentsIdVoteOK
*/
type PutCommentsIDVoteOK struct {

	/*
	  In: Body
	*/
	Payload *models.Rating `json:"body,omitempty"`
}

// NewPutCommentsIDVoteOK creates PutCommentsIDVoteOK with default headers values
func NewPutCommentsIDVoteOK() *PutCommentsIDVoteOK {

	return &PutCommentsIDVoteOK{}
}

// WithPayload adds the payload to the put comments Id vote o k response
func (o *PutCommentsIDVoteOK) WithPayload(payload *models.Rating) *PutCommentsIDVoteOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put comments Id vote o k response
func (o *PutCommentsIDVoteOK) SetPayload(payload *models.Rating) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutCommentsIDVoteOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PutCommentsIDVoteForbiddenCode is the HTTP code returned for type PutCommentsIDVoteForbidden
const PutCommentsIDVoteForbiddenCode int = 403

/*PutCommentsIDVoteForbidden access denied

swagger:response putCommentsIdVoteForbidden
*/
type PutCommentsIDVoteForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPutCommentsIDVoteForbidden creates PutCommentsIDVoteForbidden with default headers values
func NewPutCommentsIDVoteForbidden() *PutCommentsIDVoteForbidden {

	return &PutCommentsIDVoteForbidden{}
}

// WithPayload adds the payload to the put comments Id vote forbidden response
func (o *PutCommentsIDVoteForbidden) WithPayload(payload *models.Error) *PutCommentsIDVoteForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put comments Id vote forbidden response
func (o *PutCommentsIDVoteForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutCommentsIDVoteForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PutCommentsIDVoteNotFoundCode is the HTTP code returned for type PutCommentsIDVoteNotFound
const PutCommentsIDVoteNotFoundCode int = 404

/*PutCommentsIDVoteNotFound Comment not found

swagger:response putCommentsIdVoteNotFound
*/
type PutCommentsIDVoteNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewPutCommentsIDVoteNotFound creates PutCommentsIDVoteNotFound with default headers values
func NewPutCommentsIDVoteNotFound() *PutCommentsIDVoteNotFound {

	return &PutCommentsIDVoteNotFound{}
}

// WithPayload adds the payload to the put comments Id vote not found response
func (o *PutCommentsIDVoteNotFound) WithPayload(payload *models.Error) *PutCommentsIDVoteNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put comments Id vote not found response
func (o *PutCommentsIDVoteNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutCommentsIDVoteNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
