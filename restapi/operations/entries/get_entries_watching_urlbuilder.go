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

// GetEntriesWatchingURL generates an URL for the get entries watching operation
type GetEntriesWatchingURL struct {
	After  *string
	Before *string
	Limit  *int64

	_basePath string
	// avoid unkeyed usage
	_ struct{}
}

// WithBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *GetEntriesWatchingURL) WithBasePath(bp string) *GetEntriesWatchingURL {
	o.SetBasePath(bp)
	return o
}

// SetBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *GetEntriesWatchingURL) SetBasePath(bp string) {
	o._basePath = bp
}

// Build a url path and query string
func (o *GetEntriesWatchingURL) Build() (*url.URL, error) {
	var _result url.URL

	var _path = "/entries/watching"

	_basePath := o._basePath
	if _basePath == "" {
		_basePath = "/api/v1"
	}
	_result.Path = golangswaggerpaths.Join(_basePath, _path)

	qs := make(url.Values)

	var after string
	if o.After != nil {
		after = *o.After
	}
	if after != "" {
		qs.Set("after", after)
	}

	var before string
	if o.Before != nil {
		before = *o.Before
	}
	if before != "" {
		qs.Set("before", before)
	}

	var limit string
	if o.Limit != nil {
		limit = swag.FormatInt64(*o.Limit)
	}
	if limit != "" {
		qs.Set("limit", limit)
	}

	_result.RawQuery = qs.Encode()

	return &_result, nil
}

// Must is a helper function to panic when the url builder returns an error
func (o *GetEntriesWatchingURL) Must(u *url.URL, err error) *url.URL {
	if err != nil {
		panic(err)
	}
	if u == nil {
		panic("url can't be nil")
	}
	return u
}

// String returns the string representation of the path with query string
func (o *GetEntriesWatchingURL) String() string {
	return o.Must(o.Build()).String()
}

// BuildFull builds a full url with scheme, host, path and query string
func (o *GetEntriesWatchingURL) BuildFull(scheme, host string) (*url.URL, error) {
	if scheme == "" {
		return nil, errors.New("scheme is required for a full url on GetEntriesWatchingURL")
	}
	if host == "" {
		return nil, errors.New("host is required for a full url on GetEntriesWatchingURL")
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
func (o *GetEntriesWatchingURL) StringFull(scheme, host string) string {
	return o.Must(o.BuildFull(scheme, host)).String()
}
