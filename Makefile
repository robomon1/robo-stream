# RoboStream Build Makefile
# Builds server, client, and mobile apps for all platforms

.PHONY: all clean help
.PHONY: server server-mac server-windows server-linux
.PHONY: client client-mac client-windows client-linux
.PHONY: mobile android ios
.PHONY: zoomosc zoomosc-mac zoomosc-windows zoomosc-linux
.PHONY: release package

# Load environment variables from .env file if it exists
-include .env
export

# Version (update this for releases)
VERSION ?= 1.0.0

# Auto-generated build number: Unix timestamp minus 1,000,000,000
# Current value ~740M, max Android/iOS limit is 2.1B (~43 years of headroom)
BUILD := $(shell echo $$(($$(date +%s) - 1000000000)))

# Directories
SERVER_DIR    = server
CLIENT_DIR    = client-web
ZOOMOSC_DIR   = plugins/zoomosc
RELEASE_DIR   = releases/v$(VERSION)

# Build outputs
SERVER_BUILD   = $(SERVER_DIR)/build/bin
CLIENT_BUILD   = $(CLIENT_DIR)/electron-dist
ANDROID_BUILD  = $(CLIENT_DIR)/android/app/build/outputs/apk/release
IOS_BUILD      = $(CLIENT_DIR)/ios/build
ZOOMOSC_BUILD  = $(ZOOMOSC_DIR)/dist

# ZoomOSC plugin version (read from its version.txt)
ZOOMOSC_VERSION ?= $(shell cat $(ZOOMOSC_DIR)/version.txt 2>/dev/null || echo "dev")

#==============================================================================
# Help
#==============================================================================

help:
	@echo "RoboStream Build System"
	@echo ""
	@echo "Usage:"
	@echo "  make all              - Build everything for all platforms"
	@echo "  make server           - Build server for all platforms"
	@echo "  make client           - Build client for all platforms"
	@echo "  make mobile           - Build Android and iOS apps"
	@echo "  make release          - Create GitHub release package"
	@echo "  make clean            - Clean all build artifacts"
	@echo ""
	@echo "Individual targets:"
	@echo "  make server-mac       - Build server for macOS (universal)"
	@echo "  make server-mac-dmg   - Build server macOS DMG (signed & notarized)"
	@echo "  make server-windows   - Build server for Windows (with NSIS installer)"
	@echo "  make server-linux     - Build server for Linux"
	@echo "  make client-mac       - Build client for macOS (unsigned)"
	@echo "  make client-mac-signed - Build client for macOS (signed)"
	@echo "  make client-windows   - Build client for Windows"
	@echo "  make client-linux     - Build client for Linux"
	@echo "  make android          - Build Android APK"
	@echo "  make android-bundle   - Build Android App Bundle (for Play Store)"
	@echo "  make ios              - Build iOS IPA (command line via xcodebuild)"
	@echo ""
	@echo "ZoomOSC plugin:"
	@echo "  make zoomosc          - Build ZoomOSC plugin for all platforms"
	@echo "  make zoomosc-mac      - Build ZoomOSC plugin for macOS (arm64 + amd64)"
	@echo "  make zoomosc-windows  - Build ZoomOSC plugin for Windows (amd64)"
	@echo "  make zoomosc-linux    - Build ZoomOSC plugin for Linux (amd64)"
	@echo ""
	@echo "Code signing (macOS):"
	@echo "  make client-mac-signed - Sign with keychain cert (may prompt)"
	@echo "  CSC_LINK=/path/to/cert.p12 CSC_KEY_PASSWORD=pwd make client-mac-signed"
	@echo "                         - Sign with .p12 file (no prompts)"
	@echo ""
	@echo "Version management:"
	@echo "  make version                                 - Show all current versions"
	@echo "  make set-client-version VERSION=x.x.x  - Set client version (build auto-generated)"
	@echo "  make set-server-version VERSION=x.x.x  - Set server version"
	@echo "  make set-zoomosc-version VERSION=x.x.x - Set ZoomOSC plugin version"
	@echo "  make set-version VERSION=x.x.x         - Set all versions (build auto-generated)"
	@echo ""
	@echo "Example: make set-version VERSION=1.0.1"
	@echo "  Build number is auto-calculated as: unix timestamp - 1,000,000,000"

