APP_NAME := streampi-server
VERSION := 1.0.0
BUILD_DIR := bin
CMD_PACKAGE := ./cmd/server
LDFLAGS := -s -w

.PHONY: all clean test deps update build run \
        linux-amd64 linux-arm64 linux-armv7 linux-armv6 \
        darwin-amd64 darwin-arm64 \
        windows-amd64 windows-arm64 \
        install help

# Default target
all: clean linux-amd64 linux-arm64 linux-armv7 darwin-amd64 darwin-arm64 windows-amd64

help:
	@echo "Stream-Pi Server Build Targets:"
	@echo ""
	@echo "  make all          - Build for all major platforms"
	@echo "  make build        - Build for current platform"
	@echo "  make run          - Run on current platform"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make deps         - Download dependencies"
	@echo "  make update       - Update dependencies"
	@echo ""
	@echo "Platform-specific builds:"
	@echo "  make linux-amd64  - Linux x86_64"
	@echo "  make linux-arm64  - Linux ARM64 (Pi 4/5)"
	@echo "  make linux-armv7  - Linux ARMv7 (Pi 3)"
	@echo "  make darwin-amd64 - macOS Intel"
	@echo "  make darwin-arm64 - macOS Apple Silicon"
	@echo "  make windows-amd64 - Windows x86_64"

clean:
	@echo "🧹 Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@mkdir -p $(BUILD_DIR)
	@echo "✓ Clean complete"

test:
	@echo "🧪 Running tests..."
	@go test -v ./...

deps:
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod verify
	@echo "✓ Dependencies ready"

update:
	@echo "⬆️  Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "✓ Dependencies updated"

# Build for current platform
build:
	@echo "🔨 Building for current platform..."
	@go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PACKAGE)
	@echo "✓ Build complete: $(BUILD_DIR)/$(APP_NAME)"

# Run on current platform
run:
	@echo "🚀 Running Stream-Pi Server..."
	@go run $(CMD_PACKAGE)

# Linux builds
linux-amd64:
	@echo "🐧 Building for Linux x86_64..."
	@GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(CMD_PACKAGE)
	@echo "✓ Built: $(BUILD_DIR)/$(APP_NAME)-linux-amd64"

linux-arm64:
	@echo "🐧 Building for Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(CMD_PACKAGE)
	@echo "✓ Built: $(BUILD_DIR)/$(APP_NAME)-linux-arm64"

linux-armv7:
	@echo "🐧 Building for Linux ARMv7..."
	@GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-armv7 $(CMD_PACKAGE)
	@echo "✓ Built: $(BUILD_DIR)/$(APP_NAME)-linux-armv7"

linux-armv6:
	@echo "🐧 Building for Linux ARMv6..."
	@GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-armv6 $(CMD_PACKAGE)
	@echo "✓ Built: $(BUILD_DIR)/$(APP_NAME)-linux-armv6"

# macOS builds
darwin-amd64:
	@echo "🍎 Building for macOS Intel..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(CMD_PACKAGE)
	@echo "✓ Built: $(BUILD_DIR)/$(APP_NAME)-darwin-amd64"

darwin-arm64:
	@echo "🍎 Building for macOS Apple Silicon..."
	@GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(CMD_PACKAGE)
	@echo "✓ Built: $(BUILD_DIR)/$(APP_NAME)-darwin-arm64"

# Windows builds
windows-amd64:
	@echo "🪟 Building for Windows x86_64..."
	@GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_PACKAGE)
	@echo "✓ Built: $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe"

windows-arm64:
	@echo "🪟 Building for Windows ARM64..."
	@GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-windows-arm64.exe $(CMD_PACKAGE)
	@echo "✓ Built: $(BUILD_DIR)/$(APP_NAME)-windows-arm64.exe"

# Install to system (macOS/Linux)
install: build
	@echo "📥 Installing to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/
	@echo "✓ Installed: /usr/local/bin/$(APP_NAME)"

# Development helpers
dev: deps build run

watch:
	@echo "👀 Watching for changes..."
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	@air
