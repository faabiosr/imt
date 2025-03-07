/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cli

import (
	"bytes"
	"fmt"
	"regexp"
)

// globToRegexp converts an rsync style glob to a regexp.
//
// Code extracted from rclone: https://github.com/rclone/rclone
//
// nolint:gocyclo
func globToRegexp(glob string) (*regexp.Regexp, error) {
	var re bytes.Buffer
	consecutiveStars := 0
	insertStars := func() error {
		if consecutiveStars > 0 {
			switch consecutiveStars {
			case 1:
				_, _ = re.WriteString(`.*`)
			default:
				return fmt.Errorf("too many stars in %q", glob)
			}
		}
		consecutiveStars = 0
		return nil
	}

	overwriteLastChar := func(c byte) {
		buf := re.Bytes()
		buf[len(buf)-1] = c
	}

	inBraces := false
	inBrackets := 0
	slashed := false
	inRegexp := false    // inside {{ ... }}
	inRegexpEnd := false // have received }} waiting for more

	var next, last rune

	for _, c := range glob {
		next, last = c, next
		if slashed {
			_, _ = re.WriteRune(c)
			slashed = false
			continue
		}
		if inRegexpEnd {
			if c == '}' {
				// Regexp is ending with }} choose longest segment
				// Replace final ) with }
				overwriteLastChar('}')
				_ = re.WriteByte(')')
				continue
			} else {
				inRegexpEnd = false
			}
		}
		if inRegexp {
			if c == '}' && last == '}' {
				inRegexp = false
				inRegexpEnd = true
				// Replace final } with )
				overwriteLastChar(')')
			} else {
				_, _ = re.WriteRune(c)
			}
			continue
		}
		if c != '*' {
			err := insertStars()
			if err != nil {
				return nil, err
			}
		}
		if inBrackets > 0 {
			_, _ = re.WriteRune(c)

			if c == '[' {
				inBrackets++
			}

			if c == ']' {
				inBrackets--
			}
			continue
		}
		switch c {
		case '\\':
			_, _ = re.WriteRune(c)
			slashed = true
		case '*':
			consecutiveStars++
		case '?':
			_, _ = re.WriteString(`.`)
		case '[':
			_, _ = re.WriteRune(c)
			inBrackets++
		case ']':
			return nil, fmt.Errorf("mismatched ']' in glob %q", glob)
		case '{':
			if inBraces {
				if last == '{' {
					inRegexp = true
					inBraces = false
				} else {
					return nil, fmt.Errorf("can't nest '{' '}' in glob %q", glob)
				}
			} else {
				inBraces = true
				_ = re.WriteByte('(')
			}
		case '}':
			if !inBraces {
				return nil, fmt.Errorf("mismatched '{' and '}' in glob %q", glob)
			}

			_ = re.WriteByte(')')
			inBraces = false
		case ',':
			if inBraces {
				_ = re.WriteByte('|')
			} else {
				_, _ = re.WriteRune(c)
			}
		case '.', '+', '(', ')', '|', '^', '$': // regexp meta characters not dealt with above
			_ = re.WriteByte('\\')
			_, _ = re.WriteRune(c)
		default:
			_, _ = re.WriteRune(c)
		}
	}

	err := insertStars()
	if err != nil {
		return nil, err
	}

	if inBrackets > 0 {
		return nil, fmt.Errorf("mismatched '[' and ']' in glob %q", glob)
	}

	if inBraces {
		return nil, fmt.Errorf("mismatched '{' and '}' in glob %q", glob)
	}

	if inRegexp {
		return nil, fmt.Errorf("mismatched '{{' and '}}' in glob %q", glob)
	}

	result, err := regexp.Compile(re.String())
	if err != nil {
		return nil, fmt.Errorf("bad glob pattern %q (regexp %q): %w", glob, re.String(), err)
	}

	return result, nil
}
