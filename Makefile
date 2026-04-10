.PHONY: help build install test test-short cover bench lint vet fmt fmt-check golden-update build-all clean ci

BINARY     := prompt-improver
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -ldflags "-X main.version=$(VERSION)"
GOTEST     := go test
GOFLAGS    := -count=1

.DEFAULT_GOAL := help

## help: Show all available targets
help:
	@echo "prompt-improver build targets ($(VERSION)):"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'

## build: Build binary with version injection
build:
	GOWORK=off go build $(LDFLAGS) -o $(BINARY) .

## install: Build, install to /usr/local/bin, codesign (macOS)
install: build
	sudo cp $(BINARY) /usr/local/bin/$(BINARY)
	@if [ "$$(uname)" = "Darwin" ]; then \
		sudo codesign -f -s - /usr/local/bin/$(BINARY); \
		echo "Installed and codesigned /usr/local/bin/$(BINARY) $(VERSION)"; \
	else \
		echo "Installed /usr/local/bin/$(BINARY) $(VERSION)"; \
	fi

## test: Run all tests with race detection
test:
	GOWORK=off $(GOTEST) -race -v $(GOFLAGS) ./...

## test-short: Run fast tests only (skip slow integration tests)
test-short:
	GOWORK=off $(GOTEST) -race -short -v $(GOFLAGS) ./...

## cover: Run tests with coverage report
cover:
	GOWORK=off $(GOTEST) -race -coverprofile=coverage.out $(GOFLAGS) ./...
	go tool cover -func=coverage.out
	@echo ""
	@echo "To view HTML report: go tool cover -html=coverage.out"

## bench: Run benchmarks
bench:
	GOWORK=off $(GOTEST) ./pkg/enhancer/ -bench=. -benchmem -count=3

## lint: Run go vet and staticcheck
lint: vet
	@command -v staticcheck >/dev/null 2>&1 || { echo "Install staticcheck: go install honnef.co/go/tools/cmd/staticcheck@latest"; exit 1; }
	staticcheck ./...

## vet: Run go vet
vet:
	GOWORK=off go vet ./...

## fmt: Format all Go files
fmt:
	gofmt -w .

## fmt-check: Check formatting (fails if unformatted)
fmt-check:
	@test -z "$$(gofmt -l .)" || { echo "Files need formatting:"; gofmt -l .; exit 1; }

## golden-update: Regenerate golden test files
golden-update:
	GOWORK=off GOLDEN_UPDATE=1 $(GOTEST) ./pkg/enhancer/ -run TestGolden -v

## build-all: Cross-compile for darwin/amd64, darwin/arm64, linux/amd64
build-all:
	@mkdir -p dist
	GOWORK=off GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 .
	GOWORK=off GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .
	GOWORK=off GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 .
	@echo "Built $(VERSION) for 3 platforms in dist/"

## clean: Remove build artifacts
clean:
	rm -f $(BINARY) coverage.out
	rm -rf dist/

## ci: Full CI pipeline (fmt-check, lint, test, cover, bench)
ci: fmt-check lint test cover bench
	@echo "CI passed ($(VERSION))"

HG_PIPELINE_MK ?= $(or $(wildcard $(abspath $(CURDIR)/../dotfiles/make/pipeline.mk)),$(wildcard $(HOME)/hairglasses-studio/dotfiles/make/pipeline.mk))
-include $(HG_PIPELINE_MK)
