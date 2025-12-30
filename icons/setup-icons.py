#!/usr/bin/env python
"""
Convert robo-stream-client.png and robo-stream-server.png to all needed icon formats
Generates icons for both server and client applications
Alternative to ImageMagick - uses PIL/Pillow
"""

import os
import sys
from pathlib import Path

try:
    from PIL import Image  # type: ignore
except ImportError:
    print("❌ Error: Pillow (PIL) is not installed")
    print("Install with: pip3 install Pillow")
    sys.exit(1)

# Icon sources (expected in robo-stream/icons/)
CLIENT_ICON_SOURCE = "robo-stream-client.png"
SERVER_ICON_SOURCE = "robo-stream-server.png"

# Target directories (relative to robo-stream root)
CLIENT_BUILD_DIR = "../client/build"
SERVER_BUILD_DIR = "../server/build"

def process_icon(source_file, build_dir, app_name):
    """Convert a source icon to all required formats and sizes"""
    
    if not os.path.exists(source_file):
        print(f"❌ Error: {source_file} not found")
        return False
    
    print(f"\n🎨 Processing {app_name}...")
    print(f"📁 Source: {source_file}")
    print(f"📁 Target: {build_dir}")
    
    # Create build directories
    Path(build_dir).mkdir(parents=True, exist_ok=True)
    Path(f"{build_dir}/darwin").mkdir(exist_ok=True)
    Path(f"{build_dir}/windows").mkdir(exist_ok=True)
    Path(f"{build_dir}/linux").mkdir(exist_ok=True)
    
    # Load source image and ensure RGBA mode for transparency
    img = Image.open(source_file)
    print(f"📊 Source mode: {img.mode}, Size: {img.size}")
    
    # Convert to RGBA to preserve/add transparency
    if img.mode != 'RGBA':
        print("⚠️  Converting to RGBA to support transparency...")
        img = img.convert('RGBA')
    else:
        print("✅ Source already has transparency (RGBA)")
    
    # Create 1024x1024 PNG for Wails (preserve transparency)
    print("  Creating appicon.png (1024x1024)...")
    icon_1024 = img.resize((1024, 1024), Image.Resampling.LANCZOS)
    icon_1024.save(f"{build_dir}/appicon.png", "PNG", optimize=True)
    
    # Verify transparency preserved
    test_img = Image.open(f"{build_dir}/appicon.png")
    if test_img.mode == 'RGBA':
        print("    ✅ Transparency preserved in appicon.png")
    else:
        print(f"    ⚠️  Warning: Image mode is {test_img.mode}, expected RGBA")
    
    # Create Linux 512x512 PNG
    print("  Creating Linux icon...")
    icon_512 = img.resize((512, 512), Image.Resampling.LANCZOS)
    icon_512.save(f"{build_dir}/linux/icon.png", "PNG")
    
    # Create Windows ICO (multiple sizes)
    print("  Creating Windows icon...")
    icon_256 = img.resize((256, 256), Image.Resampling.LANCZOS)
    icon_256.save(
        f"{build_dir}/windows/icon.ico",
        format='ICO',
        sizes=[(16, 16), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]
    )
    
    # Create macOS icon sizes (for .icns - requires iconutil on macOS)
    print("  Creating macOS icon sizes...")
    iconset_dir = Path(f"{build_dir}/darwin/icon.iconset")
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
                '-o', f'{build_dir}/darwin/icon.icns'
            ], check=True, capture_output=True)
            print("    ✅ Created icon.icns")
            # Clean up iconset directory
            import shutil
            shutil.rmtree(iconset_dir)
        except subprocess.CalledProcessError as e:
            print(f"    ⚠️  Could not create .icns (iconutil failed): {e.stderr.decode()}")
        except FileNotFoundError:
            print("    ⚠️  iconutil not found (macOS required)")
    else:
        print("    ⚠️  Skipping .icns creation (requires macOS)")
    
    print("    ✅ Created icon.ico")
    print("    ✅ Created Linux icon.png")
    
    return True

def main():
    """Process both server and client icons"""
    
    print("=" * 60)
    print("🎨 Robo-Stream Icon Generator")
    print("=" * 60)
    
    # Check if we're in the icons directory
    if not os.path.exists(CLIENT_ICON_SOURCE) or not os.path.exists(SERVER_ICON_SOURCE):
        print("\n❌ Error: Icon files not found!")
        print(f"Expected files:")
        print(f"  - {CLIENT_ICON_SOURCE}")
        print(f"  - {SERVER_ICON_SOURCE}")
        print(f"\nCurrent directory: {os.getcwd()}")
        print("\nPlease run this script from the robo-stream/icons/ directory")
        sys.exit(1)
    
    success_count = 0
    
    # Process client icon
    if process_icon(CLIENT_ICON_SOURCE, CLIENT_BUILD_DIR, "Client"):
        success_count += 1
    
    # Process server icon
    if process_icon(SERVER_ICON_SOURCE, SERVER_BUILD_DIR, "Server"):
        success_count += 1
    
    # Summary
    print("\n" + "=" * 60)
    print(f"✅ Icon generation complete! ({success_count}/2 apps)")
    print("=" * 60)
    
    if success_count == 2:
        print("\n📦 Generated icons:")
        print("\n  CLIENT:")
        print(f"    - {CLIENT_BUILD_DIR}/appicon.png (1024x1024) - Wails auto-detect")
        if os.path.exists(f"{CLIENT_BUILD_DIR}/darwin/icon.icns"):
            print(f"    - {CLIENT_BUILD_DIR}/darwin/icon.icns - macOS")
        print(f"    - {CLIENT_BUILD_DIR}/windows/icon.ico - Windows")
        print(f"    - {CLIENT_BUILD_DIR}/linux/icon.png - Linux")
        
        print("\n  SERVER:")
        print(f"    - {SERVER_BUILD_DIR}/appicon.png (1024x1024) - Wails auto-detect")
        if os.path.exists(f"{SERVER_BUILD_DIR}/darwin/icon.icns"):
            print(f"    - {SERVER_BUILD_DIR}/darwin/icon.icns - macOS")
        print(f"    - {SERVER_BUILD_DIR}/windows/icon.ico - Windows")
        print(f"    - {SERVER_BUILD_DIR}/linux/icon.png - Linux")
        
        print("\n📝 Next steps:")
        print("  1. Ensure your main.go files reference the icons (see loadIcon() function)")
        print("  2. Run 'wails build' in server/ and client/ directories")
        print("  3. Icons will be embedded in the built applications")
    else:
        print("\n⚠️  Some icons failed to generate. Check errors above.")

if __name__ == "__main__":
    main()
