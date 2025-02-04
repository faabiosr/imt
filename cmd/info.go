/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"text/template"

	"github.com/docker/go-units"
	"github.com/pterm/pterm"
	ucli "github.com/urfave/cli/v2"

	"github.com/faabiosr/imt/internal/cli"
	"github.com/faabiosr/imt/internal/client"
)

var infoTemplate = `{{if .About}}Server:
  Version: {{.About.Version}}
  Node.js: {{.About.Nodejs}}
  ImageMagick: {{.About.ImageMagick}}
  ExifTool: {{.About.ExifTool}}
  FFmpeg: {{.About.FFmpeg}}
  Build: {{.About.Build}}
{{end}}
{{- if .Storage}}Storage:
  Size: {{humansize .Storage.Size}}
  Use: {{humansize .Storage.Use}}
  Available: {{humansize .Storage.Available}}
{{end}}
{{- if .Stats}}Stats:
  Photos: {{.Stats.Photos}}
  Videos: {{.Stats.Videos}}
  Usage: {{humansize .Stats.Usage}}
{{end}}`

var funcMap = template.FuncMap{
	"humansize": func(s int64) string {
		return units.HumanSize(float64(s))
	},
}

var infoCmd = &ucli.Command{
	Name:        "info",
	Description: "Show server information",
	Action: withClient(func(cc *ucli.Context, cl *client.Client) error {
		t, err := template.New("info").Funcs(funcMap).Parse(infoTemplate)
		if err != nil {
			return err
		}

		info, err := cli.Info(cc.Context, cl)
		pterm.PrintOnError(
			t.Execute(cc.App.Writer, info),
			err,
		)

		return nil
	}),
}
