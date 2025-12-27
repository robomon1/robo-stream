# Quick Start Guide

Get up and running with Stream-Pi Server Go in 5 minutes.

## Prerequisites

- Go 1.21 or later installed
- OBS Studio 28.0+ running
- Git

## Step 1: Get the Code

```bash
cd ~/git/stream-pi
# Code should already be in server-go directory
cd server-go
```

## Step 2: Install Dependencies

```bash
go mod download
```

## Step 3: Configure OBS

1. Open OBS Studio
2. Go to **Tools** → **WebSocket Server Settings**
3. Check **Enable WebSocket server**
4. Set password (e.g., `test_password`)
5. Note the port (default: `4455`)
6. Click **Apply**

## Step 4: Test Connection

```bash
# Set your OBS password
export OBS_PASSWORD="test_password"

# Run connection test
go run test_connection.go
```

**Expected output:**
```
🔌 Connecting to OBS at localhost:4455...
✅ Connected successfully!

📊 OBS Information:
  OBS Version: 30.0.0
  WebSocket Version: 5.4.2
  ...
```

## Step 5: Run Full Test Suite

```bash
go run test_all.go
```

This will test:
- ✅ Connection
- ✅ Version check
- ✅ Scene management
- ✅ Streaming status
- ✅ Recording status
- ✅ Audio inputs
- ✅ Events
- ✅ Stats

## Step 6: Build for Your Platform

### macOS (Apple Silicon)
```bash
make darwin-arm64
# Binary at: bin/robostream-server-darwin-arm64
```

### macOS (Intel)
```bash
make darwin-amd64
# Binary at: bin/robostream-server-darwin-amd64
```

### Linux (x86_64)
```bash
make linux-amd64
# Binary at: bin/robostream-server-linux-amd64
```

### Raspberry Pi 4/5
```bash
make linux-arm64
# Binary at: bin/robostream-server-linux-arm64
```

### All Platforms
```bash
make all
# Or use the script:
./build-all.sh
```

## Step 7: Test Scene Switching

Create at least 2 scenes in OBS, then:

```bash
go run test_scenes.go
```

## Common Issues

### "connection refused"
- Check OBS is running
- Verify WebSocket server is enabled
- Check port 4455 is open

### "authentication failed"
- Check password matches OBS settings
- Try disabling authentication temporarily

### "no scenes found"
- Create some scenes in OBS first
- Verify you're connected to the right OBS instance

## Environment Variables

```bash
export OBS_HOST="localhost"      # OBS host
export OBS_PORT="4455"           # OBS WebSocket port
export OBS_PASSWORD="password"   # OBS WebSocket password
```

## Next Steps

1. ✅ Connection working? → Read [BUILD.md](BUILD.md) for deployment
2. ✅ Tests passing? → Read [TESTING.md](TESTING.md) for advanced testing
3. ✅ Ready to integrate? → Start building the server implementation
4. 📖 Learn more → Read [README.md](README.md) for full documentation

## Quick Commands

```bash
# Build current platform
make build

# Run tests
go test ./...

# Build all platforms
make all

# Clean build directory
make clean

# Show all make targets
make help
```

## Getting Help

- Check [TESTING.md](TESTING.md) for detailed testing
- Check [BUILD.md](BUILD.md) for build issues
- Check [README.md](README.md) for usage
- Review OBS WebSocket docs: https://github.com/obsproject/obs-websocket

## Success Checklist

- [ ] Go installed and working
- [ ] OBS Studio running with WebSocket enabled
- [ ] `go run test_connection.go` succeeds
- [ ] `go run test_all.go` passes all tests
- [ ] Can build for your platform
- [ ] Binary runs and connects to OBS

Once all checked, you're ready to start development! 🎉
