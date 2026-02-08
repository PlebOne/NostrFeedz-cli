#!/bin/bash
# Test script for validating tags and categories functionality
# This script helps test the tags/categories sync with NostrFeedz.com

set -e

echo "================================"
echo "NostrFeedz CLI - Tags & Categories Test"
echo "================================"
echo ""

# Check if app is built
if [ ! -f "./nostrfeedz" ]; then
    echo "Building nostrfeedz..."
    go build -o nostrfeedz ./cmd/debug
    echo "✓ Build complete"
    echo ""
fi

# Check database location
DB_PATH="$HOME/.local/share/nostrfeedz/feeds.db"
echo "Database location: $DB_PATH"

if [ -f "$DB_PATH" ]; then
    echo "✓ Database exists"
    echo ""
    
    # Show current data
    echo "Current Database Contents:"
    echo "=========================="
    
    echo -e "\nFeeds:"
    sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM feeds;" | xargs echo "  Total feeds:"
    
    echo -e "\nTags:"
    sqlite3 "$DB_PATH" "SELECT id, name FROM tags;" 2>/dev/null | while read line; do
        echo "  - $line"
    done || echo "  (no tags table or no tags)"
    
    echo -e "\nCategories:"
    sqlite3 "$DB_PATH" "SELECT id, name, icon FROM categories;" 2>/dev/null | while read line; do
        echo "  - $line"
    done || echo "  (no categories table or no categories)"
    
    echo -e "\nFeed-Tag Associations:"
    sqlite3 "$DB_PATH" "SELECT feed_id, tag_id FROM feed_tags;" 2>/dev/null | while read line; do
        echo "  - $line"
    done || echo "  (no feed_tags table or no associations)"
    
    echo -e "\nFeeds with Categories:"
    sqlite3 "$DB_PATH" "SELECT url, category_id FROM feeds WHERE category_id != '' AND category_id != 'synced';" 2>/dev/null | while read line; do
        echo "  - $line"
    done || echo "  (no feeds with categories)"
    
else
    echo "✗ Database not found - run the app first to create it"
fi

echo ""
echo "================================"
echo "Test Instructions:"
echo "================================"
echo ""
echo "1. SETUP TAGS & CATEGORIES in NostrFeedz.com:"
echo "   - Go to https://nostrfeedz.com"
echo "   - Login with your Nostr account"
echo "   - Add some feeds if you haven't already"
echo "   - Add tags to your feeds (e.g., 'tech', 'news', 'bitcoin')"
echo "   - Assign categories to feeds (e.g., 'Technology', 'News')"
echo ""
echo "2. TEST SYNC:"
echo "   - Run: ./nostrfeedz"
echo "   - Authenticate with your account"
echo "   - Press 's' to sync from Nostr"
echo "   - Check the sync message for tag/category counts"
echo ""
echo "3. TEST VIEWS:"
echo "   - Press Tab to switch to Tags view"
echo "   - Verify your tags appear in the list"
echo "   - Press Enter on a tag to see articles from all feeds with that tag"
echo "   - Press Tab again to switch to Categories view"
echo "   - Verify 'Uncategorized' appears first"
echo "   - Verify your categories appear below"
echo "   - Press Enter on a category to see articles"
echo ""
echo "4. TEST MULTI-FEED AGGREGATION:"
echo "   - Add the same tag to multiple feeds in NostrFeedz.com"
echo "   - Sync in the CLI (press 's')"
echo "   - View that tag - should show articles from all tagged feeds"
echo ""
echo "5. TEST UNCATEGORIZED:"
echo "   - Leave some feeds without a category in NostrFeedz.com"
echo "   - Sync in the CLI"
echo "   - Go to Categories view (Tab twice)"
echo "   - Select 'Uncategorized' - should show uncategorized feeds"
echo ""
echo "================================"
echo "Expected Sync Output:"
echo "================================"
echo "  Synced! Added: X feeds, Y tags, Z categories"
echo ""
echo "Where:"
echo "  - X = number of new feeds added"
echo "  - Y = number of unique tags imported"
echo "  - Z = number of feeds with categories assigned"
echo ""
echo "================================"
echo "Debugging:"
echo "================================"
echo ""
echo "If tags/categories don't show:"
echo ""
echo "1. Check Nostr event structure:"
echo "   - Your kind 30404 event should have 'tags' and 'categories' fields"
echo "   - Structure: {\"rss\": [...], \"tags\": {...}, \"categories\": {...}}"
echo ""
echo "2. Verify sync ran successfully:"
echo "   - Look for sync error messages"
echo "   - Check you're connected to the right relays"
echo ""
echo "3. Check database after sync:"
echo "   - Run: sqlite3 ~/.local/share/nostrfeedz/feeds.db"
echo "   - Query: SELECT * FROM tags;"
echo "   - Query: SELECT * FROM categories;"
echo "   - Query: SELECT * FROM feed_tags;"
echo ""
echo "4. Check debug output:"
echo "   - The sync function prints debug info to stderr"
echo "   - Look for: 'DEBUG: Sync received - ...' message"
echo ""
echo "================================"
echo "Ready to test!"
echo "================================"
