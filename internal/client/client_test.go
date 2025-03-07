/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package client

import (
	"context"
	"encoding/json"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/faabiosr/imt/internal/errors"
	"github.com/jarcoal/httpmock"
)

const (
	testHost   = "http://test.host"
	testAPIKey = "de4ba102-e5c7-414b-a3d7-0590e9ff6dbd"
)

type errReader struct {
	r   io.Reader
	err error
}

var _ io.ReadCloser = &errReader{}

func (r *errReader) Read(p []byte) (int, error) {
	c, err := r.r.Read(p)
	if err == io.EOF {
		err = r.err
	}
	return c, err
}

func (r *errReader) Close() error { return r.err }

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		apiKey  string
		err     string
	}{
		{
			name: "empty url",
			err:  "empty url is not allowed",
		},
		{
			name:    "empty api key",
			baseURL: testHost,
			err:     "empty api key is not allowed",
		},
		{
			name:    "wrong url",
			baseURL: "http://192.168.1.%1/",
			apiKey:  testAPIKey,
			err:     `parse "http://192.168.1.%1/": invalid URL escape "%1"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.baseURL, tt.apiKey)
			if e := err.Error(); e != tt.err {
				t.Errorf("unexpected error: %s (expected %s)", e, tt.err)
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	tests := []struct {
		name   string
		method string
		body   any
		err    string
	}{
		{
			name:   "invalid body",
			method: http.MethodPost,
			body:   struct{ Name float64 }{Name: math.NaN()},
			err:    "json: unsupported value: NaN",
		},
		{
			name:   "invalid method",
			method: "p@st",
			err:    `net/http: invalid method "p@st"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := New(testHost, testAPIKey)
			ctx := context.Background()
			resource, _ := url.Parse("/req")

			_, err := c.NewRequest(ctx, tt.method, resource, tt.body)
			if e := err.Error(); e != tt.err {
				t.Errorf("unexpected error: %s (expected %s)", e, tt.err)
			}
		})
	}
}