#==============================================================================
# Main Targets
#==============================================================================

all: server client mobile zoomosc
	@echo "✅ All builds complete!"
	@echo ""
	@echo "Build artifacts:"
	@echo "  Server:  $(SERVER_BUILD)/"
	@echo "  Client:  $(CLIENT_BUILD)/"
	@echo "  Android: $(ANDROID_BUILD)/"
	@echo "  iOS:     $(IOS_BUILD)/"
	@echo "  ZoomOSC: $(ZOOMOSC_BUILD)/"

server: server-mac server-windows server-linux
	@echo "✅ Server builds complete"

client: client-mac client-windows client-linux
	@echo "✅ Client builds complete"

mobile: android
	@echo "✅ Mobile builds complete (iOS requires manual Xcode archive)"
	@echo "Run 'make ios' to open Xcode for iOS build"

#==============================================================================
# Server (Wails)
#==============================================================================

server-mac:
	@echo "🔨 Building server for macOS (universal)..."
	cd $(SERVER_DIR) && wails build -platform darwin/universal
	@# Get version from wails.json
	$(eval SERVER_VERSION := $(shell grep '"productVersion"' $(SERVER_DIR)/wails.json | sed 's/.*": "//;s/",//'))
	@if [ -d "$(SERVER_BUILD)/robo-stream-server.app" ]; then \
		mv $(SERVER_BUILD)/robo-stream-server.app $(SERVER_BUILD)/robo-stream-server-$(SERVER_VERSION)-universal.app; \
		echo "📦 Renamed to: robo-stream-server-$(SERVER_VERSION)-universal.app"; \
	fi
	@echo "✅ Server macOS build complete"

server-windows:
	@echo "🔨 Building server for Windows..."
	cd $(SERVER_DIR) && wails build -platform windows/amd64
	@# Get version from wails.json
	$(eval SERVER_VERSION := $(shell grep '"productVersion"' $(SERVER_DIR)/wails.json | sed 's/.*": "//;s/",//'))
	@if [ -f "$(SERVER_BUILD)/robo-stream-server.exe" ]; then \
		mv $(SERVER_BUILD)/robo-stream-server.exe $(SERVER_BUILD)/robo-stream-server-$(SERVER_VERSION)-amd64.exe; \
		echo "📦 Renamed to: robo-stream-server-$(SERVER_VERSION)-amd64.exe"; \
	fi
	@echo "✅ Server Windows build complete"

server-linux:
	@echo "🔨 Building server for Linux..."
	cd $(SERVER_DIR) && wails build -platform linux/amd64
	@# Get version from wails.json
	$(eval SERVER_VERSION := $(shell grep '"productVersion"' $(SERVER_DIR)/wails.json | sed 's/.*": "//;s/",//'))
	@if [ -f "$(SERVER_BUILD)/robo-stream-server" ]; then \
		mv $(SERVER_BUILD)/robo-stream-server $(SERVER_BUILD)/robo-stream-server-$(SERVER_VERSION)-amd64; \
		echo "📦 Renamed to: robo-stream-server-$(SERVER_VERSION)-amd64"; \
	fi
	@echo "✅ Server Linux build complete"

