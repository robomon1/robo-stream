BINARY    := zoomosc-controller
MODULE    := github.com/robomon1/robostream-zoomosc-controller
# VERSION can be overridden from the parent Makefile (make -C plugins/zoomosc VERSION=x.x.x).
# Falls back to version.txt in this directory, then to git describe.
VERSION   ?= $(shell cat version.txt 2>/dev/null || git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -ldflags "-s -w -X main.Version=$(VERSION)"

# Output directory for release archives
DIST := dist

# Detect current platform so `make build` outputs to dist/ alongside the
# platform-specific targets rather than dropping a bare binary in the root.
NATIVE_OS   := $(shell go env GOOS)
NATIVE_ARCH := $(shell go env GOARCH)
NATIVE_EXT  := $(if $(filter windows,$(NATIVE_OS)),.exe,)

.PHONY: build build-all build-mac-arm64 build-mac-amd64 build-windows build-linux \
        release clean tidy test

## build: build for the current platform (output to dist/, consistent with platform targets)
build:
	mkdir -p $(DIST)
	go build $(LDFLAGS) -o $(DIST)/$(BINARY)-$(NATIVE_OS)-$(NATIVE_ARCH)$(NATIVE_EXT) .

## build-all: build for all supported platforms
build-all: build-mac-arm64 build-mac-amd64 build-windows build-linux

## build-mac-arm64: build for Apple Silicon macOS
build-mac-arm64:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-arm64 .

## build-mac-amd64: build for Intel macOS
build-mac-amd64:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-amd64 .

## build-windows: build for Windows (amd64)
build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-windows-amd64.exe .

## build-linux: build for Linux (amd64)
build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-amd64 .

## release: build all platforms and create release archives in dist/
release: build-all
	mkdir -p $(DIST)
	cd $(DIST) && tar czf $(BINARY)-$(VERSION)-darwin-arm64.tar.gz  $(BINARY)-darwin-arm64
	cd $(DIST) && tar czf $(BINARY)-$(VERSION)-darwin-amd64.tar.gz  $(BINARY)-darwin-amd64
	cd $(DIST) && zip    $(BINARY)-$(VERSION)-windows-amd64.zip     $(BINARY)-windows-amd64.exe
	cd $(DIST) && tar czf $(BINARY)-$(VERSION)-linux-amd64.tar.gz   $(BINARY)-linux-amd64
	@echo "Release archives written to $(DIST)/"

## test: run tests
test:
	go test ./...

## tidy: tidy go modules
tidy:
	go mod tidy

## clean: remove build artifacts
clean:
	rm -rf $(DIST) $(BINARY) $(BINARY).exe
