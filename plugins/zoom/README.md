# Robo-Stream Zoom Controller

A Robo-Stream controller plugin that controls the **Zoom** desktop client.

This plugin is included in the [robo-stream](https://github.com/robomon1/robo-stream) repository
as a reference implementation showing how to build a Robo-Stream controller plugin.
External plugins (e.g. Bitfocus Companion) follow the same pattern from a separate repository.

## How it works

The Zoom desktop client exposes a local REST API on `localhost:19421` when it is running.
This is the same API used by Elgato Stream Deck, Loupedeck, and similar hardware controllers.
No Zoom developer account or SDK is required — the API is available to any local process.

## Supported actions

| Action type       | Description                          |
|-------------------|--------------------------------------|
| `toggle_audio`    | Toggle microphone mute/unmute        |
| `mute_audio`      | Mute the microphone                  |
| `unmute_audio`    | Unmute the microphone                |
| `toggle_video`    | Toggle camera on/off                 |
| `start_video`     | Turn the camera on                   |
| `stop_video`      | Turn the camera off                  |
| `leave_meeting`   | Leave the current meeting            |
| `toggle_share`    | Toggle screen sharing                |
| `start_share`     | Start screen sharing                 |
| `stop_share`      | Stop screen sharing                  |

## Installation

### Download a pre-built binary

1. Go to the [Releases](../../releases) page and download the archive for your platform.
2. Extract the `zoom-controller` binary.
3. Create the plugin directory and move the binary into it:

   **macOS / Linux**
   ```bash
   mkdir -p ~/.robo-stream-server/plugins/zoom
   mv zoom-controller ~/.robo-stream-server/plugins/zoom/
   chmod +x ~/.robo-stream-server/plugins/zoom/zoom-controller
   ```

   **Windows**
   ```
   %APPDATA%\RoboStreamServer\plugins\zoom\zoom-controller.exe
   ```

4. Restart the Robo-Stream server. The Zoom controller will appear in the server UI.

### Build from source

Requirements: Go 1.23+

```bash
# From the repository root
cd plugins/zoom
make build
```

Place the resulting `zoom-controller` binary in the plugins directory as above.

## Building a new plugin

This plugin is a template. To create your own:

1. Copy this directory to a new repository.
2. Change the module name in `go.mod`.
3. Replace the `ZoomController` implementation with your own.
4. Implement `sdk.PluginImplementation` — every method is documented in the SDK.
5. Build and release with the included `Makefile` and GitHub Actions workflow.

The Robo-Stream SDK is at [`github.com/robomon1/robo-stream/sdk`](../../sdk/).

## Development

```bash
# Run locally (the server must also be running)
go run .

# Run tests
make test

# Build release archives for all platforms
make release VERSION=1.0.0
```

The plugin prints `PLUGIN_READY {"id":"zoom","port":XXXX,"version":"1.0.0"}` to stdout
when it is ready. The Robo-Stream server reads this line and connects to the plugin's HTTP server.
