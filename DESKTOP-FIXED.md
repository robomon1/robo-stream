# Stream-Pi Desktop App - Fixed! ✅

All issues resolved:

## ✅ Fixed Issues

1. **Config files now in OS-standard locations:**
   - macOS: `~/Library/Application Support/StreamPi/buttons.json`
   - Linux: `~/.config/streampi/buttons.json`
   - Windows: `%APPDATA%\StreamPi\buttons.json`

2. **No working directory dependency:**
   - All assets embedded in binary
   - Run from anywhere: `open ./build/bin/Stream-Pi\ Deck.app`

3. **Cross-platform builds:**
   - Build for all platforms from macOS
   - Build for all platforms from Linux/Pi
   - Automated build scripts included

4. **Distribution packages:**
   - macOS: .app bundle (+ .dmg optional)
   - Linux: .tar.gz (+ .deb instructions)
   - Windows: .zip (+ NSIS installer optional)

## 🚀 Quick Start (macOS)

```bash
cd ~/git/stream-pi-go/client-go

# Build for macOS
./build-macos.sh

# Run from anywhere!
export SERVER_URL=http://10.91.108.170:8080
open ./build/bin/Stream-Pi\ Deck.app

# Or just double-click the app
# Then configure server URL in Settings
```

## 🐧 Quick Start (Raspberry Pi)

```bash
cd ~/git/stream-pi-go/client-go

# Build for Pi
./build-linux.sh

# Run
export SERVER_URL=http://10.91.108.170:8080
./build/bin/streampi-deck

# Click fullscreen button for 8" touchscreen!
```

## 📦 Build All Platforms

From macOS or Linux:

```bash
cd ~/git/stream-pi-go/client-go
./build-all-platforms.sh
```

Creates:
- Stream-Pi Deck.app (macOS Universal)
- streampi-deck.exe (Windows)
- streampi-deck (Linux amd64, arm64, arm)
- Plus .tar.gz and .zip packages

## 🎯 What Changed from Before

**Before:**
```bash
cd client-go
./bin/streampi-client  # Must be in this directory!
# Config in ./configs/buttons.json
```

**After:**
```bash
# Run from ANYWHERE:
open ~/wherever/Stream-Pi\ Deck.app
# Config in ~/Library/Application Support/StreamPi/buttons.json
```

## 📱 For Your 8" Raspberry Pi

Perfect setup:

1. **Build on Pi:**
   ```bash
   cd ~/git/stream-pi-go/client-go
   ./build-linux.sh
   ```

2. **Auto-start:**
   ```bash
   mkdir -p ~/.config/autostart
   cat > ~/.config/autostart/streampi-deck.desktop << EOF
   [Desktop Entry]
   Type=Application
   Name=Stream-Pi Deck
   Exec=$HOME/git/stream-pi-go/client-go/build/bin/streampi-deck
   Environment="SERVER_URL=http://10.91.108.170:8080"
   X-GNOME-Autostart-enabled=true
   EOF
   ```

3. **Reboot** - app starts automatically!
4. **Click fullscreen** - full 8" display!

## 📚 Documentation

- **BUILD.md** - Complete build & packaging guide
- **DESKTOP-APP.md** - Desktop app features & usage
- **build-macos.sh** - Quick macOS build
- **build-linux.sh** - Quick Linux/Pi build
- **build-all-platforms.sh** - Build everything

## 🔧 Testing on macOS Now

```bash
cd ~/git/stream-pi-go/client-go

# Pull latest changes
git pull origin main

# Build
wails build

# Run
./build/bin/streampi-deck
```

Config will be created at:
`~/Library/Application Support/StreamPi/buttons.json`

You can now move the app anywhere and it will still work!

## 🎉 Benefits

✅ No browser chrome (save 3 lines on touchscreen!)
✅ OS-standard config locations
✅ Run from anywhere (no working directory issues)
✅ All assets embedded (no external files)
✅ Cross-compile for all platforms
✅ Distribution packages ready
✅ Auto-start capability
✅ Fullscreen mode
✅ Touch-optimized

Ready to build and deploy! 🚀
