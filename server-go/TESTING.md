# OBS WebSocket Testing Guide

Complete guide for testing OBS WebSocket 5.x integration with Stream-Pi Server.

## Table of Contents
1. [OBS Studio Setup](#obs-studio-setup)
2. [Server Configuration](#server-configuration)
3. [Testing with curl](#testing-with-curl)
4. [Testing with Go Code](#testing-with-go-code)
5. [Common Test Scenarios](#common-test-scenarios)
6. [Troubleshooting](#troubleshooting)

## OBS Studio Setup

### 1. Install OBS Studio

**macOS:**
```bash
brew install --cask obs
```

**Linux:**
```bash
# Ubuntu/Debian
sudo apt install obs-studio

# Fedora
sudo dnf install obs-studio
```

**Windows:**
Download from https://obsproject.com/download

### 2. Verify OBS Version
OBS WebSocket 5.x requires OBS Studio 28.0 or later.

```bash
# macOS
/Applications/OBS.app/Contents/MacOS/obs --version

# Linux
obs --version
```

### 3. Enable WebSocket Server

1. Launch OBS Studio
2. Go to **Tools** → **WebSocket Server Settings**
3. Check **Enable WebSocket server**
4. Set **Server Port**: `4455` (default)
5. Check **Enable Authentication** (recommended)
6. Click **Show Connect Info** to see/set password
7. Click **Apply** and **OK**

## Server Configuration

### Configuration Methods

Stream-Pi Server supports two configuration methods:

**1. Environment Variables (Recommended):**
```bash
export OBS_HOST=localhost        # or IP address
export OBS_PORT=4455             # default OBS port
export OBS_PASSWORD=your_password
export LOG_LEVEL=info            # optional: debug, info, warn, error
```

**2. Command-Line Flags:**
```bash
./streampi-server \
  --obs-host localhost \
  --obs-port 4455 \
  --obs-password your_password \
  --log-level info
```

**3. Combination (flags override environment variables):**
```bash
export OBS_PASSWORD=mypassword
./streampi-server --obs-host 192.168.1.100 --test
```

### Default Values

| Setting | Default | Environment Variable |
|---------|---------|---------------------|
| OBS Host | `localhost` | `OBS_HOST` |
| OBS Port | `4455` | `OBS_PORT` |
| OBS Password | *(none)* | `OBS_PASSWORD` |
| Log Level | `info` | `LOG_LEVEL` |

### Quick Start

```bash
# Set password
export OBS_PASSWORD="your_password"

# Run in test mode
./streampi-server --test

# Run normally
./streampi-server
```

### Network Access

To access OBS from other machines on your network:

1. In OBS: Set server to listen on all interfaces (0.0.0.0)
2. Find your IP:
   ```bash
   # macOS
   ifconfig | grep "inet " | grep -v 127.0.0.1
   
   # Linux
   ip addr show | grep "inet " | grep -v 127.0.0.1
   ```
3. Connect from other machines:
   ```bash
   ./streampi-server --obs-host 192.168.1.100
   ```

## Testing with curl

### Important Notes About curl Testing

**OBS WebSocket 5.x uses a binary WebSocket protocol**, which makes direct curl testing challenging. However, we can:
1. Test HTTP endpoint availability
2. Use WebSocket client tools
3. Use our Go test code

### 1. Test WebSocket Server is Running

```bash
# Test if port is open
nc -zv localhost 4455

# Or with telnet
telnet localhost 4455

# Or check with lsof (macOS/Linux)
lsof -i :4455

# Should show OBS listening on port 4455
```

### 2. Using websocat (Better than curl for WebSockets)

Install websocat:
```bash
# macOS
brew install websocat

# Linux
cargo install websocat
# or download from https://github.com/vi/websocat/releases
```

Basic connection test:
```bash
# Connect and see raw messages
websocat ws://localhost:4455
```

You should see a `Hello` message from OBS with authentication challenge.

### 3. Using wscat (Node.js based)

```bash
# Install
npm install -g wscat

# Connect
wscat -c ws://localhost:4455

# You'll see the Hello message
```

## Testing with Go Code

### Quick Test (Built-in)

The server has a built-in test mode:

```bash
export OBS_PASSWORD="your_password"
./streampi-server --test
```

**Expected Output:**
```
🚀 Starting Stream-Pi Server Go
Version: 1.0.0
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

### Manual Test Scripts

Create test files for specific operations:

**test_connection.go:**
```go
package main

import (
	"fmt"
	"log"

	"github.com/andreykaipov/goobs"
)

func main() {
	client, err := goobs.New(
		"localhost:4455",
		goobs.WithPassword("your_password"),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	version, err := client.General.GetVersion()
	if err != nil {
		log.Fatalf("Failed to get version: %v", err)
	}

	fmt.Printf("✅ Connected to OBS!\n")
	fmt.Printf("OBS Version: %s\n", version.ObsVersion)
	fmt.Printf("WebSocket Version: %s\n", version.ObsWebSocketVersion)
}
```

Run:
```bash
go run test_connection.go
```

## Common Test Scenarios

### Test 1: Basic Connection
```bash
export OBS_PASSWORD="your_password"
./streampi-server --test
```

**What it tests:**
- WebSocket connection
- Authentication
- Version retrieval
- Scene list
- Current scene
- Stream status
- Recording status
- Input list

### Test 2: Custom OBS Instance
```bash
./streampi-server \
  --obs-host 192.168.1.100 \
  --obs-port 4455 \
  --obs-password mypassword \
  --test
```

### Test 3: Verbose Logging
```bash
export OBS_PASSWORD="your_password"
./streampi-server --log-level debug --test
```

### Test 4: Run Server Normally
```bash
export OBS_PASSWORD="your_password"
./streampi-server
```

Press Ctrl+C to stop.

## Troubleshooting

### Connection Refused
```
Error: dial tcp 127.0.0.1:4455: connect: connection refused
```

**Solutions:**
1. Check OBS is running: `ps aux | grep obs`
2. Verify WebSocket server enabled in OBS
3. Check port: `lsof -i :4455`
4. Try different port in OBS settings

### Authentication Failed
```
Error: authentication failed
```

**Solutions:**
1. Verify password in OBS: Tools → WebSocket Server Settings → Show Connect Info
2. Check password matches:
   ```bash
   echo $OBS_PASSWORD
   ```
3. Disable authentication temporarily to test

### No Scenes Found
```
⚠️ No scenes found
```

**Solutions:**
1. Create scenes in OBS Studio
2. Verify you're connected to the right OBS instance
3. Check OBS scene collection is loaded

### Timeout
```
Error: i/o timeout
```

**Solutions:**
1. Check firewall settings
2. Verify OBS is listening on correct interface
3. Try localhost vs 127.0.0.1 vs actual IP
4. Check network connectivity

### Wrong Protocol Version
```
Error: unsupported protocol version
```

**Solutions:**
1. Update OBS to 28.0 or later
2. Update goobs library: `go get -u github.com/andreykaipov/goobs`
3. Check OBS WebSocket plugin version

## Quick Test Commands

```bash
# Check OBS is running
pgrep -f obs || echo "OBS not running"

# Check WebSocket port
lsof -i :4455 || echo "Port 4455 not listening"

# Test connection
export OBS_PASSWORD="your_password"
./streampi-server --test

# Test with verbose logging
./streampi-server --log-level debug --test
```

## Testing Checklist

- [ ] OBS Studio installed and running
- [ ] WebSocket server enabled in OBS
- [ ] Port 4455 accessible
- [ ] Password configured (if using authentication)
- [ ] At least 2 scenes created in OBS
- [ ] Audio sources configured in OBS
- [ ] Can connect via test mode
- [ ] Can list scenes
- [ ] Can get stream status
- [ ] Can get input list

## Next Steps

Once basic tests pass:
1. Test scene switching
2. Test audio control (mute/volume)
3. Test source visibility
4. Test streaming toggle (carefully!)
5. Test recording toggle (carefully!)
6. Test event handling
7. Integration test with Stream-Pi client

## Additional Resources

- [OBS WebSocket Protocol Docs](https://github.com/obsproject/obs-websocket/blob/master/docs/generated/protocol.md)
- [goobs Documentation](https://pkg.go.dev/github.com/andreykaipov/goobs)
- [OBS Studio Documentation](https://obsproject.com/wiki/)
