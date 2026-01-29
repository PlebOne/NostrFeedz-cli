# Video Player Guide

NostrFeedz CLI supports playing videos from articles (including YouTube feeds) using external video players.

## Features

### Automatic Player Detection
The app tries multiple video players in order of preference:
1. **mpv** - Lightweight, powerful, supports YouTube streaming (recommended)
2. **vlc** - Popular cross-platform player
3. **mplayer** - Classic player
4. **xdg-open** - System default

### Video Metadata Display
When viewing an article with videos, you'll see:
- **Video count** and current selection
- **Video title** (extracted from YouTube embeds when available)
- **Video URL** for reference
- **Navigation hints**

### Keyboard Shortcuts

#### Single Video
- `v` - Play the video

#### Multiple Videos
- `v` - Play currently selected video
- `Shift+←` or `H` - Previous video (auto-plays)
- `Shift+→` or `L` - Next video (auto-plays)

Navigation wraps around (last → first, first → last).

## Installation

Install at least one video player:

### mpv (Recommended)
```bash
# Ubuntu/Debian
sudo apt install mpv

# Fedora
sudo dnf install mpv

# Arch
sudo pacman -S mpv
```

### VLC
```bash
# Ubuntu/Debian
sudo apt install vlc

# Fedora
sudo dnf install vlc

# Arch
sudo pacman -S vlc
```

## Supported Video Sources
- **YouTube** (most common in RSS/Atom feeds)
  - Regular URLs: `youtube.com/watch?v=...`
  - Short URLs: `youtu.be/...`
  - Embedded players with titles extracted
- **Vimeo**
- **Direct video files** (.mp4, .webm)

## Tiling Window Manager Support

The app automatically:
- **Closes previous video** before opening next one
- **Tracks player PID** to prevent multiple windows
- Works seamlessly with rapid navigation

Just press `Shift+←` or `Shift+→` to cycle through videos without accumulating windows!

## Example Usage

1. **Browse feeds** and select an article
2. **Press Enter** to open the article
3. **Scroll to the bottom** to see media section
4. **Press `v`** to play the first video
5. If multiple videos:
   - **Press `Shift+→`** to play the next video
   - **Press `Shift+←`** to play the previous video

The media section shows:
```
─── Media ───
🎬 3 videos [2/3] - Shift+← Shift+→ to select, 'v' to play
  Title: How to Build a CLI App
  URL: https://www.youtube.com/watch?v=abc123
```

## Troubleshooting

### No video players found
Install mpv or vlc as shown above.

### Player opens but no video plays
- Check internet connection for streaming URLs
- Verify the URL is valid
- Try a different player

### Multiple windows opening
- Make sure you're using the latest version
- The app should automatically close the previous player
- Check debug output for PID tracking

### YouTube videos don't play
Most video players (especially mpv) handle YouTube URLs directly. If issues occur:
- Update your player: `sudo apt update && sudo apt upgrade mpv`
- Install youtube-dl or yt-dlp: `pip install yt-dlp`
- mpv will use these automatically

## Advanced

### Custom Player Preferences
Edit `internal/app/handlers.go` and reorder the players array in `openVideo()`:

```go
players := []struct {
    cmd  string
    args []string
}{
    {"your-preferred-player", []string{url}},
    {"mpv", []string{"--geometry=800x600", url}},
    // ...
}
```

### Player Arguments
Current defaults:
- **mpv**: `--geometry=800x600` (sized window)
- **vlc**: `--width=800 --height=600` (sized window)
- **mplayer**: Default settings

Modify in `openVideo()` function to customize.
