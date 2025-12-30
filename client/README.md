# Robo-Stream Client (Go)

Web-based client for controlling OBS Studio via the Robo-Stream server.

## Features

- ✅ **Button Grid Interface** - 4x3 customizable button grid (12 buttons)
- ✅ **Real-time Status** - See streaming, recording, and current scene status
- ✅ **Easy Configuration** - Click to configure button actions
- ✅ **Multiple Actions** - Scene switching, streaming, recording, audio control
- ✅ **Live Updates** - WebSocket connection for instant status updates
- ✅ **Modern UI** - Clean, responsive design

## Prerequisites

1. **Robo-Stream Server** must be running (server-go)
2. **OBS Studio** must be running with WebSocket enabled
3. **Go 1.21+** for building

## Quick Start

### 1. Build the Client

```bash
cd client
go mod download
go build -o bin/streampi-client ./cmd/client
```

### 2. Run the Client

```bash
# Make sure server-go is running first!
./bin/streampi-client
```

### 3. Open in Browser

```
http://localhost:3000
```

## Configuration

### Command-Line Options

```bash
./streampi-client \
  --server-url http://localhost:8080 \
  --port 3000 \
  --config configs/buttons.json \
  --log-level info
```

### Environment Variables

```bash
export SERVER_URL=http://localhost:8080
export CLIENT_PORT=3000
./streampi-client
```

### Configuration File

Button configuration is stored in `configs/buttons.json`:

```json
{
  "grid": {
    "rows": 3,
    "cols": 4
  },
  "buttons": [
    {
      "id": "btn-0-0",
      "row": 0,
      "col": 0,
      "text": "Scene 1",
      "color": "#3498db",
      "action": {
        "type": "switch_scene",
        "params": {
          "scene_name": "Main Scene"
        }
      }
    }
  ]
}
```

## Available Actions

### Scene Actions
- **switch_scene** - Switch to a specific scene
  - Parameters: `scene_name`

### Streaming Actions
- **toggle_stream** - Toggle streaming on/off
- **start_stream** - Start streaming
- **stop_stream** - Stop streaming

### Recording Actions
- **toggle_record** - Toggle recording on/off
- **start_record** - Start recording
- **stop_record** - Stop recording

### Audio Actions
- **toggle_input_mute** - Toggle audio input mute
- **mute_input** - Mute an audio input
- **unmute_input** - Unmute an audio input
  - Parameters: `input_name`

### Source Actions
- **set_source_visibility** - Show/hide a source
  - Parameters: `source_name`, `visible` (true/false)

## Using the Client

### Normal Mode

1. Click any configured button to trigger its action
2. Watch the status bar update in real-time
3. Buttons will execute their configured OBS actions

### Edit Mode

1. Click **"Edit Mode"** button
2. Click any button to configure it
3. Fill in the configuration form:
   - **Button Text**: Display text
   - **Button Color**: Background color
   - **Action**: What the button does
   - **Parameters**: Action-specific settings
4. Click **"Save"** to apply changes
5. Click **"Exit Edit Mode"** when done

### Adding New Buttons

1. Enter **Edit Mode**
2. Click **"Add Button"**
3. Configure the new button
4. Save

### Deleting Buttons

1. Enter **Edit Mode**
2. Click the button you want to delete
3. Click **"Delete"** in the configuration dialog

## API Endpoints

The client exposes these endpoints:

- `GET /` - Web interface
- `GET /api/buttons` - Get button configuration
- `POST /api/buttons` - Update all buttons
- `PUT /api/buttons/{id}` - Update single button
- `DELETE /api/buttons/{id}` - Delete button
- `POST /api/buttons/{id}/press` - Press button
- `GET /api/scenes` - Get available scenes
- `GET /api/inputs` - Get available audio inputs
- `GET /api/status` - Get OBS status
- `GET /ws` - WebSocket connection

## Architecture

```
┌─────────────────┐
│   Browser UI    │  (HTML/CSS/JS)
│  localhost:3000 │
└────────┬────────┘
         │ HTTP/WebSocket
         ▼
┌─────────────────┐
│   client     │  (Go Web Server)
│   Port 3000     │
└────────┬────────┘
         │ HTTP
         ▼
┌─────────────────┐
│   server-go     │  (OBS Connection)
│   Port 8080     │
└────────┬────────┘
         │ WebSocket
         ▼
┌─────────────────┐
│   OBS Studio    │
│   Port 4455     │
└─────────────────┘
```

## Development

### Project Structure

```
client/
├── cmd/
│   └── client/
│       └── main.go              # Entry point
├── internal/
│   ├── server/
│   │   └── server.go           # Web server
│   ├── api/
│   │   ├── handlers.go         # HTTP handlers
│   │   └── websocket.go        # WebSocket handling
│   ├── config/
│   │   └── buttons.go          # Button configuration
│   └── client/
│       └── obs_client.go       # Server communication
├── web/
│   ├── static/
│   │   ├── css/style.css       # Styling
│   │   └── js/app.js           # Frontend logic
│   └── templates/
│       └── index.html          # Main page
├── configs/
│   └── default.json            # Default configuration
└── go.mod
```

### Building

```bash
# Development build
go build -o bin/streampi-client ./cmd/client

# Production build (with optimizations)
go build -ldflags="-s -w" -o bin/streampi-client ./cmd/client

# Cross-platform builds
GOOS=darwin GOARCH=amd64 go build -o bin/streampi-client-mac-intel ./cmd/client
GOOS=darwin GOARCH=arm64 go build -o bin/streampi-client-mac-arm ./cmd/client
GOOS=linux GOARCH=amd64 go build -o bin/streampi-client-linux ./cmd/client
GOOS=windows GOARCH=amd64 go build -o bin/streampi-client-windows.exe ./cmd/client
```

## Troubleshooting

### Client won't start

**Error:** `Failed to start web server: address already in use`

**Solution:** Another process is using port 3000. Use a different port:
```bash
./streampi-client --port 3001
```

### Can't connect to server

**Error:** `Failed to send request: connection refused`

**Solution:** Make sure server-go is running:
```bash
cd ../server-go
./bin/streampi-server
```

### Buttons not working

**Check:**
1. Server-go is connected to OBS
2. OBS WebSocket is enabled
3. Scene/input names in button config match OBS exactly

### Configuration not saving

**Solution:** Check file permissions:
```bash
chmod 755 configs/
chmod 644 configs/buttons.json
```

## Tips

1. **Customize Colors**: Use the color picker to match your branding
2. **Test Actions**: Use Edit Mode to test buttons before finalizing
3. **Backup Config**: Keep a backup of `configs/buttons.json`
4. **Multiple Clients**: Run multiple clients on different ports for different setups

## Next Steps

- [ ] Add support for button icons
- [ ] Multi-page button grids
- [ ] Custom grid sizes
- [ ] Themes
- [ ] Hotkey support
- [ ] Mobile app version

## Contributing

This is part of the Robo-Stream Go project. See the main repository for contribution guidelines.

## License

MIT License - See LICENSE file for details
