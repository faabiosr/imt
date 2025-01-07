/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

// Package errors provides a single package for all error-related stuffs and
// extends the standard library's errors package.
package errors

import (
	"errors"
	"fmt"
)

// New returns an error with the supplied message.
func New(text string) error {
	return errors.New(text)
}

// As calls the standard library's errors.As function.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is calls the standard library's errors.Is function.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Unwrap calls the standard library's errors.Unwrap function.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Errorf calls the standard library's fmt.Errorf function.
func Errorf(format string, a ...any) error {
	return fmt.Errorf(format, a...)
}
