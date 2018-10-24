// Code generated by go-swagger; DO NOT EDIT.

package votes

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"
	models "github.com/sevings/mindwell-server/models"
)

// DeleteCommentsIDVoteOKCode is the HTTP code returned for type DeleteCommentsIDVoteOK
const DeleteCommentsIDVoteOKCode int = 200

/*DeleteCommentsIDVoteOK vote status

swagger:response deleteCommentsIdVoteOK
*/
type DeleteCommentsIDVoteOK struct {

	/*
	  In: Body
	*/
	Payload *models.Rating `json:"body,omitempty"`
}

// NewDeleteCommentsIDVoteOK creates DeleteCommentsIDVoteOK with default headers values
func NewDeleteCommentsIDVoteOK() *DeleteCommentsIDVoteOK {

	return &DeleteCommentsIDVoteOK{}
}

// WithPayload adds the payload to the delete comments Id vote o k response
func (o *DeleteCommentsIDVoteOK) WithPayload(payload *models.Rating) *DeleteCommentsIDVoteOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete comments Id vote o k response
func (o *DeleteCommentsIDVoteOK) SetPayload(payload *models.Rating) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteCommentsIDVoteOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeleteCommentsIDVoteForbiddenCode is the HTTP code returned for type DeleteCommentsIDVoteForbidden
const DeleteCommentsIDVoteForbiddenCode int = 403

/*DeleteCommentsIDVoteForbidden access denied

swagger:response deleteCommentsIdVoteForbidden
*/
type DeleteCommentsIDVoteForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteCommentsIDVoteForbidden creates DeleteCommentsIDVoteForbidden with default headers values
func NewDeleteCommentsIDVoteForbidden() *DeleteCommentsIDVoteForbidden {

	return &DeleteCommentsIDVoteForbidden{}
}

// WithPayload adds the payload to the delete comments Id vote forbidden response
func (o *DeleteCommentsIDVoteForbidden) WithPayload(payload *models.Error) *DeleteCommentsIDVoteForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete comments Id vote forbidden response
func (o *DeleteCommentsIDVoteForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteCommentsIDVoteForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// DeleteCommentsIDVoteNotFoundCode is the HTTP code returned for type DeleteCommentsIDVoteNotFound
const DeleteCommentsIDVoteNotFoundCode int = 404

/*DeleteCommentsIDVoteNotFound Comment not found

swagger:response deleteCommentsIdVoteNotFound
*/
type DeleteCommentsIDVoteNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteCommentsIDVoteNotFound creates DeleteCommentsIDVoteNotFound with default headers values
func NewDeleteCommentsIDVoteNotFound() *DeleteCommentsIDVoteNotFound {

	return &DeleteCommentsIDVoteNotFound{}
}

// WithPayload adds the payload to the delete comments Id vote not found response
func (o *DeleteCommentsIDVoteNotFound) WithPayload(payload *models.Error) *DeleteCommentsIDVoteNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete comments Id vote not found response
func (o *DeleteCommentsIDVoteNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteCommentsIDVoteNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
