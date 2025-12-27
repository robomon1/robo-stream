#!/bin/bash
# Build for Linux (detects architecture automatically)

set -e

ARCH=$(uname -m)

case $ARCH in
    x86_64)
        PLATFORM="linux/amd64"
        ;;
    aarch64)
        PLATFORM="linux/arm64"
        ;;
    armv7l)
        PLATFORM="linux/arm"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "🐧 Building Robo-Stream Deck for $PLATFORM..."

wails build -platform $PLATFORM -clean

echo ""
echo "✅ Build complete!"
echo ""
echo "To run:"
echo "  ./build/bin/robostream-deck"
echo ""
echo "Config file will be saved to:"
echo "  ~/.config/robostream/buttons.json"
echo ""
echo "To set server URL:"
echo "  export SERVER_URL=http://10.91.108.170:8080"
echo "  ./build/bin/robostream"
echo ""
echo "For Raspberry Pi touchscreen fullscreen:"
echo "  ./build/bin/robostream &"
echo "  # Then click the fullscreen button in the app"
