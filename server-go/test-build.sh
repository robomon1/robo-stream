#!/bin/bash

# Simple build test script
# This verifies your setup is correct

set -e

echo "🔍 Checking environment..."

# Check we're in the right place
if [ ! -f "go.mod" ]; then
    echo "❌ Error: go.mod not found"
    echo "   Please run this from the server-go directory"
    exit 1
fi

if [ ! -f "cmd/server/main.go" ]; then
    echo "❌ Error: cmd/server/main.go not found"
    exit 1
fi

echo "✓ Directory structure looks good"

# Check Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    exit 1
fi

echo "✓ Go is installed: $(go version)"

# Try a simple build
echo ""
echo "🔨 Attempting build..."
mkdir -p bin

# This is the CORRECT way to build
go build -o bin/streampi-server-test ./cmd/server

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo ""
    ls -lh bin/streampi-server-test
    echo ""
    echo "You can now run: ./bin/streampi-server-test --version"
else
    echo "❌ Build failed"
    exit 1
fi
