# Stream-Pi Desktop App - Quick Setup

## Converting from Web to Desktop App

Your Stream-Pi client is now a **native desktop application** using Wails v2!

### Why Desktop App?

You mentioned the browser takes up 3 lines on your 8" Raspberry Pi touchscreen. The desktop app:

- ✅ No browser chrome (saves screen space!)
- ✅ True fullscreen mode
- ✅ Optimized for touch
- ✅ Much lighter weight (~10MB vs ~200MB)
- ✅ Auto-start on boot (can be configured)
- ✅ Works on Mac, Windows, Linux, and Raspberry Pi

## Installation Steps

### 1. Install Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Make sure `$HOME/go/bin` is in your PATH:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### 2. Install Platform Dependencies

**On Raspberry Pi / Linux:**
```bash
sudo apt update
sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev
```

**On macOS:**
```bash
xcode-select --install
```

**On Windows:**
- Download and install [WebView2 Runtime](https://developer.microsoft.com/en-us/microsoft-edge/webview2/)

### 3. Extract and Build

```bash
cd ~/git/stream-pi-go/client-go

# Extract the desktop app files (if using tarball)
tar xzf ~/Downloads/streampi-desktop-app.tar.gz

# Or just pull from GitHub
git pull origin main

# Build the desktop app
wails build
```

### 4. Run

**Raspberry Pi / Linux:**
```bash
export SERVER_URL=http://10.91.108.170:8080
./build/bin/streampi-deck
```

**macOS:**
```bash
export SERVER_URL=http://10.91.108.170:8080
open build/bin/Stream-Pi\ Deck.app
```

**Windows:**
```bash
set SERVER_URL=http://10.91.108.170:8080
.\build\bin\streampi-deck.exe
```

## Quick Test

During development, use dev mode:

```bash
cd ~/git/stream-pi-go/client-go
export SERVER_URL=http://10.91.108.170:8080
wails dev
```

This opens the app with hot-reload for testing.

## For Your 8" Raspberry Pi Touchscreen

### Perfect Setup:

1. **Build the app:**
   ```bash
   cd ~/git/stream-pi-go/client-go
   wails build -platform linux/arm64
   ```

2. **Run fullscreen:**
   ```bash
   ./build/bin/streampi-deck
   # Click the fullscreen button in the app
   ```

3. **Auto-start on boot:**
   ```bash
   mkdir -p ~/.config/autostart
   cat > ~/.config/autostart/streampi-deck.desktop << EOF
   [Desktop Entry]
   Type=Application
   Name=Stream-Pi Deck
   Exec=/home/pi/git/stream-pi-go/client-go/build/bin/streampi-deck
   X-GNOME-Autostart-enabled=true
   EOF
   ```

4. **Configure server URL in the app:**
   - Open the app
   - Click "⚙️ Configure"
   - Scroll to "Server Configuration"
   - Enter: `http://10.91.108.170:8080`
   - Click "Update Server"

Now your Raspberry Pi will boot directly into the Stream Deck interface with NO browser chrome!

## Features You Get

### In the App:

- **Fullscreen Button** - Click to go fullscreen (saves those 3 lines!)
- **Server Configuration** - Change server URL without restarting
- **Touch Optimized** - Large 80x80px buttons
- **Two Views** - Deck view and Config view
- **All Your Buttons** - Same button config as before

### How It Works:

```
┌──────────────────────────────────┐
│   Raspberry Pi Touchscreen      │
│   (8" - Full screen, no chrome) │
│                                  │
│   Stream-Pi Deck App (Wails)    │
│   ↓ HTTP to server-go            │
└──────────────────────────────────┘
              ↓ HTTP
┌──────────────────────────────────┐
│   Computer with OBS              │
│   (10.91.108.170)                │
│                                  │
│   server-go :8080                │
│   ↓ WebSocket                    │
│   OBS Studio :4455               │
└──────────────────────────────────┘
```

## Troubleshooting

**"wails: command not found"**
```bash
export PATH=$PATH:$(go env GOPATH)/bin
# Add to ~/.bashrc to make permanent
```

**Build fails on Raspberry Pi:**
```bash
sudo apt install build-essential
sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev
```

**App starts but can't connect:**
- Check server-go is running on 10.91.108.170:8080
- Verify `curl http://10.91.108.170:8080/health` returns "OK"
- Update Server URL in app's Config view

## File Sizes

**Before (Web):**
- Chrome + page: ~200-300MB RAM
- 3 lines of screen for browser chrome

**After (Desktop):**
- Wails app: ~10-20MB RAM
- 0 lines for browser chrome (fullscreen!)
- Binary size: ~15MB

## What Didn't Change

- Button configurations (same JSON file)
- Button grid layout
- All OBS actions
- Real-time status updates
- Configuration interface

Everything works the same, just **no browser!**

## Next Steps

1. Build the app: `wails build`
2. Test it: `./build/bin/streampi-deck`
3. Click fullscreen button
4. Enjoy your full 8" display!

---

Your 8" touchscreen just got 3 lines bigger! 🎉
