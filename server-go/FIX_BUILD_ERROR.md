# How to Fix the Build Error - Step by Step

## The Problem

You're seeing this error:
```
package cmd/server/main.go is not in std
```

This means your `build-all.sh` file still has the OLD syntax.

## Quick Fix - Replace the Build Script

### Step 1: Check Your Current File

```bash
cd ~/git/stream-pi/server-go

# Show the problematic line
grep "CMD_" build-all.sh
```

**If you see:**
```
CMD_PATH="cmd/server/main.go"   # OLD - WRONG
```

**You need:**
```
CMD_PACKAGE="./cmd/server"      # NEW - CORRECT
```

### Step 2: Replace the File

**Option A: Download fresh from the tarball**
```bash
cd ~/git/stream-pi/server-go

# Extract just the build-all.sh from the tarball
tar xzf ~/Downloads/stream-pi-go.tar.gz stream-pi-go/server-go/build-all.sh --strip-components=2

# Make it executable
chmod +x build-all.sh
```

**Option B: Verify and test first**
```bash
cd ~/git/stream-pi/server-go

# Run the test script
chmod +x test-build.sh
./test-build.sh
```

If the test script works, then try the build script again.

### Step 3: Verify the Fix

```bash
# This should show CMD_PACKAGE, not CMD_PATH
grep "CMD_PACKAGE\|CMD_PATH" build-all.sh

# Should output:
# CMD_PACKAGE="./cmd/server"
```

### Step 4: Try Building

```bash
./build-all.sh
```

## Alternative: Manual Build

If the script still doesn't work, build manually:

```bash
cd ~/git/stream-pi/server-go

# Create bin directory
mkdir -p bin

# Build for macOS Apple Silicon (your Mac)
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/streampi-server-darwin-arm64 ./cmd/server

# Build for Raspberry Pi 4/5
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bin/streampi-server-linux-arm64 ./cmd/server

# Test it
./bin/streampi-server-darwin-arm64 --version
```

## The Root Cause

The issue is the difference between building a **file** vs a **package**:

**WRONG (building a file):**
```bash
go build cmd/server/main.go
```
This tells Go to compile just that one file, without module context.

**CORRECT (building a package):**
```bash
go build ./cmd/server
```
This tells Go to build the package at that path, with full module context.

## Quick Verification Checklist

Run these commands in order:

```bash
cd ~/git/stream-pi/server-go

# 1. Check you have go.mod
ls -la go.mod

# 2. Check you have the main.go
ls -la cmd/server/main.go

# 3. Check your build script
head -25 build-all.sh | grep CMD

# 4. Try simple build
go build -o bin/test ./cmd/server

# 5. If that works, try the script
./build-all.sh
```

## Still Having Issues?

Try this minimal test:

```bash
cd ~/git/stream-pi/server-go

# Most basic build possible
go build ./cmd/server

# This should create a binary called "server" in the current directory
ls -la server

# Run it
./server --version
```

If this works but the script doesn't, the script file is corrupted or cached.

## Nuclear Option: Fresh Start

```bash
# Go to parent directory
cd ~/git/stream-pi

# Remove the old server-go
rm -rf server-go

# Extract fresh from tarball
tar xzf ~/Downloads/stream-pi-go.tar.gz
cd stream-pi-go/server-go

# Download dependencies
go mod download

# Try build
./build-all.sh
```

## Contact Info

If you're still stuck, show me the output of:
```bash
cd ~/git/stream-pi/server-go
head -30 build-all.sh
```
