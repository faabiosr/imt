/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/faabiosr/imt/internal/errors"
)

const (
	userAgent = "imt/client"
	mediaType = "application/json"
)

// Client manages communication with Immich server.
type Client struct {
	hc      *http.Client
	baseURL *url.URL
	apiKey  string
}

// New returns a new Immich http client.
func New(baseURL, apiKey string) (*Client, error) {
	if baseURL == "" {
		return nil, errors.New("empty url is not allowed")
	}

	if apiKey == "" {
		return nil, errors.New("empty api key is not allowed")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return NewWithHTTPClient(parsedURL, apiKey, http.DefaultClient), nil
}

// NewWithHTTPClient returns a new Immich http client.
func NewWithHTTPClient(baseURL *url.URL, apiKey string, hc *http.Client) *Client {
	return &Client{hc: hc, baseURL: baseURL, apiKey: apiKey}
}

// NewRequest creates an API requrest. A relative URL can be provided in res URL instance. If specified, the
// value pointed to by body JSON encoded and included in as the request body.
func (c *Client) NewRequest(ctx context.Context, method string, res *url.URL, body any) (*http.Request, error) {
	url := c.baseURL.ResolveReference(res)
	buf := new(bytes.Buffer)

	if body != nil {
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url.String(), buf)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", mediaType)
	req.Header.Add("Accept", mediaType)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("x-api-key", c.apiKey)

	return req, nil
}

// Do sends an API request and returns the API response. If the HTTP response is in the 2xx range,
// unmarshal the response body into value.
func (c *Client) Do(req *http.Request, value any) (err error) {
	res, err := c.hc.Do(req)
	if err != nil {
		return netError(err)
	}
	defer func() {
		cerr := res.Body.Close()
		if err == nil {
			err = cerr
		}
	}()

	if err := c.checkResponse(res); err != nil {
		return err
	}

	if value == nil {
		return nil
	}

	return errors.HTTP(
		http.StatusInternalServerError,
		json.NewDecoder(res.Body).Decode(value),
	)
}

// checkResponse checks the API response for errors and returns them if present.
func (c *Client) checkResponse(res *http.Response) error {
	if c := res.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errRes := struct {
		Message string `json:"message"`
	}{}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.HTTP(http.StatusInternalServerError, err)
	}

	if len(data) == 0 {
		return errors.HTTP(http.StatusInternalServerError, "request error")
	}

	if err := json.Unmarshal(data, &errRes); err != nil {
		return errors.HTTP(http.StatusInternalServerError, err)
	}

	return errors.HTTP(res.StatusCode, errRes.Message)
}

// netError verifies if error is related with network.
func netError(err error) error {
	var de *net.DNSError

	if errors.As(err, &de) {
		return errors.HTTP(http.StatusServiceUnavailable, err)
	}

	return err
}
