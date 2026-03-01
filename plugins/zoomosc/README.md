# Robo-Stream ZoomOSC Controller

A [Robo-Stream](https://github.com/robomon1/robo-stream) controller plugin that lets you control the **Zoom** desktop client from your Robo-Stream remote using [ZoomOSC](https://www.liminalet.com/zoomosc) — a free/paid bridge app by Liminal.

## How It Works

ZoomOSC adds a bi-directional [OSC](https://en.wikipedia.org/wiki/Open_Sound_Control) (Open Sound Control) interface to Zoom over UDP. This plugin:

1. **Sends** OSC commands to ZoomOSC (default port **9090**)
2. **Receives** OSC feedback from ZoomOSC (default port **1234**)

```
Robo-Stream buttons
       │
       ▼
zoomosc-controller ──(OSC UDP :9090)──▶ ZoomOSC ──▶ Zoom
                   ◀─(OSC UDP :1234)── ZoomOSC
```

## Supported Actions

| Action ID       | Description                   |
|-----------------|-------------------------------|
| `toggle_audio`  | Toggle microphone mute        |
| `mute_audio`    | Mute microphone               |
| `unmute_audio`  | Unmute microphone             |
| `toggle_video`  | Toggle camera on/off          |
| `start_video`   | Turn camera on                |
| `stop_video`    | Turn camera off               |
| `leave_meeting` | Leave the current meeting     |
| `toggle_share`  | Toggle screen sharing         |
| `start_share`   | Start screen sharing          |
| `stop_share`    | Stop screen sharing           |

## Requirements

- [ZoomOSC](https://www.liminalet.com/zoomosc) (free tier works for self-control actions)
- Zoom desktop client
- Robo-Stream server 1.x or later

## Setup

### 1. Install ZoomOSC

Download ZoomOSC from [liminalet.com/zoomosc](https://www.liminalet.com/zoomosc) and install it on the same computer as Zoom.

### 2. Configure ZoomOSC

Open ZoomOSC → Settings and set:

| Setting              | Value       |
|----------------------|-------------|
| **Transmission IP**  | `127.0.0.1` |
| **Transmission Port**| `1234`      |
| **Receiving Port**   | `9090`      |

> These defaults match the plugin's defaults. If you change them, update the plugin config in the Robo-Stream server settings.

### 3. Install the Plugin

Download the binary for your platform from [Releases](../../releases) and place it in the Robo-Stream plugins directory:

| Platform       | Directory                                              |
|----------------|--------------------------------------------------------|
| macOS / Linux  | `~/.robo-stream-server/plugins/zoomosc/`               |
| Windows        | `%APPDATA%\robo-stream-server\plugins\zoomosc\`        |

The binary must be named `zoomosc-controller` (or `zoomosc-controller.exe` on Windows).

### 4. Start/Restart Robo-Stream

The ZoomOSC controller will be discovered automatically on the next server start.

## Build from Source

```bash
cd plugins/zoomosc
make build          # current platform
make build-all      # all platforms (darwin arm64/amd64, windows, linux)
```

## Configuration Options

| Key             | Default | Description                                        |
|-----------------|---------|----------------------------------------------------|
| `zoomosc_port`  | `9090`  | ZoomOSC receiving port (we send commands here)     |
| `listen_port`   | `1234`  | Our feedback port (ZoomOSC sends state updates here)|

## Troubleshooting

**Status shows "not connected"**
- Make sure ZoomOSC is running (not just Zoom)
- Verify ZoomOSC Transmission IP/Port match this plugin's `listen_port`
- Check that no firewall is blocking UDP on ports 9090 and 1234

**Commands don't work in meetings**
- ZoomOSC's free tier supports self-control (mute/video/share). These actions require you to be in a meeting.
- Some actions (spotlight, pin others) require ZoomOSC Pro

---

## Building Your Own Plugin

This plugin is also an example of how to build a Robo-Stream controller plugin. The [SDK](../../sdk/) provides everything you need:

```go
import "github.com/robomon1/robo-stream/sdk"

type MyController struct{}

func (c *MyController) Info() sdk.PluginInfo { ... }
func (c *MyController) Execute(r sdk.ExecuteRequest) error { ... }
// ... implement all sdk.PluginImplementation methods

func main() {
    sdk.RunPlugin(&MyController{})
}
```

See the [SDK README](../../sdk/) and this plugin's source code for a complete example.
