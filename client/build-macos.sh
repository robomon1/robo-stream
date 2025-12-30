#!/bin/bash
# Build for macOS only (Universal binary)

set -e

echo "🍎 Building Robo-Stream Client for macOS..."

wails build -platform darwin/universal -clean

echo ""
echo "✅ Build complete!"
echo ""
echo "To run:"
echo "  open ./build/bin/robo-stream-client.app"
echo ""
echo "Config file will be saved to:"
echo "  ~/Library/Application Support/RoboStream-Client/buttons.json"
echo ""
echo "To set server URL:"
echo "  export SERVER_URL=http://10.91.108.170:8080"
echo "  open ./build/bin/robo-stream-client.app"
