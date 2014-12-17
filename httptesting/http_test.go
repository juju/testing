// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package httptesting_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	gc "gopkg.in/check.v1"

	"github.com/juju/testing/httptesting"
)

type requestsSuite struct{}

var _ = gc.Suite(&requestsSuite{})

// handlerResponse holds the body of a testing handler response.
type handlerResponse struct {
	URL    string
	Method string
	Body   string
	Auth   bool
}

func makeHandler(c *gc.C, status int, ctype string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		c.Assert(err, gc.IsNil)

		// Create the response.
		response := handlerResponse{
			URL:    req.URL.String(),
			Method: req.Method,
			Body:   string(body),
			Auth:   req.Header.Get("Authorization") != "",
		}
		// Write the response.
		w.Header().Set("Content-Type", ctype)
		w.WriteHeader(status)
		enc := json.NewEncoder(w)
		err = enc.Encode(response)
		c.Assert(err, gc.IsNil)
	})
}

var assertJSONCallTests = []struct {
	about  string
	params httptesting.JSONCallParams
}{{
	about: "simple request",
	params: httptesting.JSONCallParams{
		Method: "GET",
		URL:    "/",
	},
}, {
	about: "method not specified",
	params: httptesting.JSONCallParams{
		URL: "/",
	},
}, {
	about: "POST request with a body",
	params: httptesting.JSONCallParams{
		Method: "POST",
		URL:    "/my/url",
		Body:   strings.NewReader("request body"),
	},
}, {
	about: "authentication",
	params: httptesting.JSONCallParams{
		URL:          "/",
		Method:       "PUT",
		Username:     "who",
		Password:     "bad-wolf",
		ExpectStatus: http.StatusOK,
	},
}, {
	about: "error status",
	params: httptesting.JSONCallParams{
		URL:          "/",
		ExpectStatus: http.StatusBadRequest,
	},
}, {
	about: "custom Do",
	params: httptesting.JSONCallParams{
		URL:          "/",
		ExpectStatus: http.StatusTeapot,
		Do: func(req *http.Request) (*http.Response, error) {
			resp, err := http.DefaultClient.Do(req)
			resp.StatusCode = http.StatusTeapot
			return resp, err
		},
	},
}, {
	about: "expect error",
	params: httptesting.JSONCallParams{
		URL:          "/",
		ExpectStatus: http.StatusTeapot,
		Do: func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("some error")
		},
		ExpectError: "some error",
	},
}, {
	about: "expect error regexp",
	params: httptesting.JSONCallParams{
		URL:          "/",
		ExpectStatus: http.StatusTeapot,
		Do: func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("some bad error")
		},
		ExpectError: "some .* error",
	},
}}

func (*requestsSuite) TestAssertJSONCall(c *gc.C) {
	for i, test := range assertJSONCallTests {
		c.Logf("test %d: %s", i, test.about)
		params := test.params

		// A missing status is assumed to be http.StatusOK.
		status := params.ExpectStatus
		if status == 0 {
			status = http.StatusOK
		}

		// Create the HTTP handler for this test.
		params.Handler = makeHandler(c, status, "application/json")

		// Populate the expected body parameter.
		expectBody := handlerResponse{
			URL:    params.URL,
			Method: params.Method,
		}

		// A missing method is assumed to be "GET".
		if expectBody.Method == "" {
			expectBody.Method = "GET"
		}

		// Handle the request body parameter.
		if params.Body != nil {
			body, err := ioutil.ReadAll(params.Body)
			c.Assert(err, gc.IsNil)
			expectBody.Body = string(body)
			params.Body = bytes.NewReader(body)
		}

		// Handle basic HTTP authentication.
		if params.Username != "" || params.Password != "" {
			expectBody.Auth = true
		}
		params.ExpectBody = expectBody
		httptesting.AssertJSONCall(c, params)
	}
}

// The TestAssertJSONCall above exercises the testing.AssertJSONCall succeeding
// calls. Failures are already massively tested in practice. DoRequest and
// AssertJSONResponse are also indirectly tested as they are called by
// AssertJSONCall.
