#!/bin/bash
# Convert robo_stream.jpg to all needed icon formats

set -e

ICON_SOURCE="./robo_stream.jpg"
BUILD_DIR="build"

if [ ! -f "$ICON_SOURCE" ]; then
    echo "❌ Error: robo_stream.jpg not found in project root"
    echo "Please place robo_stream.jpg in ~/git/stream-pi-go/"
    exit 1
fi

echo "🎨 Converting icon to all formats..."

# Create build directory if it doesn't exist
mkdir -p "$BUILD_DIR"

# Install imagemagick if not present (macOS)
if ! command -v convert &> /dev/null; then
    echo "📦 Installing ImageMagick..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        brew install imagemagick
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        sudo apt install imagemagick
    fi
fi

# Convert to PNG and resize to 1024x1024 (Wails standard)
echo "Creating appicon.png (1024x1024)..."
convert "$ICON_SOURCE" -resize 1024x1024 -background transparent -gravity center -extent 1024x1024 "$BUILD_DIR/appicon.png"

# For macOS - create icon sizes
echo "Creating macOS icons..."
mkdir -p "$BUILD_DIR/darwin"

# Create iconset directory for macOS .icns
ICONSET="$BUILD_DIR/darwin/icon.iconset"
mkdir -p "$ICONSET"

# Generate all required macOS icon sizes
convert "$ICON_SOURCE" -resize 16x16 "$ICONSET/icon_16x16.png"
convert "$ICON_SOURCE" -resize 32x32 "$ICONSET/icon_16x16@2x.png"
convert "$ICON_SOURCE" -resize 32x32 "$ICONSET/icon_32x32.png"
convert "$ICON_SOURCE" -resize 64x64 "$ICONSET/icon_32x32@2x.png"
convert "$ICON_SOURCE" -resize 128x128 "$ICONSET/icon_128x128.png"
convert "$ICON_SOURCE" -resize 256x256 "$ICONSET/icon_128x128@2x.png"
convert "$ICON_SOURCE" -resize 256x256 "$ICONSET/icon_256x256.png"
convert "$ICON_SOURCE" -resize 512x512 "$ICONSET/icon_256x256@2x.png"
convert "$ICON_SOURCE" -resize 512x512 "$ICONSET/icon_512x512.png"
convert "$ICON_SOURCE" -resize 1024x1024 "$ICONSET/icon_512x512@2x.png"

# Create .icns file (macOS only)
if [[ "$OSTYPE" == "darwin"* ]]; then
    iconutil -c icns "$ICONSET" -o "$BUILD_DIR/darwin/icon.icns"
    echo "✅ Created icon.icns"
    rm -rf "$ICONSET"
else
    echo "⚠️  Skipping .icns creation (requires macOS)"
fi

# For Windows - create .ico with multiple sizes
echo "Creating Windows icon..."
mkdir -p "$BUILD_DIR/windows"
convert "$ICON_SOURCE" -resize 256x256 \
    \( -clone 0 -resize 16x16 \) \
    \( -clone 0 -resize 32x32 \) \
    \( -clone 0 -resize 48x48 \) \
    \( -clone 0 -resize 64x64 \) \
    \( -clone 0 -resize 128x128 \) \
    \( -clone 0 -resize 256x256 \) \
    -delete 0 "$BUILD_DIR/windows/icon.ico"
echo "✅ Created icon.ico"

# For Linux - create 512x512 PNG
echo "Creating Linux icon..."
mkdir -p "$BUILD_DIR/linux"
convert "$ICON_SOURCE" -resize 512x512 "$BUILD_DIR/linux/icon.png"
echo "✅ Created Linux icon.png"

echo ""
echo "✅ Icon conversion complete!"
echo ""
echo "Created:"
echo "  - build/appicon.png (1024x1024) - Wails auto-detect"
echo "  - build/darwin/icon.icns - macOS"
echo "  - build/windows/icon.ico - Windows"
echo "  - build/linux/icon.png - Linux"
echo ""
echo "Next step: Run 'wails build' to use these icons"
