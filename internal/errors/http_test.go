/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package errors

import (
	"net/http"
	"testing"
)

func TestHTTP(t *testing.T) {
	tests := []struct {
		name string
		err  any
		want string
	}{
		{
			name: "no error",
		},
		{
			name: "error",
			err:  New("validation failed"),
			want: "validation failed",
		},
		{
			name: "string",
			err:  "validation failed",
			want: "validation failed",
		},
		{
			name: "panic",
			err:  0,
			want: "err must be error or string, got int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r != nil && r != tt.want {
					t.Errorf("want %v, got %v", tt.want, r)
				}
			}()

			err := HTTP(http.StatusBadRequest, tt.err)
			if err != nil && tt.want != err.Error() {
				t.Errorf("want %v, got %v", tt.want, err.Error())
			}
		})
	}

	t.Run("http error", func(t *testing.T) {
		err := New("validation failed")
		got := HTTP(http.StatusBadRequest, err)

		if code := got.(*httpErr).HTTPStatus(); code != http.StatusBadRequest {
			t.Errorf("status code: want %v, got %v", http.StatusBadRequest, code)
		}

		if !Is(got, err) {
			t.Errorf("error is not the same: want %v, got %v", err, got)
		}
	})
}

func TestStatusCode(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		status := StatusCode(nil)
		if status != 0 {
			t.Errorf("unexpected status code: %d, (expected 0)", status)
		}
	})

	t.Run("http error", func(t *testing.T) {
		status := StatusCode(HTTP(http.StatusBadRequest, "invalid params"))
		if status != http.StatusBadRequest {
			t.Errorf("unexpected status code: %d, (expected %d)", status, http.StatusBadRequest)
		}
	})

	t.Run("any error", func(t *testing.T) {
		status := StatusCode(New("unable to call"))
		if status != http.StatusInternalServerError {
			t.Errorf("unexpected status code: %d, (expected %d)", status, http.StatusInternalServerError)
		}
	})
}
