# Quick Start Guide

## Installation (5 minutes)

### 1. Extract Files

Extract to: `~/git/robomon1/robo-stream/server`

```bash
cd ~/git/robomon1/robo-stream
tar xzf robo-stream-server.tar.gz
cd server
```

### 2. Install Dependencies

```bash
make install
```

This will:
- Install frontend npm packages
- Download Go dependencies

### 3. Run Development Mode

```bash
make dev
```

The app will open with developer tools enabled.

## First Use (10 minutes)

### Step 1: Connect to OBS

1. Open OBS Studio
2. Tools → WebSocket Server Settings
3. Enable WebSocket server
4. Note the port (usually 4455)
5. In Robo-Stream Server:
   - Click **OBS Settings** in sidebar
   - Enter `ws://localhost:4455`
   - Enter password (if set)
   - Click **Connect**

### Step 2: Create Some Buttons

1. Click **Button Library**
2. Click **New Button**
3. Example buttons to create:
   - **Go Live**: video icon, red, action: start_stream
   - **Stop Stream**: stop-circle icon, gray, action: stop_stream  
   - **Mute Mic**: mic-off icon, orange, action: toggle_input_mute
   - **Main Scene**: layout icon, blue, action: switch_scene (params: scene_name="Main")

### Step 3: Create a Configuration

1. Click **Configurations**
2. Click **New Configuration**
3. Name it "Streamer"
4. Set grid to 4x3
5. Drag buttons from library onto grid
6. Save configuration

### Step 4: Set as Default

1. Find your "Streamer" configuration
2. Click **Set Default**

Now when clients connect, they'll automatically get this configuration!

## Building for Production

```bash
make build
```

Binary will be in `build/bin/robo-stream-server`

## Next Steps

- Create more configurations for different roles
- Connect client devices to http://YOUR_IP:8080
- Customize button icons and colors
- Set up multiple layouts for your team

## Troubleshooting

**Can't connect to OBS?**
- Make sure OBS WebSocket is enabled
- Try `ws://127.0.0.1:4455` instead of localhost
- Check if password is required

**Frontend not loading?**
- Run `cd frontend && npm install`
- Delete `frontend/node_modules` and reinstall

**Build fails?**
- Run `make clean` then `make install`
- Check Go version: `go version` (need 1.21+)
- Check Node version: `node --version` (need 18+)

## Architecture Overview

```
┌─────────────────────────────┐
│   Robo-Stream Server        │
│   (This app - Desktop)      │
│                             │
│   - Button Library          │
│   - Configurations          │
│   - Client Sessions         │
│   - OBS WebSocket           │
└──────────┬──────────────────┘
           │ HTTP API (:8080)
           ├───────────────────┐
           │                   │
    ┌──────▼──────┐    ┌──────▼──────┐
    │  Client 1   │    │  Client 2   │
    │ (MacBook)   │    │   (iPad)    │
    │             │    │             │
    │ Config:     │    │ Config:     │
    │ "Streamer"  │    │ "Assistant" │
    └──────┬──────┘    └──────┬──────┘
           │                  │
           └──────────┬───────┘
                      │
              ┌───────▼────────┐
              │   OBS Studio   │
              └────────────────┘
```

## Key Concepts

**Button Library**: Reusable buttons
- Create once, use in multiple configurations
- Edit a button, all configs update

**Configuration**: Button layout for a role
- Define grid size (e.g., 4x3, 3x2)
- Assign buttons from library to positions
- Multiple configs for different users

**Client Session**: Tracks which client uses which config
- Server remembers client preferences
- Auto-assigns default config to new clients
- Can switch configs from server UI

Enjoy your stream deck setup! 🎉
