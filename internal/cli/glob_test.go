/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cli

import (
	"fmt"
	"strings"
	"testing"
)

// Code extracted from rclone: https://github.com/rclone/rclone
func Test_globToRegexp(t *testing.T) {
	tests := []struct {
		glob string
		want string
		err  string
	}{
		{``, ``, ``},
		{`potato`, `potato`, ``},
		{`potato,sausage`, `potato,sausage`, ``},
		{`/potato`, `/potato`, ``},
		{`potato?sausage`, `potato.sausage`, ``},
		{`potat[oa]`, `potat[oa]`, ``},
		{`potat[a-z]or`, `potat[a-z]or`, ``},
		{`potat[[:alpha:]]or`, `potat[[:alpha:]]or`, ``},
		{`'.' '+' '(' ')' '|' '^' '$'`, `'\.' '\+' '\(' '\)' '\|' '\^' '\$'`, ``},
		{`*.jpg`, `.*\.jpg`, ``},
		{`a{b,c,d}e`, `a(b|c|d)e`, ``},
		{`potato**`, ``, `too many stars`},
		{`potato**sausage`, ``, `too many stars`},
		{`*.p[lm]`, `.*\.p[lm]`, ``},
		{`[\[\]]`, `[\[\]]`, ``},
		{`***potato`, ``, `too many stars`},
		{`***`, ``, `too many stars`},
		{`ab]c`, ``, `mismatched ']'`},
		{`ab[c`, ``, `mismatched '[' and ']'`},
		{`ab{x{cd`, ``, `can't nest`},
		{`ab{}}cd`, ``, `mismatched '{' and '}'`},
		{`ab}c`, ``, `mismatched '{' and '}'`},
		{`ab{c`, ``, `mismatched '{' and '}'`},
		{`*.{jpg,png,gif}`, `.*\.(jpg|png|gif)`, ``},
		{`[a--b]`, ``, `bad glob pattern`},
		{`a\*b`, `a\*b`, ``},
		{`a\\b`, `a\\b`, ``},
		{`a{{.*}}b`, `a(.*)b`, ``},
		{`a{{.*}`, ``, `mismatched '{{' and '}}'`},
		{`{{regexp}}`, `(regexp)`, ``},
		{`\{{{regexp}}`, `\{(regexp)`, ``},
		{`/{{regexp}}`, `/(regexp)`, ``},
		{`/{{\d{8}}}`, `/(\d{8})`, ``},
		{`/{{\}}}`, `/(\})`, ``},
		{`{{(?i)regexp}}`, `((?i)regexp)`, ``},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("glob %s", tt.glob), func(t *testing.T) {
			re, actual := globToRegexp(tt.glob)

			var err string
			if actual != nil {
				err = actual.Error()
			}

			if !strings.Contains(err, tt.err) {
				t.Errorf("unexpected error: %s (expected %s)", err, tt.err)
			}

			if actual != nil {
				t.SkipNow()
			}

			if got := re.String(); tt.want != got {
				t.Errorf("unexpected error: %s (expected %s)", got, tt.want)
			}
		})
	}
}