server-mac-dmg: server-mac
	@echo "📦 Creating signed DMG for server..."
	$(eval SERVER_VERSION := $(shell grep '"productVersion"' $(SERVER_DIR)/wails.json | sed 's/.*": "//;s/",//'))
	@# Sign the .app bundle
	@echo "🔐 Signing server app..."
	codesign --deep --force --verify --verbose --sign "Developer ID Application" \
		--options runtime \
		$(SERVER_BUILD)/robo-stream-server-$(SERVER_VERSION)-universal.app
	@# Create DMG
	@echo "💿 Creating DMG..."
	@mkdir -p $(SERVER_BUILD)/dmg-temp
	@cp -r $(SERVER_BUILD)/robo-stream-server-$(SERVER_VERSION)-universal.app $(SERVER_BUILD)/dmg-temp/
	@hdiutil create -volname "Robo-Stream Server" \
		-srcfolder $(SERVER_BUILD)/dmg-temp \
		-ov -format UDZO \
		$(SERVER_BUILD)/robo-stream-server-$(SERVER_VERSION)-universal.dmg
	@rm -rf $(SERVER_BUILD)/dmg-temp
	@# Notarize DMG
	@echo "📝 Submitting for notarization..."
	@if [ -n "$$APPLE_ID" ]; then \
		xcrun notarytool submit $(SERVER_BUILD)/robo-stream-server-$(SERVER_VERSION)-universal.dmg \
			--apple-id "$$APPLE_ID" \
			--password "$$APPLE_APP_SPECIFIC_PASSWORD" \
			--team-id "$$APPLE_TEAM_ID" \
			--wait; \
		xcrun stapler staple $(SERVER_BUILD)/robo-stream-server-$(SERVER_VERSION)-universal.dmg; \
		echo "✅ Notarization complete!"; \
	else \
		echo "⚠️  Skipping notarization (APPLE_ID not set)"; \
	fi
	@echo "✅ Server DMG complete: robo-stream-server-$(SERVER_VERSION)-universal.dmg"

#==============================================================================
# Client (Electron)
#==============================================================================

client-prepare:
	@echo "📦 Preparing client build..."
	cd $(CLIENT_DIR) && npm install
	cd $(CLIENT_DIR) && npm run build

client-mac: client-prepare
	@echo "🔨 Building client for macOS (unsigned)..."
	cd $(CLIENT_DIR) && CSC_IDENTITY_AUTO_DISCOVERY=false npm run electron:build:mac
	@echo "✅ Client macOS unsigned build complete"

client-mac-signed: client-prepare
	@echo "🔨 Building client for macOS (signed)..."
	@if [ -z "$$CSC_LINK" ]; then \
		echo "⚠️  CSC_LINK not set, using keychain certificate..."; \
		echo "⚠️  May require keychain password..."; \
		cd $(CLIENT_DIR) && DEBUG=electron-builder,electron-notarize  npm run electron:build:mac; \
	else \
		echo "🔐 Using certificate from: $$CSC_LINK"; \
		cd $(CLIENT_DIR) && DEBUG=electron-builder,electron-notarize  npm run electron:build:mac; \
	fi
	@echo "✅ Client macOS signed build complete"

client-windows: client-prepare
	@echo "🔨 Building client for Windows..."
	cd $(CLIENT_DIR) && CSC_IDENTITY_AUTO_DISCOVERY=false npm run electron:build:win
	@echo "✅ Client Windows build complete"

client-linux: client-prepare
	@echo "🔨 Building client for Linux..."
	cd $(CLIENT_DIR) && CSC_IDENTITY_AUTO_DISCOVERY=false npm run electron:build:linux
	@echo "✅ Client Linux build complete"

#==============================================================================
# Mobile
#==============================================================================

android: client-prepare
	@echo "🔨 Building Android APK..."
	cd $(CLIENT_DIR) && npx cap sync android
	cd $(CLIENT_DIR)/android && chmod +x gradlew
	cd $(CLIENT_DIR)/android && ./gradlew assembleRelease
	@echo "✅ Android APK build complete"
	@echo "   Output: $(ANDROID_BUILD)/app-release.apk"

android-signed: client-prepare
	@echo "🔨 Building signed Android APK..."
	cd $(CLIENT_DIR) && npx cap sync android
	cd $(CLIENT_DIR)/android && chmod +x gradlew
	cd $(CLIENT_DIR)/android && ./gradlew assembleRelease
	@echo "✅ Signed Android APK build complete"

