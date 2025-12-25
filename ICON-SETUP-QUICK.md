# Icon Setup - Quick Reference

Your `robo_stream.jpg` icon will be used for both server-go and client-go!

## 📋 Prerequisites

Place your icon file:
```bash
~/git/stream-pi-go/robo_stream.jpg  # 1136 × 896 pixels
```

## 🎨 Client-Go (Desktop App) - Automatic Icon

### Option 1: Using Bash Script (requires ImageMagick)

```bash
cd ~/git/stream-pi-go/client-go
./setup-icons.sh
```

**If ImageMagick not installed:**
```bash
# macOS
brew install imagemagick

# Linux/Raspberry Pi
sudo apt install imagemagick
```

### Option 2: Using Python Script (easier!)

```bash
cd ~/git/stream-pi-go/client-go

# Install Pillow if needed
pip3 install Pillow

# Run converter
./setup-icons.py
```

Both scripts create:
- `build/appicon.png` - 1024x1024 (Wails uses this)
- `build/darwin/icon.icns` - macOS app icon
- `build/windows/icon.ico` - Windows exe icon
- `build/linux/icon.png` - Linux 512x512

## 🏗️ Build with Icons

After running either script:

```bash
# macOS
./build-macos.sh
open build/bin/Stream-Pi\ Deck.app

# Linux/Raspberry Pi
./build-linux.sh
./build/bin/streampi-deck

# Windows
wails build -platform windows/amd64
```

Your icon will show:
- ✅ In Finder/File Manager
- ✅ In Dock/Taskbar
- ✅ In Applications folder
- ✅ When app is running
- ✅ In window title bar

## 🖥️ Server-Go (CLI App) - Optional

Server-go is a terminal app, but you can add an icon for macOS app bundle:

### macOS App Bundle (Optional)

```bash
cd ~/git/stream-pi-go/server-go

# First, generate icons for client-go
cd ../client-go
./setup-icons.sh  # or ./setup-icons.py
cd ../server-go

# Create app bundle
mkdir -p StreamPi-Server.app/Contents/{MacOS,Resources}
cp bin/streampi-server StreamPi-Server.app/Contents/MacOS/
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
</dict>
</plist>
EOF

# Now double-click StreamPi-Server.app!
```

## ✅ Verification

### Check Icons Generated

```bash
cd ~/git/stream-pi-go/client-go
ls -lh build/appicon.png           # Should be ~100KB
ls -lh build/darwin/icon.icns      # macOS only
ls -lh build/windows/icon.ico      # Should exist
ls -lh build/linux/icon.png        # Should be ~50KB
```

### Test the Icon

**macOS:**
```bash
./build-macos.sh
open build/bin/Stream-Pi\ Deck.app
# Check Dock - your robo icon should appear! 🤖
```

**Linux:**
```bash
./build-linux.sh
./build/bin/streampi-deck
# Check taskbar - your robo icon! 🤖
```

## 🔄 Change Icon Later

Want a different icon?

```bash
# 1. Replace the file
cp ~/new-icon.jpg ~/git/stream-pi-go/robo_stream.jpg

# 2. Re-run converter
cd ~/git/stream-pi-go/client-go
./setup-icons.py  # or ./setup-icons.sh

# 3. Rebuild
./build-macos.sh  # or ./build-linux.sh

# Done! New icon applied.
```

## 🚨 Troubleshooting

### "Pillow not found" (Python script)
```bash
pip3 install Pillow
```

### "convert: command not found" (Bash script)
```bash
# macOS
brew install imagemagick

# Linux
sudo apt install imagemagick
```

### Icon not showing on macOS
```bash
# Clear icon cache
sudo rm -rf /Library/Caches/com.apple.iconservices.store
killall Finder
killall Dock
```

### Icon not showing on Linux
```bash
# Update icon cache
sudo gtk-update-icon-cache /usr/share/icons/hicolor
```

## 📦 What Gets Created

```
client-go/build/
├── appicon.png          # 1024x1024 - Master icon
├── darwin/
│   └── icon.icns       # macOS bundle (16-1024px @1x and @2x)
├── windows/
│   └── icon.ico        # Windows exe (16-256px)
└── linux/
    └── icon.png        # Linux (512x512)
```

All platforms get your robot icon! 🤖✨

## 📚 Full Documentation

See `ICONS.md` for complete details including:
- Creating desktop entries for server-go
- Advanced icon customization
- Platform-specific tips
- Icon specifications

## 🎯 Summary

**Shortest path:**
```bash
# 1. Put icon in project root
cp ~/robo_stream.jpg ~/git/stream-pi-go/

# 2. Convert it
cd ~/git/stream-pi-go/client-go
pip3 install Pillow
./setup-icons.py

# 3. Build
./build-macos.sh

# 4. Run and enjoy! 🤖
open build/bin/Stream-Pi\ Deck.app
```

Your robot is now the face of Stream-Pi! 🎉
