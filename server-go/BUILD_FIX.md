# Build Script Fix - December 22, 2025

## The Problem

The build scripts were using the wrong syntax:
```bash
go build cmd/server/main.go  # WRONG - treats as file, not package
```

This causes Go to look in the standard library path instead of the module.

## The Fix

Updated all build scripts to use package syntax:
```bash
go build ./cmd/server  # CORRECT - builds the package
```

## What Changed

1. **build-all.sh** - Uses `./cmd/server` instead of `cmd/server/main.go`
2. **Makefile** - All targets updated to use package path
3. **BUILD.md** - All examples updated with correct syntax

## How to Use

### Make sure you're in the server-go directory:
```bash
cd ~/git/stream-pi/server-go
```

### Then build:
```bash
# Using the script
./build-all.sh

# Or using make
make build           # Current platform
make darwin-arm64    # Apple Silicon Mac
make linux-arm64     # Raspberry Pi 4/5
make all            # All platforms

# Or directly
go build -o bin/streampi-server ./cmd/server
```

## Quick Test

```bash
cd ~/git/stream-pi/server-go

# Verify directory structure
ls -la go.mod cmd/server/main.go

# Should show:
# go.mod exists
# cmd/server/main.go exists

# Download dependencies
go mod download

# Build for current platform
go build -o bin/streampi-server ./cmd/server

# Or use make
make build

# Or use build script
./build-all.sh
```

## The Key Difference

**Wrong (treats as file):**
```bash
go build cmd/server/main.go
```

**Correct (builds package):**
```bash
go build ./cmd/server
```

When you build a package (not a file), Go:
- Uses the module context from go.mod
- Can resolve all imports correctly
- Treats it as a proper Go package

When you build a file, Go:
- Tries to compile just that file
- Doesn't have module context
- Can't resolve imports from the module

## All Build Commands Updated

All of these now use the correct `./cmd/server` syntax:

- Individual platform builds
- build-all.sh script
- Makefile targets
- BUILD.md documentation

## Verification

After running `./build-all.sh`, you should see:
```
🚀 Building Stream-Pi Server v1.0.0 for all platforms...

📦 Linux builds:
Building for linux/amd64... ✓
Building for linux/arm64... ✓
Building for linux/armv7... ✓
Building for linux/armv6... ✓

🍎 macOS builds:
Building for darwin/amd64... ✓
Building for darwin/arm64... ✓

🪟 Windows builds:
Building for windows/amd64... ✓
Building for windows/arm64... ✓

✅ Build complete! Binaries:
  streampi-server-linux-amd64         15M
  streampi-server-linux-arm64         14M
  ...
```

This fix is permanent - all future builds will work correctly!
