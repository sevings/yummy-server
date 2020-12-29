// Code generated by go-swagger; DO NOT EDIT.

package adm

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"errors"
	"net/url"
	golangswaggerpaths "path"

	"github.com/go-openapi/swag"
)

// PostAdmGrandsonStatusURL generates an URL for the post adm grandson status operation
type PostAdmGrandsonStatusURL struct {
	Received bool

	_basePath string
	// avoid unkeyed usage
	_ struct{}
}

// WithBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *PostAdmGrandsonStatusURL) WithBasePath(bp string) *PostAdmGrandsonStatusURL {
	o.SetBasePath(bp)
	return o
}

// SetBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *PostAdmGrandsonStatusURL) SetBasePath(bp string) {
	o._basePath = bp
}

// Build a url path and query string
func (o *PostAdmGrandsonStatusURL) Build() (*url.URL, error) {
	var _result url.URL

	var _path = "/adm/grandson/status"

	_basePath := o._basePath
	if _basePath == "" {
		_basePath = "/api/v1"
	}
	_result.Path = golangswaggerpaths.Join(_basePath, _path)

	qs := make(url.Values)

	receivedQ := swag.FormatBool(o.Received)
	if receivedQ != "" {
		qs.Set("received", receivedQ)
	}

	_result.RawQuery = qs.Encode()

	return &_result, nil
}

// Must is a helper function to panic when the url builder returns an error
func (o *PostAdmGrandsonStatusURL) Must(u *url.URL, err error) *url.URL {
	if err != nil {
		panic(err)
	}
	if u == nil {
		panic("url can't be nil")
	}
	return u
}

// String returns the string representation of the path with query string
func (o *PostAdmGrandsonStatusURL) String() string {
	return o.Must(o.Build()).String()
}

// BuildFull builds a full url with scheme, host, path and query string
func (o *PostAdmGrandsonStatusURL) BuildFull(scheme, host string) (*url.URL, error) {
	if scheme == "" {
		return nil, errors.New("scheme is required for a full url on PostAdmGrandsonStatusURL")
	}
	if host == "" {
		return nil, errors.New("host is required for a full url on PostAdmGrandsonStatusURL")
	}

	base, err := o.Build()
	if err != nil {
		return nil, err
	}

	base.Scheme = scheme
	base.Host = host
	return base, nil
}

// StringFull returns the string representation of a complete url
func (o *PostAdmGrandsonStatusURL) StringFull(scheme, host string) string {
	return o.Must(o.BuildFull(scheme, host)).String()
}
