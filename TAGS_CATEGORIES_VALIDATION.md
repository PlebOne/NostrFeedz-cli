# Tags & Categories Validation Checklist

This document tracks the validation of tags and categories functionality for NostrFeedz CLI.

## Test Status

**Test Date:** 2026-02-06  
**Tester:** _____  
**Version:** 0.1.0-dev

---

## Prerequisites ✓

- [ ] NostrFeedz CLI builds successfully (`go build -o nostrfeedz ./cmd/debug`)
- [ ] Have access to NostrFeedz.com account
- [ ] Pleb_Signer or nsec authentication working
- [ ] At least 3-5 feeds added to NostrFeedz.com

---

## Test 1: Tags Sync from NostrFeedz.com

### Setup
1. - [ ] Go to https://nostrfeedz.com
2. - [ ] Add tags to at least 3 feeds
   - Example tags: `tech`, `news`, `bitcoin`, `programming`
3. - [ ] Use the same tag on multiple feeds (e.g., `tech` on 2+ feeds)

### Execute
4. - [ ] Run `./nostrfeedz` in CLI
5. - [ ] Authenticate successfully
6. - [ ] Press `s` to sync from Nostr
7. - [ ] Observe sync message

### Expected Results
- [ ] Sync message shows: `Synced! Added: X feeds, Y tags, Z categories`
- [ ] Y (tags count) matches number of unique tags added
- [ ] No error messages displayed

### Verify in Database
```bash
sqlite3 ~/.local/share/nostrfeedz/feeds.db "SELECT * FROM tags;"
```
- [ ] All tags appear in database
- [ ] Tag IDs match format: `tag_<tagname>`

### Verify in UI
8. - [ ] Press `Tab` to switch to Tags view
9. - [ ] All your tags appear in the list
10. - [ ] Tag names display correctly

**Status:** ⬜ Not Started | ⏳ In Progress | ✅ Passed | ❌ Failed

**Notes:**
```
____________________________________________________
____________________________________________________
```

---

## Test 2: Tags View Display

### Execute
1. - [ ] In Tags view (press `Tab` from Feeds view)
2. - [ ] Navigate through tags with `↑`/`↓` or `j`/`k`
3. - [ ] Observe display format

### Expected Results
- [ ] Tags displayed with `🏷️` emoji prefix
- [ ] Tag names visible and readable
- [ ] Can navigate smoothly between tags
- [ ] Current selection highlighted

**Status:** ⬜ Not Started | ⏳ In Progress | ✅ Passed | ❌ Failed

**Notes:**
```
____________________________________________________
____________________________________________________
```

---

## Test 3: Multi-Feed Article Aggregation (Tags)

### Prerequisites
- Must have same tag on multiple feeds (set up in Test 1)

### Execute
1. - [ ] In Tags view, select a tag used on multiple feeds
2. - [ ] Press `Enter` to view articles
3. - [ ] Observe article list

### Expected Results
- [ ] Articles from ALL feeds with that tag are shown
- [ ] Articles are combined into one list
- [ ] Articles sorted by date (newest first)
- [ ] Can scroll through all articles
- [ ] No duplicate articles

### Verify
4. - [ ] Press `Esc` to go back
5. - [ ] Try another tag
6. - [ ] Repeat process

**Status:** ⬜ Not Started | ⏳ In Progress | ✅ Passed | ❌ Failed

**Notes:**
```
____________________________________________________
____________________________________________________
```

---

## Test 4: Categories Sync from NostrFeedz.com

### Setup
1. - [ ] Go to https://nostrfeedz.com
2. - [ ] Assign categories to at least 3 feeds
   - Example: `Technology`, `News`, `Entertainment`
3. - [ ] Leave at least 1-2 feeds WITHOUT a category (for Uncategorized test)

### Execute
4. - [ ] Run `./nostrfeedz` in CLI (or if already running, press `s` to sync)
5. - [ ] Press `s` to sync from Nostr
6. - [ ] Observe sync message

### Expected Results
- [ ] Sync message shows: `Synced! Added: X feeds, Y tags, Z categories`
- [ ] Z (categories count) matches number of feeds with categories
- [ ] No error messages

### Verify in Database
```bash
sqlite3 ~/.local/share/nostrfeedz/feeds.db "SELECT * FROM categories;"
sqlite3 ~/.local/share/nostrfeedz/feeds.db "SELECT url, category_id FROM feeds WHERE category_id != '';"
```
- [ ] All categories appear in database
- [ ] Feeds have category_id set correctly
- [ ] Category icons and colors stored (if provided)

**Status:** ⬜ Not Started | ⏳ In Progress | ✅ Passed | ❌ Failed

**Notes:**
```
____________________________________________________
____________________________________________________
```

---

## Test 5: Categories View Display

### Execute
1. - [ ] Press `Tab` twice from Feeds view to reach Categories view
2. - [ ] Observe category list
3. - [ ] Navigate through categories with `↑`/`↓`

### Expected Results
- [ ] First item is `📋 Uncategorized`
- [ ] Your categories appear below Uncategorized
- [ ] Categories display with icons (if set in NostrFeedz.com)
- [ ] Category names visible and readable
- [ ] Can navigate smoothly

**Status:** ⬜ Not Started | ⏳ In Progress | ✅ Passed | ❌ Failed

