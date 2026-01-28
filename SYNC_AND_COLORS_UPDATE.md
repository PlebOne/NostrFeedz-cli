# Sync and Colors Update

## Changes Made

### 1. Dark Terminal Color Scheme ✅

Fixed color scheme to work properly with dark terminals (most common setup).

**Before**: Light theme colors (white background, dark text) - unreadable on dark terminals
**After**: Dark theme colors optimized for dark backgrounds

#### Color Changes:
- **Primary**: `#7C3AED` → `#A78BFA` (brighter purple)
- **Accent**: `#3B82F6` → `#60A5FA` (brighter blue)
- **Text**: `#1F2937` → `#F9FAFB` (dark → light for dark bg)
- **Background**: `#FFFFFF` → `#1F2937` (white → dark gray)
- **Selected**: Now uses white text on blue background for clear visibility
- **Border**: Adjusted for better contrast on dark terminals

### 2. Nostr Sync After Authentication ✅

The app now automatically syncs BOTH your feed subscriptions AND read status from Nostr after successful authentication.

#### How It Works:

**Step 1: Subscription Sync (kind 30404)**
1. Fetches your subscription list from Nostr
2. Extracts RSS and Nostr feeds from your published list
3. Imports any new feeds to your local database
4. Shows status: "Synced from Nostr! Added X feeds"

**Step 2: Read Status Sync (kind 30405)**
1. Fetches your read status list from Nostr
2. Extracts the GUIDs of articles you've already read
3. Marks those articles as read in your local database
4. Silently continues if read status sync fails (non-critical)

#### Subscription Format (kind 30404):

The app looks for events with:
- **Kind**: 30404 (Nostr Feed List)
- **D-Tag**: "nostr-feedz-subscriptions"
- **Author**: Your public key

Content structure:
```json
{
  "rss": ["https://example.com/feed.xml", ...],
  "nostr": ["pubkey1", "pubkey2", ...],
  "tags": {"tech": ["feed1", "feed2"]},
  "deleted": [],
  "lastUpdated": 1234567890
}
```

#### Read Status Format (kind 30405):

The app looks for events with:
- **Kind**: 30405 (Read Status)
- **D-Tag**: "nostr-feedz-read-status"
- **Author**: Your public key

Content structure:
```json
{
  "itemGuids": ["guid1", "guid2", "guid3", ...],
  "lastUpdated": 1234567890
}
```

### 3. New Database Methods ✅

- `GetFeedByURL(url string)` - Check if a feed already exists before importing
- `MarkItemReadByGUID(guid string)` - Mark an article as read by its GUID (for sync)

## Files Modified

- `pkg/styles/theme.go` - Updated color scheme for dark terminals
- `internal/app/app.go` - Added sync message handling
- `internal/app/handlers.go` - Added `syncFromNostr()` function with read status sync
- `internal/db/sqlite.go` - Added `GetFeedByURL()` and `MarkItemReadByGUID()` methods

## User Experience

### Before:
1. Authenticate → Empty feed list, no read status
2. User must manually add feeds
3. All articles show as unread
4. No sync with Nostr

### After:
1. Authenticate → "Syncing from Nostr..."
2. Status message: "Synced from Nostr! Added X feeds"
3. Feeds automatically loaded from your Nostr subscription list
4. **Articles you've read on other devices are already marked as read**
5. Better readable menu with dark terminal colors

## Cross-Device Sync

This implementation enables **true cross-device synchronization**:

- Read an article on the web app → CLI shows it as read
- Subscribe to a feed in the CLI → Web app shows the subscription
- Mark items read in the CLI → Web app syncs the read status
- Works across desktop, mobile, and any Nostr-Feedz client

## Testing

```bash
# Rebuild
go build -o nostrfeedz ./cmd/nostrfeedz/

# Run
./nostrfeedz

# After authentication, you should see:
# - "Successfully authenticated! Syncing from Nostr..."
# - "Synced from Nostr! Added X feeds"
# - Your feeds list populated from Nostr
# - Articles you've read elsewhere already marked as read
```

## Implementation Details

### Sync Flow:

```
1. User authenticates with Pleb_Signer
2. App gets user's public key
3. Fetches kind 30404 (subscriptions) from relays
4. Imports RSS feeds to local DB
5. Imports Nostr feeds to local DB
6. Fetches kind 30405 (read status) from relays
7. Marks articles as read by GUID
8. Displays sync results
```

### Error Handling:

- **Subscription sync failure** → Shows error, no feeds imported
- **Read status sync failure** → Logs warning, continues (non-critical)
- **Missing articles** → Skips marking (article might not be fetched yet)
- **Duplicate feeds** → Checks URL before importing (prevents duplicates)

## Next Steps

To publish your local subscriptions and read status to Nostr:

### Publishing Subscriptions:
```go
// Get all feeds from DB
feeds := m.db.GetFeeds()

// Build subscription list
subs := &nostr.SubscriptionList{
    RSS:   []string{},
    Nostr: []string{},
    Tags:  map[string][]string{},
}

for _, feed := range feeds {
    if feed.Type == "rss" {
        subs.RSS = append(subs.RSS, feed.URL)
    } else if feed.Type == "nostr" {
        subs.Nostr = append(subs.Nostr, feed.NPUB)
    }
}

// Publish to Nostr
m.nostr.PublishSubscriptions(subs)
```

### Publishing Read Status:
```go
// Get all read items from DB
readItems := m.db.GetReadItems()

// Build read status list
status := &nostr.ReadStatusList{
    ItemGuids: []string{},
}

for _, item := range readItems {
    status.ItemGuids = append(status.ItemGuids, item.GUID)
}

// Publish to Nostr
m.nostr.PublishReadStatus(status)
```

## Notes

- Sync only runs on initial authentication (not continuous)
- Duplicate feeds are skipped (checks URL)
- Synced feeds are tagged with category "synced"
- RSS feeds use URL as-is
- Nostr feeds use `nostr:<pubkey>` format
- Read status uses article GUIDs (unique identifiers)
- Both sync events use **replaceable events** (kind 3xxxx)

