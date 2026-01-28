# Read Status Sync Implementation

## Overview

Added **kind 30405** sync to import read status from Nostr after authentication.

## What Was Added

### Sync Flow Update

The sync process now has two phases:

**Phase 1: Subscription Sync (kind 30404)** - Already implemented
- Fetches subscription list
- Imports RSS and Nostr feeds

**Phase 2: Read Status Sync (kind 30405)** - NEW
- Fetches read status list
- Marks articles as read by GUID
- Continues silently if sync fails (non-critical)

### Code Changes

**File: `internal/app/handlers.go`**
- Extended `syncFromNostr()` to fetch kind 30405
- Added loop to mark items as read by GUID
- Added error handling (non-fatal if read status fails)

**File: `internal/db/sqlite.go`**
- Added `MarkItemReadByGUID(guid string)` method
- Uses GUID to find articles (works across feeds)
- Updates `is_read` column to 1

### Read Status Event Format (kind 30405)

```json
{
  "kind": 30405,
  "pubkey": "<your-hex-pubkey>",
  "created_at": 1732645747,
  "tags": [
    ["d", "nostr-feedz-read-status"],
    ["client", "nostrfeedz-cli"]
  ],
  "content": "{\"itemGuids\":[\"guid1\",\"guid2\",\"guid3\",...],\"lastUpdated\":1732645747}"
}
```

## Benefits

### Cross-Device Sync ✅

- Read an article on **web app** → Shows as read in **CLI**
- Read an article on **mobile** → Shows as read everywhere
- Read an article in **CLI** → Will sync to other devices (when publishing is implemented)

## Files Changed

- `internal/app/handlers.go` - Added read status sync
- `internal/db/sqlite.go` - Added `MarkItemReadByGUID()` method
- `SYNC_AND_COLORS_UPDATE.md` - Updated documentation
