# Update: Working Entry Point Added

## What Changed

Added `cmd/server/main.go` - the missing entry point that allows the build scripts to work!

## What It Does

The new main.go provides:
- ✅ Working executable that connects to OBS
- ✅ Command-line flags for configuration
- ✅ Test mode to verify OBS integration
- ✅ Proper signal handling (Ctrl+C)
- ✅ Uses all the OBS WebSocket 5.x code we created

## Quick Test

```bash
cd server-go

# Install dependencies
go mod download

# Test mode - connects to OBS and shows info
export OBS_PASSWORD="your_password"
go run cmd/server/main.go --test

# Regular mode - stays running
go run cmd/server/main.go
```

## Build Now Works

```bash
# Build for current platform
go build -o bin/streampi-server cmd/server/main.go

# Or use make
make build

# Or build all platforms
./build-all.sh
```

## Command-Line Options

```bash
streampi-server [options]

Options:
  --obs-host string      OBS WebSocket host (default "localhost")
  --obs-port int         OBS WebSocket port (default 4455)
  --obs-password string  OBS WebSocket password (env: OBS_PASSWORD)
  --log-level string     Log level: debug, info, warn, error (default "info")
  --test                 Run in test mode (connect and exit)
  --version              Show version information
```

## Example Usage

```bash
# Connect to OBS with password
./streampi-server --obs-password "mypassword"

# Connect to remote OBS
./streampi-server --obs-host "192.168.1.100" --obs-password "secret"

# Debug mode
./streampi-server --log-level debug

# Test OBS connection
./streampi-server --test --obs-password "mypassword"
```

## Test Mode Output

When you run with `--test`, you'll see:
```
🚀 Starting Stream-Pi Server Go
Connecting to OBS at localhost:4455
✅ Connected to OBS!
OBS Version: 30.0.0
🧪 Running OBS integration tests...
📋 Getting scene list...
Found 3 scenes:
  - Scene 1
  - Scene 2
  - Scene 3
🎬 Getting current scene...
Current scene: Scene 1
📡 Getting stream status...
Streaming: false
🔴 Getting recording status...
Recording: false (paused: false)
🎤 Getting input list...
Found 2 inputs:
  - Microphone
  - Desktop Audio
✅ All tests completed!
Test mode complete, exiting
```

## Environment Variables

```bash
export OBS_HOST="localhost"
export OBS_PORT="4455"
export OBS_PASSWORD="your_password"

./streampi-server
```

## What's Next

This is a minimal but functional server that:
1. Connects to OBS ✅
2. Can query OBS state ✅
3. Uses our OBS WebSocket 5.x integration ✅

Still needed:
- WebSocket server for clients
- Profile management
- Action execution system
- Client implementation

But now you can:
- Build and deploy the binary
- Test OBS connection
- Verify the integration works on your setup
