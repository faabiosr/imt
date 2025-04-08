/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pterm/pterm"
	ucli "github.com/urfave/cli/v2"

	"github.com/faabiosr/imt/internal/cli"
	"github.com/faabiosr/imt/internal/client"
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

	app.Before = func(_ *ucli.Context) error {
		pterm.DisableColor()
		return nil
	}

	app.Action = func(cc *ucli.Context) error {
		tpl := fmt.Sprintf(rootCommandTemplate, helpHeaderTemplate)
		ucli.HelpPrinterCustom(cc.App.Writer, tpl, cc.App, nil)

		return nil
	}

	app.Commands = commands(loginCmd, logoutCmd, infoCmd, albumCmd)

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

// action represents urfave/cli.ActionFunc.
type action func(*ucli.Context, *client.Client) error

// withClient wraps action func with internal/client.
func withClient(fn action) ucli.ActionFunc {
	return func(cc *ucli.Context) error {
		creds, err := cli.Session(cc.String("config"))
		if err != nil {
			return err
		}
		cl, err := client.New(creds.Host, creds.Key)
		if err != nil {
			return err
		}

		return fn(cc, cl)
	}
}

const kvSize = 2

// pairs reads flag string slice as a key/value pairs.
func pairs(cc *ucli.Context, name string) (map[string]string, error) {
	items := cc.StringSlice(name)

	m := make(map[string]string, len(items))

	for _, pair := range items {
		kv := strings.SplitN(pair, "=", kvSize)
		if len(kv) != kvSize {
			return m, fmt.Errorf("%s '%s' must be formatted as key=value", name, pair)
		}

		m[kv[0]] = kv[1]
	}

	return m, nil
}

// spinner creates pterm spinner.
func spinner(w io.Writer, text string) *pterm.SpinnerPrinter {
	return pterm.DefaultSpinner.
		WithSequence([]string{"⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ "}...).
		WithText(text).
		WithShowTimer(false).
		WithRemoveWhenDone(true).
		WithWriter(w)
}
