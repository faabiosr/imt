/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	ucli "github.com/urfave/cli/v2"

	"github.com/faabiosr/imt/internal/cli"
	"github.com/faabiosr/imt/internal/client"
	"github.com/faabiosr/imt/internal/errors"
)

var albumCmd = &ucli.Command{
	Name:        "album",
	Description: "Manages albums",
	Subcommands: commands(autoCreateAlbums),
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
	},
	Action: withClient(func(cc *ucli.Context, cl *client.Client) error {
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

		spin, err := spinner(cc.App.Writer, "creating albums...").Start()
		if err != nil {
			return err
		}

		if err := cli.AutoCreateAlbums(cc.Context, cl, opts); err != nil {
			return err
		}

		_ = spin.Stop()

		return nil
	}),
}
