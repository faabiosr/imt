/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package errors

import (
	"fmt"
	"net/http"
)

type httpErr struct {
	code int
	err  error
}

func (e *httpErr) Error() string {
	return e.err.Error()
}

func (e *httpErr) Unwrap() error {
	return e.err
}

func (e *httpErr) HTTPStatus() int {
	return e.code
}

// HTTP wraps err with the specified HTTP status code. err may be either
// an error or a string.
// Any other type will cause a panic. If err is a nil error, the return
// value will also be nil.
func HTTP(status int, err any) error {
	if err == nil {
		return nil
	}

	var e error
	switch t := err.(type) {
	case error:
		e = t
	case string:
		e = New(t)
	default:
		panic(fmt.Sprintf("err must be error or string, got %T", err))
	}

	return &httpErr{
		code: status,
		err:  e,
	}
}

type httpStatuser interface {
	HTTPStatus() int
}

// StatusCode returns the HTTP status code embedded in the error.
func StatusCode(err error) int {
	if err == nil {
		return 0
	}

	var he httpStatuser

	if As(err, &he) {
		return he.HTTPStatus()
	}

	return http.StatusInternalServerError
}
