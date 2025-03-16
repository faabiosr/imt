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

func TestAsset_fetchAssetByOriginalPath(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		ctx := context.Background()
		path := "/media/trip"
		hc := http.DefaultClient

		httpmock.ActivateNonDefault(hc)
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponderWithQuery(
			http.MethodGet,
			testHost+"/api/view/folder",
			map[string]string{"path": path},
			httpmock.NewJsonResponderOrPanic(
				http.StatusInternalServerError,
				json.RawMessage(`{"message": "Failed to get assets by original path"}`),
			),
		)

		baseURL, _ := url.Parse(testHost)

		cl := client.NewWithHTTPClient(baseURL, testAPIKey, hc)

		_, err := fetchAssetsIDsByOriginalPath(ctx, cl, path)
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		path := "/media/shows"
		hc := http.DefaultClient

		httpmock.ActivateNonDefault(hc)
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponderWithQuery(
			http.MethodGet,
			testHost+"/api/view/folder",
			map[string]string{"path": path},
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				json.RawMessage(`[{"id": "dff78948-b5b2-4d04-a493-ad65df879286"}, {"id": "8dba92a5-753b-4bee-be4f-f7a59ba20762"}]`),
			),
		)

		baseURL, _ := url.Parse(testHost)

		cl := client.NewWithHTTPClient(baseURL, testAPIKey, hc)

		ids, err := fetchAssetsIDsByOriginalPath(ctx, cl, path)
		if err != nil {
			t.Error(err)
		}

		if n := len(ids); n != 2 {
			t.Errorf("unexpected number of assets ids: %d (expected %d)", n, 2)
		}
	})
}

func TestAsset_fetchAssetsIDsByOriginalPaths(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		ctx := context.Background()
		path1 := "/media/cars"
		path2 := "/media/food"
		paths := []string{path1, path2}
		hc := http.DefaultClient

		httpmock.ActivateNonDefault(hc)
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponderWithQuery(
			http.MethodGet,
			testHost+"/api/view/folder",
			map[string]string{"path": path1},
			httpmock.NewJsonResponderOrPanic(
				http.StatusInternalServerError,
				json.RawMessage(`{"message": "Failed to get assets by original path"}`),
			),
		)

		httpmock.RegisterResponderWithQuery(
			http.MethodGet,
			testHost+"/api/view/folder",
			map[string]string{"path": path2},
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				json.RawMessage(`[{"id": "8dba92a5-753b-4bee-be4f-f7a59ba20762"}]`),
			),
		)

		baseURL, _ := url.Parse(testHost)

		cl := client.NewWithHTTPClient(baseURL, testAPIKey, hc)

		_, err := fetchAssetsIDsByOriginalPaths(ctx, cl, paths)
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		path1 := "/media/series"
		path2 := "/media/music"
		paths := []string{path1, path2}
		hc := http.DefaultClient

		httpmock.ActivateNonDefault(hc)
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponderWithQuery(
			http.MethodGet,
			testHost+"/api/view/folder",
			map[string]string{"path": path1},
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				json.RawMessage(`[{"id": "dff78948-b5b2-4d04-a493-ad65df879286"}]`),
			),
		)

		httpmock.RegisterResponderWithQuery(
			http.MethodGet,
			testHost+"/api/view/folder",
			map[string]string{"path": path2},
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				json.RawMessage(`[{"id": "8dba92a5-753b-4bee-be4f-f7a59ba20762"}]`),
			),
		)

		baseURL, _ := url.Parse(testHost)

		cl := client.NewWithHTTPClient(baseURL, testAPIKey, hc)

		ids, err := fetchAssetsIDsByOriginalPaths(ctx, cl, paths)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		if n := len(ids); n != 2 {
			t.Errorf("unexpected number of assets ids: %d (expected %d)", n, 2)
		}
	})
}
