# Robo-Stream Client UI Improvements

Visual improvements for 8" touchscreen display.

## Changes Made

### 1. ✅ Smaller Title
- Reduced from 1.8em to 1.2em
- More space for button grid
- Better for small touchscreens

### 2. ✅ Status Bar Removed
- No more "Streaming Recording Scene" section at top
- Saves valuable screen space
- Status now shown directly on buttons

### 3. ✅ Button State Indicators
Active buttons now show a **red pulsing dot** in the top-right corner:

**When Active:**
- 🔴 Streaming button (when stream is on)
- 🔴 Recording button (when recording)
- 🔴 Scene buttons (when that scene is current)

**Visual Effects:**
- Red dot pulses (breathing animation)
- Button has red glow around edges
- Instantly shows active state

## Current Button States

### Streaming Actions
```javascript
// These buttons show red dot when streaming:
- toggle_stream
- start_stream  
- stop_stream
```

### Recording Actions
```javascript
// These buttons show red dot when recording:
- toggle_record
- start_record
- stop_record
```

### Scene Actions
```javascript
// Scene buttons show red dot when that scene is active:
- switch_scene (compares button's scene_name to current_scene)
```

## Alternative Visual Indicator Options

You can choose from these alternatives by editing `frontend/css/style.css`:

### Option 1: Red Dot (Current - Recommended for Touchscreen)
**Pros:**
- ✅ Very visible
- ✅ Doesn't obstruct button text
- ✅ Works on any button color
- ✅ Clear on/off state

**Cons:**
- ❌ Small on low DPI screens

**CSS:**
```css
.deck-button.active::before {
    content: '';
    position: absolute;
    top: 6px;
    right: 6px;
    width: 12px;
    height: 12px;
    background: #e74c3c;
    border-radius: 50%;
    box-shadow: 0 0 8px rgba(231, 76, 60, 0.8);
    animation: pulse-dot 2s infinite;
}
```

### Option 2: Border Glow
**Pros:**
- ✅ Surrounds entire button
- ✅ Very noticeable
- ✅ Professional look

**Cons:**
- ❌ May clash with button colors

**CSS:**
```css
.deck-button.active {
    border: 3px solid #e74c3c;
    box-shadow: 0 0 20px rgba(231, 76, 60, 0.8);
}
```

### Option 3: Corner Triangle
**Pros:**
- ✅ Larger than dot
- ✅ Distinctive shape

**Cons:**
- ❌ May cover button text

**CSS:**
```css
.deck-button.active::before {
    content: '';
    position: absolute;
    top: 0;
    right: 0;
    width: 0;
    height: 0;
    border-style: solid;
    border-width: 0 20px 20px 0;
    border-color: transparent #e74c3c transparent transparent;
}
```

### Option 4: Brightness Increase
**Pros:**
- ✅ Subtle
- ✅ Works with button color

**Cons:**
- ❌ Less obvious
- ❌ Harder to see on bright colors

**CSS:**
```css
.deck-button.active {
    filter: brightness(1.3);
}
```

### Option 5: Background Pulse
**Pros:**
- ✅ Entire button pulses
- ✅ Very noticeable

**Cons:**
- ❌ May be distracting
- ❌ Affects button color

**CSS:**
```css
.deck-button.active {
    animation: button-pulse 2s infinite;
}

@keyframes button-pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.7; }
}
```

### Option 6: Checkmark Icon
**Pros:**
- ✅ Clear "active" meaning
- ✅ Professional

**Cons:**
- ❌ Takes more space
- ❌ May confuse with "success"

**CSS:**
```css
.deck-button.active::before {
    content: '✓';
    position: absolute;
    top: 4px;
    right: 6px;
    font-size: 16px;
    color: #27ae60;
    font-weight: bold;
}
```

### Option 7: LED Bar (Bottom)
**Pros:**
- ✅ Doesn't obscure text
- ✅ Like physical buttons

**Cons:**
- ❌ Smaller visibility

**CSS:**
```css
.deck-button.active::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 4px;
    background: #e74c3c;
    border-radius: 0 0 8px 8px;
}
```

### Option 8: Dual Indicator (Dot + Glow)
**Pros:**
- ✅ Most visible
- ✅ Combines best of both

**Cons:**
- ❌ May be too much

**This is the current default!**

## Customizing the Indicator

Edit `client/frontend/css/style.css`:

**Change dot color:**
```css
.deck-button.active::before {
    background: #27ae60;  /* Green instead of red */
}
```

**Change dot size:**
```css
.deck-button.active::before {
    width: 16px;   /* Larger */
    height: 16px;
}
```

**Change dot position:**
```css
.deck-button.active::before {
    top: 50%;      /* Center vertically */
    right: 4px;    /* Closer to edge */
}
```

**Remove pulsing:**
```css
.deck-button.active::before {
    animation: none;  /* Static dot */
}
```

**Change glow color:**
```css
.deck-button.active {
    box-shadow: 0 0 15px rgba(46, 204, 113, 0.6);  /* Green glow */
}
```

## Multi-State Indicators

For buttons that might have multiple states:

### Example: Recording with Pause
```css
/* Recording (red dot) */
.deck-button.recording::before {
    background: #e74c3c;
}

/* Paused (yellow dot) */
.deck-button.paused::before {
    background: #f39c12;
}
```

### Example: Scene Transitions
```css
/* Current scene (solid red) */
.deck-button.current-scene::before {
    background: #e74c3c;
}

/* Next scene (hollow red) */
.deck-button.next-scene::before {
    background: transparent;
    border: 2px solid #e74c3c;
}
```

## Recommendations

**For 8" Touchscreen:**
1. ✅ **Red Dot + Glow** (current) - Best visibility
2. **Border Glow** - Good alternative if dots are too small
3. **LED Bar** - Clean, minimal

**For Larger Displays:**
1. **Red Dot** - Clean and professional
2. **Border Glow** - More noticeable
3. **Brightness** - Subtle and elegant

**For High DPI (Retina):**
1. **Red Dot** - Crisp and clear
2. **Checkmark** - Professional look
3. **Corner Triangle** - Distinctive

## Testing Different Options

To test an option:

1. **Edit CSS file:**
   ```bash
   cd ~/git/robo-stream/client
   nano frontend/css/style.css
   ```

2. **Find `.deck-button.active` section** (around line 120)

3. **Replace with option CSS** from above

4. **Rebuild and test:**
   ```bash
   wails build
   ./build/bin/streampi-deck
   ```

5. **Toggle streaming/recording** to see the effect

## Current Settings

**Title Size:** 1.2em (was 1.8em)
**Status Bar:** Hidden
**Active Indicator:** Red pulsing dot (top-right) + glow
**Animation:** 2s pulse cycle

## Screen Space Saved

**Before:**
- Title: ~45px
- Status bar: ~40px  
- Total: ~85px

**After:**
- Title: ~30px
- Status bar: 0px
- Total: ~30px

**Result:** ~55px more space for buttons! 🎉

## Future Enhancements

Possible additions:
- **Battery indicator** (for portable setups)
- **Network status** (connection quality)
- **Time indicator** (stream duration)
- **CPU/Memory usage** (performance monitoring)
- **Custom button badges** (user-defined icons)

All can be added as small overlays like the current dot indicator!
