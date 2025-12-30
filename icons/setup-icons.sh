#!/bin/bash
# Setup icons for both Robo-Stream Server and Client
# This script should be run from the robo-stream/icons/ directory

set -e

echo "=========================================="
echo "🎨 Robo-Stream Icon Setup"
echo "=========================================="
echo ""

# Check if we're in the right directory
if [ ! -f "robo-stream-client.png" ] || [ ! -f "robo-stream-server.png" ]; then
    echo "❌ Error: Icon files not found!"
    echo "Please run this script from the robo-stream/icons/ directory"
    echo ""
    echo "Expected files:"
    echo "  - robo-stream-client.png"
    echo "  - robo-stream-server.png"
    exit 1
fi

# Check if Python 3 is available
if ! command -v python3 &> /dev/null; then
    echo "❌ Error: python3 not found"
    echo "Please install Python 3"
    exit 1
fi

# Check if Pillow is installed
if ! python3 -c "import PIL" 2> /dev/null; then
    echo "⚠️  Pillow (PIL) not found"
    echo "Installing Pillow..."
    pip3 install Pillow
    echo ""
fi

# Run the Python script
echo "Running icon generator..."
python3 setup-icons.py

echo ""
echo "✅ Icon setup complete!"
echo ""
echo "To build with icons:"
echo "  cd ../server && wails build"
echo "  cd ../client && wails build"
