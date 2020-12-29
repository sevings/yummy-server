// Code generated by go-swagger; DO NOT EDIT.

package design

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/sevings/mindwell-server/models"
)

// GetDesignOKCode is the HTTP code returned for type GetDesignOK
const GetDesignOKCode int = 200

/*GetDesignOK Design of your tlog

swagger:response getDesignOK
*/
type GetDesignOK struct {

	/*
	  In: Body
	*/
	Payload *models.Design `json:"body,omitempty"`
}

// NewGetDesignOK creates GetDesignOK with default headers values
func NewGetDesignOK() *GetDesignOK {

	return &GetDesignOK{}
}

// WithPayload adds the payload to the get design o k response
func (o *GetDesignOK) WithPayload(payload *models.Design) *GetDesignOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get design o k response
func (o *GetDesignOK) SetPayload(payload *models.Design) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetDesignOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
