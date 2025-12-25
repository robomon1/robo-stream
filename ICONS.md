# Adding Icons to Stream-Pi

Guide for adding your robo_stream.jpg icon to both server-go and client-go.

## Step 1: Place Your Icon

Place `robo_stream.jpg` in the project root:

```bash
# Your icon should be here:
~/git/stream-pi-go/robo_stream.jpg
```

## Step 2: Convert Icon for Client-Go (Desktop App)

```bash
cd ~/git/stream-pi-go/client-go
./setup-icons.sh
```

This script will:
1. Convert your JPG to PNG (1024x1024)
2. Create macOS .icns file (all required sizes)
3. Create Windows .ico file (all required sizes)
4. Create Linux .png file (512x512)

**Output:**
```
build/
├── appicon.png         # 1024x1024 - Wails auto-detect
├── darwin/
│   └── icon.icns      # macOS bundle icon
├── windows/
│   └── icon.ico       # Windows executable icon
└── linux/
    └── icon.png       # Linux 512x512
```

## Step 3: Build with Icons

### macOS
```bash
./build-macos.sh
```

Your app will now have the custom icon:
- In Finder
- In Dock
- In Applications folder
- When running

### Linux/Raspberry Pi
```bash
./build-linux.sh
```

The icon will show:
- In file manager
- In taskbar
- In application menu
- When running

### Windows
```bash
wails build -platform windows/amd64
```

The .exe will have your icon.

## For Server-Go (CLI App)

Server-go is a command-line app, but we can still add an icon for when it's packaged.

### macOS - Create App Bundle (Optional)

If you want server-go to have an icon on macOS:

```bash
cd ~/git/stream-pi-go/server-go

# Create app bundle structure
mkdir -p StreamPi-Server.app/Contents/MacOS
mkdir -p StreamPi-Server.app/Contents/Resources

# Copy binary
cp bin/streampi-server StreamPi-Server.app/Contents/MacOS/

# Copy icon (after running client-go/setup-icons.sh)
cp ../client-go/build/darwin/icon.icns StreamPi-Server.app/Contents/Resources/

# Create Info.plist
cat > StreamPi-Server.app/Contents/Info.plist << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>streampi-server</string>
    <key>CFBundleIconFile</key>
    <string>icon.icns</string>
    <key>CFBundleIdentifier</key>
    <string>com.streampi.server</string>
    <key>CFBundleName</key>
    <string>Stream-Pi Server</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
</dict>
</plist>
EOF

# Now you can double-click StreamPi-Server.app
open StreamPi-Server.app
```

### Linux - Desktop Entry

```bash
cd ~/git/stream-pi-go/server-go

# Copy icon (512x512 PNG)
sudo cp ../client-go/build/linux/icon.png /usr/share/icons/hicolor/512x512/apps/streampi-server.png

# Update icon cache
sudo gtk-update-icon-cache /usr/share/icons/hicolor

# Create desktop entry
cat > ~/.local/share/applications/streampi-server.desktop << EOF
[Desktop Entry]
Name=Stream-Pi Server
Exec=$HOME/git/stream-pi-go/server-go/bin/streampi-server
Icon=streampi-server
Type=Application
Categories=AudioVideo;
Terminal=true
EOF

# Update desktop database
update-desktop-database ~/.local/share/applications
```

## Verification

### Check Icons Are Generated

```bash
cd ~/git/stream-pi-go/client-go

# Should see all icon files
ls -lh build/appicon.png
ls -lh build/darwin/icon.icns
ls -lh build/windows/icon.ico
ls -lh build/linux/icon.png
```

### Test on macOS

```bash
# Build and check
./build-macos.sh

# View app info (should show your icon)
mdls -name kMDItemDisplayName -name kMDItemFSSize build/bin/Stream-Pi\ Deck.app

# Open app
open build/bin/Stream-Pi\ Deck.app
# Check Dock - your icon should appear!
```

### Test on Linux

```bash
# Build
./build-linux.sh

# Run (icon should show in taskbar)
./build/bin/streampi-deck
```

## Troubleshooting

### "convert: command not found"

**macOS:**
```bash
brew install imagemagick
```

**Linux:**
```bash
sudo apt install imagemagick
```

### Icon doesn't show on macOS

```bash
# Clear icon cache
sudo rm -rfv /Library/Caches/com.apple.iconservices.store
killall Finder
killall Dock
```

### Icon doesn't show on Linux

```bash
# Update icon cache
sudo gtk-update-icon-cache /usr/share/icons/hicolor
```

### Want to change the icon later

1. Replace `robo_stream.jpg` in project root
2. Run `./setup-icons.sh` again
3. Rebuild: `./build-macos.sh` or `./build-linux.sh`

## Icon Specifications

### Your Source Icon
- **File**: robo_stream.jpg
- **Size**: 1136 × 896 pixels
- **Format**: JPEG

### Generated Icons

| Platform | Format | Sizes | Purpose |
|----------|--------|-------|---------|
| **macOS** | .icns | 16, 32, 128, 256, 512, 1024px (@1x and @2x) | App bundle |
| **Windows** | .ico | 16, 32, 48, 64, 128, 256px | Executable |
| **Linux** | .png | 512x512 | Desktop entry |
| **Wails** | .png | 1024x1024 | Auto-conversion source |

## Summary

**Quick setup:**
```bash
# 1. Place icon
cp ~/Downloads/robo_stream.jpg ~/git/stream-pi-go/

# 2. Convert to all formats
cd ~/git/stream-pi-go/client-go
./setup-icons.sh

# 3. Build with icons
./build-macos.sh  # or ./build-linux.sh

# 4. Enjoy your branded app!
open build/bin/Stream-Pi\ Deck.app
```

Your app will now have your custom icon everywhere! 🎨