android-bundle: client-prepare
	@echo "🔨 Building Android App Bundle (AAB)..."
	cd $(CLIENT_DIR) && npx cap sync android
	cd $(CLIENT_DIR)/android && chmod +x gradlew
	cd $(CLIENT_DIR)/android && ./gradlew bundleRelease
	@echo "✅ Android App Bundle build complete"
	@echo "   Output: $(CLIENT_DIR)/android/app/build/outputs/bundle/release/app-release.aab"

ios: client-prepare
	@echo "📱 Building iOS IPA from command line..."
	cd $(CLIENT_DIR) && npx cap sync ios
	mkdir -p $(IOS_BUILD)
	cd $(CLIENT_DIR)/ios/App && xcodebuild archive \
		-workspace App.xcworkspace \
		-scheme App \
		-configuration Release \
		-sdk iphoneos \
		-archivePath ../build/App.xcarchive
	xcodebuild -exportArchive \
		-archivePath $(IOS_BUILD)/App.xcarchive \
		-exportPath $(IOS_BUILD)/ \
		-exportOptionsPlist $(CLIENT_DIR)/ios/ExportOptions.plist
	@echo "✅ iOS IPA build complete"
	@echo "   Output: $(IOS_BUILD)/"

#==============================================================================
# ZoomOSC Plugin
#==============================================================================

zoomosc: zoomosc-mac zoomosc-windows zoomosc-linux
	@echo "✅ ZoomOSC plugin builds complete"
	@echo "   Output: $(ZOOMOSC_BUILD)/"

zoomosc-mac:
	@echo "🔨 Building ZoomOSC plugin for macOS (arm64 + amd64)..."
	$(MAKE) -C $(ZOOMOSC_DIR) build-mac-arm64 build-mac-amd64 VERSION=$(ZOOMOSC_VERSION)
	@echo "✅ ZoomOSC macOS builds complete"

zoomosc-windows:
	@echo "🔨 Building ZoomOSC plugin for Windows (amd64)..."
	$(MAKE) -C $(ZOOMOSC_DIR) build-windows VERSION=$(ZOOMOSC_VERSION)
	@echo "✅ ZoomOSC Windows build complete"

zoomosc-linux:
	@echo "🔨 Building ZoomOSC plugin for Linux (amd64)..."
	$(MAKE) -C $(ZOOMOSC_DIR) build-linux VERSION=$(ZOOMOSC_VERSION)
	@echo "✅ ZoomOSC Linux build complete"

#==============================================================================
# Release Packaging
#==============================================================================

