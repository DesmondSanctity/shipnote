# shipnote — developer Makefile
#
# Targets are dependency-free (no make recursion across files) and avoid
# wall-clock time so output of `make build` is reproducible from a clean
# tree.

SHELL := /bin/bash

MODULE       := github.com/DesmondSanctity/shipnote
BINARY       := shipnote
PKG_VERSION  := $(MODULE)/internal/version
VERSION      ?= $(shell git describe --tags --always --dirty 2>/dev/null | head -n1)
VERSION      := $(if $(VERSION),$(VERSION),dev)
COMMIT       ?= $(shell git rev-parse --verify HEAD 2>/dev/null | head -n1)
COMMIT       := $(if $(COMMIT),$(COMMIT),none)
DATE         ?= $(shell git log -1 --format=%cI 2>/dev/null | head -n1)
DATE         := $(if $(DATE),$(DATE),unknown)

LDFLAGS := -s -w \
	-X $(PKG_VERSION).Version=$(VERSION) \
	-X $(PKG_VERSION).Commit=$(COMMIT) \
	-X $(PKG_VERSION).Date=$(DATE)

.PHONY: all build test lint fmt vet tidy clean run golden casts help

all: lint test build

build: ## Build the shipnote binary into ./bin
	@mkdir -p bin
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/shipnote

test: ## Run unit tests
	go test -race -count=1 ./...

golden: ## Run determinism / golden-file tests (alias for now)
	go test -race -count=1 -run Golden ./...

lint: ## Run golangci-lint
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not installed. See https://golangci-lint.run/usage/install/"; exit 1; }
	golangci-lint run ./...

fmt: ## Format with gofumpt (falls back to gofmt)
	@command -v gofumpt >/dev/null 2>&1 \
		&& gofumpt -l -w . \
		|| gofmt -l -w .

vet: ## go vet
	go vet ./...

tidy: ## go mod tidy
	go mod tidy

run: build ## Build then run with arguments: `make run ARGS="--help"`
	./bin/$(BINARY) $(ARGS)

clean: ## Remove build artifacts
	rm -rf bin dist

casts: ## Render docs/casts/*.cast → docs/img/*.gif via agg
	@command -v agg >/dev/null 2>&1 || { \
		echo "agg not installed. See https://github.com/asciinema/agg"; exit 1; }
	@mkdir -p docs/img
	@for cast in docs/casts/*.cast; do \
		[ -e "$$cast" ] || { echo "no .cast files in docs/casts/"; exit 0; }; \
		name=$$(basename "$$cast" .cast); \
		echo "agg $$cast -> docs/img/$$name.gif"; \
		agg --cols 100 --rows 25 "$$cast" "docs/img/$$name.gif"; \
	done

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	  | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
