package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"

	"github.com/faabiosr/imt/internal/client"
)

func TestAlbum_excludeFilter(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		exclude, err := excludeFilter([]string{"***"})
		if err == nil {
			t.Error("expected an error, got nil")
		}

		if exclude("") {
			t.Error("expected false, got true")
		}
	})

	t.Run("success match", func(t *testing.T) {
		exclude, err := excludeFilter([]string{"/trip/*"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		if !exclude("/media/trip/img_001.jpg") {
			t.Error("expected true, got false")
		}
	})

	t.Run("success not match", func(t *testing.T) {
		exclude, err := excludeFilter([]string{"/trip/*"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		if exclude("/media/food/img_0001.jpg") {
			t.Error("expected false, got true")
		}
	})
}

func TestAlbum_groupAlbums(t *testing.T) {
	t.Run("only folder option", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.MkdirAll(tmp+"/food/vegetables/", 0o755); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		f, err := os.CreateTemp(tmp+"/food", "tt")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		_, _ = f.WriteString("testing")
		_ = f.Close()

		opts := &AutoCreateAlbumsOptions{
			Folder: tmp + string(os.PathSeparator),
		}

		albums, err := groupAlbums(opts)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		expected := map[string][]string{"food": {"/food"}}

		if !reflect.DeepEqual(albums, expected) {
			t.Errorf(
				"unexpected albums: '%v' (expected '%v')",
				albums,
				expected,
			)
		}
	})

	t.Run("all options", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.MkdirAll(tmp+"/1/food/fruit/", 0o755); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if err := os.MkdirAll(tmp+"/2/food/fruit/", 0o755); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if err := os.MkdirAll(tmp+"/nope", 0o755); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		opts := &AutoCreateAlbumsOptions{
			Folder:            tmp + string(os.PathSeparator),
			Recursive:         true,
			ParentGroupAssets: true,
			SkipLevels:        2,
			OriginalPath:      "/m/",
			Exclude:           []string{"/nope*"},
			Albums: map[string]string{
				"food":  "Food",
				"fruit": "Fruit",
			},
		}

		albums, err := groupAlbums(opts)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		expected := map[string][]string{
			"Food": {
				"/m/1/food",
				"/m/1/food/fruit",
				"/m/2/food",
				"/m/2/food/fruit",
			},
			"Fruit": {
				"/m/1/food/fruit",
				"/m/2/food/fruit",
			},
		}

		if !reflect.DeepEqual(albums, expected) {
			t.Errorf("unexpected albums: '%v' (expected '%v')", albums, expected)
		}
	})
}

func TestAlbumAutoCreateAlbums(t *testing.T) {
	t.Run("group failed with invalid pattern", func(t *testing.T) {
		ctx := context.Background()
		hc := http.DefaultClient
		baseURL, _ := url.Parse(testHost)
		cl := client.NewWithHTTPClient(baseURL, testAPIKey, hc)

		opts := &AutoCreateAlbumsOptions{
			Exclude: []string{"***"},
		}

		err := AutoCreateAlbums(ctx, cl, opts)
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("no albums", func(t *testing.T) {
		ctx := context.Background()
		hc := http.DefaultClient
		baseURL, _ := url.Parse(testHost)
		cl := client.NewWithHTTPClient(baseURL, testAPIKey, hc)

		tmp := t.TempDir()
		if err := os.MkdirAll(tmp+"/food/", 0o755); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		opts := &AutoCreateAlbumsOptions{
			Folder:  tmp + string(os.PathSeparator),
			Exclude: []string{"/food*"},
		}

		err := AutoCreateAlbums(ctx, cl, opts)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("fetch albums failed", func(t *testing.T) {
		ctx := context.Background()
		hc := http.DefaultClient

		httpmock.ActivateNonDefault(hc)
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(
			http.MethodGet,
			testHost+"/api/albums",
			httpmock.NewJsonResponderOrPanic(
				http.StatusInternalServerError,
				json.RawMessage(`{"message": "failed"}`),
			),
		)

		baseURL, _ := url.Parse(testHost)
		cl := client.NewWithHTTPClient(baseURL, testAPIKey, hc)

		tmp := t.TempDir()
		if err := os.MkdirAll(tmp+"/food/", 0o755); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		opts := &AutoCreateAlbumsOptions{
			Folder: tmp + string(os.PathSeparator),
		}

		err := AutoCreateAlbums(ctx, cl, opts)
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("create albums failed", func(t *testing.T) {
		ctx := context.Background()
		hc := http.DefaultClient

		httpmock.ActivateNonDefault(hc)
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(
			http.MethodGet,
			testHost+"/api/albums",
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				json.RawMessage(`[{"albumName": "testing", "id": "821256df-77e9-4616-91b9-57465995a01b"}]`),
			),
		)

		httpmock.RegisterResponder(
			http.MethodPost,
			testHost+"/api/albums",
			httpmock.NewJsonResponderOrPanic(
				http.StatusInternalServerError,
				json.RawMessage(`{"message": ["failed"]}`),
			),
		)

		baseURL, _ := url.Parse(testHost)
		cl := client.NewWithHTTPClient(baseURL, testAPIKey, hc)

		tmp := t.TempDir()
		if err := os.MkdirAll(tmp+"/food/", 0o755); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		opts := &AutoCreateAlbumsOptions{
			Folder: tmp + string(os.PathSeparator),
		}

		err := AutoCreateAlbums(ctx, cl, opts)
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("fetch assets failed", func(t *testing.T) {
		ctx := context.Background()
		hc := http.DefaultClient

		httpmock.ActivateNonDefault(hc)
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(
			http.MethodGet,
			testHost+"/api/albums",
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				json.RawMessage(`[{"albumName": "testing", "id": "821256df-77e9-4616-91b9-57465995a01b"}]`),
			),
		)

		httpmock.RegisterResponder(
			http.MethodPost,
			testHost+"/api/albums",
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				json.RawMessage(`{"albumName": "food", "id": "4cbd308b-ed70-4fe9-92f3-ad4ac3ee8710"}`),
			),
		)

		httpmock.RegisterResponder(
			http.MethodGet,
			testHost+"/api/view/folder",
			httpmock.NewJsonResponderOrPanic(
				http.StatusInternalServerError,
				json.RawMessage(`{"message": "Failed to get assets by original path"}`),
			),
		)

		baseURL, _ := url.Parse(testHost)
		cl := client.NewWithHTTPClient(baseURL, testAPIKey, hc)

		tmp := t.TempDir()
		if err := os.MkdirAll(tmp+"/food/", 0o755); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		opts := &AutoCreateAlbumsOptions{
			Folder: tmp + string(os.PathSeparator),
		}

		err := AutoCreateAlbums(ctx, cl, opts)
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("add assets to album failed", func(t *testing.T) {
		ctx := context.Background()
		hc := http.DefaultClient

		httpmock.ActivateNonDefault(hc)
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(
			http.MethodGet,
			testHost+"/api/albums",
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				json.RawMessage(`[{"albumName": "testing", "id": "821256df-77e9-4616-91b9-57465995a01b"}]`),
			),
		)

		httpmock.RegisterResponder(
			http.MethodPost,
			testHost+"/api/albums",
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				json.RawMessage(`{"albumName": "food", "id": "4cbd308b-ed70-4fe9-92f3-ad4ac3ee8710"}`),
			),
		)

		httpmock.RegisterResponder(
			http.MethodGet,
			testHost+"/api/view/folder",
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				json.RawMessage(`[{"id": "dff78948-b5b2-4d04-a493-ad65df879286"}, {"id": "8dba92a5-753b-4bee-be4f-f7a59ba20762"}]`),
			),
		)

		httpmock.RegisterResponder(
			http.MethodPut,
			testHost+"/api/albums/4cbd308b-ed70-4fe9-92f3-ad4ac3ee8710/assets",
			httpmock.NewJsonResponderOrPanic(
				http.StatusNotFound,
				json.RawMessage(`{"message": "albums was not found"}`),
			),
		)

		baseURL, _ := url.Parse(testHost)
		cl := client.NewWithHTTPClient(baseURL, testAPIKey, hc)

		tmp := t.TempDir()
		if err := os.MkdirAll(tmp+"/food/", 0o755); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		opts := &AutoCreateAlbumsOptions{
			Folder: tmp + string(os.PathSeparator),
		}

		err := AutoCreateAlbums(ctx, cl, opts)
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		hc := http.DefaultClient

		httpmock.ActivateNonDefault(hc)
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(
			http.MethodGet,
			testHost+"/api/albums",
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				json.RawMessage(`[{"albumName": "testing", "id": "821256df-77e9-4616-91b9-57465995a01b"}]`),
			),
		)

		httpmock.RegisterResponder(
			http.MethodPost,
			testHost+"/api/albums",
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				json.RawMessage(`{"albumName": "food", "id": "4cbd308b-ed70-4fe9-92f3-ad4ac3ee8710"}`),
			),
		)

		httpmock.RegisterResponder(
			http.MethodGet,
			testHost+"/api/view/folder",
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				json.RawMessage(`[{"id": "dff78948-b5b2-4d04-a493-ad65df879286"}, {"id": "8dba92a5-753b-4bee-be4f-f7a59ba20762"}]`),
			),
		)

		httpmock.RegisterResponder(
			http.MethodPut,
			testHost+"/api/albums/4cbd308b-ed70-4fe9-92f3-ad4ac3ee8710/assets",
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				json.RawMessage(`[{"id": "821256df-77e9-4616-91b9-57465995a01b"}]`),
			),
		)

		baseURL, _ := url.Parse(testHost)
		cl := client.NewWithHTTPClient(baseURL, testAPIKey, hc)

		tmp := t.TempDir()
		if err := os.MkdirAll(tmp+"/1/food/fruit/", 0o755); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if err := os.MkdirAll(tmp+"/2/food/fruit/", 0o755); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		opts := &AutoCreateAlbumsOptions{
			Folder:    tmp + string(os.PathSeparator),
			Recursive: true,
		}

		err := AutoCreateAlbums(ctx, cl, opts)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}
