# Robo-Stream Plugin System

Robo-Stream supports external **controller plugins** — standalone binaries that add support for new applications (Zoom via ZoomOSC, hardware controllers, etc.). The server discovers and manages plugins automatically.

## How It Works

```
Robo-Stream Server
  ├── OBS Controller       (built-in, compiled into server)
  └── Plugin Manager
        ├── zoomosc-controller  ← spawns binary, talks HTTP/JSON
        └── (your-plugin)       ← spawns binary, talks HTTP/JSON
```

- Plugins are **standalone binaries** (Go, Python, anything)
- Server spawns each binary and communicates via a simple HTTP/JSON protocol
- Plugin prints `PLUGIN_READY {"id":"myid","port":PORT,"version":"1.0.0"}` to stdout
- Server connects to the plugin's HTTP server for actions, config, and status

## Plugin Directory

Plugins live in subdirectories of the server data directory:

| Platform       | Plugins directory                                       |
|----------------|---------------------------------------------------------|
| macOS / Linux  | `~/.robo-stream-server/plugins/<id>/<binary>`           |
| Windows        | `%USERPROFILE%\.robo-stream-server\plugins\<id>\<binary>`|

Each plugin gets its own subdirectory named after its plugin ID.

```
~/.robo-stream-server/
  plugins/
    zoomosc/
      zoomosc-controller      ← macOS/Linux binary
    myplugin/
      myplugin-controller     ← must be executable
```

## Installing a Plugin

### Manual Install (current method)

```bash
# 1. Build the plugin binary
cd plugins/zoomosc
make build

# 2. Create the plugin directory
mkdir -p ~/.robo-stream-server/plugins/zoomosc

# 3. Copy the binary
cp zoomosc-controller ~/.robo-stream-server/plugins/zoomosc/

# 4. Restart the server — plugin is discovered automatically
```

### Check Plugin Is Running

```bash
# List running plugin processes
ps aux | grep controller

# Expected output:
# robo  12345  0.0  ...  ~/.robo-stream-server/plugins/zoomosc/zoomosc-controller
```

## Installed Plugins (ZoomOSC)

### ZoomOSC Prerequisites

