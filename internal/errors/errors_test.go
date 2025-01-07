/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package errors

import (
	"os"
	"testing"
)

func TestUnwrap(t *testing.T) {
	err := Errorf("foo: %w", New("foo"))
	got := Unwrap(err)

	if expected := "foo"; expected != got.Error() {
		t.Errorf("unexpected error: want %s, got %s", expected, got)
	}
}

func TestIs(t *testing.T) {
	err := Errorf("foo: %w", os.ErrExist)
	got := Is(err, os.ErrExist)
	if !got {
		t.Errorf("want true, got %t", got)
	}
}

type FooErr interface {
	Foo()
}

type fooErr string

func (e fooErr) Error() string { return string(e) }

func (e fooErr) Foo() {}

func TestAs(t *testing.T) {
	target := new(FooErr)

	err := Errorf("foo: %w", fooErr("failed"))
	got := As(err, target)

	if !got {
		t.Errorf("want true, got %t", got)
	}
}
