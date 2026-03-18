BINARY     := gitlens
MODULE     := gitlens
GO         := go
CGO_FLAGS  := CGO_ENABLED=1
NOCGO_FLAGS := CGO_ENABLED=0
GOFLAGS    :=
LDFLAGS    := -ldflags="-s -w"
BUILD_DIR  := .
COVERAGE   := coverage.out

# Detect OS for install path
ifeq ($(shell uname), Darwin)
  INSTALL_DIR := /usr/local/bin
else
  INSTALL_DIR := /usr/local/bin
endif

.PHONY: all build build-nocgo test test-cover test-nocgo vet lint fmt install clean help

all: build

## build: Build binary with CGo (syntax highlighting enabled)
build:
	$(CGO_FLAGS) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY) $(BUILD_DIR)

## build-nocgo: Build binary without CGo (no C compiler required)
build-nocgo:
	$(NOCGO_FLAGS) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY)-nocgo $(BUILD_DIR)

## test: Run all tests with CGo enabled
test:
	$(CGO_FLAGS) $(GO) test ./... -count=1

## test-cover: Run tests with coverage report (opens coverage.html)
test-cover:
	$(CGO_FLAGS) $(GO) test ./... -coverprofile=$(COVERAGE) -count=1
	$(GO) tool cover -html=$(COVERAGE) -o coverage.html
	@echo "Coverage report: coverage.html"

## test-nocgo: Verify tests also pass without CGo
test-nocgo:
	$(NOCGO_FLAGS) $(GO) test ./... -count=1

## vet: Run go vet
vet:
	$(GO) vet ./...

## lint: Run golangci-lint (install: https://golangci-lint.run/usage/install/)
lint:
	golangci-lint run ./...

## fmt: Format all Go source files
fmt:
	$(GO) fmt ./...

## install: Install binary to $(INSTALL_DIR)
install: build
	install -m 755 $(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "Installed to $(INSTALL_DIR)/$(BINARY)"

## clean: Remove build artifacts
clean:
	rm -f $(BINARY) $(BINARY)-nocgo $(COVERAGE) coverage.html

## help: Print this help message
help:
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | column -t -s ':'
