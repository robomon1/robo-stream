#!/bin/bash

# Robo-Stream Client Build Script
# Builds the client for multiple platforms

set -e

VERSION=${VERSION:-"1.0.0"}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

echo "🚀 Building Robo-Stream Client v${VERSION} for all platforms..."

# Clean previous builds
rm -rf bin/
mkdir -p bin/

# Build for multiple platforms
echo ""
echo "📦 Linux builds:"
echo "Building for linux/amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/robostream-client-linux-amd64 ./cmd/client

echo "Building for linux/arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o bin/robostream-client-linux-arm64 ./cmd/client

echo ""
echo "🍎 macOS builds:"
echo "Building for darwin/amd64 (Intel)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/robostream-client-darwin-amd64 ./cmd/client

echo "Building for darwin/arm64 (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o bin/robostream-client-darwin-arm64 ./cmd/client

echo ""
echo "🪟 Windows builds:"
echo "Building for windows/amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/robostream-client-windows-amd64.exe ./cmd/client

echo ""
echo "✅ Build complete! Binaries in bin/"
ls -lh bin/
