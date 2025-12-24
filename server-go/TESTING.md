# OBS WebSocket Testing Guide

Complete guide for testing OBS WebSocket 5.x integration with Stream-Pi Server.

## Table of Contents
1. [OBS Studio Setup](#obs-studio-setup)
2. [WebSocket Server Configuration](#websocket-server-configuration)
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

## WebSocket Server Configuration

### Default Settings
```yaml
Host: localhost (or 127.0.0.1)
Port: 4455
Password: your_password_here
Protocol: ws:// (not wss://)
```

### Test Connection Info
```
WebSocket URL: ws://localhost:4455
Password: test_password
```

### Network Access
To access from other machines on your network:
```yaml
Host: 0.0.0.0  # Listen on all interfaces
Port: 4455
```

Then connect from other machines using:
```
ws://YOUR_MAC_IP:4455
```

Find your IP:
```bash
# macOS
ifconfig | grep "inet " | grep -v 127.0.0.1

# Linux
ip addr show | grep "inet " | grep -v 127.0.0.1
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

### 4. Python WebSocket Test Script

Save as `test_obs_ws.py`:
```python
#!/usr/bin/env python3
import websocket
import json
import hashlib
import base64
import uuid

# Configuration
WS_URL = "ws://localhost:4455"
PASSWORD = "your_password"

def generate_auth_response(password, salt, challenge):
    """Generate authentication response for OBS WebSocket 5.x"""
    secret = base64.b64encode(
        hashlib.sha256((password + salt).encode()).digest()
    ).decode()
    
    auth = base64.b64encode(
        hashlib.sha256((secret + challenge).encode()).digest()
    ).decode()
    
    return auth

def test_connection():
    ws = websocket.WebSocket()
    ws.connect(WS_URL)
    
    # Receive Hello message
    hello_msg = json.loads(ws.recv())
    print("Hello message:", json.dumps(hello_msg, indent=2))
    
    # Extract auth challenge
    auth_data = hello_msg['d']['authentication']
    salt = auth_data['salt']
    challenge = auth_data['challenge']
    
    # Generate auth response
    auth_response = generate_auth_response(PASSWORD, salt, challenge)
    
    # Send Identify message
    identify = {
        "op": 1,
        "d": {
            "rpcVersion": 1,
            "authentication": auth_response,
            "eventSubscriptions": 33  # All events
        }
    }
    ws.send(json.dumps(identify))
    
    # Receive Identified message
    identified = json.loads(ws.recv())
    print("Identified:", json.dumps(identified, indent=2))
    
    # Send a request (GetVersion)
    request = {
        "op": 6,
        "d": {
            "requestType": "GetVersion",
            "requestId": str(uuid.uuid4())
        }
    }
    ws.send(json.dumps(request))
    
    # Receive response
    response = json.loads(ws.recv())
    print("Response:", json.dumps(response, indent=2))
    
    ws.close()
    print("Connection test successful!")

if __name__ == "__main__":
    test_connection()
```

Run:
```bash
chmod +x test_obs_ws.py
./test_obs_ws.py
```

## Testing with Go Code

### 1. Simple Connection Test

Save as `test_connection.go`:
```go
package main

import (
	"fmt"
	"log"

	"github.com/andreykaipov/goobs"
)

func main() {
	// Connect to OBS
	client, err := goobs.New(
		"localhost:4455",
		goobs.WithPassword("your_password"),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Get version
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

### 2. Test All Scene Operations

Save as `test_scenes.go`:
```go
package main

import (
	"fmt"
	"log"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/scenes"
)

func main() {
	client, err := goobs.New("localhost:4455", goobs.WithPassword("your_password"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect()

	// Get scene list
	sceneList, err := client.Scenes.GetSceneList()
	if err != nil {
		log.Fatalf("Failed to get scene list: %v", err)
	}

	fmt.Printf("✅ Available Scenes:\n")
	for i, scene := range sceneList.Scenes {
		fmt.Printf("  %d. %s\n", i+1, scene.SceneName)
	}

	if len(sceneList.Scenes) == 0 {
		fmt.Println("⚠️  No scenes found. Create some scenes in OBS first.")
		return
	}

	// Get current scene
	current, err := client.Scenes.GetCurrentProgramScene()
	if err != nil {
		log.Fatalf("Failed to get current scene: %v", err)
	}
	fmt.Printf("\n✅ Current Scene: %s\n", current.CurrentProgramSceneName)

	// Test scene switching (if multiple scenes exist)
	if len(sceneList.Scenes) > 1 {
		targetScene := sceneList.Scenes[1].SceneName
		fmt.Printf("\n🔄 Switching to: %s\n", targetScene)

		req := &scenes.SetCurrentProgramSceneParams{
			SceneName: &targetScene,
		}
		_, err := client.Scenes.SetCurrentProgramScene(req)
		if err != nil {
			log.Fatalf("Failed to set scene: %v", err)
		}

		fmt.Printf("✅ Successfully switched to: %s\n", targetScene)
	}
}
```

### 3. Test Streaming Operations

Save as `test_streaming.go`:
```go
package main

import (
	"fmt"
	"log"

	"github.com/andreykaipov/goobs"
)

func main() {
	client, err := goobs.New("localhost:4455", goobs.WithPassword("your_password"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect()

	// Get stream status
	status, err := client.Stream.GetStreamStatus()
	if err != nil {
		log.Fatalf("Failed to get stream status: %v", err)
	}

	fmt.Printf("✅ Stream Status:\n")
	fmt.Printf("  Active: %v\n", status.OutputActive)
	fmt.Printf("  Reconnecting: %v\n", status.OutputReconnecting)
	if status.OutputActive {
		fmt.Printf("  Duration: %d ms\n", status.OutputDuration)
		fmt.Printf("  Bytes: %d\n", status.OutputBytes)
	}

	// Get recording status
	recStatus, err := client.Record.GetRecordStatus()
	if err != nil {
		log.Fatalf("Failed to get record status: %v", err)
	}

	fmt.Printf("\n✅ Recording Status:\n")
	fmt.Printf("  Active: %v\n", recStatus.OutputActive)
	fmt.Printf("  Paused: %v\n", recStatus.OutputPaused)
	if recStatus.OutputActive {
		fmt.Printf("  Duration: %d ms\n", recStatus.OutputDuration)
		fmt.Printf("  Bytes: %d\n", recStatus.OutputBytes)
	}

	// ⚠️  Uncomment to test toggle (will actually start/stop streaming!)
	// fmt.Println("\n🔄 Toggling stream...")
	// _, err = client.Stream.ToggleStream(nil)
	// if err != nil {
	//     log.Fatalf("Failed to toggle stream: %v", err)
	// }
	// fmt.Println("✅ Stream toggled!")
}
```

### 4. Test Audio/Sources

Save as `test_audio.go`:
```go
package main

import (
	"fmt"
	"log"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/inputs"
)

func main() {
	client, err := goobs.New("localhost:4455", goobs.WithPassword("your_password"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect()

	// Get input list
	inputList, err := client.Inputs.GetInputList(nil)
	if err != nil {
		log.Fatalf("Failed to get input list: %v", err)
	}

	fmt.Printf("✅ Available Inputs:\n")
	for i, input := range inputList.Inputs {
		fmt.Printf("  %d. %s (%s)\n", i+1, input.InputName, input.InputKind)
	}

	if len(inputList.Inputs) == 0 {
		fmt.Println("⚠️  No inputs found. Add some audio sources in OBS first.")
		return
	}

	// Test with first input
	inputName := inputList.Inputs[0].InputName
	fmt.Printf("\n🎤 Testing with input: %s\n", inputName)

	// Get mute status
	muteReq := &inputs.GetInputMuteParams{
		InputName: &inputName,
	}
	muteStatus, err := client.Inputs.GetInputMute(muteReq)
	if err != nil {
		log.Fatalf("Failed to get mute status: %v", err)
	}
	fmt.Printf("  Muted: %v\n", muteStatus.InputMuted)

	// Get volume
	volReq := &inputs.GetInputVolumeParams{
		InputName: &inputName,
	}
	volume, err := client.Inputs.GetInputVolume(volReq)
	if err != nil {
		log.Fatalf("Failed to get volume: %v", err)
	}
	fmt.Printf("  Volume: %.2f dB (mul: %.2f)\n", volume.InputVolumeDb, volume.InputVolumeMul)
}
```

## Common Test Scenarios

### Test 1: Basic Connection
```bash
go run test_connection.go
```

**Expected Output:**
```
✅ Connected to OBS!
OBS Version: 30.0.0
WebSocket Version: 5.4.2
```

### Test 2: Scene Management
```bash
go run test_scenes.go
```

**Expected Output:**
```
✅ Available Scenes:
  1. Scene 1
  2. Scene 2
  3. Scene 3

✅ Current Scene: Scene 1

🔄 Switching to: Scene 2
✅ Successfully switched to: Scene 2
```

### Test 3: Check Streaming Status
```bash
go run test_streaming.go
```

**Expected Output:**
```
✅ Stream Status:
  Active: false
  Reconnecting: false

✅ Recording Status:
  Active: false
  Paused: false
```

### Test 4: Audio Sources
```bash
go run test_audio.go
```

**Expected Output:**
```
✅ Available Inputs:
  1. Microphone (coreaudio_input_capture)
  2. Desktop Audio (coreaudio_output_capture)

🎤 Testing with input: Microphone
  Muted: false
  Volume: 0.00 dB (mul: 1.00)
```

## Comprehensive Test Suite

Create `test_all.go`:
```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/andreykaipov/goobs"
)

func main() {
	fmt.Println("🧪 OBS WebSocket Test Suite\n")

	// Test 1: Connection
	fmt.Println("1️⃣  Testing Connection...")
	client, err := goobs.New("localhost:4455", goobs.WithPassword("your_password"))
	if err != nil {
		log.Fatalf("❌ Connection failed: %v", err)
	}
	defer client.Disconnect()
	fmt.Println("✅ Connected!\n")

	// Test 2: Version
	fmt.Println("2️⃣  Testing Version...")
	version, err := client.General.GetVersion()
	if err != nil {
		log.Fatalf("❌ Version check failed: %v", err)
	}
	fmt.Printf("✅ OBS %s, WebSocket %s\n\n", version.ObsVersion, version.ObsWebSocketVersion)

	// Test 3: Scene List
	fmt.Println("3️⃣  Testing Scene List...")
	scenes, err := client.Scenes.GetSceneList()
	if err != nil {
		log.Fatalf("❌ Scene list failed: %v", err)
	}
	fmt.Printf("✅ Found %d scenes\n\n", len(scenes.Scenes))

	// Test 4: Stream Status
	fmt.Println("4️⃣  Testing Stream Status...")
	streamStatus, err := client.Stream.GetStreamStatus()
	if err != nil {
		log.Fatalf("❌ Stream status failed: %v", err)
	}
	fmt.Printf("✅ Streaming: %v\n\n", streamStatus.OutputActive)

	// Test 5: Recording Status
	fmt.Println("5️⃣  Testing Recording Status...")
	recStatus, err := client.Record.GetRecordStatus()
	if err != nil {
		log.Fatalf("❌ Record status failed: %v", err)
	}
	fmt.Printf("✅ Recording: %v\n\n", recStatus.OutputActive)

	// Test 6: Input List
	fmt.Println("6️⃣  Testing Input List...")
	inputs, err := client.Inputs.GetInputList(nil)
	if err != nil {
		log.Fatalf("❌ Input list failed: %v", err)
	}
	fmt.Printf("✅ Found %d inputs\n\n", len(inputs.Inputs))

	// Test 7: Event Handling
	fmt.Println("7️⃣  Testing Events (will wait 5 seconds)...")
	eventReceived := false
	client.AddEventHandler("CurrentProgramSceneChanged", func(event any) {
		eventReceived = true
		fmt.Println("✅ Received scene change event!")
	})
	
	fmt.Println("   Try switching scenes in OBS...")
	time.Sleep(5 * time.Second)
	if !eventReceived {
		fmt.Println("⚠️  No events received (try switching scenes manually)")
	}
	fmt.Println()

	// Test 8: Stats
	fmt.Println("8️⃣  Testing Stats...")
	stats, err := client.General.GetStats()
	if err != nil {
		log.Fatalf("❌ Stats failed: %v", err)
	}
	fmt.Printf("✅ FPS: %.2f, CPU: %.2f%%, Memory: %.2f MB\n\n", 
		stats.ActiveFps, stats.CpuUsage, stats.MemoryUsage)

	fmt.Println("✅ All tests passed!")
}
```

Run:
```bash
go run test_all.go
```

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
2. Check password in your code matches
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
echo "Testing connection..." && \
go run test_connection.go && \
echo "✅ Success!" || echo "❌ Failed!"

# Full test suite
go run test_all.go
```

## Testing Checklist

- [ ] OBS Studio installed and running
- [ ] WebSocket server enabled in OBS
- [ ] Port 4455 accessible
- [ ] Password configured (if using authentication)
- [ ] At least 2 scenes created in OBS
- [ ] Audio sources configured in OBS
- [ ] Can connect via Go test
- [ ] Can list scenes
- [ ] Can get stream status
- [ ] Can get input list
- [ ] Events are received

## Next Steps

Once basic tests pass:
1. Test scene switching
2. Test audio control (mute/volume)
3. Test source visibility
4. Test streaming toggle (carefully!)
5. Test recording toggle (carefully!)
6. Test event handling
7. Integration test with Stream-Pi server

## Additional Resources

- [OBS WebSocket Protocol Docs](https://github.com/obsproject/obs-websocket/blob/master/docs/generated/protocol.md)
- [goobs Documentation](https://pkg.go.dev/github.com/andreykaipov/goobs)
- [OBS Studio Documentation](https://obsproject.com/wiki/)
