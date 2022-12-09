PROJECT_VERSION ?= $(shell git describe --abbrev=0 --tags)
PROJECT_NAME = cov
PROJECT_SHA ?= $(shell git rev-parse HEAD)
PROJECT_RELEASE ?= dev

default: lint build

lint:
	golangci-lint run \
		--deadline=5m \
		--disable-all \
		--exclude-use-default=false \
		--exclude=package-comments \
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

build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -trimpath

build_linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -trimpath

build_darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -trimpath

.PHONY: build
