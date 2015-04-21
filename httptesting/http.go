// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package httptesting

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	gc "gopkg.in/check.v1"

	jc "github.com/juju/testing/checkers"
)

// BodyAsserter represents a function that can assert the correctness of
// a JSON reponse.
type BodyAsserter func(c *gc.C, body json.RawMessage)

// JSONCallParams holds parameters for AssertJSONCall.
// If left empty, some fields will automatically be filled with defaults.
type JSONCallParams struct {
	// Do is used to make the HTTP request.
	// If it is nil, http.DefaultClient.Do will be used.
	// If the body reader implements io.Seeker,
	// req.Body will also implement that interface.
	Do func(req *http.Request) (*http.Response, error)

	// ExpectError holds the error regexp to match
	// against the error returned from the HTTP Do
	// request. If it is empty, the error is expected to be
	// nil.
	ExpectError string

	// Handler holds the handler to use to make the request.
	Handler http.Handler

	// Method holds the HTTP method to use for the call.
	// GET is assumed if this is empty.
	Method string

	// URL holds the URL to pass when making the request.
	URL string

	// JSONBody specifies a JSON value to marshal to use
	// as the body of the request. If this is specified, Body will
	// be ignored and the Content-Type header will
	// be set to application/json. The request
	// body will implement io.Seeker.
	JSONBody interface{}

	// Body holds the body to send in the request.
	Body io.Reader

	// Header specifies the HTTP headers to use when making
	// the request.
	Header http.Header

	// ContentLength specifies the length of the body.
	// It may be zero, in which case the default net/http
	// content-length behaviour will be used.
	ContentLength int64

	// Username, if specified, is used for HTTP basic authentication.
	Username string

	// Password, if specified, is used for HTTP basic authentication.
	Password string

	// ExpectStatus holds the expected HTTP status code.
	// http.StatusOK is assumed if this is zero.
	ExpectStatus int

	// ExpectBody holds the expected JSON body.
	// This may be a function of type BodyAsserter in which case it
	// will be called with the http response body to check the
	// result.
	ExpectBody interface{}

	// Cookies, if specified, are added to the request.
	Cookies []*http.Cookie
}

// AssertJSONCall asserts that when the given handler is called with
// the given parameters, the result is as specified.
func AssertJSONCall(c *gc.C, p JSONCallParams) {
	c.Logf("JSON call, url %q", p.URL)
	if p.ExpectStatus == 0 {
		p.ExpectStatus = http.StatusOK
	}
	rec := DoRequest(c, DoRequestParams{
		Do:            p.Do,
		ExpectError:   p.ExpectError,
		Handler:       p.Handler,
		Method:        p.Method,
		URL:           p.URL,
		Body:          p.Body,
		JSONBody:      p.JSONBody,
		Header:        p.Header,
		ContentLength: p.ContentLength,
		Username:      p.Username,
		Password:      p.Password,
		Cookies:       p.Cookies,
	})
	if p.ExpectError != "" {
		return
	}
	AssertJSONResponse(c, rec, p.ExpectStatus, p.ExpectBody)
}

// AssertJSONResponse asserts that the given response recorder has
// recorded the given HTTP status, response body and content type. If
// expectBody is of type BodyAsserter it will be called with the response
// body to ensure the response is correct.
func AssertJSONResponse(c *gc.C, rec *httptest.ResponseRecorder, expectStatus int, expectBody interface{}) {
	c.Assert(rec.Code, gc.Equals, expectStatus, gc.Commentf("body: %s", rec.Body.Bytes()))

	// Ensure the response includes the expected body.
	if expectBody == nil {
		c.Assert(rec.Body.Bytes(), gc.HasLen, 0)
		return
	}
	c.Assert(rec.Header().Get("Content-Type"), gc.Equals, "application/json")
	if assertBody, ok := expectBody.(BodyAsserter); ok {
		var data json.RawMessage
		err := json.Unmarshal(rec.Body.Bytes(), &data)
		c.Assert(err, jc.ErrorIsNil, gc.Commentf("body: %s", rec.Body.Bytes()))
		assertBody(c, data)
		return
	}
	c.Assert(rec.Body.String(), jc.JSONEquals, expectBody)
}

