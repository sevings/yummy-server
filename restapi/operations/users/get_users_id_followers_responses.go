// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/yummy-server/models"
)

// GetUsersIDFollowersOKCode is the HTTP code returned for type GetUsersIDFollowersOK
const GetUsersIDFollowersOKCode int = 200

/*GetUsersIDFollowersOK User list

swagger:response getUsersIdFollowersOK
*/
type GetUsersIDFollowersOK struct {

	/*
	  In: Body
	*/
	Payload *models.UserList `json:"body,omitempty"`
}

// NewGetUsersIDFollowersOK creates GetUsersIDFollowersOK with default headers values
func NewGetUsersIDFollowersOK() *GetUsersIDFollowersOK {
	return &GetUsersIDFollowersOK{}
}

// WithPayload adds the payload to the get users Id followers o k response
func (o *GetUsersIDFollowersOK) WithPayload(payload *models.UserList) *GetUsersIDFollowersOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users Id followers o k response
func (o *GetUsersIDFollowersOK) SetPayload(payload *models.UserList) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersIDFollowersOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetUsersIDFollowersForbiddenCode is the HTTP code returned for type GetUsersIDFollowersForbidden
const GetUsersIDFollowersForbiddenCode int = 403

/*GetUsersIDFollowersForbidden access denied

swagger:response getUsersIdFollowersForbidden
*/
type GetUsersIDFollowersForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetUsersIDFollowersForbidden creates GetUsersIDFollowersForbidden with default headers values
func NewGetUsersIDFollowersForbidden() *GetUsersIDFollowersForbidden {
	return &GetUsersIDFollowersForbidden{}
}

// WithPayload adds the payload to the get users Id followers forbidden response
func (o *GetUsersIDFollowersForbidden) WithPayload(payload *models.Error) *GetUsersIDFollowersForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users Id followers forbidden response
func (o *GetUsersIDFollowersForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersIDFollowersForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetUsersIDFollowersNotFoundCode is the HTTP code returned for type GetUsersIDFollowersNotFound
const GetUsersIDFollowersNotFoundCode int = 404

/*GetUsersIDFollowersNotFound User not found

swagger:response getUsersIdFollowersNotFound
*/
type GetUsersIDFollowersNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetUsersIDFollowersNotFound creates GetUsersIDFollowersNotFound with default headers values
func NewGetUsersIDFollowersNotFound() *GetUsersIDFollowersNotFound {
	return &GetUsersIDFollowersNotFound{}
}

// WithPayload adds the payload to the get users Id followers not found response
func (o *GetUsersIDFollowersNotFound) WithPayload(payload *models.Error) *GetUsersIDFollowersNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users Id followers not found response
func (o *GetUsersIDFollowersNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersIDFollowersNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}