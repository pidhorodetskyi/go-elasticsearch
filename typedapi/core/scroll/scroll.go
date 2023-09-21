// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Code generated from the elasticsearch-specification DO NOT EDIT.
// https://github.com/elastic/elasticsearch-specification/tree/5260ec5b7c899ab1a7939f752218cae07ef07dd7

// Allows to retrieve a large numbers of results from a single search request.
package scroll

import (
	gobytes "bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

const (
	scrollidMask = iota + 1
)

// ErrBuildPath is returned in case of missing parameters within the build of the request.
var ErrBuildPath = errors.New("cannot build path, check for missing path parameters")

type Scroll struct {
	transport elastictransport.Interface

	headers http.Header
	values  url.Values
	path    url.URL

	buf *gobytes.Buffer

	req      *Request
	deferred []func(request *Request) error
	raw      io.Reader

	paramSet int

	scrollid string
}

// NewScroll type alias for index.
type NewScroll func() *Scroll

// NewScrollFunc returns a new instance of Scroll with the provided transport.
// Used in the index of the library this allows to retrieve every apis in once place.
func NewScrollFunc(tp elastictransport.Interface) NewScroll {
	return func() *Scroll {
		n := New(tp)

		return n
	}
}

// Allows to retrieve a large numbers of results from a single search request.
//
// https://www.elastic.co/guide/en/elasticsearch/reference/master/search-request-body.html#request-body-search-scroll
func New(tp elastictransport.Interface) *Scroll {
	r := &Scroll{
		transport: tp,
		values:    make(url.Values),
		headers:   make(http.Header),
		buf:       gobytes.NewBuffer(nil),

		req: NewRequest(),
	}

	return r
}

// Raw takes a json payload as input which is then passed to the http.Request
// If specified Raw takes precedence on Request method.
func (r *Scroll) Raw(raw io.Reader) *Scroll {
	r.raw = raw

	return r
}

// Request allows to set the request property with the appropriate payload.
func (r *Scroll) Request(req *Request) *Scroll {
	r.req = req

	return r
}

// HttpRequest returns the http.Request object built from the
// given parameters.
func (r *Scroll) HttpRequest(ctx context.Context) (*http.Request, error) {
	var path strings.Builder
	var method string
	var req *http.Request

	var err error

	if len(r.deferred) > 0 {
		for _, f := range r.deferred {
			deferredErr := f(r.req)
			if deferredErr != nil {
				return nil, deferredErr
			}
		}
	}

	if r.raw != nil {
		r.buf.ReadFrom(r.raw)
	} else if r.req != nil {

		data, err := json.Marshal(r.req)

		if err != nil {
			return nil, fmt.Errorf("could not serialise request for Scroll: %w", err)
		}

		r.buf.Write(data)

	}

	r.path.Scheme = "http"

	switch {
	case r.paramSet == 0:
		path.WriteString("/")
		path.WriteString("_search")
		path.WriteString("/")
		path.WriteString("scroll")

		method = http.MethodPost
	case r.paramSet == scrollidMask:
		path.WriteString("/")
		path.WriteString("_search")
		path.WriteString("/")
		path.WriteString("scroll")
		path.WriteString("/")

		path.WriteString(r.scrollid)

		method = http.MethodPost
	}

	r.path.Path = path.String()
	r.path.RawQuery = r.values.Encode()

	if r.path.Path == "" {
		return nil, ErrBuildPath
	}

	if ctx != nil {
		req, err = http.NewRequestWithContext(ctx, method, r.path.String(), r.buf)
	} else {
		req, err = http.NewRequest(method, r.path.String(), r.buf)
	}

	req.Header = r.headers.Clone()

	if req.Header.Get("Content-Type") == "" {
		if r.buf.Len() > 0 {
			req.Header.Set("Content-Type", "application/vnd.elasticsearch+json;compatible-with=8")
		}
	}

	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/vnd.elasticsearch+json;compatible-with=8")
	}

	if err != nil {
		return req, fmt.Errorf("could not build http.Request: %w", err)
	}

	return req, nil
}

// Perform runs the http.Request through the provided transport and returns an http.Response.
func (r Scroll) Perform(ctx context.Context) (*http.Response, error) {
	req, err := r.HttpRequest(ctx)
	if err != nil {
		return nil, err
	}

	res, err := r.transport.Perform(req)
	if err != nil {
		return nil, fmt.Errorf("an error happened during the Scroll query execution: %w", err)
	}

	return res, nil
}

// Do runs the request through the transport, handle the response and returns a scroll.Response
func (r Scroll) Do(ctx context.Context) (*Response, error) {

	response := NewResponse()

	res, err := r.Perform(ctx)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 299 {
		err = json.NewDecoder(res.Body).Decode(response)
		if err != nil {
			return nil, err
		}

		return response, nil
	}

	errorResponse := types.NewElasticsearchError()
	err = json.NewDecoder(res.Body).Decode(errorResponse)
	if err != nil {
		return nil, err
	}

	if errorResponse.Status == 0 {
		errorResponse.Status = res.StatusCode
	}

	return nil, errorResponse
}

// Header set a key, value pair in the Scroll headers map.
func (r *Scroll) Header(key, value string) *Scroll {
	r.headers.Set(key, value)

	return r
}

// ScrollId The scroll ID
// API Name: scrollid
func (r *Scroll) ScrollId(scrollid string) *Scroll {
	r.paramSet |= scrollidMask
	r.scrollid = scrollid

	return r
}

// RestTotalHitsAsInt If true, the API response’s hit.total property is returned as an integer. If
// false, the API response’s hit.total property is returned as an object.
// API name: rest_total_hits_as_int
func (r *Scroll) RestTotalHitsAsInt(resttotalhitsasint bool) *Scroll {
	r.values.Set("rest_total_hits_as_int", strconv.FormatBool(resttotalhitsasint))

	return r
}

// Scroll Period to retain the search context for scrolling.
// API name: scroll
func (r *Scroll) Scroll(duration types.Duration) *Scroll {
	r.req.Scroll = duration

	return r
}
