#!/bin/bash
# Build Stream-Pi Deck for all platforms

set -e

VERSION="1.0.0"
BUILD_DIR="build"

echo "🚀 Building Stream-Pi Deck v${VERSION} for all platforms..."

# Clean previous builds
rm -rf ${BUILD_DIR}

# macOS Intel
echo "📦 Building for macOS (Intel)..."
wails build -platform darwin/amd64 -clean

# macOS Apple Silicon
echo "📦 Building for macOS (Apple Silicon)..."
wails build -platform darwin/arm64

# macOS Universal
echo "📦 Building for macOS (Universal)..."
wails build -platform darwin/universal

# Windows
echo "📦 Building for Windows..."
wails build -platform windows/amd64

# Linux amd64
echo "📦 Building for Linux (amd64)..."
wails build -platform linux/amd64

# Linux arm64 (for Raspberry Pi 64-bit)
echo "📦 Building for Linux (arm64)..."
wails build -platform linux/arm64

# Linux arm (for Raspberry Pi 32-bit)
echo "📦 Building for Linux (arm)..."
wails build -platform linux/arm

echo ""
echo "✅ Build complete! Binaries are in ${BUILD_DIR}/bin/"
echo ""
echo "macOS:"
echo "  - Stream-Pi Deck.app (Universal)"
echo ""
echo "Windows:"
echo "  - streampi-deck.exe"
echo ""
echo "Linux:"
echo "  - streampi-deck (amd64)"
echo "  - streampi-deck (arm64)"
echo "  - streampi-deck (arm)"
echo ""

# Create distribution packages
echo "📦 Creating distribution packages..."

# macOS DMG (requires macOS)
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "Creating macOS DMG..."
    # TODO: Add DMG creation
fi

# Create Linux tar.gz
echo "Creating Linux packages..."
cd ${BUILD_DIR}/bin

# Linux amd64
if [ -f "streampi-deck" ]; then
    tar czf ../streampi-deck-${VERSION}-linux-amd64.tar.gz streampi-deck
    echo "  ✅ streampi-deck-${VERSION}-linux-amd64.tar.gz"
fi

# Linux arm64
if [ -f "streampi-deck-arm64" ]; then
    tar czf ../streampi-deck-${VERSION}-linux-arm64.tar.gz streampi-deck-arm64
    echo "  ✅ streampi-deck-${VERSION}-linux-arm64.tar.gz"
fi

# Linux arm
if [ -f "streampi-deck-arm" ]; then
    tar czf ../streampi-deck-${VERSION}-linux-arm.tar.gz streampi-deck-arm
    echo "  ✅ streampi-deck-${VERSION}-linux-arm.tar.gz"
fi

# Windows zip
if [ -f "streampi-deck.exe" ]; then
    zip -q ../streampi-deck-${VERSION}-windows-amd64.zip streampi-deck.exe
    echo "  ✅ streampi-deck-${VERSION}-windows-amd64.zip"
fi

cd ../..

echo ""
echo "🎉 All done! Distribution packages are in ${BUILD_DIR}/"
echo ""
echo "Config file locations by OS:"
echo "  macOS:   ~/Library/Application Support/StreamPi/buttons.json"
echo "  Linux:   ~/.config/streampi/buttons.json"
echo "  Windows: %APPDATA%\\StreamPi\\buttons.json"
