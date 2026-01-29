# Setting Up Tags and Categories in Nostr

Currently, NostrFeedz CLI syncs tags and categories from your Nostr subscription list (kind 30404 event). However, **you need to set these up in your Nostr profile first** using a compatible client or by publishing the event yourself.

## Current Status

When you run NostrFeedz CLI and press `s` to sync, it will now show:
```
Synced! Added: 5 feeds, 3 tags, 2 categories
```

If you see:
```
Synced from Nostr! (No new data)
```

This means your Nostr subscription list doesn't currently include tags or categories.

## Understanding the Data Structure

Your Nostr subscription list (kind 30404) can contain:

```json
{
  "rss": ["https://example.com/feed.xml", ...],
  "nostr": ["npub1...", ...],
  "tags": {
    "https://example.com/feed.xml": ["tech", "news"],
    "npub1...": ["programming", "tutorials"]
  },
  "categories": {
    "https://example.com/feed.xml": {
      "name": "Technology",
      "icon": "🔧",
      "color": "#3498db"
    }
  }
}
```

## Ways to Add Tags and Categories

### Option 1: Use NostrFeedz Web App (Recommended)

The main NostrFeedz web application (if available) should provide a UI for managing tags and categories. Changes will automatically sync to your CLI.

### Option 2: Manually Publish a Kind 30404 Event

You can create and publish a kind 30404 event with tags and categories. Here's an example structure:

**Event Structure:**
- **Kind:** 30404 (subscription list)
- **d tag:** "nostr-feedz-subscriptions"
- **Content:** JSON with your feeds, tags, and categories

### Option 3: Use a Script to Update Your Subscription List

Create a script that:
1. Fetches your current subscription list
2. Adds tags and categories
3. Republishes the event

## Example Tags Structure

```json
{
  "tags": {
    "https://hnrss.org/frontpage": ["tech", "news", "startup"],
    "https://xkcd.com/rss.xml": ["comics", "tech", "fun"],
    "npub1...": ["nostr", "updates"]
  }
}
```

- **Keys:** Feed URLs or npubs
- **Values:** Array of tag names
- Tags are case-insensitive
- A feed can have multiple tags

## Example Categories Structure

```json
{
  "categories": {
    "https://hnrss.org/frontpage": {
      "name": "Technology",
      "icon": "💻",
      "color": "#3498db"
    },
    "https://techcrunch.com/feed": {
      "name": "News",
      "icon": "📰",
      "color": "#e74c3c"
    }
  }
}
```

- **Keys:** Feed URLs or npubs
- **name:** Category name (required)
- **icon:** Emoji icon (optional, default: 📁)
- **color:** Hex color code (optional)
- A feed can only have ONE category

## Using NostrFeedz CLI Without Tags/Categories

You can still use NostrFeedz CLI effectively without tags or categories:

1. **All Feeds View** - See all your feeds in one list
2. **Uncategorized Category** - Automatically shows all feeds without categories
3. **Search/Filter** - Use keyboard navigation to find feeds quickly

## Testing Your Setup

After setting up tags and categories in Nostr:

1. **Open NostrFeedz CLI**
2. **Press `s`** to sync
3. **Check the status message** - Should show "Added: X tags, Y categories"
4. **Press `Tab`** to switch to Tags view
5. **Press `Tab` again** to switch to Categories view
6. **Press `Enter`** on any tag/category to see articles

## Troubleshooting

### "No tags/categories shown after sync"

**Possible causes:**
- Your Nostr subscription list doesn't include tags/categories yet
- The event hasn't been published correctly
- You're connected to relays that don't have your subscription list

**Solutions:**
- Verify your kind 30404 event contains tags/categories
- Check you're connected to the same relays where you published
- Re-publish your subscription list with tags/categories

### "Tags show but no articles"

**Possible causes:**
- Feeds with those tags haven't been synced yet
- No articles fetched from those feeds

**Solutions:**
- Wait for feed sync to complete (happens automatically)
- Check that feeds are actually subscribed in your list

### "Categories shows only Uncategorized"

This is normal if:
- You haven't assigned categories to your feeds yet
- Your subscription list doesn't include categories

The **Uncategorized** category is always shown and contains feeds without assigned categories.

## Future Enhancements

Planned features for tags and categories:
- [ ] Edit tags/categories directly in CLI
- [ ] Create new tags/categories from CLI
- [ ] Drag-and-drop organization
- [ ] Tag/category suggestions
- [ ] Bulk editing
- [ ] Import/export tag configurations

## Example: Publishing Tags via CLI

Here's a simple example of how you might update your subscription list (requires nostr libraries):

```bash
# This is pseudocode - actual implementation depends on your Nostr client
nostr publish --kind 30404 --d-tag "nostr-feedz-subscriptions" \
  --content '{
    "rss": ["https://example.com/feed"],
    "tags": {
      "https://example.com/feed": ["tech", "news"]
    },
    "categories": {
      "https://example.com/feed": {
        "name": "Technology",
        "icon": "💻"
      }
    }
  }'
```

## Getting Help

If you're having trouble setting up tags and categories:

1. Check the sync status message to see what was imported
2. Verify your Nostr connection is working
3. Ensure you're using the correct d-tag: "nostr-feedz-subscriptions"
4. Check your kind 30404 event content format
5. Make sure feed URLs match exactly (including https://)

## Related Documentation

- [TAGS_AND_CATEGORIES.md](./TAGS_AND_CATEGORIES.md) - How to use tags and categories in the CLI
- [DEVELOPMENT_SUMMARY.md](./DEVELOPMENT_SUMMARY.md) - Technical implementation details
- [NIP55_GUIDE.md](./NIP55_GUIDE.md) - Nostr authentication setup
