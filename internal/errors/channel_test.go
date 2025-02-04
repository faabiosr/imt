/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package errors

import (
	"sync"
	"testing"
)

func TestReadChannel(t *testing.T) {
	msgs := []string{"one", "two", "three"}
	errs := make(chan error, len(msgs))

	var wg sync.WaitGroup

	for _, m := range msgs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- New(m)
		}()
	}

	wg.Wait()
	close(errs)

	err := ReadChannel(errs)
	if err == nil {
		t.Error("expected an error, got nil")
	}
}
