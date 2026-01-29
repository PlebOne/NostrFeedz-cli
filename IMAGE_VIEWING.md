# Image Viewing in NostrFeedz CLI

## Overview
NostrFeedz CLI supports inline image display in compatible terminals using Kitty, Sixel, or iTerm2 protocols.

## Supported Terminals
- **Kitty** - Full support (best experience)
- **WezTerm** - Via Kitty protocol
- **iTerm2** - Native inline images (macOS)
- **xterm, mlterm, mintty** - Via Sixel protocol
- **Others** - Fallback to external viewers

## How to View Images

### In an Article:
1. Navigate to an article with images (you'll see "📷 N image(s)" at the bottom)
2. Press **`i`** - Toggle image viewer (show/hide first image)
3. Press **`i`** again - Close image and return to article
4. Press **`ESC`** - Go back to articles list (closes image if open)
5. Press **`I`** (Shift+i) - Open in external viewer (feh/eog/xdg-open)

### Image Caching:
- Images are **automatically downloaded in the background** when articles are fetched
- Cached at: `~/.config/nostrfeedz/cache/images/`
- **Instant loading** - no wait when you press 'i'
- **500MB cache limit** - oldest images auto-removed
- **30-day expiration** - for non-favorited articles
- **Favorited articles** - images kept forever

## Terminal Setup

### Kitty Users:
Make sure your `TERM` variable is set correctly:
```bash
export TERM=xterm-kitty
./nostrfeedz
```

Or add to your `~/.bashrc` or `~/.zshrc`:
```bash
export TERM=xterm-kitty
```

### Testing Terminal Support:
Run this to check what protocols your terminal supports:
```bash
echo $TERM
# Should show: xterm-kitty, xterm-256color, etc.
```

## Keyboard Shortcuts

### Article View:
- **i** - Toggle image viewer (show/hide)
- **I** - Force external viewer
- **v** - Play video in mpv
- **o** - Open article in browser
- **ESC** - Go back to articles list

### Image Viewer:
- **i** - Close image (toggle off)
- **ESC** - Go back to articles list
- **I** - Open in external viewer instead

## Troubleshooting

### Images don't show inline:
1. Check your terminal supports inline images (see supported list above)
2. Verify `TERM` variable: `echo $TERM`
3. Try setting: `export TERM=xterm-kitty` (for Kitty)
4. Images will fall back to external viewer automatically

### ESC doesn't close image:
- Use **`i`** to toggle the image viewer on/off
- **`ESC`** goes back to the articles list (and will close the image as a side effect)
- Check that you're in the image viewer (you should see "IMAGE VIEWER" header)

### Images load slowly:
- First load downloads and caches
- Subsequent loads are instant from cache
- Check cache at: `~/.config/nostrfeedz/cache/images/`

### Cache is too large:
Cache auto-manages itself:
- Max 500MB (configurable in future)
- Auto-removes oldest images
- Favorited articles protected

## External Viewers

If inline images don't work, the app will try these viewers in order:
1. **feh** - Lightweight image viewer
2. **eog** - GNOME image viewer  
3. **xdg-open** - System default

For videos, it uses **mpv**.

## Future Enhancements
- [ ] Multiple image viewing (arrow keys to navigate)
- [ ] Zoom in/out
- [ ] Image slideshow mode
- [ ] Configurable cache limits
- [ ] Manual cache cleanup command

## Technical Details

### Terminal Image Protocols
The app uses terminal graphics protocols to display images inline:
- **Kitty Graphics Protocol** - Best quality, most features
- **Sixel** - Older but widely supported
- **iTerm2 Inline Images** - macOS specific

### Image Persistence Issue
Terminal graphics protocols render directly to the terminal buffer and persist even after the application updates its view. When closing an image:

1. **termimg.ClearAll()** - Sends protocol-specific clear commands to terminal
2. **tea.ClearScreen** - Forces Bubble Tea to redraw the entire screen
3. Both steps ensure images are completely removed from display

### Cache Implementation
- **Location**: `~/.config/nostrfeedz/cache/images/`
- **Naming**: SHA-256 hash of image URL
- **Background Downloads**: Goroutines download images without blocking UI
- **Cache-First**: Always checks cache before network request
- **Cleanup**: Runs on startup and periodically during runtime
