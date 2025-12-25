# Building & Packaging Stream-Pi Deck

Complete guide for building and distributing Stream-Pi Deck on all platforms.

## Config File Locations (OS-Standard)

The app now uses OS-standard configuration directories:

| OS | Config Location |
|---|---|
| **macOS** | `~/Library/Application Support/StreamPi/buttons.json` |
| **Linux** | `~/.config/streampi/buttons.json` |
| **Windows** | `%APPDATA%\StreamPi\buttons.json` |

The app will automatically create these directories on first run.

## Prerequisites

### All Platforms
```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### macOS
```bash
xcode-select --install
```

### Linux/Raspberry Pi
```bash
sudo apt update
sudo apt install build-essential libgtk-3-dev libwebkit2gtk-4.0-dev
```

### Windows
- Install [WebView2 Runtime](https://developer.microsoft.com/en-us/microsoft-edge/webview2/)
- Install [Go](https://golang.org/dl/)

## Quick Build (Current Platform Only)

### macOS
```bash
./build-macos.sh
open ./build/bin/Stream-Pi\ Deck.app
```

### Linux/Raspberry Pi
```bash
./build-linux.sh
./build/bin/streampi-deck
```

### Windows
```bash
wails build
.\build\bin\streampi-deck.exe
```

## Build All Platforms (Cross-Compilation)

From macOS or Linux, you can build for all platforms:

```bash
./build-all-platforms.sh
```

This creates:
- `build/bin/Stream-Pi Deck.app` (macOS Universal)
- `build/bin/streampi-deck.exe` (Windows)
- `build/bin/streampi-deck` (Linux amd64)
- `build/bin/streampi-deck-arm64` (Linux ARM64 - Pi 4/5)
- `build/bin/streampi-deck-arm` (Linux ARM - Pi 3)

Plus distribution packages:
- `build/streampi-deck-1.0.0-linux-amd64.tar.gz`
- `build/streampi-deck-1.0.0-linux-arm64.tar.gz`
- `build/streampi-deck-1.0.0-linux-arm.tar.gz`
- `build/streampi-deck-1.0.0-windows-amd64.zip`

## Building for Specific Platforms

### macOS (from macOS)

**Intel only:**
```bash
wails build -platform darwin/amd64
```

**Apple Silicon only:**
```bash
wails build -platform darwin/arm64
```

**Universal (both):**
```bash
wails build -platform darwin/universal
```

### Windows (from any platform)

```bash
wails build -platform windows/amd64
```

### Linux

**Desktop (amd64):**
```bash
wails build -platform linux/amd64
```

**Raspberry Pi 4/5 (64-bit):**
```bash
wails build -platform linux/arm64
```

**Raspberry Pi 3/Zero (32-bit):**
```bash
wails build -platform linux/arm
```

## Development Build (with DevTools)

```bash
wails build
./build/bin/streampi-deck
```

DevTools will open automatically (Cmd+Option+I won't work in production builds).

## Running the Built App

### macOS

**From anywhere:**
```bash
open ~/git/stream-pi-go/client-go/build/bin/Stream-Pi\ Deck.app
```

**Set server URL:**
```bash
export SERVER_URL=http://10.91.108.170:8080
open ~/git/stream-pi-go/client-go/build/bin/Stream-Pi\ Deck.app
```

Or configure it in the app: Settings → Server Configuration

### Linux/Raspberry Pi

**Run directly:**
```bash
cd ~/git/stream-pi-go/client-go
./build/bin/streampi-deck
```

**With server URL:**
```bash
export SERVER_URL=http://10.91.108.170:8080
./build/bin/streampi-deck
```

**Fullscreen on Pi:**
```bash
./build/bin/streampi-deck &
# Then click fullscreen button in app
```

### Windows

```bash
cd C:\git\stream-pi-go\client-go
.\build\bin\streampi-deck.exe
```

## Distribution Packages

### macOS - .app Bundle

The `.app` bundle contains everything needed:
```
Stream-Pi Deck.app/
├── Contents/
│   ├── MacOS/
│   │   └── streampi-deck (binary with embedded assets)
│   ├── Resources/
│   └── Info.plist
```

**Share it:**
1. Zip the .app bundle
2. Recipients just unzip and double-click
3. Config will be created automatically in `~/Library/Application Support/StreamPi/`

**Create DMG (macOS only):**
```bash
# Install create-dmg
brew install create-dmg

# Create DMG
create-dmg \
  --volname "Stream-Pi Deck" \
  --window-pos 200 120 \
  --window-size 800 400 \
  --icon-size 100 \
  --icon "Stream-Pi Deck.app" 200 190 \
  --hide-extension "Stream-Pi Deck.app" \
  --app-drop-link 600 185 \
  "Stream-Pi-Deck-1.0.0.dmg" \
  "build/bin/Stream-Pi Deck.app"
