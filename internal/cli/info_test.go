/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/jarcoal/httpmock"

	"github.com/faabiosr/imt/internal/client"
)

const (
	testHost   = "http://test.host"
	testAPIKey = "de4ba102-e5c7-414b-a3d7-0590e9ff6dbd"
)

func TestInfo(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		ctx := context.Background()
		hc := http.DefaultClient

		httpmock.ActivateNonDefault(hc)
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, testHost+"/api/server/about",
			httpmock.NewJsonResponderOrPanic(http.StatusOK, json.RawMessage(`{}`)))

		httpmock.RegisterResponder(http.MethodGet, testHost+"/api/server/storage",
			httpmock.NewJsonResponderOrPanic(http.StatusOK, json.RawMessage(`{}`)))

		httpmock.RegisterResponder(http.MethodGet, testHost+"/api/server/statistics",
			httpmock.NewJsonResponderOrPanic(http.StatusForbidden, json.RawMessage(`{"message": "Forbidden"}`)))

		baseURL, _ := url.Parse(testHost)

		cl := client.NewWithHTTPClient(baseURL, testAPIKey, hc)

		_, err := Info(ctx, cl)
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		hc := http.DefaultClient

		httpmock.ActivateNonDefault(hc)
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, testHost+"/api/server/about",
			httpmock.NewJsonResponderOrPanic(http.StatusOK, json.RawMessage(`{"version": "1.0"}`)))

		httpmock.RegisterResponder(http.MethodGet, testHost+"/api/server/storage",
			httpmock.NewJsonResponderOrPanic(http.StatusOK, json.RawMessage(`{"diskSizeRaw": 10}`)))

		httpmock.RegisterResponder(http.MethodGet, testHost+"/api/server/statistics",
			httpmock.NewJsonResponderOrPanic(http.StatusOK, json.RawMessage(`{"photos": 10}`)))

		baseURL, _ := url.Parse(testHost)

		cl := client.NewWithHTTPClient(baseURL, testAPIKey, hc)

		si, err := Info(ctx, cl)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		if v := "1.0"; v != si.About.Version {
			t.Errorf("unexpected version: %s (expected %s)", si.About.Version, v)
		}

		if s := 10; int64(s) != si.Storage.Size {
			t.Errorf("unexpected size: %d (expected %d)", si.Storage.Size, s)
		}

		if p := 10; int64(p) != si.Stats.Photos {
			t.Errorf("unexpected photos: %d (expected %d)", si.Stats.Photos, p)
		}
	})
}
