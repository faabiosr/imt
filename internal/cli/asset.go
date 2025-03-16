/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cli

import (
	"context"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/faabiosr/imt/internal/client"
	"github.com/faabiosr/imt/internal/errors"
)

func fetchAssetsIDsByOriginalPath(ctx context.Context, cl *client.Client, path string) ([]string, error) {
	resource, _ := url.Parse("/api/view/folder")
	query := resource.Query()
	query.Add("path", path)

	resource.RawQuery = query.Encode()

	ids := []string{}

	req, err := cl.NewRequest(ctx, http.MethodGet, resource, nil)
	if err != nil {
		return ids, err
	}

	res := []struct {
		ID string `json:"id"`
	}{}

	if err := cl.Do(req, &res); err != nil {
		return ids, err
	}

	for _, asset := range res {
		ids = append(ids, asset.ID)
	}

	return ids, nil
}

func fetchAssetsIDsByOriginalPaths(ctx context.Context, cl *client.Client, paths []string) ([]string, error) {
	assets := []string{}
	m := sync.Mutex{}

	g, ctx := errgroup.WithContext(ctx)

	fn := func(path string) func() error {
		return func() error {
			ids, err := fetchAssetsIDsByOriginalPath(ctx, cl, path)
			if err != nil {
				return err
			}

			m.Lock()
			assets = append(assets, ids...)
			m.Unlock()

			return nil
		}
	}

	for _, path := range paths {
		g.Go(fn(path))
	}

	if err := g.Wait(); err != nil {
		return []string{}, errors.Errorf("one of the paths failed to retrieve assets: %w", err)
	}

	return assets, nil
}