func TestDo(t *testing.T) {
	tests := []struct {
		name   string
		mock   func()
		value  any
		status int
		err    string
	}{
		{
			name: "network error",
			mock: func() {
				httpmock.RegisterResponder(http.MethodGet, testHost+"/req",
					httpmock.NewErrorResponder(&net.DNSError{Err: "failed", IsTimeout: true}))
			},
			status: http.StatusServiceUnavailable,
			err:    `Get "http://test.host/req": lookup : failed`,
		},
		{
			name: "any error",
			mock: func() {
				httpmock.RegisterResponder(http.MethodGet, testHost+"/req",
					httpmock.NewErrorResponder(errors.New("failed")))
			},
			status: http.StatusInternalServerError,
			err:    `Get "http://test.host/req": failed`,
		},
		{
			name: "resource not found",
			mock: func() {
				res := json.RawMessage(`{"message": "resource was not found"}`)
				httpmock.RegisterResponder(http.MethodGet, testHost+"/req",
					httpmock.NewJsonResponderOrPanic(http.StatusNotFound, res))
			},
			status: http.StatusNotFound,
			err:    "resource was not found",
		},
		{
			name: "error message as array",
			mock: func() {
				res := json.RawMessage(`{"message": ["resource was not found"]}`)
				httpmock.RegisterResponder(http.MethodGet, testHost+"/req",
					httpmock.NewJsonResponderOrPanic(http.StatusNotFound, res))
			},
			status: http.StatusNotFound,
			err:    "resource was not found",
		},
		{
			name: "failed to parse error message as int",
			mock: func() {
				res := json.RawMessage(`{"message": 1000}`)
				httpmock.RegisterResponder(http.MethodGet, testHost+"/req",
					httpmock.NewJsonResponderOrPanic(http.StatusNotFound, res))
			},
			status: http.StatusInternalServerError,
			err:    "unable to parse message, invalid value type",
		},
		{
			name: "closed response",
			mock: func() {
				res := &http.Response{
					Body: &errReader{
						r:   strings.NewReader(""),
						err: errors.New("failed"),
					},
				}
				httpmock.RegisterResponder(http.MethodGet, testHost+"/req",
					httpmock.ResponderFromResponse(res))
			},
			status: http.StatusInternalServerError,
			err:    "failed",
		},
		{
			name: "request error",
			mock: func() {
				httpmock.RegisterResponder(http.MethodGet, testHost+"/req",
					httpmock.NewStringResponder(http.StatusNotFound, ""))
			},
			status: http.StatusInternalServerError,
			err:    "request error",
		},
		{
			name: "invalid json when not found error",
			mock: func() {
				httpmock.RegisterResponder(http.MethodGet, testHost+"/req",
					httpmock.NewStringResponder(http.StatusNotFound, "{"))
			},
			status: http.StatusInternalServerError,
			err:    "unexpected end of JSON input",
		},
		{
			name: "invalid json error",
			mock: func() {
				httpmock.RegisterResponder(http.MethodGet, testHost+"/req",
					httpmock.NewStringResponder(http.StatusOK, "{"))
			},
			value:  &map[string]string{},
			status: http.StatusInternalServerError,
			err:    "unexpected EOF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := http.DefaultClient

			httpmock.ActivateNonDefault(hc)
			defer httpmock.DeactivateAndReset()

			baseURL, _ := url.Parse(testHost)

			c := &Client{
				hc:      hc,
				baseURL: baseURL,
				apiKey:  testAPIKey,
			}

			ctx := context.Background()
			resource, _ := url.Parse("/req")

			req, _ := c.NewRequest(ctx, http.MethodGet, resource, nil)

			tt.mock()

			err := c.Do(req, tt.value)
			status := errors.StatusCode(err)

			if err.Error() != tt.err {
				t.Errorf("unexpected error: %s (expected %s)", err, tt.err)
			}

			if status != tt.status {
				t.Errorf("unexpected status code: %d (expected %d)", status, tt.status)
			}
		})
	}

	t.Run("success with no value", func(t *testing.T) {
		hc := http.DefaultClient

		httpmock.ActivateNonDefault(hc)
		defer httpmock.DeactivateAndReset()

		baseURL, _ := url.Parse(testHost)

		c := &Client{
			hc:      hc,
			baseURL: baseURL,
			apiKey:  testAPIKey,
		}

		ctx := context.Background()
		resource, _ := url.Parse("/req")

		req, _ := c.NewRequest(ctx, http.MethodGet, resource, nil)

		httpmock.RegisterResponder(http.MethodGet, testHost+"/req",
			httpmock.NewStringResponder(http.StatusNoContent, ""))

		err := c.Do(req, nil)
		if err != nil {
			t.Errorf("unexpected error: %s (expected nil)", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		hc := http.DefaultClient

		httpmock.ActivateNonDefault(hc)
		defer httpmock.DeactivateAndReset()

		baseURL, _ := url.Parse(testHost)

		c := &Client{
			hc:      hc,
			baseURL: baseURL,
			apiKey:  testAPIKey,
		}

		ctx := context.Background()
		resource, _ := url.Parse("/req")

		req, _ := c.NewRequest(ctx, http.MethodGet, resource, nil)

		res := json.RawMessage(`{"name": "immich"}`)
		httpmock.RegisterResponder(http.MethodGet, testHost+"/req",
			httpmock.NewJsonResponderOrPanic(http.StatusOK, res))

		data := struct {
			Name string
		}{}

		err := c.Do(req, &data)
		if err != nil {
			t.Errorf("unexpected error: %d (expected nil)", errors.StatusCode(err))
		}

		if data.Name != "immich" {
			t.Errorf("unexpected data name: %s (expected immich)", data.Name)
		}
	})
}
