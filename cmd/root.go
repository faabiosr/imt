/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/faabiosr/imt/internal/cli"
	ucli "github.com/urfave/cli/v2"
)

// Execute runs root cmd.
func Execute(ctx context.Context, args []string) {
	if err := newCmd().RunContext(ctx, args); err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "%v\n\n", err)
		os.Exit(1)
	}
}

const unknown = "unknown"

// variables are expected to be set at build time.
var (
	releaseVersion = unknown
	releaseCommit  = unknown
	releaseOS      = unknown
)

// variables that defines custom app templates.
var (
	helpHeaderTemplate = `{{template "helpNameTemplate" .}}

Usage: {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[options]{{end}}{{if .Commands}} command [command options]{{end}}{{end}}{{ if .Description}}

{{wrap .Description 0}}{{end}}`

	rootCommandTemplate = `%s

For listing options and commands, use '{{.HelpName}} --help or {{.HelpName}} -h'.
`

	appHelpTemplate = `%s{{if .VisibleFlags}}

Options: {{template "visibleFlagTemplate" .}}{{end}}{{if .VisibleCommands}}

Commands:{{template "visibleCommandCategoryTemplate" .}}{{end}}

For more information on a command, use '{{.HelpName}} [command] --help'.
`

	commandHelpTemplate = `{{.HelpName}}{{if .Description}} - {{template "descriptionTemplate" .}}{{end}}

Usage: {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}{{if .Args}}[arguments...]{{end}}[arguments...]{{end}}{{end}}{{if .VisibleFlags}}

Options: {{template "visibleFlagTemplate" .}}{{end}}{{if .VisibleCommands}}

Commands:{{template "visibleCommandCategoryTemplate" .}}{{end}}

`
)

// newCmd creates cli application defining custom help templates and default values.
func newCmd() *ucli.App {
	app := &ucli.App{}
	app.Name = "imt"
	app.Usage = "Immich tools"
	app.Description = "A collection of command-line tools for Immich."
	app.Version = fmt.Sprintf("%s, build: %s, os: %s", releaseVersion, releaseCommit, releaseOS)
	app.CustomAppHelpTemplate = fmt.Sprintf(appHelpTemplate, helpHeaderTemplate)
	app.HideHelpCommand = true
	app.Suggest = true

	app.EnableBashCompletion = true
	app.BashComplete = func(ctx *ucli.Context) {
		for _, cmd := range ctx.App.Commands {
			_, _ = fmt.Fprintln(ctx.App.Writer, cmd.Name)
		}
	}

	app.Flags = []ucli.Flag{
		&ucli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "imt config auth file",
			Value:   must(cli.DefaultCredentialsPath()),
		},
	}

	app.Action = func(cc *ucli.Context) error {
		tpl := fmt.Sprintf(rootCommandTemplate, helpHeaderTemplate)
		ucli.HelpPrinterCustom(cc.App.Writer, tpl, cc.App, nil)

		return nil
	}

	app.Commands = commands(loginCmd, logoutCmd)

	return app
}

// commands sets custom help templates and default values.
func commands(cmds ...*ucli.Command) []*ucli.Command {
	for _, cmd := range cmds {
		cmd.Usage = cmd.Description
		cmd.HideHelpCommand = true
		cmd.CustomHelpTemplate = commandHelpTemplate
	}

	return cmds
}

// must is a helper that wraps a cal to a function returns (string, error)
// and panics if the error is non-nil.
func must(s string, err error) string {
	if err != nil {
		panic(err)
	}

	return s
}
