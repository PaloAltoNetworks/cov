PROJECT_VERSION ?= $(shell git describe --abbrev=0 --tags)
PROJECT_NAME = cov
PROJECT_SHA ?= $(shell git rev-parse HEAD)
PROJECT_RELEASE ?= dev

export GO111MODULE = on
export GOPRIVATE = go.aporeto.io,github.com/aporeto-inc

define VERSIONS_FILE
package configuration

// Various version information.
var (
	ProjectName    = "$(PROJECT_NAME)"
	ProjectVersion = "$(PROJECT_VERSION)"
	ProjectSha     = "$(PROJECT_SHA)"
	ProjectRelease = "$(PROJECT_RELEASE)"
)
endef
export VERSIONS_FILE

init:
	# go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	@echo generating versions.go
	@echo "$$VERSIONS_FILE" > ./internal/configuration/versions.go

remod:
	@cd /tmp && go get go.aporeto.io/remod@master
	@ case "${PROJECT_BRANCH}" in \
	release-*) remod up go.aporeto.io --version "${PROJECT_BRANCH}" ;; \
	*) remod up go.aporeto.io --version "master" ;; \
	esac;

lint: init
	golangci-lint run \
		--deadline=5m \
		--disable-all \
		--exclude-use-default=false \
		--enable=errcheck \
		--enable=goimports \
		--enable=ineffassign \
		--enable=revive \
		--enable=unused \
		--enable=structcheck \
		--enable=staticcheck \
		--enable=varcheck \
		--enable=deadcode \
		--enable=unconvert \
		--enable=misspell \
		--enable=prealloc \
		--enable=nakedret \
		--enable=typecheck \
		./...

test: lint
	go test ./... -race -cover -covermode=atomic -coverprofile=unit_coverage.cov

build_linux: test
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

build: test
	go build

.PHONY: build
