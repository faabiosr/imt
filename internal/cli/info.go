/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cli

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/faabiosr/imt/internal/client"
	"github.com/faabiosr/imt/internal/errors"
)

// About represents server information.
type About struct {
	Version     string `json:"version"`
	Nodejs      string `json:"nodejs"`
	ImageMagick string `json:"imagemagick"`
	ExifTool    string `json:"exiftool"`
	FFmpeg      string `json:"ffmpeg"`
	Libvips     string `json:"libvips"`
	Build       string `json:"build"`
}

// Storage represents the server storage information.
type Storage struct {
	Size      int64 `json:"diskSizeRaw"`
	Use       int64 `json:"diskUseRaw"`
	Available int64 `json:"diskAvailableRaw"`
}

// Stats represents the server statistics.
type Stats struct {
	Photos int64 `json:"photos"`
	Videos int64 `json:"videos"`
	Usage  int64 `json:"usage"`
}

// ServerInfo represents the server information.
type ServerInfo struct {
	About   *About
	Storage *Storage
	Stats   *Stats
}

// Info retrieves server information, like version, statistics and storage.
func Info(ctx context.Context, cl *client.Client) (*ServerInfo, error) {
	si := &ServerInfo{}

	var wg sync.WaitGroup
	const num = 3
	errs := make(chan error, num)

	wg.Add(1)

	go func() {
		defer wg.Done()
		errs <- about(ctx, cl, si)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		errs <- storage(ctx, cl, si)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		errs <- stats(ctx, cl, si)
	}()

	wg.Wait()
	close(errs)

	err := errors.ReadChannel(errs)
	if err != nil {
		return si, fmt.Errorf("one of the server information requests failed: %w", err)
	}

	return si, nil
}

func about(ctx context.Context, cl *client.Client, si *ServerInfo) error {
	resource, _ := url.Parse("/api/server/about")

	req, err := cl.NewRequest(ctx, http.MethodGet, resource, nil)
	if err != nil {
		return err
	}

	return cl.Do(req, &si.About)
}

func storage(ctx context.Context, cl *client.Client, si *ServerInfo) error {
	resource, _ := url.Parse("/api/server/storage")

	req, err := cl.NewRequest(ctx, http.MethodGet, resource, nil)
	if err != nil {
		return err
	}

	return cl.Do(req, &si.Storage)
}

func stats(ctx context.Context, cl *client.Client, si *ServerInfo) error {
	resource, _ := url.Parse("/api/server/statistics")

	req, err := cl.NewRequest(ctx, http.MethodGet, resource, nil)
	if err != nil {
		return err
	}

	return cl.Do(req, &si.Stats)
}
