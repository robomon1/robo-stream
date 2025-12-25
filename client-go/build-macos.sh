#!/bin/bash
# Build for macOS only (Universal binary)

set -e

echo "🍎 Building Stream-Pi Deck for macOS..."

wails build -platform darwin/universal -clean

echo ""
echo "✅ Build complete!"
echo ""
echo "To run:"
echo "  open ./build/bin/Stream-Pi\\ Deck.app"
echo ""
echo "Config file will be saved to:"
echo "  ~/Library/Application Support/StreamPi/buttons.json"
echo ""
echo "To set server URL:"
echo "  export SERVER_URL=http://10.91.108.170:8080"
echo "  open ./build/bin/Stream-Pi\\ Deck.app"
