/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/pterm/pterm"
	ucli "github.com/urfave/cli/v2"

	"github.com/faabiosr/imt/internal/cli"
	"github.com/faabiosr/imt/internal/client"
	"github.com/faabiosr/imt/internal/errors"
)

var albumCmd = &ucli.Command{
	Name:        "album",
	Description: "Manages albums",
	Subcommands: commands(autoCreateAlbums, listAlbums),
}

var autoCreateAlbums = &ucli.Command{
	Name:        "auto-create",
	Description: "create albums automatically based on folder structure",
	Flags: []ucli.Flag{
		&ucli.BoolFlag{
			Name:  "recursive",
			Usage: "reads the photos folder recursively",
		},
		&ucli.IntFlag{
			Name:  "skip-levels",
			Usage: "skip folder levels names of group creation from root path.",
		},
		&ucli.StringFlag{
			Name:  "original-path",
			Usage: "sets the original path where the photos is stored in Immich",
		},
		&ucli.StringSliceFlag{
			Name:  "exclude",
			Usage: "exclude files matching pattern",
		},
		&ucli.StringSliceFlag{
			Name:  "rename",
			Usage: "set a key/value album to be renamed",
		},
		&ucli.StringFlag{
			Name:  "from-config",
			Usage: "load parameters from config file",
		},
	},
	Action: withClient(func(cc *ucli.Context, cl *client.Client) error {
		cfg := cc.String("from-config")
		if cfg != "" {
			opts, err := loadAutoCreateConfigFile(cfg)
			if err != nil {
				return err
			}

			return autoCreateAlbumsAction(cc, cl, opts)
		}

		if cc.Args().Len() != 1 {
			return errors.New("Empty path is not allowed")
		}

		albums, err := pairs(cc, "rename")
		if err != nil {
			return err
		}

		opts := &cli.AutoCreateAlbumsOptions{
			Folder:       cc.Args().First(),
			Recursive:    cc.Bool("recursive"),
			SkipLevels:   cc.Int("skip-levels"),
			OriginalPath: cc.String("original-path"),
			Exclude:      cc.StringSlice("exclude"),
			Albums:       albums,
		}

		return autoCreateAlbumsAction(cc, cl, opts)
	}),
}

func autoCreateAlbumsAction(cc *ucli.Context, cl *client.Client, opts *cli.AutoCreateAlbumsOptions) error {
	spin, err := spinner(cc.App.Writer, "creating albums...").Start()
	if err != nil {
		return err
	}

	if err := cli.AutoCreateAlbums(cc.Context, cl, opts); err != nil {
		return err
	}

	return spin.Stop()
}

func loadAutoCreateConfigFile(name string) (*cli.AutoCreateAlbumsOptions, error) {
	if name == "" {
		return nil, errors.New("empty filename is not allowed")
	}

	content, err := os.ReadFile(name)
	if err != nil {
		return nil, errors.Errorf("unable to read auto create albums config file: %w", err)
	}

	var opts cli.AutoCreateAlbumsOptions
	return &opts, json.Unmarshal(content, &opts)
}

var listAlbums = &ucli.Command{
	Name:        "list",
	Description: "list albums stored",
	Action: withClient(func(cc *ucli.Context, cl *client.Client) error {
		albums, err := cli.FetchAlbums(cc.Context, cl)
		if err != nil {
			return err
		}

		data := pterm.TableData{
			{"ID", "NAME", "NUMBER OF ASSETS"},
		}

		for _, album := range albums {
			data = append(data, []string{album.ID, album.Name, strconv.FormatInt(album.AssetCount, 10)})
		}

		return pterm.DefaultTable.
			WithHasHeader().
			WithData(data).
			Render()
	}),
}
