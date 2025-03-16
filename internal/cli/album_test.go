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
	t.Run("invalid exclusion pattern", func(t *testing.T) {
		opts := &AutoCreateAlbumsOptions{
			Exclude: []string{"***"},
		}
		_, err := groupAlbums(opts)
		if err == nil {
			t.Error("expected an error, got nil")
		}
	})

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

func TestAlbum_createAlbum(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		ctx := context.Background()
		hc := http.DefaultClient

		httpmock.ActivateNonDefault(hc)
		defer httpmock.DeactivateAndReset()

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

		_, err := createAlbum(ctx, cl, "testing")
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
			http.MethodPost,
			testHost+"/api/albums",
			httpmock.NewJsonResponderOrPanic(
				http.StatusOK,
				json.RawMessage(`{"albumName": "testing", "id": "821256df-77e9-4616-91b9-57465995a01b"}`),
			),
		)

		baseURL, _ := url.Parse(testHost)

		cl := client.NewWithHTTPClient(baseURL, testAPIKey, hc)

		album, err := createAlbum(ctx, cl, "testing")
		if err != nil {
			t.Error(err)
		}

		if album.Name != "testing" {
			t.Errorf("unexpected album name: %s (expected testing)", album.Name)
		}
	})
}

func TestAlbum_fetchAlbums(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
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

		_, err := fetchAlbums(ctx, cl)
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

		baseURL, _ := url.Parse(testHost)

		cl := client.NewWithHTTPClient(baseURL, testAPIKey, hc)

		albums, err := fetchAlbums(ctx, cl)
		if err != nil {
			t.Error(err)
		}

		if n := len(albums); n != 1 {
			t.Errorf("unexpected number of assets ids: %d (expected %d)", n, 1)
		}
	})
}
