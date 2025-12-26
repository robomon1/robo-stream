#!/bin/bash

# Robo-Stream Server - Cross-Platform Build Script
# Builds binaries for all supported platforms

set -e

# Check we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: go.mod not found. Please run this script from the server-go directory."
    exit 1
fi

if [ ! -d "cmd/server" ]; then
    echo "❌ Error: cmd/server directory not found."
    exit 1
fi

# Configuration
APP_NAME="robostream-server"
VERSION="${VERSION:-1.0.0}"
BUILD_DIR="bin"
CMD_PACKAGE="./cmd/server"

# Build flags for smaller binaries
LDFLAGS="-s -w"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🚀 Building Robo-Stream Server v${VERSION} for all platforms..."
echo ""

# Create build directory
mkdir -p ${BUILD_DIR}

# Function to build for a platform
build_platform() {
    local goos=$1
    local goarch=$2
    local goarm=$3
    local output_name=$4
    
    printf "Building for ${YELLOW}${goos}/${goarch}${NC}"
    if [ -n "${goarm}" ]; then
        printf " (ARM v${goarm})"
    fi
    printf "... "
    
    if [ -n "${goarm}" ]; then
        GOOS=${goos} GOARCH=${goarch} GOARM=${goarm} go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${output_name} ${CMD_PACKAGE}
    else
        GOOS=${goos} GOARCH=${goarch} go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${output_name} ${CMD_PACKAGE}
    fi
    
    if [ $? -eq 0 ]; then
        printf "${GREEN}✓${NC}\n"
    else
        printf "${RED}✗${NC}\n"
        exit 1
    fi
}

# Linux builds
echo "📦 Linux builds:"
build_platform "linux" "amd64" "" "${APP_NAME}-linux-amd64"
build_platform "linux" "arm64" "" "${APP_NAME}-linux-arm64"
build_platform "linux" "arm" "7" "${APP_NAME}-linux-armv7"
build_platform "linux" "arm" "6" "${APP_NAME}-linux-armv6"
echo ""

# macOS builds
echo "🍎 macOS builds:"
build_platform "darwin" "amd64" "" "${APP_NAME}-darwin-amd64"
build_platform "darwin" "arm64" "" "${APP_NAME}-darwin-arm64"
echo ""

# Windows builds
echo "🪟 Windows builds:"
build_platform "windows" "amd64" "" "${APP_NAME}-windows-amd64.exe"
build_platform "windows" "arm64" "" "${APP_NAME}-windows-arm64.exe"
echo ""

# Display results
echo "✅ Build complete! Binaries:"
echo ""
ls -lh ${BUILD_DIR}/ | tail -n +2 | awk '{printf "  %-40s %8s\n", $9, $5}'
echo ""

# Calculate total size
total_size=$(du -sh ${BUILD_DIR} | awk '{print $1}')
echo "Total size: ${total_size}"
