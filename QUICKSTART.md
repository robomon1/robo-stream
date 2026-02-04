# Robo-Stream Quick Start Guide

Get up and running with Robo-Stream in 10 minutes! This guide walks you through setting up OBS Studio, Robo-Stream Server, and Robo-Stream Client.

## Table of Contents

- [What You'll Need](#what-youll-need)
- [Step 1: Install OBS Studio](#step-1-install-obs-studio)
- [Step 2: Configure OBS WebSocket](#step-2-configure-obs-websocket)
- [Step 3: Install Robo-Stream Server](#step-3-install-robo-stream-server)
- [Step 4: Install Robo-Stream Client](#step-4-install-robo-stream-client)
- [Step 5: Connect and Test](#step-5-connect-and-test)
- [Troubleshooting](#troubleshooting)
- [Next Steps](#next-steps)

---

## What You'll Need

**Hardware:**
- Computer for streaming (with OBS Studio)
- Device for controlling (phone, tablet, or another computer)
- Both devices on the same WiFi network

**Software to Download:**
- OBS Studio (free)
- Robo-Stream Server (free)
- Robo-Stream Client (free)

**Time Required:** 10-15 minutes

---

## Step 1: Install OBS Studio

### Download OBS Studio

**Already have OBS?** Skip to [Step 2](#step-2-configure-obs-websocket)

**1. Download OBS Studio:**
- Go to: https://obsproject.com/download
- Select your operating system:
  - **Windows:** Download installer
  - **macOS:** Download DMG
  - **Linux:** Follow distribution instructions

**2. Install OBS Studio:**

**Windows:**
- Run the downloaded `.exe` file
- Follow installation wizard
- Launch OBS Studio

**macOS:**
- Open the downloaded `.dmg` file
- Drag OBS to Applications folder
- Launch OBS Studio from Applications
- If prompted, allow access to screen recording and audio

**Linux:**
```bash
# Ubuntu/Debian
sudo apt install obs-studio

# Fedora
sudo dnf install obs-studio

# Arch
sudo pacman -S obs-studio
```

**3. First-Time Setup:**
- Launch OBS Studio
- If Auto-Configuration Wizard appears:
  - Select "Optimize just for recording" or your preferred option
  - Complete the wizard
- You should see OBS Studio with a preview window

✅ **OBS Studio is now installed!**

---

## Step 2: Configure OBS WebSocket

Robo-Stream uses OBS WebSocket to communicate with OBS Studio.

### Enable OBS WebSocket

**1. Open OBS Studio** (if not already open)

**2. Open Settings:**
- Click **Tools** menu → **WebSocket Server Settings**
- (If you don't see this option, update OBS to version 28.0+)

**3. Enable WebSocket Server:**
- ✅ Check **"Enable WebSocket server"**

**4. Configure Server Settings:**
- **Server Port:** Leave as default (`4455`)
- **Server Password:** 
  - Option A: Leave blank (no password) - easier for local network
  - Option B: Set a password - more secure

**For this Quick Start, we recommend no password for simplicity.**

**5. Click "OK"**

**6. Note Your Server Details:**
- Port: `4455`
- Password: (none, or your chosen password)

✅ **OBS WebSocket is now enabled!**

---

## Step 3: Install Robo-Stream Server

The server runs on your streaming computer and connects to OBS.

### Download Robo-Stream Server

**1. Go to Releases:**
- Visit: https://github.com/robomon1/robo-stream/releases
- Find the latest release

**2. Download Server for Your Platform:**

**macOS:**
- Download: `robo-stream-server-macos-universal.zip`
- Works on both Intel and Apple Silicon Macs

**Windows:**
- Download: `robo-stream-server-windows-amd64.exe`

**Linux:**
- Download: `robo-stream-server-linux-amd64`

### Install and Launch

**macOS:**
```bash
# 1. Download and extract
unzip robo-stream-server-macos-universal.zip

# 2. Open the app
open robo-stream-server.app
```

**First time opening:**
- macOS may show "cannot be opened because it is from an unidentified developer"
- **Right-click** → **Open**
- Click **"Open"** in the dialog
- This only needs to be done once

**Windows:**
```bash
# 1. Download robo-stream-server-windows-amd64.exe

# 2. Double-click to run
```

**First time opening:**
- Windows may show "Windows protected your PC"
- Click **"More info"**
- Click **"Run anyway"**
- This only needs to be done once

**Linux:**
```bash
# 1. Download robo-stream-server-linux-amd64

# 2. Make executable
chmod +x robo-stream-server-linux-amd64

# 3. Run
./robo-stream-server-linux-amd64
```

### Verify Server is Running

**When the server starts, you should see:**

```
Robo-Stream Server v1.0.0
OBS Connection: Connected
Server URL: http://192.168.1.100:8080
```

**Important:** Write down your **Server URL** - you'll need it in Step 4!

Example: `http://192.168.1.100:8080`

**The IP address will be different on your network** - use the one shown in your server window.

✅ **Robo-Stream Server is now running!**

**Troubleshooting:**
- **"Cannot connect to OBS"** - Make sure OBS Studio is running first
- **"WebSocket connection failed"** - Check OBS WebSocket is enabled (Step 2)
- **No server URL shown** - Check your firewall settings

---

## Step 4: Install Robo-Stream Client

The client is the control interface you use from any device.

### Download Robo-Stream Client

**1. Choose Your Platform:**

**Mobile (Recommended for remote control):**

**Android:**
- **Option A:** Download from [Google Play Store](https://play.google.com/store/apps/details?id=com.robostream.robostreamclient)
- **Option B:** Download APK from [GitHub Releases](https://github.com/robomon1/robo-stream/releases)
  - Download: `robo-stream-client-v1.0.0.apk`
  - Install APK (requires "Install from Unknown Sources" enabled)

**iOS:**
- Download from [Apple App Store](https://apps.apple.com/app/robo-stream-client)

**Desktop (For controlling from another computer):**

**macOS:**
- Download: `robo-stream-client-macos-arm64.zip` (Apple Silicon - M1/M2/M3/M4)
- OR: `robo-stream-client-macos-intel.zip` (Intel Macs)

**Windows:**
- Download: `robo-stream-client-windows-amd64.exe`

**Linux:**
- Download: `robo-stream-client-linux-amd64.AppImage`

### Install Client

**Android (APK):**
1. Enable "Install from Unknown Sources" in Settings
2. Open downloaded APK
3. Tap "Install"
4. Launch "Robo-Stream Client"

**Android (Play Store) / iOS (App Store):**
1. Install from store
2. Launch app

**macOS:**
1. Extract zip file
2. Drag to Applications folder
3. Right-click → Open (first time only)

**Windows:**
1. Download and run `.exe`
2. Click "Run anyway" if SmartScreen appears (first time only)

**Linux:**
1. Download `.AppImage`
2. Make executable: `chmod +x robo-stream-client-*.AppImage`
3. Run: `./robo-stream-client-*.AppImage`

✅ **Robo-Stream Client is now installed!**

---

## Step 5: Connect and Test

Now let's connect the client to your server and test it out!

### Connect Client to Server

**1. Launch Robo-Stream Client**

**2. Settings Screen Appears:**

On first launch, you'll see the settings screen:
```
┌─────────────────────────────────┐
│      Robo-Stream Client         │
├─────────────────────────────────┤
│                                 │
│  Server URL:                    │
│  [http://192.168.1.100:8080  ] │
│                                 │
│         [Connect]               │
│                                 │
└─────────────────────────────────┘
```

**3. Enter Server URL:**
- Type or paste the Server URL from Step 3
- Example: `http://192.168.1.100:8080`
- **Make sure both devices are on the same WiFi network!**

**4. Click "Connect"**

**5. Connection Success!**

You should see:
- Blue banner: "Connected to http://192.168.1.100:8080"
- Button grid appears with control buttons
- OBS status indicators (streaming, recording, etc.)

✅ **You're connected!**

### Test the Connection

**Try these buttons:**

**1. Stream Control:**
- Click **"Start Streaming"** button
- Watch OBS Studio start streaming
- Click **"Stop Streaming"** button
- Confirm OBS stops streaming

**2. Recording Control:**
- Click **"Start Recording"** button
- OBS starts recording
- Click **"Stop Recording"** button
- OBS stops recording

**3. Scene Switching:**
- Click any scene button
- Watch OBS switch to that scene
- Try different scenes

**4. Source Control:**
- Click source visibility toggles
- Watch sources show/hide in OBS

🎉 **Everything works!**

---

## Troubleshooting

### Client Can't Connect to Server

**"Connection failed: Failed to connect to server"**

**Check:**
1. ✅ Both devices on same WiFi network
2. ✅ Server is running (check server window)
3. ✅ OBS Studio is running
4. ✅ Correct server URL entered (check IP address)
5. ✅ Firewall not blocking (see below)

**Firewall Issues:**

**Windows:**
- Allow Robo-Stream Server through Windows Firewall
- Windows Security → Firewall → Allow an app

**macOS:**
- System Preferences → Security & Privacy → Firewall
- Allow Robo-Stream Server

**Linux:**
```bash
# Allow port 8080
sudo ufw allow 8080/tcp
```

### OBS WebSocket Connection Failed

**"Cannot connect to OBS WebSocket"**

**Check:**
1. ✅ OBS Studio is running
2. ✅ OBS WebSocket is enabled (Tools → WebSocket Server Settings)
3. ✅ Port is 4455 (default)
4. ✅ Password matches (if you set one)

**Fix:**
- Restart OBS Studio
- Restart Robo-Stream Server
- Check OBS version (needs 28.0+)

### Server URL Not Showing

**Server starts but shows no URL**

**Find your IP manually:**

**Windows:**
```cmd
ipconfig
```
Look for "IPv4 Address" under your WiFi adapter

**macOS/Linux:**
```bash
ifconfig | grep inet
```
Look for IP address starting with 192.168 or 10.x

Then use: `http://YOUR_IP_ADDRESS:8080`

### Wrong IP Address

**Server shows IP but client can't connect**

Your computer might have multiple network interfaces.

**Try these IPs:**
- Look for IP starting with `192.168.x.x`
- Or `10.x.x.x`
- Avoid `127.0.0.1` (localhost - only works on same computer)

**Test from phone browser:**
- Open browser on your phone
- Go to: `http://YOUR_IP:8080/api/health`
- Should show: `{"status":"ok"}`
- If this works, use this IP in the client

### Android App Shows Black Screen

**App launches but screen is black**

**Fix:**
1. Press power button to lock screen
2. Unlock screen
3. App should appear

**Permanent fix:**
- Update to latest version from Play Store
- This was fixed in v1.0.1+

### iOS App Wrong Orientation

**App launches in portrait instead of landscape**

**Temporary fix:**
1. Rotate device once (any direction)
2. Rotate back to desired orientation
3. App should adjust

**Permanent fix:**
- Update to latest version from App Store
- This was fixed in v1.0.1+

### Buttons Don't Work

**Buttons appear but don't do anything**

**Check:**
1. ✅ OBS Studio is running
2. ✅ Server shows "OBS Connection: Connected"
3. ✅ Client shows "Connected" banner

**Fix:**
- Restart OBS Studio
- Restart Robo-Stream Server
- Reconnect client

### Still Having Issues?

**Get Help:**
- [GitHub Issues](https://github.com/robomon1/robo-stream/issues) - Report bugs
- [GitHub Discussions](https://github.com/robomon1/robo-stream/discussions) - Ask questions
- [Documentation](https://github.com/robomon1/robo-stream) - Read full docs

---

## Next Steps

### Customize Your Setup

**Create Custom Configurations:**
- Each configuration can have different button layouts
- Organize buttons by scene, function, or workflow
- Switch between configs on the fly

**Add More Clients:**
- Install client on multiple devices
- Control from phone, tablet, or computer
- All clients work simultaneously

**Configure Buttons:**
- Customize button labels and colors
- Map to different OBS functions
- Create scene-specific controls

### Advanced Features

**Multi-Scene Workflows:**
- Create buttons for scene transitions
- Control source visibility per scene
- Set up complex streaming workflows

**Status Monitoring:**
- Real-time stream status
- Recording indicators
- Virtual camera status
- Replay buffer status

**Network Setup:**
- Use from anywhere on your local network
- Add multiple streaming computers
- Control multiple OBS instances

### Learn More

**Documentation:**
- [README.md](README.md) - Project overview
- [RELEASE.md](RELEASE.md) - Release and publishing guide
- [GitHub Wiki](https://github.com/robomon1/robo-stream/wiki) - Detailed guides

**Community:**
- [Discussions](https://github.com/robomon1/robo-stream/discussions) - Share tips and tricks
- [Issues](https://github.com/robomon1/robo-stream/issues) - Report bugs or request features

---

## Quick Reference

### Startup Sequence

Every time you want to use Robo-Stream:

1. **Start OBS Studio** on streaming computer
2. **Start Robo-Stream Server** on streaming computer
3. **Note the Server URL** displayed
4. **Launch Robo-Stream Client** on control device
5. **Enter Server URL** and connect
6. **Start controlling!**

### URLs to Remember

- **OBS Studio:** https://obsproject.com
- **Robo-Stream Releases:** https://github.com/robomon1/robo-stream/releases
- **Documentation:** https://github.com/robomon1/robo-stream
- **Support:** https://github.com/robomon1/robo-stream/issues

### Default Ports

- **OBS WebSocket:** 4455
- **Robo-Stream Server:** 8080

### System Requirements

**Server:**
- OBS Studio 28.0+
- Windows 10+, macOS 10.15+, or Linux
- Same computer running OBS

**Client:**
- **Mobile:** Android 5.0+, iOS 12.0+
- **Desktop:** Windows 10+, macOS 10.15+, Linux
- Same WiFi network as server

---

## Tips for Best Experience

### Network Tips

✅ **Use WiFi 5 GHz** - Faster and more reliable than 2.4 GHz
✅ **Keep devices close to router** - Better signal strength
✅ **Avoid VPNs** - Can block local network communication
✅ **Use static IP** - Prevents server URL from changing

### Performance Tips

✅ **Close unnecessary apps** - Free up resources on streaming computer
✅ **Update regularly** - Get latest features and bug fixes
✅ **Use landscape mode** - Better experience on mobile devices
✅ **Charge devices** - Battery drain when streaming

### Workflow Tips

✅ **Test before streaming** - Verify all buttons work
✅ **Create backups** - Save your configurations
✅ **Use descriptive names** - Easy to identify buttons/configs
✅ **Practice transitions** - Get familiar with controls

---

## Success!

🎉 You're all set up and ready to control your streams remotely!

**What you've accomplished:**
- ✅ Installed and configured OBS Studio
- ✅ Enabled OBS WebSocket
- ✅ Installed Robo-Stream Server
- ✅ Installed Robo-Stream Client
- ✅ Connected and tested the system

**Now you can:**
- Start/stop streaming from anywhere in your house
- Switch scenes without being at your computer
- Control your entire streaming setup remotely
- Focus on your content, not the technical details

Happy streaming! 🎬✨

---

Need help? Found a bug? Have a suggestion?  
Visit: https://github.com/robomon1/robo-stream/issues

---

