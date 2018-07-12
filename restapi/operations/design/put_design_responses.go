// Code generated by go-swagger; DO NOT EDIT.

package design

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/sevings/mindwell-server/models"
)

// PutDesignOKCode is the HTTP code returned for type PutDesignOK
const PutDesignOKCode int = 200

/*PutDesignOK Design of your tlog

swagger:response putDesignOK
*/
type PutDesignOK struct {

	/*
	  In: Body
	*/
	Payload *models.Design `json:"body,omitempty"`
}

// NewPutDesignOK creates PutDesignOK with default headers values
func NewPutDesignOK() *PutDesignOK {

	return &PutDesignOK{}
}

// WithPayload adds the payload to the put design o k response
func (o *PutDesignOK) WithPayload(payload *models.Design) *PutDesignOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put design o k response
func (o *PutDesignOK) SetPayload(payload *models.Design) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutDesignOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
