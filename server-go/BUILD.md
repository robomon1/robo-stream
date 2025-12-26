# Build Instructions - Stream-Pi Server Go

## Prerequisites

### All Platforms
- Go 1.21 or later
- Git
- Make (optional, for using Makefile)

### macOS
```bash
# Install Go via Homebrew
brew install go

# Verify installation
go version
```

### Linux
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# Verify installation
go version
```

### Windows
Download and install Go from https://go.dev/dl/

## Building on macOS

### 1. Clone and Setup
```bash
# Navigate to the project
cd ~/git/stream-pi/server-go

# Download dependencies
go mod download

# Verify dependencies
go mod verify
```

### 2. Build for macOS (Native)
```bash
# Build for your current architecture (Apple Silicon M1/M2/M3 or Intel)
go build -o bin/robostream-server ./cmd/server

# Run the binary
./bin/robostream-server
```

### 3. Build for Specific macOS Architecture
```bash
# For Apple Silicon (M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -o bin/robostream-server-darwin-arm64 ./cmd/server

# For Intel Macs
GOOS=darwin GOARCH=amd64 go build -o bin/robostream-server-darwin-amd64 ./cmd/server
```

## Cross-Platform Compilation

### Build All Platforms at Once

Create a build script `build-all.sh`:

```bash
#!/bin/bash

# Server binary name
APP_NAME="robostream-server"
VERSION="1.0.0"
BUILD_DIR="bin"

# Create build directory
mkdir -p ${BUILD_DIR}

# Build flags for smaller binaries
LDFLAGS="-s -w"

echo "Building Stream-Pi Server for all platforms..."

# Linux x86_64
echo "Building for Linux x86_64..."
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-linux-amd64 cmd/server/main.go

# Linux ARM64 (Raspberry Pi 4, AWS Graviton, etc.)
echo "Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-linux-arm64 cmd/server/main.go

# Linux ARM v7 (Raspberry Pi 3)
echo "Building for Linux ARMv7..."
GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-linux-armv7 cmd/server/main.go

# macOS x86_64 (Intel)
echo "Building for macOS x86_64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-darwin-amd64 cmd/server/main.go

# macOS ARM64 (Apple Silicon M1/M2/M3)
echo "Building for macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-darwin-arm64 cmd/server/main.go

# Windows x86_64
echo "Building for Windows x86_64..."
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-windows-amd64.exe cmd/server/main.go

# Windows ARM64
echo "Building for Windows ARM64..."
GOOS=windows GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME}-windows-arm64.exe cmd/server/main.go

echo "Build complete! Binaries in ${BUILD_DIR}/"
ls -lh ${BUILD_DIR}/
```

Make it executable and run:
```bash
chmod +x build-all.sh
./build-all.sh
```

### Individual Platform Builds

#### Linux x86_64
```bash
GOOS=linux GOARCH=amd64 go build -o bin/robostream-server-linux-amd64 ./cmd/server
```

#### Linux ARM64 (Raspberry Pi 4, 5)
```bash
GOOS=linux GOARCH=arm64 go build -o bin/robostream-server-linux-arm64 ./cmd/server
```

#### Linux ARMv7 (Raspberry Pi 3)
```bash
GOOS=linux GOARCH=arm GOARM=7 go build -o bin/robostream-server-linux-armv7 ./cmd/server
```

#### macOS Intel (x86_64)
```bash
GOOS=darwin GOARCH=amd64 go build -o bin/robostream-server-darwin-amd64 ./cmd/server
```

#### macOS Apple Silicon (ARM64)
```bash
GOOS=darwin GOARCH=arm64 go build -o bin/robostream-server-darwin-arm64 ./cmd/server
```

#### Windows x86_64
```bash
GOOS=windows GOARCH=amd64 go build -o bin/robostream-server-windows-amd64.exe ./cmd/server
```

#### Windows ARM64
```bash
GOOS=windows GOARCH=arm64 go build -o bin/robostream-server-windows-arm64.exe ./cmd/server
```

## Using Makefile

Create a `Makefile`:

```makefile
APP_NAME := robostream-server
VERSION := 1.0.0
BUILD_DIR := bin
LDFLAGS := -s -w

.PHONY: all clean test linux-amd64 linux-arm64 linux-armv7 darwin-amd64 darwin-arm64 windows-amd64 windows-arm64

all: clean linux-amd64 linux-arm64 linux-armv7 darwin-amd64 darwin-arm64 windows-amd64 windows-arm64

clean:
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)

test:
	go test -v ./...

# Linux builds
linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 cmd/server/main.go

linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 cmd/server/main.go

linux-armv7:
	GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-armv7 cmd/server/main.go

# macOS builds
darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 cmd/server/main.go

darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 cmd/server/main.go

# Windows builds
windows-amd64:
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe cmd/server/main.go

windows-arm64:
	GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-windows-arm64.exe cmd/server/main.go

# Build for current platform
build:
	go build -o $(BUILD_DIR)/$(APP_NAME) cmd/server/main.go

# Run on current platform
run:
	go run cmd/server/main.go

# Install dependencies
deps:
	go mod download
	go mod verify

# Update dependencies
update:
	go get -u ./...
	go mod tidy
```

Usage:
```bash
# Build all platforms
make all

# Build specific platform
make darwin-arm64

# Build for current platform
make build

# Run tests
make test

# Clean build directory
make clean
```

## Build Optimization

### Smaller Binaries
```bash
# Strip debug info and reduce binary size
go build -ldflags="-s -w" -o bin/robostream-server ./cmd/server

