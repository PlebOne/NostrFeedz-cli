# NostrFeedz CLI

A beautiful command-line interface for Nostr-Feedz - your RSS and Nostr feed reader with cross-device synchronization.

## Features

### 🔐 Authentication
- **Pleb_Signer** (NIP-55 via D-Bus) - Recommended for Linux desktop
- **NIP-46** remote signer support
- **nsec** private key (fallback)

### 📰 Feed Management
- RSS & Nostr Feeds - Subscribe to RSS feeds and Nostr long-form content (NIP-23)
- Cross-Device Sync - Sync subscriptions and read status via Nostr (kinds 30404, 30405)
- Tags & Categories - Organize feeds with tags and categories
- Uncategorized view - Find feeds without organization

### 🎨 Media Support
- **Images** - External viewer integration (sxiv, feh, imv, eog)
  - Image caching with background downloads
  - Navigation with arrow keys
  - Automatic cache cleanup
- **Videos** - Video player support (mpv, vlc, mplayer)
  - YouTube video and Shorts playback
  - Navigation with Shift+arrow keys
  - Tiling window manager friendly

### 💾 Offline & Storage
- Local SQLite database for offline reading
- Image caching (500MB limit, 30-day expiration)
- Favorites persistence

### 🎯 Organization
- Feed list with unread counts
- Tags for flexible grouping
- Categories for hierarchical organization
- Article aggregation across multiple feeds

## Installation

### Prerequisites

- Go 1.21 or later
- SQLite3
- Optional: Image viewer (sxiv, feh, imv, eog)
- Optional: Video player (mpv, vlc, mplayer)
- Optional: Pleb_Signer for NIP-55 authentication

### Build from Source

```bash
git clone https://github.com/plebone/nostrfeedz-cli
cd nostrfeedz-cli
go build -o nostrfeedz ./cmd/nostrfeedz
```

### Install

```bash
go install github.com/plebone/nostrfeedz-cli/cmd/nostrfeedz@latest
```

## Quick Start

1. **Launch the app:**
   ```bash
   nostrfeedz
   ```

