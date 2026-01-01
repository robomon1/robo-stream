# Robo-Stream Client Makefile

.PHONY: all dev build build-all setup-icons clean help

# Default target
all: build

# Setup icons - copy from icons/ to build/ directories
setup-icons:
	@echo "📦 Setting up client icons..."
	@mkdir -p build/darwin
	@mkdir -p build/windows
	@mkdir -p build/linux
	@if [ -f icons/appicon.png ]; then \
		cp icons/appicon.png build/; \
		echo "  ✓ appicon copied"; \
	fi
	@if [ -f icons/darwin/icon.icns ]; then \
		cp icons/darwin/icon.icns build/darwin/; \
		echo "  ✓ macOS icon copied"; \
	fi
	@if [ -f icons/windows/icon.ico ]; then \
		cp icons/windows/icon.ico build/windows/; \
		echo "  ✓ Windows icon copied"; \
	fi
	@if [ -f icons/linux/icon.png ]; then \
		cp icons/linux/icon.png build/linux/; \
		echo "  ✓ Linux icon copied"; \
	fi
	@echo "✅ Client icons ready"

# Development mode
dev: setup-icons
	@echo "🚀 Starting client in dev mode..."
	wails dev

# Build for current platform
build: setup-icons
	@echo "🔨 Building client for current platform..."
	wails build

# Build for all platforms
build-all: setup-icons
	@echo "🔨 Building client for all platforms..."
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 wails build
	@echo "Building for macOS ARM64..."
	GOOS=darwin GOARCH=arm64 wails build
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 wails build
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 wails build
	@echo "Building for Linux ARM64 (Raspberry Pi)..."
	GOOS=linux GOARCH=arm64 wails build
	@echo "✅ All client builds complete"

# Clean build artifacts (but keep icons/)
clean:
	@echo "🧹 Cleaning client build artifacts..."
	rm -rf build
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	@echo "  Note: icons/ directory preserved"
	@echo "✅ Client cleaned"

# Help
help:
	@echo "Robo-Stream Client - Available targets:"
	@echo "  make dev        - Run in development mode with hot reload"
	@echo "  make build      - Build for current platform"
	@echo "  make build-all  - Build for all platforms (macOS, Windows, Linux, ARM)"
	@echo "  make clean      - Remove build artifacts (preserves icons/)"
	@echo "  make setup-icons - Copy icons from icons/ to build/"
	@echo ""
	@echo "Icon locations:"
	@echo "  Source:  client/icons/"
	@echo "  Build:   client/build/{darwin,windows,linux}/"