# Further compress with UPX (optional)
upx --best --lzma bin/robostream-server
```

### Static Linking (Linux)
```bash
# Create fully static binary (no external dependencies)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -extldflags '-static'" -o bin/robostream-server-linux-amd64-static ./cmd/server
```

## Development Builds

### With Debug Info
```bash
# Build with debug symbols and race detector
go build -race -o bin/robostream-server-debug ./cmd/server
```

### With Version Info
```bash
# Embed version information
VERSION="1.0.0"
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD)

go build -ldflags="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" -o bin/robostream-server ./cmd/server
```

## Distribution

### Create Release Archives

```bash
#!/bin/bash

VERSION="1.0.0"
DIST_DIR="dist"

mkdir -p ${DIST_DIR}

# Function to create archive
create_archive() {
    local platform=$1
    local binary=$2
    local archive_name="robostream-server-${VERSION}-${platform}"
    
    mkdir -p ${DIST_DIR}/${archive_name}
    cp bin/${binary} ${DIST_DIR}/${archive_name}/
    cp README.md ${DIST_DIR}/${archive_name}/
    cp config.example.yaml ${DIST_DIR}/${archive_name}/
    
    if [[ $platform == *"windows"* ]]; then
        cd ${DIST_DIR} && zip -r ${archive_name}.zip ${archive_name}
    else
        cd ${DIST_DIR} && tar czf ${archive_name}.tar.gz ${archive_name}
    fi
    cd ..
    rm -rf ${DIST_DIR}/${archive_name}
}

# Create archives for each platform
create_archive "linux-amd64" "robostream-server-linux-amd64"
create_archive "linux-arm64" "robostream-server-linux-arm64"
create_archive "darwin-amd64" "robostream-server-darwin-amd64"
create_archive "darwin-arm64" "robostream-server-darwin-arm64"
create_archive "windows-amd64" "robostream-server-windows-amd64.exe"

echo "Release archives created in ${DIST_DIR}/"
```

## Testing Builds

### Verify Binary
```bash
# Check binary info
file bin/robostream-server-darwin-arm64

# Check size
ls -lh bin/

# Test run
./bin/robostream-server-darwin-arm64 --version
```

### Test Cross-Compiled Binary

For Raspberry Pi:
```bash
# Copy to Pi
scp bin/robostream-server-linux-arm64 pi@raspberrypi:~/

# SSH and test
ssh pi@raspberrypi
chmod +x robostream-server-linux-arm64
./robostream-server-linux-arm64 --version
```

## Troubleshooting

### Issue: "command not found: go"
**Solution**: Ensure Go is installed and in your PATH
```bash
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc  # or ~/.bashrc
```

### Issue: "cannot find package"
**Solution**: Download dependencies
```bash
go mod download
go mod tidy
```

### Issue: Build fails with "undefined reference"
**Solution**: Ensure CGO is properly configured or disable it
```bash
CGO_ENABLED=0 go build ...
```

### Issue: Binary too large
**Solution**: Use build flags to reduce size
```bash
go build -ldflags="-s -w" -o bin/robostream-server cmd/server/main.go
```

### Issue: Cross-compilation fails for ARM
**Solution**: Install cross-compilation tools
```bash
# macOS
brew install FiloSottile/musl-cross/musl-cross

# Set environment
export CC=aarch64-linux-musl-gcc
```

## CI/CD Integration

### GitHub Actions Example

`.github/workflows/build.yml`:
```yaml
name: Build

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Build
      env:
        GOOS: ${{ matrix.goos }}
        GOARCH: ${{ matrix.goarch }}
      run: |
        go build -v -o bin/robostream-server-${{ matrix.goos }}-${{ matrix.goarch }} cmd/server/main.go
    
    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: robostream-server-${{ matrix.goos }}-${{ matrix.goarch }}
        path: bin/
```

## Platform-Specific Notes

### macOS
- Apple Silicon (M1/M2/M3) uses `arm64`
- Intel Macs use `amd64`
- Binaries may need to be signed for distribution
- Use `lipo` to create universal binaries:
  ```bash
  lipo -create bin/robostream-server-darwin-amd64 bin/robostream-server-darwin-arm64 -output bin/robostream-server-darwin-universal
  ```

### Windows
- Executables require `.exe` extension
- May need to be signed for distribution
- Windows Defender may flag unsigned binaries

### Linux
- Static binaries recommended for maximum compatibility
- Different distros may require different flags
- ARM builds: specify `GOARM=7` for Raspberry Pi 3, omit for Pi 4/5

### Raspberry Pi
- Pi 4/5: Use `arm64` build
- Pi 3: Use `armv7` build (with `GOARM=7`)
- Pi Zero: Use `armv6` build (with `GOARM=6`)

## Quick Reference

| Platform | GOOS | GOARCH | GOARM | Command |
|----------|------|--------|-------|---------|
| macOS Intel | darwin | amd64 | - | `GOOS=darwin GOARCH=amd64 go build` |
| macOS Apple Silicon | darwin | arm64 | - | `GOOS=darwin GOARCH=arm64 go build` |
| Linux x86_64 | linux | amd64 | - | `GOOS=linux GOARCH=amd64 go build` |
| Linux ARM64 | linux | arm64 | - | `GOOS=linux GOARCH=arm64 go build` |
| Raspberry Pi 4/5 | linux | arm64 | - | `GOOS=linux GOARCH=arm64 go build` |
| Raspberry Pi 3 | linux | arm | 7 | `GOOS=linux GOARCH=arm GOARM=7 go build` |
| Windows x86_64 | windows | amd64 | - | `GOOS=windows GOARCH=amd64 go build` |
| Windows ARM64 | windows | arm64 | - | `GOOS=windows GOARCH=arm64 go build` |
