/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package errors

// ReadChannel reads errors from an errors channel until is closed, and join
// them.
func ReadChannel(errChan <-chan error) error {
	errs := []error{}
	for err := range errChan {
		errs = append(errs, err)
	}

	return Join(errs...)
}
