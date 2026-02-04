# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Robo-Stream is a client-server application for remotely controlling OBS Studio. The server runs on the streaming computer and communicates with OBS via WebSocket. The client provides a web/mobile/desktop interface for remote control.

**Key Architecture:**
- **Server:** Go/Wails desktop application that interfaces with OBS Studio via WebSocket
- **Client:** Web application built with Vite that runs on Electron (desktop), Capacitor (mobile), or web browsers
- **Communication:** REST API + WebSocket for real-time status updates between client and server

## Build Commands

### Development

```bash
# Run server in dev mode (with hot reload)
make dev-server
# OR
cd server && wails dev

# Run client in dev mode (web browser)
make dev-client
# OR
cd client-web && npm run dev

# Open mobile simulators
cd client-web && npx cap open ios      # Opens Xcode
cd client-web && npx cap open android  # Opens Android Studio
```

### Testing

```bash
# Test server
make test-server
# OR
cd server && go test ./...

# Test client
make test-client
# OR
cd client-web && npm test
```

### Building Release Artifacts

```bash
# Build everything for all platforms
make all

# Build individual components
make server              # All platforms
make server-mac          # macOS universal binary
make server-windows      # Windows
make server-linux        # Linux

make client              # All desktop platforms
make client-mac          # macOS (Intel + Apple Silicon)
make client-windows      # Windows
make client-linux        # Linux (AppImage)

make android             # Android APK (for GitHub releases)
make android-bundle      # Android App Bundle (for Play Store)
make ios                 # Opens Xcode for manual archive

# Package everything into releases/v{VERSION}/
make release VERSION=1.0.0
```

### Version Management

```bash
# View all current versions
make version

# Update client version (updates package.json, Android, iOS)
make set-client-version VERSION=1.0.1 BUILD=2

# Update server version (updates wails.json)
make set-server-version VERSION=1.1.0

# Update both to same version
make set-version VERSION=1.0.0 BUILD=1
```

Note: BUILD number must increment for each store upload (Android/iOS), independent of VERSION.

### Quick Development Builds

```bash
make quick-server-mac    # Fast macOS-only build
make quick-client-mac    # Fast macOS-only build
make quick-android       # Debug APK
```

## Architecture Details

### Server Architecture (Go/Wails)

**Entry Point:** `server/main.go` → `server/app.go`

**Key Managers:**
- `internal/manager/obs_manager.go` - OBS WebSocket communication (gorilla/websocket)
- `internal/manager/config_manager.go` - Configuration loading/persistence
- `internal/manager/button_manager.go` - Button action execution
- `internal/manager/session_manager.go` - Client session tracking

**API Layer:**
- `internal/api/server.go` - HTTP REST API (gorilla/mux)
- Exposes endpoints for configuration, button actions, OBS status

**Models:**
- `internal/models/configuration.go` - Button layouts and configurations
- `internal/models/button.go` - Button actions (start streaming, switch scene, etc.)
- `internal/models/obs_config.go` - OBS connection settings
- `internal/models/session.go` - Client session data

**Storage:**
- `internal/storage/storage.go` - JSON file persistence in user data directory
- Configurations stored as JSON files

**Wails Integration:**
- Frontend is served from `server/frontend/` (built from separate source)
- Server exposes methods to Wails runtime for desktop UI
- Also runs HTTP server on port 8080 for remote clients

### Client Architecture (Web/Capacitor/Electron)

**Entry Points:**
- Web: `client-web/src/index.html` + `client-web/src/js/app.js`
- Electron: `client-web/electron-main.cjs` → serves dist/
- Capacitor: `client-web/capacitor.config.ts` → native apps

**Key Modules:**
- `src/js/app.js` - Main application logic, UI rendering, state management
- `src/js/api.js` - HTTP client for server communication (fetch-based)
- `src/js/native.js` - Platform detection and Capacitor plugin integration

**Communication:**
1. Client registers with server (`POST /api/client/register`) to get session ID
2. Fetches configuration (`GET /api/configurations/{id}`)
3. Polls or receives OBS status updates
4. Sends button actions (`POST /api/buttons/{buttonID}/execute`)

**Platform Support:**
- **Web:** Runs directly in browsers via Vite dev server
- **Desktop:** Electron bundles dist/ folder, runs on macOS/Windows/Linux
- **Mobile:** Capacitor wraps dist/ as native iOS/Android apps

**Build Outputs:**
- Web: `client-web/dist/` (Vite build)
- Electron: `client-web/electron-dist/` (electron-builder)
- Android: `client-web/android/app/build/outputs/`
- iOS: `client-web/ios/build/` (manual Xcode archive)

### Configuration System

Configurations define button layouts and are stored as JSON files.

**Server Side:**
- Stored in: `{userData}/configurations/*.json`
- Managed by: `internal/manager/config_manager.go`
- Format: JSON with button grid definitions

**Client Side:**
- Fetched via REST API: `GET /api/configurations/{id}`
- Rendered dynamically in `app.js`
- Each button has: label, color, action type, parameters

**Button Action Types:**
- `start_streaming`, `stop_streaming`
- `start_recording`, `stop_recording`
- `switch_scene` (with scene name parameter)
- `toggle_source` (with source name parameter)
- `start_virtual_camera`, `stop_virtual_camera`
- And more (see `internal/models/button.go`)

### OBS Communication

**Protocol:** OBS WebSocket v5 (obs-websocket plugin)

**Connection Flow:**
1. Server reads OBS config from storage (host, port, password)
2. OBSManager establishes WebSocket connection
3. Authenticates with password (SHA256 challenge-response)
4. Subscribes to events (StreamStateChanged, RecordingStateChanged, etc.)
5. Maintains connection with ping/pong heartbeat

