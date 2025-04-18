.DEFAULT_GOAL := help

build: ## builds the app (only for testing purpose)
	@go build -v -o ./build/imt .
.PHONY: build

clean: ## clean up files generated by coverage or go mod
	@rm -fR ./vendor/ ./cover.* ./build/ ./dist/
.PHONY: clean

configure: ## creates folders and download dependencies
	@mkdir -p ./build
	@go mod download
.PHONY: configure

cover: test ## run unit tests and generates the html coverage file
	@go tool cover -html=./cover.text -o ./cover.html
	@test -f ./cover.text && rm ./cover.text;
.PHONY: cover

help: ## display help screen
	@sed \
        -e '/^[a-zA-Z0-9_\-]*:.*##/!d' \
        -e 's/:.*##\s*/:/' \
        -e 's/^\(.\+\):\(.*\)/$(shell tput setaf 6)\1$(shell tput sgr0):\2/' \
        $(MAKEFILE_LIST) | column -c2 -t -s :
	@echo ''
.PHONY: help

lint: ## golangci linters
	@golangci-lint run ./...
.PHONY: lint

release: ## runs the goreleaser and creates the release for local testing.
	@goreleaser release --snapshot
.PHONY: release

test: ## run unit tests
	@go test -v -race -coverprofile=./cover.text -covermode=atomic $(shell go list ./...)
.PHONY: test
