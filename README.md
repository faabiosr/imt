# imt

[![Build Status](https://img.shields.io/github/actions/workflow/status/faabiosr/imt/test.yaml?logo=github&style=flat-square)](https://github.com/faabiosr/imt/actions?query=workflow:test)
[![Codecov branch](https://img.shields.io/codecov/c/github/faabiosr/imt/master.svg?style=flat-square)](https://codecov.io/gh/faabiosr/imt)
[![Go Report Card](https://goreportcard.com/badge/github.com/faabiosr/imt?style=flat-square)](https://goreportcard.com/report/github.com/faabiosr/imt)
[![Release](https://img.shields.io/github/v/release/faabiosr/imt?display_name=tag&style=flat-square)](https://github.com/faabiosr/imt/releases)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](https://github.com/faabiosr/imt/blob/master/LICENSE)

## :tada: Overview
A collection of command-line tools for [Immich](https://immich.app/).

## :relaxed: Motivation
Immich is a great tool for managing photos, however when you have a big collection of pictures, it is hard to manage albums, especially if you a different way to organize like me.

## :dart: Installation

### Unix-like

#### Manual installation
```sh
# by default will install into ~/.local/bin folder.
curl -sSL https://raw.githubusercontent.com/faabiosr/imt/main/install.sh | bash 

# install into /usr/local/bin
curl -sSL https://raw.githubusercontent.com/faabiosr/imt/main/install.sh | sudo INSTALL_PATH=/usr/local/bin bash
```

### go
```sh
go install github.com/faabiosr/imt@latest
```

## :gem: Usage

### Login using Immich API Key (please generate one before use)
```sh
imt login http://your-immich-server
```

### Logout (remove the stored credentials)
```sh
imt logout
```

### Create albums based on folder structure
```sh
# will create albums for the folders inside the `/home/user/photos`.
imt album auto-create /home/user/photos/

# will create albums recursivelly for the folders inside the `/home/user/photos`.
imt album auto-create --recursive /home/user/photos/

# will create albums recursivelly and skip levels size for the folders inside the `/home/user/photos`.
imt album auto-create --recursive --skip-levels 2 -/home/user/photos/

# will create albums from config file.
imt album auto-create --from-config example_auto_create.json

# for more option please run:
imt album auto-create -h
```

### Server info
```sh
# Shows server info
imt info
```

## :toolbox: Development

### Requirements

The entire environment is based on Golang, and you need to install the tools below:
- Install [Go](https://golang.org)
- Install [GolangCI-Lint](https://github.com/golangci/golangci-lint#install) - Linter

### Makefile

Please run the make target below to see the provided targets.

```sh
$ make help
```

## :page_with_curl: License

This project is released under the MIT licence. See [LICENSE](https://github.com/faabiosr/imt/blob/master/LICENSE) for more details.
