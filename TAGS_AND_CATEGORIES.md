# Tags and Categories Guide

NostrFeedz CLI supports organizing your feeds using **Tags** and **Categories** from your Nostr profile.

## Overview

You can view your feeds in three different modes:
1. **📰 All Feeds** - List all feeds
2. **🏷️ Tags** - Group feeds by tags
3. **📂 Categories** - Group feeds by categories

Press `Tab` to cycle between these views.

## Tags View

Tags allow you to group related feeds together. When you select a tag, you'll see all articles from feeds that have that tag.

### How it works:
1. Press `Tab` to switch to **Tags** view
2. Use `↑`/`↓` to navigate through your tags
3. Press `Enter` to see all articles from feeds with that tag
4. Articles from multiple feeds are combined and shown together

### Example:
If you have a "tech" tag applied to:
- Hacker News feed
- TechCrunch feed
- Dev.to feed

Selecting the "tech" tag will show articles from all three feeds in one list, sorted by date.

## Categories View

Categories are broader groupings for your feeds, often representing topics or sections.

### Special Category: Uncategorized

The **Uncategorized** category is automatically added at the top of the list. It shows all feeds that don't have a category assigned.

This is useful for:
- Finding feeds you haven't organized yet
- Temporary feeds you're trying out
- Feeds that don't fit into your existing categories

### How it works:
1. Press `Tab` twice to switch to **Categories** view
2. Use `↑`/`↓` to navigate through your categories
3. The first option is always **📋 Uncategorized**
4. Press `Enter` to see all articles from feeds in that category
5. Articles from multiple feeds are combined and shown together

### Example Categories:
```
▸ 📋 Uncategorized (23)
  📰 News (12)
  🔧 Tech (45)
  🎮 Gaming (8)
  📚 Learning (15)
```

## Keyboard Shortcuts

When in Tags or Categories view:
- `↑` or `k` - Move up
- `↓` or `j` - Move down
- `Enter` - View articles from all feeds with that tag/category
- `Tab` - Switch between Feeds/Tags/Categories
- `Esc` - Go back to main view
- `s` - Sync from Nostr (refreshes tags and categories)

## Unread Counts

Just like the Feeds view, Tags and Categories views will show unread counts (when implemented) to help you see where you have new content.

## Setting Up Tags and Categories

Tags and categories are synced from your Nostr profile:

1. **Configure your feeds on Nostr** - Use a Nostr client that supports feed management
2. **Add tags to feeds** - Apply tags like "tech", "news", "podcasts"
3. **Assign categories** - Group feeds into broader categories
4. **Sync in NostrFeedz CLI** - Press `s` to sync from Nostr
5. **Access your organization** - Use `Tab` to switch views

## Tips

- **Combine organization methods**: You can use both tags and categories together
- **Use Uncategorized to find unorganized feeds**: Start here to sort new feeds
- **Tags are more flexible**: One feed can have multiple tags, but only one category
- **Categories are hierarchical**: Better for top-level organization
- **Sync regularly**: Press `s` to keep your organization up to date

## Workflow Example

1. Add a new RSS feed via Nostr
2. Press `s` in NostrFeedz CLI to sync
3. Press `Tab` twice to go to Categories view
4. Select **Uncategorized** to see the new feed
5. Go back to Nostr and assign it to a category
6. Press `s` again to sync
7. The feed now appears in its category!

## Benefits

- **Reduce noise**: Focus on specific topics when you want
- **Combine related content**: See all tech news together, regardless of source
- **Quick filtering**: Jump to exactly what you want to read
- **Better overview**: See all your organization at a glance
- **Find forgotten feeds**: Use Uncategorized to discover feeds you haven't organized