**Command Execution:**
- Client → HTTP API → ButtonManager → OBSManager → OBS WebSocket
- Response propagates back through chain

**Status Updates:**
- OBS sends events via WebSocket
- OBSManager updates internal state
- Clients poll `/api/obs/status` for current state

### Cross-Platform Builds

**Server (Wails):**
- Uses Docker for cross-compilation (Linux builds on macOS)
- macOS: Universal binary (Intel + Apple Silicon)
- Windows: AMD64 executable
- Linux: AMD64 binary

**Client (Electron):**
- electron-builder handles cross-platform packaging
- Code signing disabled by default (`CSC_IDENTITY_AUTO_DISCOVERY=false` in Makefile)
- macOS: DMG + ZIP for both architectures
- Windows: NSIS installer + portable EXE
- Linux: AppImage + DEB

**Client (Mobile):**
- Android: Gradle builds APK (release) or AAB (Play Store)
- iOS: Xcode manual archive required (opens with `make ios`)
- Signing requires keystore (Android) and Apple Developer cert (iOS)

## Important Development Notes

### Code Signing Requirements

**Android:**
- Requires keystore: `client-web/android/app/robostream-release.keystore`
- Config: `client-web/android/keystore.properties` (not in git)
- See RELEASE.md for keystore setup instructions

**iOS:**
- Requires Apple Developer Program membership ($99/year)
- Automatic signing configured in Xcode (preferred)
- App ID: `com.robostream.robostreamclient`

**Electron (macOS):**
- Signing disabled in Makefile to avoid build hangs
- Users bypass Gatekeeper with right-click → Open

### Network Configuration

**Server:**
- Listens on `0.0.0.0:8080` by default
- Auto-detects local IP address and displays in UI
- CORS enabled for all origins (local network use)

**Client:**
- Connects to user-provided server URL (`http://{IP}:8080`)
- Android requires `allowMixedContent: true` for HTTP connections (see capacitor.config.ts)
- Client ID persisted in localStorage for session tracking

### OBS Requirements

- OBS Studio 28.0+ required
- WebSocket plugin must be enabled (Tools → WebSocket Server Settings)
- Default port: 4455
- Password optional but recommended

### File Locations

**Server Data Directory:**
- macOS: `~/Library/Application Support/robo-stream-server/`
- Windows: `%APPDATA%/robo-stream-server/`
- Linux: `~/.config/robo-stream-server/`

Contains:
- `configurations/*.json` - Button layouts
- `obs_config.json` - OBS connection settings
- `sessions.json` - Active client sessions

## Common Development Tasks

### Adding a New Button Action

1. Add action type constant to `server/internal/models/button.go`
2. Implement handler in `server/internal/manager/button_manager.go`
3. Add OBS command execution in `server/internal/manager/obs_manager.go`
4. Update client UI to support new action in `client-web/src/js/app.js`

### Modifying Configuration Schema

1. Update structs in `server/internal/models/configuration.go`
2. Update JSON persistence in `server/internal/storage/storage.go`
3. Update client parsing in `client-web/src/js/api.js` and `app.js`
4. Provide migration logic for existing configurations

### Testing Mobile Apps

**iOS Simulator:**
```bash
cd client-web
npm run build
npx cap sync ios
npx cap open ios
# Click Run in Xcode
```

**Android Emulator:**
```bash
cd client-web
npm run build
npx cap sync android
npx cap open android
# Click Run in Android Studio
```

**Physical Device Testing:**
- iOS: Requires Apple Developer account, add device UDID in Xcode
- Android: Enable USB debugging, install debug APK directly

### Debugging Tips

**Server:**
- Run `wails dev` for hot reload and console output
- Logs printed to terminal
- Use Go debugger (dlv) for breakpoints

**Client:**
- Run `npm run dev` for hot reload
- Browser DevTools for web version
- Electron DevTools: Cmd+Option+I (macOS) or F12 (Windows/Linux)
- Mobile: Chrome DevTools remote debugging (chrome://inspect for Android)

**OBS Connection Issues:**
- Check OBS WebSocket settings (port, password)
- Verify OBS is running before starting server
- Check firewall settings on both machines

### Release Process

See RELEASE.md for complete instructions. Quick summary:

1. Update versions: `make set-client-version VERSION=x.x.x BUILD=n`
2. Commit version changes and tag
3. Build: `make release VERSION=x.x.x`
4. Create GitHub release with artifacts from `releases/v{VERSION}/`
5. Upload Android AAB to Play Console
6. Archive and upload iOS to App Store Connect

## Documentation Files

- **README.md** - Project overview, setup instructions, features
- **QUICKSTART.md** - User guide for installing and using Robo-Stream
- **RELEASE.md** - Complete release and publishing guide
- **BUILD.md** - Legacy build documentation
- **FEATURES.md** - Feature specifications and roadmap
- **PRIVACY.md** - Privacy policy for app stores

## Technology Stack

**Server:**
- Go 1.21+
- Wails v2 (desktop framework)
- gorilla/websocket (OBS communication)
- gorilla/mux (HTTP routing)

**Client:**
- Vanilla JavaScript (no framework)
- Vite (build tool)
- Capacitor 5 (mobile wrapper)
- Electron 40 (desktop wrapper)

**Build Tools:**
- Make (orchestration)
- npm (JavaScript package management)
- Wails CLI (Go desktop builds)
- electron-builder (desktop packaging)
- Gradle (Android builds)
- Xcode (iOS builds)
