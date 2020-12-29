// Code generated by go-swagger; DO NOT EDIT.

package entries

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"errors"
	"net/url"
	golangswaggerpaths "path"

	"github.com/go-openapi/swag"
)

// GetEntriesBestURL generates an URL for the get entries best operation
type GetEntriesBestURL struct {
	Category *string
	Limit    *int64
	Query    *string
	Tag      *string

	_basePath string
	// avoid unkeyed usage
	_ struct{}
}

// WithBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *GetEntriesBestURL) WithBasePath(bp string) *GetEntriesBestURL {
	o.SetBasePath(bp)
	return o
}

// SetBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *GetEntriesBestURL) SetBasePath(bp string) {
	o._basePath = bp
}

// Build a url path and query string
func (o *GetEntriesBestURL) Build() (*url.URL, error) {
	var _result url.URL

	var _path = "/entries/best"

	_basePath := o._basePath
	if _basePath == "" {
		_basePath = "/api/v1"
	}
	_result.Path = golangswaggerpaths.Join(_basePath, _path)

	qs := make(url.Values)

	var categoryQ string
	if o.Category != nil {
		categoryQ = *o.Category
	}
	if categoryQ != "" {
		qs.Set("category", categoryQ)
	}

	var limitQ string
	if o.Limit != nil {
		limitQ = swag.FormatInt64(*o.Limit)
	}
	if limitQ != "" {
		qs.Set("limit", limitQ)
	}

	var queryQ string
	if o.Query != nil {
		queryQ = *o.Query
	}
	if queryQ != "" {
		qs.Set("query", queryQ)
	}

	var tagQ string
	if o.Tag != nil {
		tagQ = *o.Tag
	}
	if tagQ != "" {
		qs.Set("tag", tagQ)
	}

	_result.RawQuery = qs.Encode()

	return &_result, nil
}

// Must is a helper function to panic when the url builder returns an error
func (o *GetEntriesBestURL) Must(u *url.URL, err error) *url.URL {
	if err != nil {
		panic(err)
	}
	if u == nil {
		panic("url can't be nil")
	}
	return u
}

// String returns the string representation of the path with query string
func (o *GetEntriesBestURL) String() string {
	return o.Must(o.Build()).String()
}

// BuildFull builds a full url with scheme, host, path and query string
func (o *GetEntriesBestURL) BuildFull(scheme, host string) (*url.URL, error) {
	if scheme == "" {
		return nil, errors.New("scheme is required for a full url on GetEntriesBestURL")
	}
	if host == "" {
		return nil, errors.New("host is required for a full url on GetEntriesBestURL")
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
func (o *GetEntriesBestURL) StringFull(scheme, host string) string {
	return o.Must(o.BuildFull(scheme, host)).String()
}
