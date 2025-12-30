# Icon Transparency Fix

Fixed icon conversion scripts to preserve transparency from your PNG source.

## ✅ What Was Fixed

### Problem
Scripts were adding white background when converting `robo_stream.png` (transparent) to app icons.

### Solution
Both scripts now:
1. **Preserve RGBA mode** - Keeps alpha channel (transparency)
2. **Use `-background none`** (ImageMagick) or keep RGBA (Pillow)
3. **Verify transparency** after conversion

## 🔧 Updated Scripts

### Python Script (`setup-icons.py`)
**Changes:**
```python
# Before
img = img.convert('RGBA')  # Might add white bg

# After  
if img.mode != 'RGBA':
    img = img.convert('RGBA')  # Only convert if needed
else:
    print("✅ Source already has transparency (RGBA)")

# Save with optimization but preserve alpha
icon.save("icon.png", "PNG", optimize=True)
```

### Bash Script (`setup-icons.sh`)
**Changes:**
```bash
# Before
convert source.png -resize 1024x1024 output.png

# After
convert source.png -background none -resize 1024x1024 output.png
```

The `-background none` tells ImageMagick to preserve transparency.

## 📦 How to Use

### 1. Place Your PNG
```bash
# Your PNG should have transparent background
cp ~/robo_stream.png ~/git/robo-stream/
```

### 2. Convert with Updated Scripts

**Option A: Python (Recommended)**
```bash
cd ~/git/robo-stream/client
./setup-icons.py
```

**Option B: Bash**
```bash
cd ~/git/robo-stream/client
./setup-icons.sh
```

### 3. Verify Transparency Preserved

**Check with Python:**
```bash
python3 << EOF
from PIL import Image
img = Image.open('build/appicon.png')
print(f"Mode: {img.mode}")
if img.mode == 'RGBA':
    print("✅ Transparency preserved!")
else:
    print("❌ No transparency - mode should be RGBA")
EOF
```

**Check with ImageMagick:**
```bash
identify -format "%[channels]\n" build/appicon.png
# Should output: srgba (includes alpha channel)
```

**Check with Preview (macOS):**
```bash
open build/appicon.png
# In Preview, you should see checkerboard pattern in transparent areas
```

## 🧪 Testing

After conversion, check your icons:

### 1. Check the Files
```bash
cd ~/git/robo-stream/client

# Should see RGBA mode
file build/appicon.png
# Output: PNG image data, 1024 x 1024, 8-bit/color RGBA

# Check all icons
file build/darwin/icon.iconset/*.png
file build/linux/icon.png
```

### 2. Visual Check (macOS)
```bash
# Open in Preview - should see checkerboard in transparent areas
open build/appicon.png
open build/linux/icon.png
```

### 3. Build and Check App
```bash
wails build
open build/bin/Robo-Stream\ Client.app

# The icon should have transparent background
# Not white background
```

## 🔍 Troubleshooting

### Still Seeing White Background?

**1. Check source file:**
```bash
# Verify your source PNG has transparency
python3 << EOF
from PIL import Image
img = Image.open('../robo_stream.png')
print(f"Source mode: {img.mode}")
if img.mode == 'RGBA':
    print("✅ Source has transparency")
    # Check if it actually uses transparency
    if img.getextrema()[3][0] < 255:
        print("✅ Source has transparent pixels")
    else:
        print("⚠️  Source is RGBA but has no transparent pixels")
else:
    print("❌ Source has no transparency")
EOF
```

**2. Re-export your source:**
If your `robo_stream.png` doesn't have transparency:

**Using Preview (macOS):**
1. Open original image
2. Select all (Cmd+A)
3. Copy (Cmd+C)
4. File → New from Clipboard
5. File → Export → Format: PNG
6. Save as `robo_stream.png`

**Using GIMP:**
1. Image → Mode → RGB
2. Layer → Transparency → Add Alpha Channel
3. Use Magic Wand to select background
4. Press Delete
5. File → Export As → robo_stream.png

**Using Photoshop:**
1. Select background with Magic Wand
2. Delete
3. File → Export → Export As → PNG
4. Enable Transparency

**3. Check ImageMagick version:**
```bash
convert -version
# Make sure it's recent (7.x)

# If old version, upgrade:
brew upgrade imagemagick  # macOS
sudo apt upgrade imagemagick  # Linux
```

**4. Force transparency in conversion:**

If still having issues, use this manual command:

```bash
cd ~/git/robo-stream/client

# Explicitly preserve transparency
convert ../robo_stream.png \
    -background none \
    -alpha set \
    -channel RGBA \
    -resize 1024x1024 \
    -gravity center \
    -extent 1024x1024 \
    build/appicon.png
```

### Icons Show Transparent in Preview but Not in App?

This might be a Wails/macOS issue:

**1. Check wails.json:**
Make sure icon path is correct:
```json
"info": {
    "icon": "build/appicon.png"
}
```

**2. Clean and rebuild:**
```bash
wails build -clean
```

**3. Check app bundle:**
```bash
# Check if icon is embedded correctly
open build/bin/Robo-Stream\ Client.app/Contents/Resources/
# Should see iconfile.icns
```

## 🎨 Alternative: Add Custom Background

If you want a specific background color instead of transparent:

### Python
```python
# Create new image with colored background
from PIL import Image

# Load transparent icon
icon = Image.open('robo_stream.png')

# Create background
bg_color = (10, 10, 10)  # Dark gray RGB
background = Image.new('RGB', icon.size, bg_color)

# Paste icon on background (using icon as mask)
background.paste(icon, (0, 0), icon)

# Save
background.save('robo_stream_bg.png')
```

### ImageMagick
```bash
# Add dark background
convert robo_stream.png \
    -background '#0a0a0a' \
    -alpha remove \
    -alpha off \
    robo_stream_bg.png
```

## 📊 Before & After

### Before (White Background Added)
```
robo_stream.png (transparent) 
    ↓ [conversion]
appicon.png (white background) ❌
```

### After (Transparency Preserved)
```
robo_stream.png (transparent)
    ↓ [conversion with -background none]
appicon.png (transparent) ✅
```

## ✅ Verification Checklist

- [ ] Source PNG has transparent background
- [ ] Ran updated `setup-icons.py` or `setup-icons.sh`
- [ ] Verified output with `identify` or Preview
- [ ] Rebuilt app with `wails build -clean`
- [ ] Icon shows transparent background in app

## 🚀 Quick Fix

**If in doubt, just run this:**

```bash
cd ~/git/robo-stream/client

# Make sure you have robo_stream.png with transparency in parent dir
ls -l ../robo_stream.png

# Run conversion
./setup-icons.py

# Check it worked
python3 << 'EOF'
from PIL import Image
img = Image.open('build/appicon.png')
if img.mode == 'RGBA':
    print("✅ SUCCESS - Transparency preserved!")
else:
    print("❌ FAILED - No transparency")
EOF

# Rebuild
wails build -clean

# Test
open build/bin/Robo-Stream\ Client.app
```

Your transparent background should now be preserved! 🎨✨
