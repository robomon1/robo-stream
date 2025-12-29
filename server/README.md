# Robo-Stream Server

Desktop application for managing OBS Stream Deck configurations across multiple clients.

## Features

- **Configuration Management**: Create and manage multiple button layouts for different roles
- **Button Library**: Reusable button components that can be shared across configurations
- **Client Tracking**: Monitor connected clients and assign configurations
- **OBS Integration**: Control OBS Studio via WebSocket
- **Multi-Platform**: Runs on macOS, Windows, and Linux

## Architecture

```
Server (Desktop App)
├── Button Library (reusable buttons)
├── Configurations (button layouts)
├── Session Manager (client tracking)
└── OBS Manager (WebSocket control)
    ↓
HTTP API (:8080)
    ↓
Multiple Clients (different configs)
    ↓
OBS Studio
```

## Installation

### Prerequisites

- Go 1.21 or later
- Node.js 18 or later
- Wails v2.8.0 or later

### Install Wails

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Setup

```bash
# Navigate to server directory
cd ~/git/robomon1/robo-stream/server

# Install frontend dependencies
cd frontend
npm install
cd ..

# Download Go dependencies
go mod download
```

## Development

### Run in Dev Mode

```bash
wails dev
```

This will:
- Start the frontend dev server with hot reload
- Run the Go backend
- Open the application window
- Enable developer tools (Cmd+Option+I on macOS)

### Build for Production

```bash
# Build for current platform
wails build

# Build for specific platform
wails build -platform darwin/arm64
wails build -platform darwin/amd64
wails build -platform windows/amd64
wails build -platform linux/amd64
```

Built binaries will be in `build/bin/`

## Usage

### 1. Start the Server

Launch the Robo-Stream Server application.

### 2. Connect to OBS

1. Open OBS Studio
2. Go to **Tools** → **WebSocket Server Settings**
3. Enable WebSocket server (default port: 4455)
4. Set a password (optional)
5. In Robo-Stream Server, go to **OBS Settings**
6. Enter connection details and click **Connect**

### 3. Create Buttons

1. Go to **Button Library**
2. Click **New Button**
3. Configure:
   - Name
   - Icon
   - Color
   - Action (switch scene, mute input, etc.)
4. Save button

### 4. Create Configurations

1. Go to **Configurations**
2. Click **New Configuration**
3. Set grid size (e.g., 4x3)
4. Drag buttons from library onto grid
5. Save configuration

### 5. Set Default Configuration

New clients will automatically receive the default configuration.

1. Select a configuration
2. Click **Set Default**

### 6. Connect Clients

On client devices:

1. Open the Robo-Stream client app
2. Connect to server (http://SERVER_IP:8080)
3. Client automatically receives assigned configuration

## API Endpoints

The server provides a REST API for clients on port 8080:

### Client Registration
```
POST /api/client/register
Body: { client_id, client_name }
Response: { session_id, config_id, config }
```

### Get Configuration
```
GET /api/client/config
Headers: X-Session-ID
Response: { id, name, grid, buttons[] }
```

### Execute Action
```
POST /api/action
Headers: X-Session-ID
Body: { type, params }
Response: { success }
```

### OBS Status
```
GET /api/obs/status
Response: { connected, streaming, recording, current_scene }
```

## Data Storage

Configuration files are stored in:

- **macOS**: `~/.robo-stream-server/`
- **Windows**: `%APPDATA%/RoboStreamServer/`
- **Linux**: `~/.robo-stream-server/`

Files:
- `buttons.json` - Button library
- `configs.json` - Configurations
- `sessions.json` - Client sessions

## Example Use Cases

### Streamer + Assistant Setup

**Streamer Configuration**:
- 4x3 grid (12 buttons)
- Full control: scenes, audio, streaming, recording

**Assistant Configuration**:
- 2x2 grid (4 buttons)
- Limited: scene switching and mute only

Both configurations share the same underlying buttons from the library.

### Multi-Role Team

- **Director**: All controls
- **Camera Operator**: Scene switches only
- **Audio Tech**: Audio controls only

Each role gets a tailored configuration built from shared buttons.

## Development Structure

```
server/
├── main.go                 # Entry point
├── app.go                  # Wails app bindings
├── internal/
│   ├── models/            # Data models
│   ├── storage/           # JSON persistence
│   ├── manager/           # Business logic
│   └── api/               # HTTP API server
└── frontend/              # Svelte UI
    ├── src/
    │   ├── App.svelte
    │   └── lib/           # Components
    └── public/
```

## Troubleshooting

### OBS Connection Failed

- Verify OBS WebSocket server is enabled
- Check port number (default: 4455)
- Verify password if set
- Try `ws://localhost:4455` for local OBS

### Client Can't Connect

- Check server is running
- Verify firewall allows port 8080
- Use server's IP address, not localhost
- Check client is on same network

### Build Errors

```bash
# Clean build cache
wails build -clean

# Update dependencies
go mod tidy
cd frontend && npm install
```

## Contributing

1. Create feature branch
2. Make changes
3. Test with `wails dev`
4. Build with `wails build`
5. Submit pull request

## License

MIT
