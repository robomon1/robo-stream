#!/bin/bash
set -e

echo "🧪 Testing build..."
go mod tidy
go build -o bin/robostream-client ./cmd/client

if [ -f bin/robostream-client ]; then
    echo "✅ Build successful!"
    ./bin/robostream-client --version
else
    echo "❌ Build failed!"
    exit 1
fi
