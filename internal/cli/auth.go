/*
 * Copyright (c) Fabio da Silva Ribeiro <faabiosr@gmail.com>
 * SPDX-License-Identifier: MIT
 */

package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

const perm = 0o700

// Credentials holds the Immich credentials.
type Credentials struct {
	Host string `json:"host"`
	Key  string `json:"key"`
}

// DefaultCredentialsPath returns the path of the credentials file.
func DefaultCredentialsPath() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve user information: %w", err)
	}

	return filepath.Join(u.HomeDir, ".config", "imt", "auth.json"), nil
}

// Login received the credentials and stores into file. The credentials will be
// used to communicate with Immich server.
func Login(cred *Credentials, filename string) (err error) {
	if filename == "" {
		filename, err = DefaultCredentialsPath()
		if err != nil {
			return err
		}
	}

	filename = filepath.Clean(filename)

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create config folder: %w", err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		cerr := f.Close()
		if err == nil {
			err = cerr
		}
	}()

	return json.NewEncoder(f).Encode(cred)
}

// Logout removes stored credentials.
func Logout(filename string) error {
	if filename != "" {
		return os.Remove(filepath.Clean(filename))
	}

	filename, err := DefaultCredentialsPath()
	if err != nil {
		return err
	}

	return os.Remove(filename)
}

// Session returns the credentials stored.
func Session(filename string) (_ *Credentials, err error) {
	if filename == "" {
		filename, err = DefaultCredentialsPath()
		if err != nil {
			return nil, err
		}
	}

	filename = filepath.Clean(filename)

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open config file: %w", err)
	}

	var cred Credentials
	return &cred, json.NewDecoder(f).Decode(&cred)
}
