# Robo-Stream Release Process

This guide covers the complete process for creating and publishing Robo-Stream releases.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Version Management](#version-management)
- [Building Release Artifacts](#building-release-artifacts)
- [Creating GitHub Releases](#creating-github-releases)
- [Publishing to App Stores](#publishing-to-app-stores)
- [Screenshots Guide](#screenshots-guide)
- [Post-Release](#post-release)

---

## Prerequisites

### Code Signing Setup

Before creating releases, you need to set up code signing for each platform.

#### Android Code Signing

**1. Create Release Keystore:**

```bash
cd client-web/android/app
keytool -genkey -v -keystore robostream-release.keystore \
  -alias robostream \
  -keyalg RSA \
  -keysize 2048 \
  -validity 10000
```

**What you'll be asked:**
- Keystore password (SAVE THIS!)
- Key password (can be same as keystore, SAVE THIS!)
- First and last name: Your Name
- Organizational unit: Robo-Stream
- Organization: Your Company
- City, State, Country

**2. Save Passwords Securely:**

⚠️ **CRITICAL:** Without these passwords, you cannot update your app!

Store passwords in:
- Password manager (1Password, LastPass, etc.)
- Secure notes
- Encrypted file

**3. Create keystore.properties:**

```bash
cd client-web/android
cat > keystore.properties << EOF
storeFile=app/robostream-release.keystore
storePassword=YOUR_KEYSTORE_PASSWORD
keyAlias=robostream
keyPassword=YOUR_KEY_PASSWORD
EOF
```

Replace `YOUR_KEYSTORE_PASSWORD` and `YOUR_KEY_PASSWORD` with your actual passwords.

**4. Backup Keystore:**

```bash
# CRITICAL: Without this keystore, you cannot update your app!
cp app/robostream-release.keystore ~/Backups/

# Also save to cloud storage (Google Drive, Dropbox, etc.)
# Keep multiple backups in different locations
```

**5. Verify Signing:**

```bash
# Test build
cd android
./gradlew bundleRelease

# Verify signature
jarsigner -verify -verbose -certs app/build/outputs/bundle/release/app-release.aab

# Should show: jar verified.
```

#### iOS Code Signing

**1. Apple Developer Account:**

- Join Apple Developer Program: https://developer.apple.com/programs/
- Cost: $99/year
- Processing time: 1-2 days (individual), 1-2 weeks (organization)

**2. Create App ID:**

1. Go to: https://developer.apple.com/account/resources/identifiers/list
2. Click **+** to create new identifier
3. Select **App IDs** → **App**
4. Description: `Robo-Stream Client`
5. Bundle ID: **Explicit** → `com.robostream.robostreamclient`
6. Select capabilities if needed
7. Click **Continue** → **Register**

**3. Code Signing (Automatic - Recommended):**

```bash
cd client-web
npx cap open ios
```

In Xcode:
1. Select **App** target
2. **Signing & Capabilities** tab
3. ✅ Check **"Automatically manage signing"**
4. Select your **Team**
5. Xcode handles certificates and profiles automatically

**4. Code Signing (Manual - Advanced):**

If you need manual signing:

1. **Create Distribution Certificate:**
   - https://developer.apple.com/account/resources/certificates/list
   - Create **Apple Distribution** certificate
   - Download and install

2. **Create Provisioning Profile:**
   - https://developer.apple.com/account/resources/profiles/list
   - **App Store** distribution type
   - Select your App ID
   - Select Distribution certificate
   - Download and install

3. **Configure in Xcode:**
   - Uncheck "Automatically manage signing"
   - Select provisioning profile manually

#### macOS Code Signing (Electron - Optional)

**For unsigned builds:**
- Set `CSC_IDENTITY_AUTO_DISCOVERY=false` (already done in Makefile)
- Users get Gatekeeper warning but can bypass with right-click → Open

**For signed builds:**
- Apple Developer ID certificate required
- Notarization recommended
- Beyond scope of this guide

---

## Version Management

Robo-Stream uses separate versioning for server and client components.

### Versioning Strategy

**Semantic Versioning:** MAJOR.MINOR.PATCH

- **MAJOR:** Breaking changes
- **MINOR:** New features (backward compatible)
- **PATCH:** Bug fixes

**Example:**
- Server: `1.2.0` (new feature added)
- Client: `1.0.1` (bug fix)

### View Current Versions

```bash
make version
```

Output shows versions for:
- Client (package.json)
- Android (versionName, versionCode)
- iOS (MARKETING_VERSION, CURRENT_PROJECT_VERSION)
- Server (wails.json)

### Update Client Version

```bash
make set-client-version VERSION=1.0.1 BUILD=2
```

**Parameters:**
- `VERSION`: Semantic version (1.0.1)
- `BUILD`: Integer build number (always increment)

**Updates:**
- `client-web/package.json` → version
- `client-web/android/app/build.gradle` → versionName, versionCode
- `client-web/ios/App/App.xcodeproj/project.pbxproj` → MARKETING_VERSION, CURRENT_PROJECT_VERSION

**Build Number Rules:**
- Always increment for each upload to stores
- Can't reuse build numbers
- Independent of version number

**Examples:**
```bash
# Bug fix release
make set-client-version VERSION=1.0.1 BUILD=2

# New feature release
make set-client-version VERSION=1.1.0 BUILD=3

# Breaking change
make set-client-version VERSION=2.0.0 BUILD=4
```

### Update Server Version

```bash
make set-server-version VERSION=1.1.0
```

**Parameters:**
- `VERSION`: Semantic version (1.1.0)

**Updates:**
- `server/wails.json` → productVersion

**No build number needed** - desktop apps use semantic versioning only.

### Update Both Versions

```bash
make set-version VERSION=1.0.0 BUILD=1
```

Updates both server and client to the same version. Use when releasing coordinated major versions.

### Commit Version Changes

```bash
# After setting versions
git add -A
git commit -m "Bump client to v1.0.1 and server to v1.1.0"
git tag client-v1.0.1
git tag server-v1.1.0
git push origin main --tags
```

---

## Building Release Artifacts

### Build All Platforms

```bash
# Build everything
make all

# Or build and package for release
make release VERSION=1.0.0
```

This creates `releases/v1.0.0/` directory with all binaries.

### Build Individual Components

**Server:**
```bash
# All platforms
make server

# Individual platforms
make server-mac       # macOS universal binary
make server-windows   # Windows
make server-linux     # Linux
```

**Client Desktop:**
```bash
# All platforms
make client

# Individual platforms
make client-mac       # macOS (Intel + Apple Silicon)
make client-windows   # Windows
make client-linux     # Linux
```

**Client Mobile:**
```bash
# Android APK (for GitHub releases)
make android

# Android App Bundle (for Google Play)
make android-bundle

# iOS (opens Xcode for manual archive)
make ios
```

### Build Output Locations

**Server:**
```
server/build/bin/
├── robo-stream-server.app      # macOS universal
├── robo-stream-server.exe      # Windows
└── robo-stream-server          # Linux
```

**Client Desktop:**
```
client-web/electron-dist/
├── mac-arm64/Robo-Stream-Client.app
├── mac/Robo-Stream-Client.app
├── Robo-Stream-Client-*.exe
└── Robo-Stream-Client-*.AppImage
```

**Client Mobile:**
```
client-web/android/app/build/outputs/
├── apk/release/app-release.apk          # For GitHub/sideload
└── bundle/release/app-release.aab       # For Play Store

client-web/ios/build/
└── App.ipa                              # For TestFlight/App Store
```

**Packaged Release:**
```
releases/v1.0.0/
├── robo-stream-server-macos-universal.zip
├── robo-stream-server-windows-amd64.exe
├── robo-stream-server-linux-amd64
├── robo-stream-client-macos-arm64.zip
├── robo-stream-client-macos-intel.zip
├── robo-stream-client-windows-amd64.exe
├── robo-stream-client-linux-amd64.AppImage
├── robo-stream-client-v1.0.0.apk
└── robo-stream-client-v1.0.0.ipa
```

---

## Creating GitHub Releases

### Manual Process

**1. Create Tag (if not already done):**

```bash
git tag v1.0.0
git push origin v1.0.0
```

**2. Create Release on GitHub:**

1. Go to: https://github.com/robomon1/robo-stream/releases/new
2. Click **"Draft a new release"**
3. **Choose a tag:** Select `v1.0.0` (or create new)
4. **Release title:** `Robo-Stream v1.0.0`
5. **Describe this release:**

```markdown
## What's New

### Server v1.1.0
- Added automatic reconnection to OBS
- Improved WebSocket stability
- Fixed memory leak in long-running sessions
- Updated dependencies

### Client v1.0.1
- Fixed Android app orientation on startup
- Improved iOS landscape mode detection
- Added connection retry logic
- Updated button styling

### Bug Fixes
- Resolved issue with configuration switching (#12)
- Fixed rare crash on server disconnect (#15)

### Known Issues
- Configuration export not yet implemented (#18)

## Downloads

### Server
Download the server for your platform and run on your streaming computer:
- **macOS (Universal):** `robo-stream-server-macos-universal.zip` (Intel + Apple Silicon)
- **Windows:** `robo-stream-server-windows-amd64.exe`
- **Linux:** `robo-stream-server-linux-amd64`

### Client Desktop
Download the client for your platform:
- **macOS (Apple Silicon):** `robo-stream-client-macos-arm64.zip` (M1/M2/M3/M4)
- **macOS (Intel):** `robo-stream-client-macos-intel.zip`
- **Windows:** `robo-stream-client-windows-amd64.exe`
- **Linux:** `robo-stream-client-linux-amd64.AppImage`

### Client Mobile
- **Android:** Available on [Google Play](https://play.google.com/store/apps/details?id=com.robostream.robostreamclient) or download APK for sideload
- **iOS:** Available on [App Store](https://apps.apple.com/app/robo-stream-client)

## Installation

### Server
1. Download for your platform
2. Extract (macOS) or run (Windows/Linux)
3. Start OBS Studio first
4. Launch Robo-Stream Server
5. Note the server URL displayed

### Client
1. Download and install for your platform
2. Launch the app
3. Enter your server URL
4. Start controlling!

## Requirements
- OBS Studio 28.0+
- Server and client on same network
- WebSocket plugin enabled in OBS
```

6. **Upload files** from `releases/v1.0.0/`
7. **Set as latest release** (check box)
8. Click **"Publish release"**

### Release Notes Template

Keep release notes:
- **Clear** - Users understand what changed
- **Organized** - Server vs Client vs Bug fixes
- **User-focused** - What it means for them, not technical details
- **Linked** - Reference issues with `#123`

---

## Publishing to App Stores

### Google Play Store

#### Prerequisites

- Google Play Developer account ($25 one-time)
- Completed app listing (first release only)
- Signed Android App Bundle (.aab)

#### Build App Bundle

```bash
make android-bundle
```

Output: `client-web/android/app/build/outputs/bundle/release/app-release.aab`

#### Upload to Play Console

**1. Go to Play Console:**
- https://play.google.com/console
- Select **Robo-Stream Client** (or create app if first release)

**2. Create Release:**
- Navigate to **Production** → **Releases**
- Click **"Create new release"**

**3. Upload AAB:**
- Upload `app-release.aab`
- Wait for processing (1-2 minutes)

**4. Release Notes:**
```
What's new in this version:

• Fixed orientation issue on Android devices
• Improved connection stability
• Updated button layouts
• Bug fixes and performance improvements
```

**5. Review and Roll Out:**
- Click **"Review release"**
- Check everything looks correct
- Click **"Start rollout to Production"**

**6. Wait for Review:**
- First release: 1-7 days
- Updates: 1-3 days
- You'll receive email notifications

#### First-Time App Setup

If this is your first release, complete these sections first:

**Store Listing:**
- App name: Robo-Stream Client
- Short description (80 chars): "Control your OBS streams from anywhere"
- Full description: See template in [PUBLISH_ANDROID_PLAY_STORE.md]
- Screenshots: See [Screenshots Guide](#screenshots-guide)
- App icon: 512×512 PNG
- Feature graphic: 1024×500 PNG

**Store Settings:**
- App category: Utilities
- Content rating: Complete questionnaire (likely Everyone)
- Target audience: 13+
- Privacy policy: Required if collecting data

**Data Safety:**
- Fill out questionnaire about data collection
- For local-only app: "No data collection"

See detailed guide: [PUBLISH_ANDROID_PLAY_STORE.md](../PUBLISH_ANDROID_PLAY_STORE.md)

---

### Apple App Store

#### Prerequisites

- Apple Developer Program membership ($99/year)
- Xcode installed
- App created in App Store Connect (first release only)
- Completed app listing (first release only)

#### Build and Archive

**1. Open Project:**
```bash
cd client-web
npx cap open ios
```

**2. In Xcode:**
- Select **"Any iOS Device (arm64)"** as destination (not simulator)
- **Product** → **Clean Build Folder** (Shift+Cmd+K)
- **Product** → **Archive**
- Wait for archive to complete (2-5 minutes)

**3. Upload to App Store Connect:**
- Organizer window opens automatically
- Select your archive
- Click **"Distribute App"**
- Select **"App Store Connect"**
- Click **"Next"** through remaining screens
- Click **"Upload"**
- Wait for upload (5-20 minutes)

#### Submit for Review

**1. Go to App Store Connect:**
- https://appstoreconnect.apple.com/
- Select **Robo-Stream Client**

**2. Wait for Processing:**
- Go to **TestFlight** tab
- Build will show "Processing" (10-60 minutes)
- Wait until status is "Ready to Submit"

**3. Add Build to App Store:**
- Go to **App Store** tab
- Under **Build** section, click **"+ "** or **"Select a build"**
- Choose your uploaded build
- Click **"Done"**

**4. What's New in This Version:**
```
Bug fixes and improvements:

• Fixed landscape orientation on iPad
• Improved connection reliability
• Updated UI for better usability
• Performance enhancements
```

**5. Submit for Review:**
- Click **"Add for Review"**
- Answer export compliance: "No" (uses standard HTTPS encryption)
- Add reviewer notes if needed:
```
This app requires Robo-Stream Server running on the same network.

For testing, the app will show a settings screen where any URL can be entered.
Server software is available at: https://github.com/robomon1/robo-stream

No login required - app connects to local server only.
```
- Click **"Submit for Review"**

**6. Wait for Review:**
- Typical review time: 1-3 days
- You'll receive email notifications
- Status visible in App Store Connect

#### First-Time App Setup

**Create App:**
1. App Store Connect → **Apps** → **+**
2. Platform: iOS
3. Name: Robo-Stream Client
4. Primary Language: English (U.S.)
5. Bundle ID: `com.robostream.robostreamclient`
6. SKU: `robostream-client-1`

**App Information:**
- Name: Robo-Stream Client
- Subtitle: Stream Control Made Easy
- Category: Utilities
- Privacy Policy URL: Required if collecting data
- Description: See template in [PUBLISH_IOS_APP_STORE.md]
- Keywords: streaming, obs, remote, control
- Screenshots: See [Screenshots Guide](#screenshots-guide)

**Age Rating:**
- Complete questionnaire
- Result: Likely 4+ (suitable for all ages)

See detailed guide: [PUBLISH_IOS_APP_STORE.md](../PUBLISH_IOS_APP_STORE.md)

---

## Screenshots Guide

### Required Screenshots

#### Google Play Console

**Phone Screenshots** (REQUIRED):
- Minimum: 2 screenshots
- Maximum: 8 screenshots
- Dimensions: 1920×1080 (16:9 landscape) recommended
- Format: PNG or JPG

**Also required:**
- App icon: 512×512 PNG
- Feature graphic: 1024×500 PNG

#### Apple App Store

**iPhone 6.7" Display** (REQUIRED):
- Dimensions: 2796×1290 pixels (landscape)
- Minimum: 3 screenshots
- Maximum: 10 screenshots
- Format: PNG or JPG

**iPad 12.9"** (REQUIRED if iPad supported):
- Dimensions: 2732×2048 pixels (landscape)
- Minimum: 3 screenshots
- Format: PNG or JPG

**Also required:**
- App icon: 1024×1024 PNG

### Capturing Screenshots

**Android:**
```bash
# Open in Android Studio
cd client-web
npx cap open android

# Select Pixel 7 Pro or similar
# Run app (Cmd+R or click Run)
# Navigate through app
# Press Cmd+S to save screenshot
```

**iOS:**
```bash
# Open in Xcode
cd client-web
npx cap open ios

# Select iPhone 15 Plus (6.7")
# Run app (Cmd+R)
# Navigate through app
# Press Cmd+S to save screenshot
# Screenshots saved to Desktop
```

### Recommended Screenshots

Capture 5-6 screenshots showing:

1. **Connection Screen** - Settings with server URL input
2. **Main Interface** - Button grid with controls
3. **Active State** - Buttons highlighted or in use
4. **Configuration** - Multiple configs or settings
5. **Status Display** - Stream status indicators
6. **Key Feature** - Scene switching or unique feature

### Screenshot Best Practices

- **Use landscape orientation** (your app is landscape-focused)
- **Show real functionality** (actual buttons, real UI)
- **Keep it clean** (remove test data, use realistic content)
- **Consistent branding** (same colors, fonts)
- **Optional:** Add device frames with tools like Figma or Shotbot

### Quick Checklist

- [ ] Captured on correct device sizes
- [ ] Landscape orientation
- [ ] Shows key features
- [ ] Clean, professional appearance
- [ ] No test/debug data visible
- [ ] Meets dimension requirements

---

## Post-Release

### Announce Release

**Update Documentation:**
```bash
# Update README with new version
git commit -m "Update README for v1.0.0 release"
git push
```

**Announce on:**
- GitHub Discussions
- Social media (Twitter/X, Reddit, Discord)
- Project website
- Email newsletter (if applicable)

### Monitor Release

**GitHub:**
- Watch for issues related to new release
- Monitor download counts
- Respond to questions in Discussions

**Play Console:**
- Monitor crash reports
- Check user reviews
- Respond to reviews
- Track installation metrics

**App Store Connect:**
- Monitor crash reports (Xcode Organizer)
- Check reviews and ratings
- Respond to reviews
- Track downloads and metrics

**TestFlight (iOS beta):**
- Invite beta testers for next version
- Gather feedback before App Store submission

### Handle Issues

**Critical Bugs:**
1. Fix immediately
2. Increment patch version (1.0.1 → 1.0.2)
3. Build and release hotfix
4. Fast-track through stores if critical

**Minor Issues:**
1. Log in GitHub Issues
2. Plan for next minor release
3. Communicate timeline to users

### Plan Next Release

**Track:**
- Feature requests
- Bug reports
- User feedback
- Metrics and analytics

**Schedule:**
- Bug fix releases: As needed
- Minor releases: Every 2-4 weeks
- Major releases: Every 2-3 months

---

## Troubleshooting Releases

### Android

**"Version code already used"**
- Solution: Increment BUILD number
- `make set-client-version VERSION=1.0.1 BUILD=3`

**"Upload failed: invalid signature"**
- Solution: Check keystore.properties is correct
- Verify keystore file exists
- Ensure passwords are correct

**"App Bundle contains code targeting API level X"**
- Solution: Update `targetSdk` in variables.gradle
- Current requirement: API 35+

### iOS

**"No signing certificate"**
- Solution: Enable automatic signing in Xcode
- Or manually create Distribution certificate

**"Build processing failed"**
- Solution: Check for invalid assets or configurations
- Review build logs in App Store Connect

**"Missing compliance information"**
- Solution: Answer export compliance
- For standard HTTPS: Select "No"

### Release Build Issues

**Electron hangs on signing**
- Solution: Already handled in Makefile with `CSC_IDENTITY_AUTO_DISCOVERY=false`

**Wails build fails**
- Solution: Check Docker is running (for cross-compilation)
- Ensure all dependencies installed

**Android Gradle errors**
- Solution: `cd android && ./gradlew clean`
- Update Gradle version if needed

---

## Quick Reference

### Version Bump and Release

```bash
# 1. Update versions
make set-client-version VERSION=1.0.1 BUILD=2
make set-server-version VERSION=1.1.0

# 2. Commit
git add -A
git commit -m "Bump to client v1.0.1 and server v1.1.0"
git tag client-v1.0.1
git tag server-v1.1.0
git push origin main --tags

# 3. Build
make release VERSION=1.0.1

# 4. Create GitHub release (manual via web UI)

# 5. Upload to stores
make android-bundle  # → Upload to Play Console
make ios            # → Archive in Xcode → Upload to App Store Connect
```

### Store Submission Checklist

**Google Play:**
- [ ] Built signed AAB
- [ ] Incremented versionCode
- [ ] Updated release notes
- [ ] Uploaded to Production track
- [ ] Started rollout

**Apple App Store:**
- [ ] Archived in Xcode
- [ ] Uploaded to App Store Connect
- [ ] Build finished processing
- [ ] Added to App Store version
- [ ] Updated "What's New"
- [ ] Submitted for review

---

## Additional Resources

- [PUBLISH_ANDROID_PLAY_STORE.md](../PUBLISH_ANDROID_PLAY_STORE.md) - Detailed Android guide
- [PUBLISH_IOS_APP_STORE.md](../PUBLISH_IOS_APP_STORE.md) - Detailed iOS guide
- [Google Play Console Help](https://support.google.com/googleplay/android-developer)
- [App Store Connect Help](https://developer.apple.com/help/app-store-connect/)
- [Semantic Versioning](https://semver.org/)

---

Need help? [Open an issue](https://github.com/robomon1/robo-stream/issues) or ask in [Discussions](https://github.com/robomon1/robo-stream/discussions)!