```

### Linux - tar.gz

Already created by `build-all-platforms.sh`:

```bash
# Extract
tar xzf streampi-deck-1.0.0-linux-amd64.tar.gz

# Run
./streampi-deck
```

**Create .deb package (Ubuntu/Debian/Raspberry Pi OS):**

```bash
mkdir -p streampi-deck-deb/DEBIAN
mkdir -p streampi-deck-deb/usr/local/bin
mkdir -p streampi-deck-deb/usr/share/applications

# Copy binary
cp build/bin/streampi-deck streampi-deck-deb/usr/local/bin/

# Create control file
cat > streampi-deck-deb/DEBIAN/control << EOF
Package: streampi-deck
Version: 1.0.0
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Stream-Pi <stream-pi@example.com>
Description: OBS Studio controller with Stream Deck interface
 Native desktop application for controlling OBS Studio
 with a touch-friendly Stream Deck interface.
EOF

# Create .desktop file
cat > streampi-deck-deb/usr/share/applications/streampi-deck.desktop << EOF
[Desktop Entry]
Name=Stream-Pi Deck
Exec=/usr/local/bin/streampi-deck
Icon=streampi-deck
Type=Application
Categories=AudioVideo;
EOF

# Build package
dpkg-deb --build streampi-deck-deb
mv streampi-deck-deb.deb streampi-deck-1.0.0-amd64.deb
```

**Install .deb:**
```bash
sudo dpkg -i streampi-deck-1.0.0-amd64.deb
```

### Windows - .zip

Already created by `build-all-platforms.sh`:

```bash
# Extract
unzip streampi-deck-1.0.0-windows-amd64.zip

# Run
streampi-deck.exe
```

**Create installer with NSIS (advanced):**
```bash
# Install NSIS first
wails build -nsis
```

## Raspberry Pi Specific

### Installation

```bash
# On Raspberry Pi
cd ~/git/stream-pi-go/client-go

# Build for Pi architecture
./build-linux.sh

# Or install from .deb
sudo dpkg -i streampi-deck-1.0.0-arm64.deb
```

### Auto-start on Boot

```bash
mkdir -p ~/.config/autostart

cat > ~/.config/autostart/streampi-deck.desktop << EOF
[Desktop Entry]
Type=Application
Name=Stream-Pi Deck
Exec=/usr/local/bin/streampi-deck
X-GNOME-Autostart-enabled=true
Environment="SERVER_URL=http://10.91.108.170:8080"
EOF
```

### Touchscreen Calibration

```bash
sudo apt install xinput-calibrator
xinput_calibrator
```

## Troubleshooting

### macOS: "App is damaged"

```bash
# Remove quarantine attribute
xattr -cr "Stream-Pi Deck.app"
```

### Linux: Missing dependencies

```bash
sudo apt install libgtk-3-0 libwebkit2gtk-4.0-37
```

### Can't find config file

The app creates it automatically. Check:
- macOS: `~/Library/Application Support/StreamPi/buttons.json`
- Linux: `~/.config/streampi/buttons.json`
- Windows: `%APPDATA%\StreamPi\buttons.json`

### Build fails with "platform not supported"

Some platforms can only be built from their native OS or require Docker.

**From macOS:**
- ✅ macOS (all)
- ✅ Windows
- ✅ Linux (all)

**From Linux:**
- ⚠️ macOS (requires osxcross)
- ✅ Windows
- ✅ Linux (all)

**From Windows:**
- ⚠️ macOS (not supported)
- ✅ Windows
- ✅ Linux (with WSL/Docker)

## File Sizes

| Platform | Binary Size | With Assets |
|----------|-------------|-------------|
| macOS Universal | ~15MB | ~15MB (embedded) |
| Windows | ~12MB | ~12MB (embedded) |
| Linux amd64 | ~14MB | ~14MB (embedded) |
| Linux arm64 | ~13MB | ~13MB (embedded) |

All assets (HTML/CSS/JS) are embedded in the binary - no external files needed!

## CI/CD (GitHub Actions)

Create `.github/workflows/build.yml`:

```yaml
name: Build

on: [push, pull_request]

jobs:
  build:
    strategy:
      matrix:
        os: [macos-latest, ubuntu-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest
      
      - name: Install Linux dependencies
        if: matrix.os == 'ubuntu-latest'
        run: sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev
      
      - name: Build
        run: |
          cd client-go
          wails build
      
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: streampi-deck-${{ matrix.os }}
          path: client-go/build/bin/*
```

## Summary

✅ **No external files needed** - Everything is embedded
✅ **OS-standard config paths** - Works correctly on all platforms  
✅ **Cross-compilation** - Build for all platforms from one machine
✅ **Distribution packages** - tar.gz, .zip, .deb, .dmg
✅ **No working directory issues** - Run from anywhere

The app is now properly packaged and ready for distribution!
