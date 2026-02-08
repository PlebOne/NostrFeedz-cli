#!/bin/bash
# Live Validation Test - Run this in your terminal

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║  NostrFeedz CLI - Live Validation Test Guide             ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""
echo "📋 PREPARATION CHECKLIST:"
echo ""
echo "Before starting, make sure you have:"
echo "  ✓ Pleb_Signer running (ps aux | grep pleb-signer)"
echo "  ✓ Access to NostrFeedz.com with your account"
echo "  ✓ At least 3-5 feeds added in NostrFeedz.com"
echo ""

# Check Pleb_Signer
if ps aux | grep -v grep | grep -q pleb-signer; then
    echo "✅ Pleb_Signer is running"
else
    echo "❌ Pleb_Signer is NOT running - please start it first"
    exit 1
fi

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "STEP 1: Set up tags and categories in NostrFeedz.com"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "1. Go to: https://nostrfeedz.com"
echo "2. Login with your Nostr account"
echo "3. If you don't have feeds yet, add a few:"
echo "   - RSS: https://hnrss.org/frontpage"
echo "   - RSS: https://xkcd.com/rss.xml"
echo "   - RSS: https://blog.rust-lang.org/feed.xml"
echo ""
echo "4. Add TAGS to your feeds:"
echo "   Example tags: 'tech', 'news', 'bitcoin', 'programming'"
echo "   TIP: Use the same tag on 2-3 different feeds for aggregation test"
echo ""
echo "5. Assign CATEGORIES to some feeds:"
echo "   Example categories: 'Technology', 'News', 'Comics'"
echo "   TIP: Leave 1-2 feeds WITHOUT a category for Uncategorized test"
echo ""
read -p "Press ENTER when you've set up tags and categories in NostrFeedz.com..."

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "STEP 2: Launch NostrFeedz CLI"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "Now open a NEW TERMINAL and run:"
echo ""
echo "  cd ~/Projects/NostrFeedz-cli"
echo "  ./nostrfeedz"
echo ""
echo "Then:"
echo "  1. Press '1' to select Pleb_Signer authentication"
echo "  2. Press ENTER to connect"
echo "  3. Approve the connection in Pleb_Signer (system tray)"
echo "  4. Wait for 'Successfully authenticated!' message"
echo ""
read -p "Press ENTER when you're authenticated and see the feeds view..."

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "STEP 3: Initial Sync Test"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "In the NostrFeedz CLI window:"
echo ""
echo "  → Press 's' to sync from Nostr"
echo ""
echo "Expected output:"
echo "  'Synced! Added: X feeds, Y tags, Z categories'"
echo ""
echo "Where:"
echo "  - X = number of feeds you have in NostrFeedz.com"
echo "  - Y = number of unique tags you added"
echo "  - Z = number of feeds you assigned categories to"
echo ""
echo "📝 Record the numbers:"
read -p "  Feeds synced (X): " feeds_count
read -p "  Tags synced (Y): " tags_count
read -p "  Categories synced (Z): " cats_count

echo ""
if [ "$tags_count" -gt 0 ] || [ "$cats_count" -gt 0 ]; then
    echo "✅ Tags or categories were synced! Continuing..."
else
    echo "⚠️  No tags or categories synced. Check:"
    echo "   - Did you add tags/categories in NostrFeedz.com?"
    echo "   - Did the sync complete successfully?"
    echo "   - Check for error messages"
    read -p "Press ENTER to continue anyway..."
fi

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "STEP 4: Verify Database"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "Let's check what's in the database..."
echo ""

sqlite3 ~/.local/share/nostrfeedz/feeds.db << 'SQL'
.mode column
.headers on
SELECT '=== TAGS ===' as info;
SELECT id, name FROM tags;
SELECT '';
SELECT '=== CATEGORIES ===' as info;
SELECT id, name, icon FROM categories;
SELECT '';
SELECT '=== FEED-TAG ASSOCIATIONS ===' as info;
SELECT feed_id, tag_id FROM feed_tags LIMIT 10;
SQL

echo ""
read -p "Press ENTER to continue to Tags view test..."

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "STEP 5: Test Tags View"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "In the NostrFeedz CLI window:"
echo ""
echo "  → Press TAB to switch to Tags view"
echo ""
echo "Expected result:"
echo "  - You should see a list of your tags with 🏷️ icons"
echo "  - Each tag should be listed"
echo ""
read -p "Can you see your tags? (y/n): " see_tags

if [ "$see_tags" = "y" ]; then
    echo "✅ Tags view working!"
else
    echo "❌ Tags not visible - check sync messages"
fi

echo ""
echo "Now test tag selection:"
echo "  → Use ↑/↓ to select a tag (preferably one on multiple feeds)"
echo "  → Press ENTER to view articles"
echo ""
echo "Expected result:"
echo "  - Articles from ALL feeds with that tag should appear"
echo "  - Articles should be combined in one list"
echo ""
read -p "Do you see articles from multiple feeds combined? (y/n): " multi_feed

if [ "$multi_feed" = "y" ]; then
    echo "✅ Multi-feed aggregation working!"
