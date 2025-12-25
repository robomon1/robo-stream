#!/usr/bin/env python3
"""
Convert robo_stream.jpg to all needed icon formats
Alternative to ImageMagick - uses PIL/Pillow
"""

import os
import sys
from pathlib import Path

try:
    from PIL import Image
except ImportError:
    print("❌ Error: Pillow (PIL) is not installed")
    print("Install with: pip3 install Pillow")
    sys.exit(1)

ICON_SOURCE = "./robo_stream.jpg"
BUILD_DIR = "build"

def create_icon_sizes():
    """Convert the source icon to all required formats and sizes"""
    
    if not os.path.exists(ICON_SOURCE):
        print(f"❌ Error: {ICON_SOURCE} not found")
        print("Please place robo_stream.jpg in the current directory")
        sys.exit(1)
    
    print("🎨 Converting icon to all formats...")
    
    # Create build directories
    Path(BUILD_DIR).mkdir(exist_ok=True)
    Path(f"{BUILD_DIR}/darwin").mkdir(exist_ok=True)
    Path(f"{BUILD_DIR}/windows").mkdir(exist_ok=True)
    Path(f"{BUILD_DIR}/linux").mkdir(exist_ok=True)
    
    # Load source image
    img = Image.open(ICON_SOURCE)
    
    # Convert to RGBA if not already
    if img.mode != 'RGBA':
        img = img.convert('RGBA')
    
    # Create 1024x1024 PNG for Wails
    print("Creating appicon.png (1024x1024)...")
    icon_1024 = img.resize((1024, 1024), Image.Resampling.LANCZOS)
    icon_1024.save(f"{BUILD_DIR}/appicon.png", "PNG")
    
    # Create Linux 512x512 PNG
    print("Creating Linux icon...")
    icon_512 = img.resize((512, 512), Image.Resampling.LANCZOS)
    icon_512.save(f"{BUILD_DIR}/linux/icon.png", "PNG")
    
    # Create Windows ICO (multiple sizes)
    print("Creating Windows icon...")
    icon_256 = img.resize((256, 256), Image.Resampling.LANCZOS)
    icon_256.save(
        f"{BUILD_DIR}/windows/icon.ico",
        format='ICO',
        sizes=[(16, 16), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]
    )
    
    # Create macOS icon sizes (for .icns - requires iconutil on macOS)
    print("Creating macOS icon sizes...")
    iconset_dir = Path(f"{BUILD_DIR}/darwin/icon.iconset")
    iconset_dir.mkdir(exist_ok=True)
    
    # Generate all required macOS icon sizes
    sizes = [
        (16, "icon_16x16.png"),
        (32, "icon_16x16@2x.png"),
        (32, "icon_32x32.png"),
        (64, "icon_32x32@2x.png"),
        (128, "icon_128x128.png"),
        (256, "icon_128x128@2x.png"),
        (256, "icon_256x256.png"),
        (512, "icon_256x256@2x.png"),
        (512, "icon_512x512.png"),
        (1024, "icon_512x512@2x.png"),
    ]
    
    for size, filename in sizes:
        icon = img.resize((size, size), Image.Resampling.LANCZOS)
        icon.save(iconset_dir / filename, "PNG")
    
    # Try to create .icns on macOS
    if sys.platform == 'darwin':
        import subprocess
        try:
            subprocess.run([
                'iconutil', '-c', 'icns',
                str(iconset_dir),
                '-o', f'{BUILD_DIR}/darwin/icon.icns'
            ], check=True)
            print("✅ Created icon.icns")
            # Clean up iconset directory
            import shutil
            shutil.rmtree(iconset_dir)
        except subprocess.CalledProcessError:
            print("⚠️  Could not create .icns (iconutil failed)")
        except FileNotFoundError:
            print("⚠️  iconutil not found (macOS required)")
    else:
        print("⚠️  Skipping .icns creation (requires macOS)")
    
    print("✅ Created icon.ico")
    print("✅ Created Linux icon.png")
    
    print("")
    print("✅ Icon conversion complete!")
    print("")
    print("Created:")
    print(f"  - {BUILD_DIR}/appicon.png (1024x1024) - Wails auto-detect")
    if os.path.exists(f"{BUILD_DIR}/darwin/icon.icns"):
        print(f"  - {BUILD_DIR}/darwin/icon.icns - macOS")
    print(f"  - {BUILD_DIR}/windows/icon.ico - Windows")
    print(f"  - {BUILD_DIR}/linux/icon.png - Linux")
    print("")
    print("Next step: Run 'wails build' to use these icons")

if __name__ == "__main__":
    create_icon_sizes()