**Notes:**
```
____________________________________________________
____________________________________________________
```

---

## Test 6: Uncategorized Category

### Prerequisites
- Must have at least 1 feed WITHOUT a category (set up in Test 4)

### Execute
1. - [ ] In Categories view (press `Tab` twice)
2. - [ ] First item should be `📋 Uncategorized`
3. - [ ] Press `Enter` on Uncategorized

### Expected Results
- [ ] Articles from uncategorized feeds are shown
- [ ] Only feeds without categories are included
- [ ] Can read articles normally
- [ ] Can navigate back with `Esc`

**Status:** ⬜ Not Started | ⏳ In Progress | ✅ Passed | ❌ Failed

**Notes:**
```
____________________________________________________
____________________________________________________
```

---

## Test 7: Multi-Feed Article Aggregation (Categories)

### Prerequisites
- Must have a category with multiple feeds

### Execute
1. - [ ] In Categories view, select a category with multiple feeds
2. - [ ] Press `Enter` to view articles
3. - [ ] Observe article list

### Expected Results
- [ ] Articles from ALL feeds in that category are shown
- [ ] Articles combined into one list
- [ ] Articles sorted by date (newest first)
- [ ] Can scroll through all articles
- [ ] No duplicate articles

**Status:** ⬜ Not Started | ⏳ In Progress | ✅ Passed | ❌ Failed

**Notes:**
```
____________________________________________________
____________________________________________________
```

---

## Test 8: View Mode Cycling

### Execute
1. - [ ] Start in Feeds view
2. - [ ] Press `Tab` - should go to Tags view
3. - [ ] Press `Tab` - should go to Categories view
4. - [ ] Press `Tab` - should cycle back to Feeds view
5. - [ ] Repeat cycle several times

### Expected Results
- [ ] View mode cycles correctly: Feeds → Tags → Categories → Feeds
- [ ] Each view displays correct content
- [ ] No crashes or UI glitches
- [ ] Selection resets appropriately when switching views

**Status:** ⬜ Not Started | ⏳ In Progress | ✅ Passed | ❌ Failed

**Notes:**
```
____________________________________________________
____________________________________________________
```

---

## Test 9: Sync Updates

### Execute
1. - [ ] With app running, go to NostrFeedz.com
2. - [ ] Add a new tag to a feed
3. - [ ] Add a new category to a feed
4. - [ ] In CLI, press `s` to sync
5. - [ ] Check if new tags/categories appear

### Expected Results
- [ ] New tags appear in Tags view after sync
- [ ] New categories appear in Categories view after sync
- [ ] Existing data not corrupted
- [ ] Sync message reflects changes

**Status:** ⬜ Not Started | ⏳ In Progress | ✅ Passed | ❌ Failed

**Notes:**
```
____________________________________________________
____________________________________________________
```

---

## Test 10: Edge Cases

### Test 10a: Feed with Multiple Tags
- [ ] Feed has 3+ tags
- [ ] Feed appears when viewing each tag
- [ ] Articles appear in all tag views

### Test 10b: Empty Tags/Categories
- [ ] Tag with no feeds shows empty article list
- [ ] Category with no feeds shows empty article list
- [ ] No crash or error

### Test 10c: Special Characters in Names
- [ ] Tag name with space (e.g., `machine learning`)
- [ ] Category name with emoji (e.g., `Tech 💻`)
- [ ] Both display and work correctly

### Test 10d: Large Number of Tags/Categories
- [ ] 10+ tags display correctly
- [ ] Can scroll through full list
- [ ] Performance is acceptable

**Status:** ⬜ Not Started | ⏳ In Progress | ✅ Passed | ❌ Failed

**Notes:**
```
____________________________________________________
____________________________________________________
```

---

## Known Issues

### Issue Template
**Issue:** _______________________________________  
**Severity:** 🔴 Critical | 🟡 Medium | 🟢 Low  
**Steps to Reproduce:**
1. ___
2. ___

**Expected:** ___  
**Actual:** ___  
**Workaround:** ___

---

## Issue #1
_[Document any issues found during testing]_

---

## Issue #2
_[Document any issues found during testing]_

---

## Summary

### Overall Status
⬜ Not Started | ⏳ In Progress | ✅ All Tests Passed | ❌ Has Failures

### Test Results
- Total Tests: 10
- Passed: __/10
- Failed: __/10
- Blocked: __/10

### Recommendations
- [ ] Ready for production use
- [ ] Needs minor fixes
- [ ] Needs major fixes
- [ ] Not ready

### Next Steps
```
1. _______________________________________________
2. _______________________________________________
3. _______________________________________________
```

---

## Test Environment

**OS:** Linux (specify distro: _______)  
**Terminal:** (e.g., Kitty, Alacritty, GNOME Terminal)  
**Go Version:** `go version` output: __________  
**Database Path:** `~/.local/share/nostrfeedz/feeds.db`  
**Config Path:** `~/.config/nostrfeedz/config.yaml`  

**Nostr Relays:**
- [ ] relay.damus.io
- [ ] nos.lol
- [ ] relay.snort.social
- [ ] Other: ____________

---

**Validation completed by:** _______________  
**Date:** _______________  
**Sign-off:** _______________
