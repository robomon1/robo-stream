#!/bin/bash
set -e

echo "🧪 Testing build..."
go mod tidy
go build -o bin/streampi-client ./cmd/client

if [ -f bin/streampi-client ]; then
    echo "✅ Build successful!"
    ./bin/streampi-client --version
else
    echo "❌ Build failed!"
    exit 1
fi
