# Quick Start Guide

## First Run

1. **Run the application:**
   ```bash
   ./nostrfeedz
   ```

2. **You'll see the authentication screen:**
   - Option `1` for Pleb_Signer (D-Bus) - **Recommended** ✨
   - Option `2` for Remote Signer (NIP-46) - Not fully implemented yet  
   - Option `3` for Private Key (nsec)

3. **Using Pleb_Signer (Option 1):**
   - Make sure [Pleb_Signer](https://github.com/PlebOne/Pleb_Signer) is running
   - Press `1` then `Enter`
   - Approve the connection in Pleb_Signer when prompted
   - Your keys stay secure in Pleb_Signer!

4. **Using Private Key (Option 3):**
   - Press `3` then enter your nsec
   - It will be stored in `~/.config/nostrfeedz/config.yaml`

5. **After successful login:**
   - You'll see the feed list (empty initially)
   - Press `q` to quit

## Testing with a Test Key

If you don't have a real nsec, you can generate one:

```bash
# In another terminal, install nak (Nostr Army Knife)
go install github.com/fiatjaf/nak@latest

# Generate a new keypair
nak key generate
```

This will output an nsec you can use for testing.

## Current Functionality

✅ **Working:**
- Authentication with nsec (private key)
- Configuration management
- Database initialization
- Basic TUI with authentication flow
- Feed list view (empty)

🚧 **In Progress:**
- Feed management (add/remove/refresh)
- Article viewing
- Nostr sync

## Next Steps

To add feeds and view articles, we need to implement:
1. RSS feed fetcher
2. Nostr feed fetcher (NIP-23)
3. Article list and reader views
4. Sync functionality

## Keyboard Shortcuts (Current)

- **Authentication Screen:**
  - `1` - Remote Signer (not yet implemented)
  - `2` - Private Key (nsec)
  - Type to enter text
  - `Enter` - Submit
  - `Backspace` - Delete character
  - `Esc` - Go back

- **Feed List:**
  - `q` / `Ctrl+C` - Quit
  - `↑` / `k` - Previous feed
  - `↓` / `j` - Next feed
  - `Enter` - Open feed (not yet implemented)
  - `a` - Add feed (not yet implemented)

## Configuration File

After first login, your config will be at: `~/.config/nostrfeedz/config.yaml`

You can manually edit this file to:
- Change relays
- Adjust display settings
- Enable/disable sync

## Database

Local data is stored in: `~/.local/share/nostrfeedz/feeds.db`

This SQLite database contains:
- Feeds
- Articles
- Read status
- Favorites
- Tags and categories

## Troubleshooting

**"Blank screen" or "Can't see authentication options"**
- Press any key - sometimes the initial render is slightly delayed
- Try resizing your terminal window slightly (this triggers a refresh)
- Ensure your terminal is at least 80x24 characters
- Check terminal size with: `echo "$(tput cols)x$(tput lines)"`
- Try running: `./test-app.sh` which includes diagnostics

**"Failed to connect to Nostr"**
- Check your internet connection
- Try different relays in config.yaml
- Verify your nsec is correct

**"Invalid private key"**
- Make sure you're using nsec format (starts with "nsec1")
- Don't use npub (that's your public key)

**"Database error"**
- Check permissions in `~/.local/share/nostrfeedz/`
- Try deleting feeds.db to reset (you'll lose local data)

## Support

- Documentation: See README.md
- Issues: Report bugs on GitHub
- Nostr: npub13hyx3qsqk3r7ctjqrr49uskut4yqjsxt8uvu4rekr55p08wyhf0qq90nt7
