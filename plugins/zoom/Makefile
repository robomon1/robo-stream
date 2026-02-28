BINARY    := zoom-controller
MODULE    := github.com/robomon1/robostream-zoom-controller
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -ldflags "-s -w -X main.Version=$(VERSION)"

# Output directory for release archives
DIST := dist

.PHONY: build build-all build-mac-arm64 build-mac-amd64 build-windows build-linux \
        release clean tidy test

## build: build for the current platform
build:
	go build $(LDFLAGS) -o $(BINARY) .

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