2. **Choose authentication method:**
   - **Option 1 (Recommended):** Pleb_Signer (NIP-55 via D-Bus)
     - Requires [Pleb_Signer](https://github.com/PlebOne/Pleb_Signer) running
     - Most secure - keys never leave the signer app
     - Press `1` and then Enter
   
   - **Option 2:** Remote Signer (NIP-46)
     - For desktop signers like nsecBunker
     - Enter your bunker URL
     - Format: `bunker://<pubkey>?relay=<relay-url>`
   
   - **Option 3:** Private Key
     - Enter your nsec (private key)
     - ⚠️ Note: Your key will be stored locally

3. **Start adding feeds:**
   - Press `a` to add a new feed
   - Enter RSS URL or Nostr npub

## Configuration

Configuration is stored in `~/.config/nostrfeedz/config.yaml`

```yaml
nostr:
  npub: "npub1..."              # Auto-filled after login
  nsec: ""                      # Only if using private key auth
  relays:
    - "wss://relay.damus.io"
    - "wss://nos.lol"
    - "wss://relay.snort.social"
  
  # Pleb_Signer (NIP-55 via D-Bus) - Recommended for Linux
  pleb_signer:
    enabled: true               # Set to true for Pleb_Signer
    key_id: ""                  # Optional: specific key to use
  
  remote_signer:
    enabled: false              # Set to true for NIP-46
    bunker_url: "bunker://..."

sync:
  enabled: true
  auto_sync_interval: "15m"

reading:
  mark_read_behavior: "on-open"
  organization_mode: "tags"

display:
  theme: "default"
  feed_list_width: 30
  article_list_width: 40
```

## Keyboard Shortcuts

### Global
- `q` / `Ctrl+C` - Quit
- `?` - Show help
- `Tab` - Cycle between panels

### Feed List
- `↑` / `k` - Previous feed
- `↓` / `j` - Next feed
- `Enter` - Open feed
- `a` - Add new feed
- `d` - Delete feed
- `r` - Refresh feed
- `s` - Sync with Nostr

### Article List
- `↑` / `k` - Previous article
- `↓` / `j` - Next article
- `Enter` - Open article
- `m` - Mark as read/unread
- `f` - Toggle favorite

### Reader
- `↑` / `k` - Scroll up
- `↓` / `j` - Scroll down
- `Space` - Page down
- `o` - Open in browser
- `Esc` - Go back

## Nostr Integration

### Pleb_Signer (NIP-55 via D-Bus) - Recommended

NostrFeedz CLI integrates with [Pleb_Signer](https://github.com/PlebOne/Pleb_Signer), a secure Nostr signer for Linux:

1. **Install Pleb_Signer**:
   ```bash
   # See: https://github.com/PlebOne/Pleb_Signer
   ```

2. **Start Pleb_Signer** and unlock it

3. **Run NostrFeedz CLI** and select option 1

**Benefits:**
- ✅ Keys never leave the signer app
- ✅ User approval for each signature
- ✅ System tray integration
- ✅ Auto-approve for trusted apps

### Remote Signer (NIP-46)

Architecture supports NIP-46 remote signers (implementation in progress):

- nsecBunker - Desktop/web remote signer
- Amber via relay bridge
- Any NIP-46 compatible signer

### Sync Events

- **Kind 30404** - Subscription list sync
- **Kind 30405** - Read status sync

Your subscriptions and read status are synced across all devices using replaceable Nostr events.

### Default Relays

```
wss://relay.damus.io
wss://nos.lol
wss://relay.snort.social
wss://relay.nostr.band
wss://nostr-pub.wellorder.net
```

## Database

Data is stored locally in `~/.local/share/nostrfeedz/feeds.db` (SQLite)

- Feeds and subscriptions
- Article cache
- Read status
- Favorites
- Tags and categories

## Development Status

### ✅ Completed
- [x] Project structure and dependencies
- [x] Database schema and operations
- [x] Configuration management
- [x] Nostr client with multi-signer support (Pleb_Signer NIP-55, NIP-46 architecture)
- [x] Pleb_Signer (NIP-55) integration via D-Bus
- [x] Basic TUI with authentication flow
- [x] Feed list view
- [x] Subscription sync (Kind 30404) - Cross-device feed list
- [x] Read status sync (Kind 30405) - Cross-device read tracking
- [x] Tag import and storage
- [x] RSS feed metadata fetching
- [x] Nostr profile metadata fetching
- [x] Dark terminal color scheme

### 🚧 In Progress
- [ ] RSS feed article fetching
- [ ] Nostr feed fetching (NIP-23 long-form content)
- [ ] Article list and reader views
- [ ] Tag display in TUI

### 📋 Planned
- [ ] Remote Signer (NIP-46) implementation
- [ ] Publish local changes back to Nostr
- [ ] Continuous background sync
- [ ] Search functionality
- [ ] Category management UI
- [ ] Guide directory integration
- [ ] Video feed support
- [ ] Markdown rendering with syntax highlighting

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Credits

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [go-nostr](https://github.com/nbd-wtf/go-nostr) - Nostr protocol
- [gofeed](https://github.com/mmcdole/gofeed) - RSS parsing

## Links

- **Web App**: https://nostrfeedz.com
- **GitHub**: https://github.com/PlebOne/Nostr-Feedz
- **Nostr**: npub13hyx3qsqk3r7ctjqrr49uskut4yqjsxt8uvu4rekr55p08wyhf0qq90nt7

## Keyboard Shortcuts

### Navigation
- `↑/↓` or `k/j` - Navigate up/down
- `Enter` - Open selected item
- `Esc` - Go back
- `Tab` - Switch view (Feeds/Tags/Categories)
- `q` - Quit

### Feeds View
- `s` - Sync from Nostr
- Unread counts shown next to each feed

### Articles View
- `↑/↓` - Scroll through articles
- `Enter` - Read article (marks as read)
- `f` - Toggle favorite

### Reader View
- `↑/↓` - Scroll article
- `i` - View image (if available)
- `←/→` - Navigate between images (if multiple)
- `v` - Play video (if available)
- `Shift+←/→` - Navigate between videos (if multiple)

### Tags & Categories
- Navigate with `↑/↓`
- Press `Enter` to see all articles from feeds with that tag/category
- **Uncategorized** category shows feeds without assigned category

## Documentation

- [Quick Start Guide](./QUICKSTART.md) - Get started quickly
- [NIP-55 Setup Guide](./NIP55_GUIDE.md) - Pleb_Signer authentication
- [Image Viewing](./IMAGE_VIEWING_SIMPLE.md) - External image viewer setup
- [Video Player](./VIDEO_PLAYER_GUIDE.md) - Video playback configuration
- [Tags and Categories](./TAGS_AND_CATEGORIES.md) - Organization features
- [Setting Up Tags & Categories](./SETTING_UP_TAGS_AND_CATEGORIES.md) - Configure in Nostr
- [TODO & Roadmap](./TODO.md) - Future plans and tasks
- [Troubleshooting](./TROUBLESHOOTING.md) - Common issues and solutions

## Status

**Version:** 0.1.0-dev (Active Development)

### ✅ Completed Features
- Authentication (Pleb_Signer NIP-55, NIP-46 architecture, nsec)
- Feed syncing from Nostr (kind 30404)
- Read status sync (kind 30405)
- RSS and Nostr feed fetching
- Article reading with markdown rendering
- Image viewing (external viewers)
- Video playback (mpv, vlc)
- YouTube video and Shorts support
- Tags and Categories organization
- Unread count badges
- Image caching
- Process management for external apps

### 🚧 Known Issues
- Unread counts require ESC to feeds view to refresh
- Debug output still present (will be removed)
- Tags/Categories need to be set up in NostrFeedz.com first

See [TODO.md](./TODO.md) for complete roadmap and task list.