else
    echo "⚠️  Check if the tag is on multiple feeds"
fi

echo ""
echo "  → Press ESC to go back to tags list"
echo ""
read -p "Press ENTER to continue to Categories test..."

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "STEP 6: Test Categories View"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "In the NostrFeedz CLI window:"
echo ""
echo "  → Press TAB to switch to Categories view"
echo ""
echo "Expected result:"
echo "  - First item should be '📋 Uncategorized'"
echo "  - Your categories should appear below"
echo ""
read -p "Do you see 'Uncategorized' as the first item? (y/n): " see_uncat

if [ "$see_uncat" = "y" ]; then
    echo "✅ Uncategorized category present!"
else
    echo "❌ Uncategorized missing - possible UI issue"
fi

read -p "Do you see your categories below it? (y/n): " see_cats

if [ "$see_cats" = "y" ]; then
    echo "✅ Categories view working!"
else
    echo "❌ Categories not visible"
fi

echo ""
echo "Now test Uncategorized:"
echo "  → Select 'Uncategorized' (should be first)"
echo "  → Press ENTER"
echo ""
read -p "Do you see feeds that DON'T have a category? (y/n): " uncat_works

if [ "$uncat_works" = "y" ]; then
    echo "✅ Uncategorized filter working!"
else
    echo "⚠️  Check if you have feeds without categories"
fi

echo ""
echo "  → Press ESC to go back"
echo "  → Select a category"
echo "  → Press ENTER"
echo ""
read -p "Do articles from that category appear? (y/n): " cat_articles

if [ "$cat_articles" = "y" ]; then
    echo "✅ Category filtering working!"
fi

echo ""
read -p "Press ENTER to continue to final test..."

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "STEP 7: Test View Cycling"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "In the NostrFeedz CLI window:"
echo ""
echo "  → Press ESC to go back to main view"
echo "  → Press TAB - should go to Tags view"
echo "  → Press TAB - should go to Categories view"
echo "  → Press TAB - should cycle back to Feeds view"
echo ""
read -p "Does view cycling work smoothly? (y/n): " cycling

if [ "$cycling" = "y" ]; then
    echo "✅ View cycling working!"
fi

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "STEP 8: Test Sync Update"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "Let's test live updates:"
echo ""
echo "1. Leave NostrFeedz CLI running"
echo "2. Go to NostrFeedz.com in your browser"
echo "3. Add a NEW tag to an existing feed (e.g., 'test-tag')"
echo "4. Go back to NostrFeedz CLI"
echo "5. Press 's' to sync"
echo "6. Press TAB to go to Tags view"
echo ""
read -p "Does the new tag appear in the list? (y/n): " new_tag

if [ "$new_tag" = "y" ]; then
    echo "✅ Sync updates working!"
else
    echo "⚠️  May need to restart app or check sync"
fi

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "VALIDATION COMPLETE!"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "📊 RESULTS SUMMARY:"
echo ""
echo "  Initial sync:     Feeds=$feeds_count, Tags=$tags_count, Categories=$cats_count"
echo "  Tags view:        $([ "$see_tags" = "y" ] && echo "✅ PASS" || echo "❌ FAIL")"
echo "  Multi-feed agg:   $([ "$multi_feed" = "y" ] && echo "✅ PASS" || echo "⚠️  WARN")"
echo "  Categories view:  $([ "$see_cats" = "y" ] && echo "✅ PASS" || echo "❌ FAIL")"
echo "  Uncategorized:    $([ "$uncat_works" = "y" ] && echo "✅ PASS" || echo "⚠️  WARN")"
echo "  View cycling:     $([ "$cycling" = "y" ] && echo "✅ PASS" || echo "❌ FAIL")"
echo "  Sync updates:     $([ "$new_tag" = "y" ] && echo "✅ PASS" || echo "⚠️  WARN")"
echo ""

# Calculate pass rate
passes=0
[ "$see_tags" = "y" ] && passes=$((passes + 1))
[ "$see_cats" = "y" ] && passes=$((passes + 1))
[ "$cycling" = "y" ] && passes=$((passes + 1))

echo "Core tests passed: $passes/3"
echo ""

if [ $passes -eq 3 ]; then
    echo "🎉 EXCELLENT! All core functionality working!"
    echo ""
    echo "✅ Tags & Categories validation PASSED"
    echo ""
    echo "You can now mark TODO item #1 as complete!"
else
    echo "⚠️  Some tests did not pass. Review the results above."
    echo ""
    echo "Issues to investigate:"
    [ "$see_tags" != "y" ] && echo "  - Tags not appearing in view"
    [ "$see_cats" != "y" ] && echo "  - Categories not appearing in view"
    [ "$cycling" != "y" ] && echo "  - View cycling not working"
fi

echo ""
echo "📝 For detailed tracking, see: TAGS_CATEGORIES_VALIDATION.md"
echo "🔍 For any issues found, see: TAGS_CATEGORIES_CODE_REVIEW.md"
echo ""
echo "Thank you for testing! 🚀"
