# Robo-Stream

> Control your OBS Studio streams from anywhere

Robo-Stream is a client-server application that lets you remotely control OBS Studio from any device on your network. Perfect for streamers who want wireless control of their streaming setup.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- 🎮 **Remote OBS Control** - Start, stop, and manage streams from any device
- 🎨 **Custom Button Layouts** - Configure controls to match your workflow
- 📱 **Cross-Platform** - Available on macOS, Windows, Linux, iOS, and Android
- 🔄 **Multiple Configurations** - Switch between different setups instantly
- 📊 **Real-Time Status** - See your stream status at a glance
- 🎬 **Scene Management** - Change scenes with a tap
- 👁️ **Source Control** - Show or hide sources easily

## Architecture

Robo-Stream consists of two components:

- **Server** (Wails/Go) - Runs on your streaming computer, interfaces with OBS
- **Client** (Capacitor/Electron) - Control interface for desktop and mobile

## Download

### Latest Releases

Download the latest version from [GitHub Releases](https://github.com/robomon1/robo-stream/releases)

**Server:**
- macOS (Universal)
- Windows
- Linux

**Client:**
- macOS (Intel & Apple Silicon)
- Windows
- Linux
- [Android (Google Play)](https://play.google.com/store/apps/details?id=com.robostream.robostreamclient)
- [iOS (App Store)](https://apps.apple.com/app/robo-stream-client/idXXXXXXXXX)

## Quick Start

### Server Setup

1. Download and install Robo-Stream Server for your platform
2. Launch OBS Studio
3. Start Robo-Stream Server
4. Note the server URL displayed (e.g., `http://192.168.1.100:8080`)

### Client Setup

1. Install Robo-Stream Client on your device
2. Launch the app
3. Enter your server URL
4. Start controlling your stream!

---

## Development Setup

### Prerequisites

Install the following tools:

#### Required for All Development

- **Git** - Version control
  ```bash
  # macOS (Homebrew)
  brew install git
  ```

- **Node.js 20+** - JavaScript runtime
  ```bash
  # macOS (Homebrew)
  brew install node
  
  # Verify installation
  node --version  # Should be 20.x or higher
  npm --version
  ```

#### Server Development (Go/Wails)

- **Go 1.21+** - Programming language
  ```bash
  # macOS (Homebrew)
  brew install go
  
  # Verify installation
  go version  # Should be 1.21 or higher
  ```

- **Wails CLI** - Desktop app framework
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  
  # Add Go bin to PATH (add to ~/.zshrc or ~/.bash_profile)
  export PATH=$PATH:~/go/bin
  
  # Verify installation
  wails version
  ```

#### Client Development (Electron/Capacitor)

- **Android Studio** - For Android builds
  - Download from: https://developer.android.com/studio
  - Install Android SDK (API 35)
  - Install command-line tools

- **Xcode** - For iOS/macOS builds (macOS only)
  ```bash
  # Install from Mac App Store
  # Then install command-line tools:
  xcode-select --install
  ```

- **Docker** - For cross-platform builds
  ```bash
  # macOS (Homebrew)
  brew install --cask docker
  
  # Start Docker Desktop
  open -a Docker
  ```

### Clone Repository

```bash
git clone https://github.com/robomon1/robo-stream.git
cd robo-stream
```

### Install Dependencies

#### Server Dependencies

```bash
cd server
go mod download
cd ..
```

#### Client Dependencies

```bash
cd client-web
npm install
cd ..
```

### Development Workflow

#### Run Server (Development Mode)

```bash
# From project root
make dev-server

# Or manually:
cd server
wails dev
```

Server will launch with live reload at `http://localhost:34115`

#### Run Client (Development Mode)

**Web Browser:**
```bash
# From project root
make dev-client

# Or manually:
cd client-web
npm run dev
```

Client will open at `http://localhost:5173`

**Mobile Simulators:**
```bash
# iOS Simulator
cd client-web
npx cap open ios
# Click Run in Xcode

# Android Emulator
cd client-web
npx cap open android
# Click Run in Android Studio
```

### Code Structure

```
robo-stream/
├── server/              # Wails/Go server application
│   ├── app.go          # Main application logic
│   ├── frontend/       # Server UI (built from client)
│   └── wails.json      # Wails configuration
│
├── client-web/         # Capacitor/Electron client
│   ├── src/           # Source code
│   │   ├── js/        # JavaScript logic
│   │   ├── css/       # Stylesheets
│   │   └── index.html # Main HTML
│   ├── android/       # Android project
│   ├── ios/           # iOS project
│   └── package.json   # Node dependencies
│
├── Makefile           # Build automation
└── README.md          # This file
```

---

## Building for Release

### Prerequisites for Release Builds

#### Code Signing Setup

**Android:**

1. **Create Release Keystore:**
   ```bash
   cd client-web/android/app
   keytool -genkey -v -keystore robostream-release.keystore \
     -alias robostream \
     -keyalg RSA \
     -keysize 2048 \
     -validity 10000
   ```

2. **Save Passwords Securely** - You'll need these for every release

3. **Create keystore.properties:**
   ```bash
   cd client-web/android
   cat > keystore.properties << EOF
   storeFile=app/robostream-release.keystore
   storePassword=YOUR_KEYSTORE_PASSWORD
   keyAlias=robostream
   keyPassword=YOUR_KEY_PASSWORD
   EOF
   ```

4. **Backup Keystore:**
   ```bash
   # CRITICAL: Without this keystore, you cannot update your app!
   cp robostream-release.keystore ~/Backups/
   # Also save to cloud storage
   ```

**iOS:**

1. **Apple Developer Account** - $99/year membership required
   - Join at: https://developer.apple.com/programs/

2. **Create App ID:**
   - Go to: https://developer.apple.com/account/resources/identifiers/list
   - Create App ID: `com.robostream.robostreamclient`

3. **Code Signing:**
   - Use automatic signing in Xcode (recommended)
   - Or manually create Distribution certificate and provisioning profile

**macOS (Electron):**

- Optional: Apple Developer ID for notarization
- Unsigned builds work but show Gatekeeper warning

### Version Management

Robo-Stream uses separate versioning for server and client components.

#### View Current Versions

```bash
make version
```

Output:
```
Client package.json:
  "version": "1.0.0",

Android build.gradle:
  versionName = "1.0.0"
  versionCode = 1

iOS (project.pbxproj):
  MARKETING_VERSION = 1.0.0;
  CURRENT_PROJECT_VERSION = 1;

Server (wails.json):
  "productVersion": "1.0.0",
```

#### Update Client Version

```bash
make set-client-version VERSION=1.0.1 BUILD=2
```

This updates:
- package.json
- Android versionName and versionCode
- iOS MARKETING_VERSION and CURRENT_PROJECT_VERSION

**Note:** Always increment BUILD number for each store upload

#### Update Server Version

```bash
make set-server-version VERSION=1.1.0
```

This updates:
- server/wails.json productVersion

#### Update Both (Same Version)

```bash
make set-version VERSION=1.0.0 BUILD=1
```

### Build Commands

#### Build Everything

```bash
# Build all platforms
make all

# Build for specific version
make all VERSION=1.0.0
```

#### Build Server Only

```bash
# All platforms
make server

# Individual platforms
make server-mac       # macOS universal binary
make server-windows   # Windows
make server-linux     # Linux
```

#### Build Client Only

```bash
# All platforms
make client

# Individual platforms
make client-mac       # macOS (Intel + Apple Silicon)
make client-windows   # Windows
make client-linux     # Linux
```

#### Build Mobile Apps

```bash
# Android APK (for GitHub releases)
make android

# Android App Bundle (for Google Play)
make android-bundle

# iOS (opens Xcode for manual archive)
make ios
```

### Build Outputs

After building, artifacts are located:

**Server:**
```
server/build/bin/
├── robo-stream-server.app      # macOS
├── robo-stream-server.exe      # Windows
└── robo-stream-server          # Linux
```

**Client (Desktop):**
```
client-web/electron-dist/
├── mac-arm64/Robo-Stream-Client.app
├── mac/Robo-Stream-Client.app
├── Robo-Stream-Client-*.exe
└── Robo-Stream-Client-*.AppImage
```

**Client (Mobile):**
```
client-web/android/app/build/outputs/
├── apk/release/app-release.apk
└── bundle/release/app-release.aab

client-web/ios/build/
└── App.ipa
```

---

## Release Process

For complete instructions on creating releases, building artifacts, and publishing to app stores, see:

**📦 [RELEASE.md](RELEASE.md)** - Complete release documentation including:
- Version management
- Building for all platforms
- Creating GitHub releases
- Publishing to Google Play Store
- Publishing to Apple App Store  
- Screenshots guide
- Troubleshooting

**Quick commands:**
```bash
# Update versions
make set-client-version VERSION=1.0.1 BUILD=2
make set-server-version VERSION=1.1.0

# Build everything
make release VERSION=1.0.0

# View current versions
make version
```

---

## Testing

### Manual Testing Checklist

Before each release, test:

**Server:**
- [ ] Launches without errors
- [ ] Connects to OBS Studio
- [ ] WebSocket server starts
- [ ] API endpoints respond

**Client Desktop:**
- [ ] Launches on macOS (Intel + Apple Silicon)
- [ ] Launches on Windows
- [ ] Launches on Linux
- [ ] Connects to server
- [ ] Button actions work
- [ ] Configuration switching works

**Client Mobile:**
- [ ] Android app installs
- [ ] iOS app installs
- [ ] Connects to server on network
- [ ] Landscape orientation works
- [ ] Buttons respond correctly

### Automated Testing

```bash
# Server tests
make test-server

# Client tests
make test-client
```

---

## Troubleshooting

### Build Issues

**"wails: command not found"**
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
export PATH=$PATH:~/go/bin
```

**"npm: command not found"**
```bash
brew install node
```

**Android build fails with signing error**
- Check `keystore.properties` exists in `client-web/android/`
- Verify keystore file path is correct
- Ensure passwords are correct

**Electron build hangs on signing**
- Add `CSC_IDENTITY_AUTO_DISCOVERY=false` environment variable
- Or set `"identity": null` in package.json build config

### Runtime Issues

**Server won't connect to OBS**
- Ensure OBS Studio is running
- Check OBS WebSocket plugin is enabled
- Verify WebSocket password matches

**Client can't find server**
- Check both devices on same network
- Verify server URL (IP:port)
- Check firewall settings
- For Android: Ensure network_security_config.xml allows HTTP

**Mobile app shows black screen**
- Check for JavaScript errors in logs
- Verify orientation settings
- Rebuild with `npx cap sync`

---

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow existing code style
- Test on multiple platforms before submitting
- Update documentation for new features
- Keep commits atomic and well-described

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Support

- **Issues:** https://github.com/robomon1/robo-stream/issues
- **Discussions:** https://github.com/robomon1/robo-stream/discussions

---

## Acknowledgments

- Built with [Wails](https://wails.io/) - Desktop app framework
- Built with [Capacitor](https://capacitorjs.com/) - Mobile app framework
- Uses [OBS WebSocket](https://github.com/obsproject/obs-websocket) protocol

---

## Roadmap

See [GitHub Issues](https://github.com/robomon1/robo-stream/issues) for planned features and known issues.

**Planned Features:**
- [ ] Cloud sync for configurations
- [ ] Multi-server support
- [ ] Custom themes
- [ ] Plugin system
- [ ] Twitch/YouTube integration

---

## FAQ

**Q: Do I need OBS Studio?**  
A: Yes, Robo-Stream controls OBS Studio via WebSocket protocol.

**Q: Can I use this over the internet?**  
A: Currently designed for local network use. VPN or port forwarding required for remote access.

**Q: Is it free?**  
A: Yes, Robo-Stream is free and open source under MIT license.

**Q: What OBS version is required?**  
A: OBS Studio 28.0+ with WebSocket plugin enabled.

**Q: Can I customize the button layouts?**  
A: Yes, configurations are stored as JSON and fully customizable.

---

