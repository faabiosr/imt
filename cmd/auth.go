/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cmd

import (
	"errors"

	"github.com/faabiosr/imt/internal/cli"
	"github.com/pterm/pterm"
	ucli "github.com/urfave/cli/v2"
)

var loginCmd = &ucli.Command{
	Name:        "login",
	Description: "Login using API key",
	ArgsUsage:   "[host]",
	Flags: []ucli.Flag{
		&ucli.StringFlag{
			Name:    "filename",
			Aliases: []string{"f"},
			Usage:   "filename path of credentials to store",
		},
	},
	Action: func(cc *ucli.Context) (err error) {
		if cc.Args().Len() != 1 {
			return errors.New("Empty host is not allowed")
		}

		input := pterm.DefaultInteractiveTextInput.WithMask("*")

		key, err := input.Show("Enter Immich API key")
		if err != nil {
			return err
		}

		cred := &cli.Credentials{
			Host: cc.Args().First(),
			Key:  key,
		}

		return cli.Login(cred, cc.String("filename"))
	},
}

var logoutCmd = &ucli.Command{
	Name:        "logout",
	Description: "Remove stored credentials",
	ArgsUsage:   "[host]",
	Flags: []ucli.Flag{
		&ucli.StringFlag{
			Name:    "filename",
			Aliases: []string{"f"},
			Usage:   "filename path of credentials to read",
		},
	},
	Action: func(cc *ucli.Context) (err error) {
		result, _ := pterm.DefaultInteractiveConfirm.Show("Do you really want to remove the credentials?")
		pterm.Println()

		if result {
			return cli.Logout(cc.String("filename"))
		}

		return nil
	},
}
