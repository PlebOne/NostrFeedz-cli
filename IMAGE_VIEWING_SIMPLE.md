# Image Viewing - Simple External Viewer Approach

## Overview
Due to terminal graphics protocol reliability issues, NostrFeedz CLI now uses **external image viewers** for the best experience.

## How It Works

### Viewing Single Image:
1. Read an article
2. See "📷 1 image - Press 'i' to view"
3. Press **`i`** → Image opens in lightweight viewer
4. Press **`q`** in the viewer → Image closes
5. Continue reading article

### Viewing Multiple Images:
1. Read an article with multiple images
2. See "📷 3 images [1/3] - ← → to select, 'i' to view"
3. Press **`←`** or **`→`** → Navigate between images
4. Press **`i`** → View current image in viewer
5. Press **`q`** → Close viewer
6. Press **`←`** or **`→`** → Select next/previous image
7. Press **`i`** again → View next image

Images wrap around - after the last image, ← goes to first.

### Recommended Image Viewers:

#### Best Options (Lightweight & Fast):
1. **sxiv** - Simple X Image Viewer
   ```bash
   sudo apt install sxiv  # Debian/Ubuntu
   sudo pacman -S sxiv    # Arch
   ```
   - Super lightweight
   - Press 'q' to close
   - Borderless mode

2. **feh** - Fast image viewer
   ```bash
   sudo apt install feh
   ```
   - Auto-zoom to fit
   - Borderless window
   - Press 'q' to close

3. **imv** - Wayland image viewer
   ```bash
   sudo apt install imv
   ```
   - Modern Wayland support
   - Lightweight and fast

#### Fallback Options:
- **eog** - GNOME Eye of GNOME (pre-installed on GNOME)
- **eom** - MATE Eye of MATE
- **xdg-open** - System default

## Keyboard Shortcuts

### In Article:
- **i** - View current image
- **← →** - Navigate between images (when multiple)
- **v** - Play video in mpv
- **o** - Open article in browser
- **ESC** - Back to articles list
- **↑↓** - Scroll article
- **Space** - Page down

### In Image Viewer:
- **q** - Close viewer (most viewers)
- **ESC** - Close viewer (most viewers)
- **Click X** - Close window

## Why External Viewers?

### Advantages:
✅ **Reliable** - No terminal protocol issues
✅ **Fast** - Dedicated image rendering
✅ **Universal** - Works in any terminal
✅ **Familiar** - Standard viewer controls
✅ **Quality** - Full resolution display
✅ **Multi-monitor** - Can move to different screen
✅ **Cached** - Images load from local cache instantly

### How Caching Works:
1. **Background Download** - Images download when article loads
2. **Local Storage** - Cached at `~/.config/nostrfeedz/cache/images/`
3. **Instant Display** - Viewer opens cached file immediately
4. **No Re-download** - Once cached, always fast

### Installation:

```bash
# Install sxiv (recommended)
sudo apt install sxiv      # Debian/Ubuntu
sudo dnf install sxiv      # Fedora
sudo pacman -S sxiv        # Arch

# Or feh
sudo apt install feh

# Or imv (for Wayland)
sudo apt install imv
```

## Configuration

The app tries viewers in this order:
1. sxiv (with borderless, 800x600)
2. feh (with auto-zoom, borderless)
3. imv
4. eog (GNOME default)
5. eom (MATE default)
6. xdg-open (system default)

First available viewer is used automatically.

## Tips

### sxiv Tips:
- Press `q` - Close
- Press `+`/`-` - Zoom in/out
- Mouse wheel - Zoom
- Drag - Pan image

### feh Tips:
- Press `q` - Close
- Press `+`/`-` - Zoom
- Mouse wheel - Zoom
- Right-click drag - Pan

### Multi-Image Articles:
Articles with multiple images show selection:
- "📷 5 images [2/5]" - Shows you're on image 2 of 5
- Press ← → to cycle through
- Press 'i' to view current selection
- Selection wraps around (after last goes to first)

Example workflow:
1. Article has 4 images
2. Press → → → to get to image 3
3. Press 'i' to view image 3
4. Close viewer with 'q'
5. Press → to select image 4
6. Press 'i' to view image 4

## Future Plans

- [x] Multiple image navigation with arrow keys
- [ ] Slideshow mode (auto-advance through images)
- [ ] Gallery view (thumbnails)
- [ ] Configurable viewer preference
- [ ] Thumbnail previews in article list
- [ ] Optional inline images (when more reliable)