package: clean-release
	@echo "📦 Packaging release v$(VERSION)..."
	@mkdir -p $(RELEASE_DIR)
	
	@# Server binaries
	@echo "📁 Copying server binaries..."
	@if [ -f "$(SERVER_BUILD)/robo-stream-server.app/Contents/MacOS/robo-stream-server" ]; then \
		cd $(SERVER_BUILD) && zip -r ../../../$(RELEASE_DIR)/robo-stream-server-macos-universal.zip robo-stream-server.app; \
	fi
	@if [ -f "$(SERVER_BUILD)/robo-stream-server.exe" ]; then \
		cp $(SERVER_BUILD)/robo-stream-server.exe $(RELEASE_DIR)/robo-stream-server-windows-amd64.exe; \
	fi
	@if [ -f "$(SERVER_BUILD)/robo-stream-server" ]; then \
		cp $(SERVER_BUILD)/robo-stream-server $(RELEASE_DIR)/robo-stream-server-linux-amd64; \
	fi
	
	@# Client binaries
	@echo "📁 Copying client binaries..."
	@if [ -d "$(CLIENT_BUILD)/mac-arm64" ]; then \
		cd $(CLIENT_BUILD)/mac-arm64 && zip -r ../../../$(RELEASE_DIR)/robo-stream-client-macos-arm64.zip robo-stream-client.app; \
	fi
	@if [ -d "$(CLIENT_BUILD)/mac" ]; then \
		cd $(CLIENT_BUILD)/mac && zip -r ../../../$(RELEASE_DIR)/robo-stream-client-macos-intel.zip robo-stream-client.app; \
	fi
	@if [ -f "$(CLIENT_BUILD)/robo-stream-client Setup *.exe" ]; then \
		cp $(CLIENT_BUILD)/robo-stream-client*.exe $(RELEASE_DIR)/robo-stream-client-windows-amd64.exe; \
	fi
	@if [ -f "$(CLIENT_BUILD)/*.AppImage" ]; then \
		cp $(CLIENT_BUILD)/*.AppImage $(RELEASE_DIR)/robo-stream-client-linux-amd64.AppImage; \
	fi
	
	@# Mobile
	@echo "📁 Copying mobile apps..."
	@if [ -f "$(ANDROID_BUILD)/app-release.apk" ]; then \
		cp $(ANDROID_BUILD)/app-release.apk $(RELEASE_DIR)/robo-stream-client-v$(VERSION).apk; \
	fi
	@if [ -f "$(IOS_BUILD)/App.ipa" ]; then \
		cp $(IOS_BUILD)/App.ipa $(RELEASE_DIR)/robo-stream-client-v$(VERSION).ipa; \
	fi

	@# ZoomOSC plugin
	@echo "📁 Copying ZoomOSC plugin archives..."
	@for f in $(ZOOMOSC_BUILD)/zoomosc-controller-*.tar.gz $(ZOOMOSC_BUILD)/zoomosc-controller-*.zip; do \
		[ -f "$$f" ] && cp "$$f" $(RELEASE_DIR)/ && echo "  $$f"; \
	done || true

	@echo ""
	@echo "✅ Release package created: $(RELEASE_DIR)/"
	@echo ""
	@echo "Contents:"
	@ls -lh $(RELEASE_DIR)/
	@echo ""
	@echo "Next steps:"
	@echo "  1. git tag v$(VERSION)"
	@echo "  2. git push origin v$(VERSION)"
	@echo "  3. Create GitHub release and upload files from $(RELEASE_DIR)/"

release: all package
	@echo "✅ Release v$(VERSION) ready!"

#==============================================================================
# Development
#==============================================================================

dev-server:
	@echo "🚀 Starting server in dev mode..."
	cd $(SERVER_DIR) && wails dev

dev-client:
	@echo "🚀 Starting client in dev mode..."
	cd $(CLIENT_DIR) && npm run dev

#==============================================================================
# Testing
#==============================================================================

test-server:
	@echo "🧪 Testing server..."
	cd $(SERVER_DIR) && go test ./...

test-client:
	@echo "🧪 Testing client..."
	cd $(CLIENT_DIR) && npm test

test-zoomosc:
	@echo "🧪 Testing ZoomOSC plugin..."
	cd $(ZOOMOSC_DIR) && go test ./...

#==============================================================================
# Clean
#==============================================================================

clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(SERVER_BUILD)
	rm -rf $(CLIENT_BUILD)
	rm -rf $(CLIENT_DIR)/android/app/build
	rm -rf $(CLIENT_DIR)/ios/build
	rm -rf $(CLIENT_DIR)/dist
	rm -rf $(CLIENT_DIR)/node_modules/.vite
	rm -rf $(ZOOMOSC_BUILD)
	@echo "✅ Clean complete"

clean-release:
	@echo "🧹 Cleaning release directory..."
	rm -rf releases/
	@echo "✅ Release directory cleaned"

clean-all: clean clean-release
	@echo "🧹 Deep cleaning..."
	rm -rf $(CLIENT_DIR)/node_modules
	@echo "✅ Deep clean complete"

#==============================================================================
# Version Management
#==============================================================================

version:
	@echo "Main (version.txt):"
	@cat version.txt 2>/dev/null | sed 's/^/  /' || echo "  Not found"
	@echo ""
	@echo "Server (wails.json):"
	@grep '"productVersion"' $(SERVER_DIR)/wails.json | sed 's/^[[:space:]]*/  /' || echo "  Not found"
	@echo ""
	@echo "Client package.json:"
	@grep '"version"' $(CLIENT_DIR)/package.json | sed 's/^[[:space:]]*/  /' || echo "  Not found"
	@echo ""
	@echo "Android build.gradle:"
	@grep 'versionName' $(CLIENT_DIR)/android/app/build.gradle | sed 's/^[[:space:]]*/  /' || echo "  Not found"
	@grep 'versionCode' $(CLIENT_DIR)/android/app/build.gradle | sed 's/^[[:space:]]*/  /' || echo "  Not found"
	@echo ""
	@echo "iOS (project.pbxproj):"
	@grep 'MARKETING_VERSION = ' $(CLIENT_DIR)/ios/App/App.xcodeproj/project.pbxproj | head -1 | sed 's/^[[:space:]]*/  /' || echo "  Not found"
	@grep 'CURRENT_PROJECT_VERSION = ' $(CLIENT_DIR)/ios/App/App.xcodeproj/project.pbxproj | head -1 | sed 's/^[[:space:]]*/  /' || echo "  Not found"
	@echo ""
	@echo "ZoomOSC plugin (plugins/zoomosc/version.txt):"
	@cat $(ZOOMOSC_DIR)/version.txt 2>/dev/null | sed 's/^/  /' || echo "  Not found"