// DoRequestParams holds parameters for DoRequest.
// If left empty, some fields will automatically be filled with defaults.
type DoRequestParams struct {
	// Do is used to make the HTTP request.
	// If it is nil, http.DefaultClient.Do will be used.
	// If the body reader implements io.Seeker,
	// req.Body will also implement that interface.
	Do func(req *http.Request) (*http.Response, error)

	// ExpectError holds the error regexp to match
	// against the error returned from the HTTP Do
	// request. If it is empty, the error is expected to be
	// nil.
	ExpectError string

	// Handler holds the handler to use to make the request.
	Handler http.Handler

	// Method holds the HTTP method to use for the call.
	// GET is assumed if this is empty.
	Method string

	// URL holds the URL to pass when making the request.
	URL string

	// JSONBody specifies a JSON value to marshal to use
	// as the body of the request. If this is specified, Body will
	// be ignored and the Content-Type header will
	// be set to application/json. The request
	// body will implement io.Seeker.
	JSONBody interface{}

	// Body holds the body to send in the request.
	Body io.Reader

	// Header specifies the HTTP headers to use when making
	// the request.
	Header http.Header

	// ContentLength specifies the length of the body.
	// It may be zero, in which case the default net/http
	// content-length behaviour will be used.
	ContentLength int64

	// Username, if specified, is used for HTTP basic authentication.
	Username string

	// Password, if specified, is used for HTTP basic authentication.
	Password string

	// Cookies, if specified, are added to the request.
	Cookies []*http.Cookie
}

// DoRequest invokes a request on the given handler with the given
// parameters.
func DoRequest(c *gc.C, p DoRequestParams) *httptest.ResponseRecorder {
	if p.Method == "" {
		p.Method = "GET"
	}
	if p.Do == nil {
		p.Do = http.DefaultClient.Do
	}
	srv := httptest.NewServer(p.Handler)
	defer srv.Close()

	if p.JSONBody != nil {
		data, err := json.Marshal(p.JSONBody)
		c.Assert(err, jc.ErrorIsNil)
		p.Body = bytes.NewReader(data)
	}
	// Note: we avoid NewRequest's odious reader wrapping by using
	// a custom nopCloser function.
	req, err := http.NewRequest(p.Method, srv.URL+p.URL, nopCloser(p.Body))
	c.Assert(err, jc.ErrorIsNil)
	if p.JSONBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, val := range p.Header {
		req.Header[key] = val
	}
	if p.ContentLength != 0 {
		req.ContentLength = p.ContentLength
	} else {
		req.ContentLength = bodyContentLength(p.Body)
	}
	if p.Username != "" || p.Password != "" {
		req.SetBasicAuth(p.Username, p.Password)
	}
	for _, cookie := range p.Cookies {
		req.AddCookie(cookie)
	}
	resp, err := p.Do(req)
	if p.ExpectError != "" {
		c.Assert(err, gc.ErrorMatches, p.ExpectError)
		return nil
	}
	c.Assert(err, jc.ErrorIsNil)
	defer resp.Body.Close()

	// TODO(rog) don't return a ResponseRecorder because we're not actually
	// using httptest.NewRecorder ?
	var rec httptest.ResponseRecorder
	rec.HeaderMap = resp.Header
	rec.Code = resp.StatusCode
	rec.Body = new(bytes.Buffer)
	_, err = io.Copy(rec.Body, resp.Body)
	c.Assert(err, jc.ErrorIsNil)
	return &rec
}

// bodyContentLength returns the Content-Length
// to use for the given body. Usually http.NewRequest
// would infer this (and the cases here come directly
// from the logic in that function) but unfortunately
// there's no way to avoid the NopCloser wrapping
// for any of the types mentioned here.
func bodyContentLength(body io.Reader) int64 {
	n := 0
	switch v := body.(type) {
	case *bytes.Buffer:
		n = v.Len()
	case *bytes.Reader:
		n = v.Len()
	case *strings.Reader:
		n = v.Len()
	}
	return int64(n)
}

// nopCloser is like ioutil.NopCloser except that
// the returned value implements io.Seeker if
// r implements io.Seeker
func nopCloser(r io.Reader) io.ReadCloser {
	if r == nil {
		return nil
	}
	rc, ok := r.(io.ReadCloser)
	if ok {
		return rc
	}
	rs, ok := r.(io.ReadSeeker)
	if ok {
		return readSeekNopCloser{rs}
	}
	return ioutil.NopCloser(r)
}

type readSeekNopCloser struct {
	io.ReadSeeker
}

func (readSeekNopCloser) Close() error {
	return nil
}
