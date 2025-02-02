/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAuth(t *testing.T) {
	t.Run("login failed to create config folder", func(t *testing.T) {
		tmp, err := os.MkdirTemp("", "testing")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		f, err := os.CreateTemp(tmp, "tt")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		_, _ = f.WriteString("testing")
		_ = f.Close()

		t.Cleanup(func() {
			_ = os.Remove(f.Name())
			_ = os.RemoveAll(tmp)
		})

		filename := filepath.Join(f.Name(), "auth.json")

		cred := &Credentials{}

		if err := Login(cred, filename); err == nil {
			t.Error("expected an error, got nil")
		}
	})

	t.Run("login and logout success", func(t *testing.T) {
		tmp, err := os.MkdirTemp("", "testing")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		t.Cleanup(func() {
			_ = os.RemoveAll(tmp)
		})

		filename := filepath.Join(tmp, "auth.json")

		cred := &Credentials{
			Host: "https://immich.app",
			Key:  "da32e327-43c3-4578-a3b8-fd1dfea33d58",
		}

		if err := Login(cred, filename); err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		res, err := Session(filename)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		if res.Host != cred.Host {
			t.Errorf("unexpected host: %s (expected %s)", res.Host, cred.Host)
		}

		if res.Key != cred.Key {
			t.Errorf("unexpected Key: %s (expected %s)", res.Key, cred.Key)
		}

		if err := Logout(filename); err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		if _, err := os.Stat(filename); os.IsExist(err) {
			t.Errorf("file should not exists: %v", err)
		}
	})

	t.Run("login and logout default path", func(t *testing.T) {
		cred := &Credentials{
			Host: "https://immich.app",
			Key:  "da32e327-43c3-4578-a3b8-fd1dfea33d58",
		}

		if err := Login(cred, ""); err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		res, err := Session("")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		if res.Host != cred.Host {
			t.Errorf("unexpected host: %s (expected %s)", res.Host, cred.Host)
		}

		if res.Key != cred.Key {
			t.Errorf("unexpected Key: %s (expected %s)", res.Key, cred.Key)
		}

		if err := Logout(""); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}

func TestAuthSession(t *testing.T) {
	_, err := Session("/tmp/invalid.json")
	if err == nil {
		t.Error("expected an error, got nil")
	}
}