set-client-version:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: VERSION not specified"; \
		echo "Usage: make set-client-version VERSION=1.0.0"; \
		exit 1; \
	fi
	@echo "Setting client version to $(VERSION) (build $(BUILD))..."
	@echo ""

	@# Update client package.json
	@echo "📦 Updating package.json..."
	cd $(CLIENT_DIR) && npm version $(VERSION) --no-git-tag-version --allow-same-version

	@# Update Android
	@echo "🤖 Updating Android..."
	@sed -i.bak 's/versionName = ".*"/versionName = "$(VERSION)"/' $(CLIENT_DIR)/android/app/build.gradle
	@sed -i.bak 's/versionCode = [0-9]*/versionCode = $(BUILD)/' $(CLIENT_DIR)/android/app/build.gradle
	@rm $(CLIENT_DIR)/android/app/build.gradle.bak

	@# Update iOS via xcodeproj
	@echo "🍎 Updating iOS..."
	@sed -i.bak 's/MARKETING_VERSION = .*/MARKETING_VERSION = $(VERSION);/' $(CLIENT_DIR)/ios/App/App.xcodeproj/project.pbxproj
	@sed -i.bak 's/CURRENT_PROJECT_VERSION = .*/CURRENT_PROJECT_VERSION = $(BUILD);/' $(CLIENT_DIR)/ios/App/App.xcodeproj/project.pbxproj
	@rm -f $(CLIENT_DIR)/ios/App/App.xcodeproj/project.pbxproj.bak

	@echo ""
	@echo "✅ Client version updated to $(VERSION) (build $(BUILD))"
	@echo ""
	@echo "Updated files:"
	@echo "  • $(CLIENT_DIR)/package.json"
	@echo "  • $(CLIENT_DIR)/android/app/build.gradle"
	@echo "  • $(CLIENT_DIR)/ios/App/App.xcodeproj/project.pbxproj"
	@echo ""
	@echo "Next steps:"
	@echo "  1. git add -A"
	@echo "  2. git commit -m 'Bump client version to $(VERSION)'"
	@echo "  3. git tag client-v$(VERSION)"