1. Install [ZoomOSC](https://www.liminalet.com/zoomosc) by Liminal (free tier available)
2. Open ZoomOSC → Settings:
   - **Transmission IP:** `127.0.0.1`
   - **Transmission Port:** `1234`  ← ZoomOSC sends status to this port
   - **Receiving Port:** `9090`     ← ZoomOSC listens for commands on this port
3. **Always start or join meetings through ZoomOSC** using its built-in Start/Join buttons — not through the Zoom app. ZoomOSC uses the Zoom SDK and must own the meeting session. Joining via the Zoom app directly causes commands to fail with "error 4".

### Build & Install ZoomOSC Plugin

```bash
cd plugins/zoomosc
make build
mkdir -p ~/.robo-stream-server/plugins/zoomosc
cp zoomosc-controller ~/.robo-stream-server/plugins/zoomosc/
# Restart server
```

---

## REST API Reference

### List all controllers

```bash
curl -s http://localhost:8080/api/controllers | jq .
```

Response:
```json
[
  {
    "connected": true,
    "description": "Controls OBS Studio via the OBS WebSocket v5 protocol",
    "id": "obs",
    "name": "OBS Studio",
    "version": "1.0.0"
  },
  {
    "connected": true,
    "description": "Controls Zoom via ZoomOSC (mute, camera, leave, share).",
    "id": "zoomosc",
    "name": "ZoomOSC",
    "version": "1.0.0"
  }
]
```

### Get controller status (detailed)

```bash
# OBS status
curl -s http://localhost:8080/api/controllers/obs/status | jq .

# ZoomOSC status (shows muted, video, sharing state)
curl -s http://localhost:8080/api/controllers/zoomosc/status | jq .
```

ZoomOSC status response when connected:
```json
{
  "connected": true,
  "muted": false,
  "video": true,
  "sharing": false
}
```

### Get controller config schema

```bash
curl -s http://localhost:8080/api/controllers/zoomosc/config/schema | jq .
```

### Get current controller config

```bash
curl -s http://localhost:8080/api/controllers/zoomosc/config | jq .
```

### Update controller config

```bash
# Update ZoomOSC ports
curl -s -X PUT http://localhost:8080/api/controllers/zoomosc/config \
  -H "Content-Type: application/json" \
  -d '{"zoomosc_port": 9090, "listen_port": 1234}'

# Connect OBS
curl -s -X PUT http://localhost:8080/api/controllers/obs/config \
  -H "Content-Type: application/json" \
  -d '{"url": "localhost:4455", "password": "mypassword"}'
```

### Execute a button action

The action endpoint is `POST /api/action` and requires a registered session.
First register a throwaway session, then use its ID:

```bash
# 1. Register a session (one-time, reuse the ID for multiple commands)
SESSION=$(curl -s -X POST http://localhost:8080/api/client/register \
  -H "Content-Type: application/json" \
  -d '{"name":"curl-test"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['session_id'])")

# 2. Toggle ZoomOSC mute
curl -s -X POST http://localhost:8080/api/action \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION" \
  -d '{"controller":"zoomosc","type":"toggle_audio"}'

# Mute ZoomOSC microphone
curl -s -X POST http://localhost:8080/api/action \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION" \
  -d '{"controller":"zoomosc","type":"mute_audio"}'

# Leave Zoom meeting
curl -s -X POST http://localhost:8080/api/action \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION" \
  -d '{"controller":"zoomosc","type":"leave_meeting"}'

# Toggle screen share
curl -s -X POST http://localhost:8080/api/action \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION" \
  -d '{"controller":"zoomosc","type":"toggle_share"}'

# OBS: switch scene
curl -s -X POST http://localhost:8080/api/action \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: $SESSION" \
  -d '{"controller":"obs","type":"switch_scene","params":{"scene_name":"Main Scene"}}'
```

> **Note:** The remote client app registers its own session automatically. The manual `curl` registration above is only needed for direct API testing.

### Restart a plugin

```bash
curl -s -X POST http://localhost:8080/api/controllers/zoomosc/restart
```

### Uninstall a plugin

```bash
curl -s -X DELETE http://localhost:8080/api/controllers/zoomosc
```

---

## Building a New Plugin (Go)

### 1. Create your module

```bash
mkdir plugins/myplugin && cd plugins/myplugin
go mod init github.com/yourname/robostream-myplugin-controller
```

Add the SDK dependency to `go.mod`:
```
require github.com/robomon1/robo-stream/sdk v0.0.0
replace github.com/robomon1/robo-stream/sdk => ../../sdk
```

### 2. Implement the plugin

```go
package main

import (
    "fmt"
    "github.com/robomon1/robo-stream/sdk"
)

type MyController struct{}

func (c *MyController) Info() sdk.PluginInfo {
    return sdk.PluginInfo{
        ID:          "myplugin",
        Name:        "My Plugin",
        Description: "Controls my application",
        Version:     "1.0.0",
        Author:      "Your Name",
    }
}

func (c *MyController) Initialize(cfg map[string]interface{}) error { return nil }
func (c *MyController) Status() sdk.PluginStatus {
    return sdk.PluginStatus{Connected: true}
}
func (c *MyController) ConfigSchema() []sdk.ConfigField { return nil }
func (c *MyController) GetConfig() map[string]interface{} { return nil }
func (c *MyController) UpdateConfig(cfg map[string]interface{}) error { return nil }

func (c *MyController) SupportedActions() []sdk.ActionTypeDef {
    return []sdk.ActionTypeDef{
        {Type: "do_something", Name: "Do Something", Description: "Does the thing"},
    }
}

func (c *MyController) Execute(action sdk.ExecuteRequest) error {
    switch action.Type {
    case "do_something":
        fmt.Println("Doing the thing!")
        return nil
    }
    return fmt.Errorf("unknown action: %s", action.Type)
}

func (c *MyController) DefaultButtons() []sdk.Button {
    return []sdk.Button{
        {
            Name:   "Do Something",
            Color:  "#3b82f6",
            Action: sdk.ExecuteRequest{Controller: "myplugin", Type: "do_something"},
        },
    }
}

func main() {
    sdk.RunPlugin(&MyController{})
}
```

### 3. Build and install

```bash
go build -o myplugin-controller .
mkdir -p ~/.robo-stream-server/plugins/myplugin
cp myplugin-controller ~/.robo-stream-server/plugins/myplugin/
# Restart server — your plugin appears automatically
```

### 4. Test with curl

```bash
# Check it was discovered
curl -s http://localhost:8080/api/controllers | jq '.[] | select(.id == "myplugin")'

# Trigger an action
curl -s -X POST http://localhost:8080/api/actions/execute \
  -H "Content-Type: application/json" \
  -d '{"controller": "myplugin", "type": "do_something"}'
```

## Plugin Protocol Reference

Plugins implement a simple HTTP/JSON protocol on a random port:

| Endpoint         | Method | Description                          |
|-----------------|--------|--------------------------------------|
| `/info`         | GET    | Returns `sdk.PluginInfo`            |
| `/status`       | GET    | Returns `sdk.PluginStatus`          |
| `/initialize`   | POST   | Receives config, returns 200 or error|
| `/config/schema`| GET    | Returns `[]sdk.ConfigField`         |
| `/config`       | GET    | Returns current config map           |
| `/config`       | PUT    | Updates config                       |
| `/actions`      | GET    | Returns `[]sdk.ActionTypeDef`       |
| `/execute`      | POST   | Receives `sdk.ExecuteRequest`       |
| `/buttons`      | GET    | Returns `[]sdk.Button` (defaults)   |

On startup, the plugin prints exactly one line to stdout:
```
PLUGIN_READY {"id":"myplugin","port":12345,"version":"1.0.0"}
```

This tells the server which port to connect to. After that, stdout can be used for logging.

## ZoomOSC Action Reference

| Action ID       | OSC Path Sent          | Description                 |
|-----------------|------------------------|-----------------------------|
| `toggle_audio`  | `/zoom/me/toggleMute`  | Toggle mic mute/unmute      |
| `mute_audio`    | `/zoom/me/mute`        | Mute microphone             |
| `unmute_audio`  | `/zoom/me/unMute`      | Unmute microphone           |
| `toggle_video`  | `/zoom/me/toggleVideo` | Toggle camera               |
| `start_video`   | `/zoom/me/videoOn`     | Turn camera on              |
| `stop_video`    | `/zoom/me/videoOff`    | Turn camera off             |
| `leave_meeting` | `/zoom/leaveMeeting`   | Leave current meeting       |
| `toggle_share`  | `/zoom/me/toggleShare` | Toggle screen share         |
| `start_share`   | `/zoom/me/startShare`  | Start screen share          |
| `stop_share`    | `/zoom/me/stopShare`   | Stop screen share           |
