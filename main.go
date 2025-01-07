/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package main

import (
	"context"
	"os"

	"github.com/faabiosr/imt/cmd"
)

func main() {
	cmd.Execute(context.Background(), os.Args)
}