set-server-version:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: VERSION not specified"; \
		echo "Usage: make set-server-version VERSION=1.0.0"; \
		exit 1; \
	fi
	@echo "Setting server version to $(VERSION)..."
	@echo ""
	
	@# Update Server wails.json
	@echo "🖥️  Updating Server..."
	@if [ -f "$(SERVER_DIR)/wails.json" ]; then \
		sed -i.bak 's/"productVersion": ".*"/"productVersion": "$(VERSION)"/' $(SERVER_DIR)/wails.json 2>/dev/null || true; \
		rm -f $(SERVER_DIR)/wails.json.bak; \
	fi

	@# Update server/version.go
	@if [ -f "$(SERVER_DIR)/version.go" ]; then \
		sed -i.bak 's/var Version = ".*"/var Version = "$(VERSION)"/' $(SERVER_DIR)/version.go 2>/dev/null || true; \
		rm -f $(SERVER_DIR)/version.go.bak; \
	fi

	@echo ""
	@echo "✅ Server version updated to $(VERSION)"
	@echo ""
	@echo "Updated files:"
	@echo "  • $(SERVER_DIR)/wails.json"
	@echo "  • $(SERVER_DIR)/version.go"
	@echo ""
	@echo "Next steps:"
	@echo "  1. git add -A"
	@echo "  2. git commit -m 'Bump server version to $(VERSION)'"
	@echo "  3. git tag server-v$(VERSION)"

set-zoomosc-version:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: VERSION not specified"; \
		echo "Usage: make set-zoomosc-version VERSION=1.0.0"; \
		exit 1; \
	fi
	@echo "Setting ZoomOSC plugin version to $(VERSION)..."
	@echo "$(VERSION)" > $(ZOOMOSC_DIR)/version.txt
	@echo ""
	@echo "✅ ZoomOSC plugin version updated to $(VERSION)"
	@echo ""
	@echo "Updated files:"
	@echo "  • $(ZOOMOSC_DIR)/version.txt"

set-version: set-client-version set-server-version set-zoomosc-version
	@echo "$(VERSION)" > version.txt
	@echo ""
	@echo "✅ All versions updated to $(VERSION)!"
	@echo "  • version.txt (root)"

#==============================================================================
# Quick Builds (for testing)
#==============================================================================

quick-server-mac:
	@echo "⚡ Quick server build (macOS only)..."
	cd $(SERVER_DIR) && wails build -platform darwin/universal -clean=false

quick-client-mac:
	@echo "⚡ Quick client build (macOS only, unsigned)..."
	cd $(CLIENT_DIR) && npm run build && CSC_IDENTITY_AUTO_DISCOVERY=false npm run electron:build:mac

quick-client-mac-signed:
	@echo "⚡ Quick client build (macOS only, signed)..."
	@if [ -z "$$CSC_LINK" ]; then \
		echo "⚠️  CSC_LINK not set, using keychain certificate..."; \
		cd $(CLIENT_DIR) && npm run build &&  npm run electron:build:mac; \
	else \
		echo "🔐 Using certificate from: $$CSC_LINK"; \
		cd $(CLIENT_DIR) && npm run build &&  npm run electron:build:mac; \
	fi

quick-android:
	@echo "⚡ Quick Android build..."
	cd $(CLIENT_DIR) && npm run build && npx cap sync android
	cd $(CLIENT_DIR)/android && ./gradlew assembleDebug
	@echo "✅ Debug APK: $(CLIENT_DIR)/android/app/build/outputs/apk/debug/app-debug.apk"

#==============================================================================
# Info
#==============================================================================

info:
	@echo "RoboStream Build Information"
	@echo "============================="
	@echo ""
	@echo "Version: $(VERSION)"
	@echo ""
	@echo "Directories:"
	@echo "  Server:  $(SERVER_DIR)"
	@echo "  Client:  $(CLIENT_DIR)"
	@echo "  Release: $(RELEASE_DIR)"
	@echo ""
	@echo "Tools:"
	@which wails || echo "  ❌ wails not found"
	@which node || echo "  ❌ node not found"
	@which npm || echo "  ❌ npm not found"
	@which npx || echo "  ❌ npx not found"
	@echo ""
	@echo "Node version:"
	@node --version || echo "  ❌ node not installed"
	@echo ""
	@echo "Go version:"
	@go version || echo "  ❌ go not installed"
