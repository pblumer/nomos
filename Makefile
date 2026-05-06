.PHONY: help test vet fmt build run demo validate-demo serve-demo clean version version-json release-build

VERSION ?= dev
BUILT_BY ?= source
MODULE := $(shell go list -m)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
DIRTY := $(shell if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then test -n "$$(git status --porcelain 2>/dev/null)" && echo true || echo false; else echo unknown; fi)

LDFLAGS := -X $(MODULE)/internal/version.Version=$(VERSION) \
	-X $(MODULE)/internal/version.Commit=$(COMMIT) \
	-X $(MODULE)/internal/version.Date=$(DATE) \
	-X $(MODULE)/internal/version.Dirty=$(DIRTY) \
	-X $(MODULE)/internal/version.BuiltBy=$(BUILT_BY)

help:
	@echo "Targets: fmt vet test build version version-json release-build demo validate-demo serve-demo clean"
fmt:
	gofmt -w .
vet:
	go vet ./...
test:
	go test ./...
build:
	go build -ldflags "$(LDFLAGS)" -o ./bin/nomos ./cmd/nomos
version: build
	./bin/nomos version
version-json: build
	./bin/nomos version --format json
release-build:
	@if [ "$(VERSION)" = "dev" ] || [ -z "$(VERSION)" ]; then echo "VERSION must be set to a release value, e.g. VERSION=v0.1.0"; exit 1; fi
	$(MAKE) build VERSION=$(VERSION) BUILT_BY=release
demo: build
	rm -rf ./tmp/demo-cosmos
	NOMOS_BIN=./bin/nomos COSMOS_PATH=./tmp/demo-cosmos sh scripts/create-demo-cosmos.sh
validate-demo: build
	./bin/nomos validate --path ./tmp/demo-cosmos
serve-demo: build
	./bin/nomos serve --path ./tmp/demo-cosmos --listen 127.0.0.1:8080
clean:
	rm -rf ./bin ./tmp
