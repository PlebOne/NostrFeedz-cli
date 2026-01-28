# NostrFeedz CLI

A beautiful command-line interface for Nostr-Feedz - your RSS and Nostr feed reader with cross-device synchronization.

## Features

- 🔐 **Multiple Authentication Methods**:
  - **Pleb_Signer** (NIP-55 via D-Bus) - Recommended for Linux desktop
  - **NIP-46** remote signer support architecture
  - **nsec** private key (fallback)
- 📰 **RSS & Nostr Feeds** - Subscribe to RSS feeds and Nostr long-form content (NIP-23)
- 🎨 **Beautiful TUI** - Built with Bubble Tea for a modern terminal experience
- 🔄 **Cross-Device Sync** - Sync subscriptions and read status via Nostr (kinds 30404, 30405)
- 💾 **Offline Support** - Local SQLite database for offline reading
- 🏷️ **Organization** - Organize feeds with tags and categories
- ⭐ **Favorites** - Star articles for later reading
- 🔍 **Search** - Find articles across all your feeds

## Installation

### Prerequisites

- Go 1.21 or later
- SQLite3

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
