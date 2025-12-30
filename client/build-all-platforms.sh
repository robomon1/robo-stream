#!/bin/bash
# Build Robo-Stream Client for all platforms

set -e

VERSION="1.0.0"
BUILD_DIR="build"

echo "🚀 Building Robo-Stream Client v${VERSION} for all platforms..."

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
echo "  - robo-stream-client.app (Universal)"
echo ""
echo "Windows:"
echo "  - robo-stream-client.exe"
echo ""
echo "Linux:"
echo "  - robo-stream-client (amd64)"
echo "  - robo-stream-client (arm64)"
echo "  - robo-stream-client (arm)"
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
if [ -f "robo-stream-client" ]; then
    tar czf ../robo-stream-client-${VERSION}-linux-amd64.tar.gz robo-stream-client
    echo "  ✅ robo-stream-client-${VERSION}-linux-amd64.tar.gz"
fi

# Linux arm64
if [ -f "robo-stream-client-arm64" ]; then
    tar czf ../robo-stream-client-${VERSION}-linux-arm64.tar.gz robo-stream-client-arm64
    echo "  ✅ robo-stream-client-${VERSION}-linux-arm64.tar.gz"
fi

# Linux arm
if [ -f "robo-stream-client-arm" ]; then
    tar czf ../streampi-deck-${VERSION}-linux-arm.tar.gz robo-stream-client-arm
    echo "  ✅ streampi-deck-${VERSION}-linux-arm.tar.gz"
fi

# Windows zip
if [ -f "robo-stream-client.exe" ]; then
    zip -q ../robo-stream-client-${VERSION}-windows-amd64.zip robo-stream-client.exe
    echo "  ✅ robo-stream-client-${VERSION}-windows-amd64.zip"
fi

cd ../..

echo ""
echo "🎉 All done! Distribution packages are in ${BUILD_DIR}/"
echo ""
echo "Config file locations by OS:"
echo "  macOS:   ~/Library/Application Support/RoboStream-Client/buttons.json"
echo "  Linux:   ~/.config/robo-stream-client/buttons.json"
echo "  Windows: %APPDATA%\\RoboStream-Client\\buttons.json"
