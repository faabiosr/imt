/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cli

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/faabiosr/imt/internal/client"
)

// AutoCreateAlbumOptions handles the options to auto create albums.
type AutoCreateAlbumsOptions struct {
	Folder            string            `json:"folder"`
	Recursive         bool              `json:"recursive"`
	SkipLevels        int               `json:"skip_levels"`
	OriginalPath      string            `json:"original_path,omitempty"`
	Exclude           []string          `json:"exclude,omitempty"`
	ParentGroupAssets bool              `json:"parent_group_assets"`
	Albums            map[string]string `json:"albums,omitempty"`
}

// album represents an album stored in Immich.
type album struct {
	ID   string `json:"id"`
	Name string `json:"albumName"`
}

// albums represents a collection of albums.
type albums []album

// AutoCreateAlbums will create albums based on folders.
func AutoCreateAlbums(ctx context.Context, cl *client.Client, opts *AutoCreateAlbumsOptions) error {
	groups, err := groupAlbums(opts)
	if err != nil {
		return err
	}

	if len(groups) == 0 {
		return nil
	}

	as, err := fetchAlbums(ctx, cl)
	if err != nil {
		return err
	}

	items := make(map[string][]string)

	for name, folders := range groups {
		i := slices.IndexFunc(as, func(a album) bool {
			return a.Name == name
		})

		if i >= 0 {
			items[as[i].ID] = append(items[as[i].ID], folders...)
			continue
		}

		a, err := createAlbum(ctx, cl, name)
		if err != nil {
			return err
		}

		items[a.ID] = folders
		as = append(as, a)
	}

	for id, paths := range items {
		assets, err := fetchAssetsIDsByOriginalPaths(ctx, cl, paths)
		if err != nil {
			return err
		}

		if err := addAssetsToAlbum(ctx, cl, id, assets); err != nil {
			return err
		}
	}

	return nil
}

// createAlbum creates an album with name.
func createAlbum(ctx context.Context, cl *client.Client, name string) (album, error) {
	resource, _ := url.Parse("/api/albums")

	body := map[string]string{
		"albumName": name,
	}

	a := album{}

	req, err := cl.NewRequest(ctx, http.MethodPost, resource, body)
	if err != nil {
		return a, err
	}

	return a, cl.Do(req, &a)
}

// fetchAlbums returns all albums stored.
func fetchAlbums(ctx context.Context, cl *client.Client) (albums, error) {
	resource, _ := url.Parse("/api/albums")

	var as albums

	req, err := cl.NewRequest(ctx, http.MethodGet, resource, nil)
	if err != nil {
		return as, err
	}

	return as, cl.Do(req, &as)
}

// addAssetsToAlbum adds a list of assets into an album.
func addAssetsToAlbum(ctx context.Context, cl *client.Client, id string, assets []string) error {
	resource, _ := url.Parse(fmt.Sprintf("/api/albums/%s/assets", id))
	body := map[string]any{
		"ids": assets,
	}

	req, err := cl.NewRequest(ctx, http.MethodPut, resource, body)
	if err != nil {
		return err
	}

	return cl.Do(req, nil)
}

// excludeFilter apply a glob/regexp filter to remove folders path.
func excludeFilter(excludes []string) (func(path string) bool, error) {
	rules := make([]*regexp.Regexp, 0, len(excludes))
	fn := func(string) bool { return false }

	for _, e := range excludes {
		r, err := globToRegexp(e)
		if err != nil {
			return fn, err
		}

		rules = append(rules, r)
	}

	return func(path string) bool {
		for _, r := range rules {
			if r.MatchString(path) {
				return true
			}
		}

		return false
	}, nil
}

// groupAlbums reads the folder tree and group albums by folder name.
// Internally also exclude folders that should not be created as album based on
// the options set.
func groupAlbums(opts *AutoCreateAlbumsOptions) (map[string][]string, error) {
	folder := filepath.Dir(opts.Folder)
	depth := strings.Count(folder, string(os.PathSeparator))

	op := ""
	if opts.OriginalPath != "" {
		op = filepath.Dir(opts.OriginalPath)
	}

	albums := map[string][]string{}

	excludes, err := excludeFilter(opts.Exclude)
	if err != nil {
		return albums, err
	}

	err = filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !opts.Recursive {
			if d.IsDir() && strings.Count(path, string(os.PathSeparator)) > depth+1 {
				return fs.SkipDir
			}
		}

		if !d.IsDir() {
			return nil
		}

		path = strings.Replace(path, folder, op, 1)

		if excludes(path) {
			return nil
		}

		segments := strings.Split(path, string(os.PathSeparator))
		skip := opts.SkipLevels + 1
		if len(segments) <= skip {
			return nil
		}

		segments = segments[skip:]

		if !opts.ParentGroupAssets {
			segments = segments[len(segments)-1:]
		}

		for _, s := range segments {
			if v, ok := opts.Albums[s]; ok {
				s = v
			}

			if _, ok := albums[s]; !ok {
				albums[s] = append(albums[s], path)
				return nil
			}

			albums[s] = append(albums[s], path)
		}

		return nil
	})

	return albums, err
}
