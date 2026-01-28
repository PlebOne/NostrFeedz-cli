# Category Sync and View Modes

## Overview
NostrFeedz CLI now supports category synchronization from Nostr and provides three different view modes for organizing and browsing your feeds.

## Features Implemented

### 1. Category Sync (Kind 30404)
The webapp now publishes categories as part of the subscription list event. The CLI imports this data:

**Nostr Event Format:**
```json
{
  "kind": 30404,
  "content": {
    "rss": ["https://feed1.com", ...],
    "nostr": ["npub1...", ...],
    "tags": {
      "tech": ["https://feed1.com", "npub1..."],
      "bitcoin": ["https://feed2.com"]
    },
    "categories": {
      "https://feed1.com": "Technology",
      "npub1...": "Bitcoin",
      "https://feed2.com": "News"
    },
    "deleted": [],
    "lastUpdated": 1769626318
  }
}
```

**Category Assignment:**
- Maps feed URLs/npubs to category names
- Categories are created automatically if they don't exist
- Feeds are updated with their category assignment
- Remote category assignments win on conflicts during merge

### 2. Three View Modes

Users can now toggle between three different ways to view their content:

#### **Feeds View** (Default)
- Shows all subscribed feeds in a flat list
- Direct access to all feeds regardless of organization
- Quick browsing of everything you're subscribed to

#### **Tags View**
- Shows all tags imported from Nostr
- Select a tag to see all feeds with that tag
- Great for topical browsing (e.g., "bitcoin", "tech", "podcasts")
- Tags are many-to-many: one feed can have multiple tags

#### **Categories View**
- Shows all categories
- Select a category to see all feeds in that category
- Hierarchical organization: each feed belongs to one category
- Categories can have icons/emojis for visual distinction

### 3. View Mode Toggle
- Press **Tab** to cycle through: Feeds → Tags → Categories → Feeds
- Current view mode displayed in header:
  - 📰 All Feeds
  - 🏷️  Tags
  - 📂 Categories
- View mode persists during navigation

### 4. Navigation
Each view mode maintains its own selection state:
- **↑/↓ or k/j** - Navigate up/down
- **Enter** - Open selected item
  - Feeds View: Opens feed to show articles
  - Tags View: Shows feeds with that tag
  - Categories View: Shows feeds in that category
- **s** - Manual sync from Nostr
- **q** - Quit

## Database Schema

### Categories Table
```sql
CREATE TABLE categories (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    color TEXT,                   -- Hex color code (optional)
    icon TEXT,                    -- Emoji icon (optional)
    sort_order INTEGER DEFAULT 0
);
```

### Updated Feeds Table
```sql
CREATE TABLE feeds (
    ...
    category_id TEXT,             -- FK to categories.id
    ...
);
```

## New Database Methods

### Category Operations
- `CreateCategory(category *Category)` - Create new category
- `GetCategories()` - Get all categories sorted by name
- `GetCategoryByName(name)` - Find category by name
- `GetFeedsByCategory(categoryID)` - Get all feeds in a category

### Tag Operations (Enhanced)
- `GetFeedsByTag(tagID)` - Get all feeds with a specific tag

## Code Changes

### Internal Structures

**ViewMode Enum:**
```go
type ViewMode int

const (
    ViewModeFeeds ViewMode = iota
    ViewModeTags
    ViewModeCategories
)
```

**SubscriptionList:**
```go
type SubscriptionList struct {
    RSS         []string            `json:"rss"`
    Nostr       []string            `json:"nostr"`
    Tags        map[string][]string `json:"tags"`        // URL/npub -> tags
    Categories  map[string]string   `json:"categories"`  // URL/npub -> category
    Deleted     []string            `json:"deleted"`
    LastUpdated int64               `json:"lastUpdated"`
}
```

### Sync Flow

1. Authenticate with Pleb_Signer
2. Fetch kind 30404 event from Nostr
3. Parse subscription list including categories
4. Create categories if they don't exist
5. Import RSS feeds with metadata fetching
6. Import Nostr feeds with profile fetching
7. Apply tags to feeds
8. **Apply categories to feeds** ← New!
9. Mark read items from kind 30405

### UI Updates

**View Rendering:**
- Dynamic header based on view mode
- Tab toggle hint displayed
- Content switches based on view mode
- Empty state messages per view mode
- Icons for categories (if provided)

## User Experience

### Typical Workflow

1. **First Launch:**
   - Authenticate with Pleb_Signer
   - App automatically syncs from Nostr
   - Feeds, tags, and categories are imported

2. **Browse by Feeds:**
   - Default view shows all feeds
   - Navigate and select any feed

3. **Browse by Tags:**
   - Press Tab to switch to Tags view
   - See all your tags (e.g., "tech", "bitcoin", "podcasts")
   - Select a tag to see related feeds

4. **Browse by Categories:**
   - Press Tab again to switch to Categories view
   - See organized categories (e.g., "Technology", "News", "Finance")
   - Categories may have custom icons
   - Select a category to see feeds within it

5. **Manual Sync:**
   - Press 's' any time to sync latest changes from Nostr
   - All views refresh with new data

## Benefits

### For Users
- **Flexible Organization:** Choose how you want to browse
- **Cross-Device Sync:** Categories sync from webapp
- **Visual Hierarchy:** Icons and clear view modes
- **Fast Switching:** Quick toggle between views

### For Development
- **Extensible:** Easy to add more view modes
- **Maintainable:** Clear separation of concerns
- **Synced:** Single source of truth on Nostr
- **Backward Compatible:** Works with old sync data (categories optional)

## Future Enhancements

Potential improvements:
- [ ] Search within each view mode
- [ ] Filter/sort options per view
- [ ] Custom category colors in TUI
- [ ] Create/edit categories from CLI
- [ ] Publish category changes back to Nostr
- [ ] Remember last view mode preference
- [ ] Show feed count per tag/category
- [ ] Multi-select for bulk operations

## Related Files

- `internal/nostr/sync.go` - SubscriptionList with Categories field
- `internal/app/app.go` - ViewMode system and rendering
- `internal/app/handlers.go` - View mode navigation and data loading
- `internal/db/sqlite.go` - Category database operations
- `internal/db/models.go` - Category model

## Testing

To test the full flow:

1. Use the webapp to organize feeds into categories
2. Run the CLI: `./nostrfeedz`
3. Authenticate with Pleb_Signer
4. Verify sync completes successfully
5. Press Tab to cycle through views:
   - Feeds view should show all feeds
   - Tags view should show imported tags
   - Categories view should show categories from webapp
6. Navigate and select items in each view
7. Press 's' to manually sync and verify updates

## Conclusion

NostrFeedz CLI now provides a flexible, organized way to browse your RSS and Nostr feeds. With three distinct view modes and full category sync, users can choose the organizational method that works best for their workflow.
