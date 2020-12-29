// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// GetUsersNameFollowingsOKCode is the HTTP code returned for type GetUsersNameFollowingsOK
const GetUsersNameFollowingsOKCode int = 200

/*GetUsersNameFollowingsOK User list

swagger:response getUsersNameFollowingsOK
*/
type GetUsersNameFollowingsOK struct {

	/*
	  In: Body
	*/
	Payload *models.FriendList `json:"body,omitempty"`
}

// NewGetUsersNameFollowingsOK creates GetUsersNameFollowingsOK with default headers values
func NewGetUsersNameFollowingsOK() *GetUsersNameFollowingsOK {

	return &GetUsersNameFollowingsOK{}
}

// WithPayload adds the payload to the get users name followings o k response
func (o *GetUsersNameFollowingsOK) WithPayload(payload *models.FriendList) *GetUsersNameFollowingsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users name followings o k response
func (o *GetUsersNameFollowingsOK) SetPayload(payload *models.FriendList) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersNameFollowingsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetUsersNameFollowingsForbiddenCode is the HTTP code returned for type GetUsersNameFollowingsForbidden
const GetUsersNameFollowingsForbiddenCode int = 403

/*GetUsersNameFollowingsForbidden access denied

swagger:response getUsersNameFollowingsForbidden
*/
type GetUsersNameFollowingsForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetUsersNameFollowingsForbidden creates GetUsersNameFollowingsForbidden with default headers values
func NewGetUsersNameFollowingsForbidden() *GetUsersNameFollowingsForbidden {

	return &GetUsersNameFollowingsForbidden{}
}

// WithPayload adds the payload to the get users name followings forbidden response
func (o *GetUsersNameFollowingsForbidden) WithPayload(payload *models.Error) *GetUsersNameFollowingsForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users name followings forbidden response
func (o *GetUsersNameFollowingsForbidden) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersNameFollowingsForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetUsersNameFollowingsNotFoundCode is the HTTP code returned for type GetUsersNameFollowingsNotFound
const GetUsersNameFollowingsNotFoundCode int = 404

/*GetUsersNameFollowingsNotFound User not found

swagger:response getUsersNameFollowingsNotFound
*/
type GetUsersNameFollowingsNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetUsersNameFollowingsNotFound creates GetUsersNameFollowingsNotFound with default headers values
func NewGetUsersNameFollowingsNotFound() *GetUsersNameFollowingsNotFound {

	return &GetUsersNameFollowingsNotFound{}
}

// WithPayload adds the payload to the get users name followings not found response
func (o *GetUsersNameFollowingsNotFound) WithPayload(payload *models.Error) *GetUsersNameFollowingsNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get users name followings not found response
func (o *GetUsersNameFollowingsNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetUsersNameFollowingsNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
