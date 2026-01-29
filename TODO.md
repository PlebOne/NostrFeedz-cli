# NostrFeedz CLI - TODO & Future Tasks

## Current Status
NostrFeedz CLI is functional with core features implemented. This document tracks remaining tasks and future enhancements.

## High Priority Tasks

### 1. Tags & Categories Validation
- [ ] Test tags sync after NostrFeedz.com changes
- [ ] Verify categories display correctly
- [ ] Ensure Uncategorized category works properly
- [ ] Test multi-feed article aggregation for tags/categories

### 2. Image Viewing Polish
- [ ] Remove debug output from production build
- [ ] Test image cycling with arrow keys thoroughly
- [ ] Verify image cache cleanup works (30-day expiration)
- [ ] Test cache size limit enforcement (500MB)

### 3. Video Player Improvements
- [ ] Test YouTube Shorts playback with mpv
- [ ] Verify video player PID tracking and cleanup
- [ ] Test video cycling with Shift+arrow keys
- [ ] Ensure mpv config (~/.config/mpv/mpv.conf) persists

## Medium Priority Tasks

### UI/UX Improvements
- [ ] Add keyboard shortcuts help screen (press `?`)
- [ ] Improve status messages (more descriptive)
- [ ] Add loading indicators for long operations
- [ ] Better error messages for sync failures
- [ ] Add confirmation dialogs for destructive actions

### Feed Management
- [ ] Add feed from CLI (without Nostr sync)
- [ ] Delete/remove feeds
- [ ] Edit feed metadata (title, description)
- [ ] Refresh single feed (fetch new articles)
- [ ] Mark all as read functionality

### Article Features
- [ ] Search within articles
- [ ] Filter articles by date range
- [ ] Export articles (markdown, text, HTML)
- [ ] Article bookmarks/favorites sync to Nostr
- [ ] Full-text search across all articles

### Tags & Categories Management
- [ ] Add/edit tags directly in CLI
- [ ] Create categories from CLI
- [ ] Assign tags to feeds from CLI
- [ ] Move feeds between categories
- [ ] Bulk tag/category operations
- [ ] Tag/category color customization

## Low Priority / Future Enhancements

### Performance Optimization
- [ ] Lazy loading for large feed lists
- [ ] Virtual scrolling for article lists
- [ ] Background feed updates (periodic sync)
- [ ] Database indexing optimization
- [ ] Memory usage profiling and optimization

### Advanced Features
- [ ] Podcast support (audio player integration)
- [ ] Article recommendations based on reading history
- [ ] Offline mode improvements
- [ ] Import/export OPML
- [ ] Multi-user support (different profiles)
- [ ] Theme customization (colors, fonts)

### Platform Support
- [ ] Windows support testing
- [ ] macOS support testing
- [ ] Package for various Linux distros (deb, rpm, AUR)
- [ ] Flatpak/Snap distribution
- [ ] Docker container

### Sync & Nostr Integration
- [ ] Two-way sync (push local changes to Nostr)
- [ ] Conflict resolution for sync
- [ ] Selective sync (choose what to sync)
- [ ] Multiple Nostr relay support
- [ ] Nostr DM notifications for new articles
- [ ] Share articles to Nostr (kind 1 posts)

### Developer Experience
- [ ] Comprehensive test suite
- [ ] CI/CD pipeline
- [ ] Automated releases
- [ ] Contribution guidelines
- [ ] API documentation
- [ ] Plugin system for extensibility

## Known Issues

### Bugs to Fix
- [ ] Unread counts not updating immediately after reading article (requires ESC back to feeds)
- [ ] Terminal image artifacts on some terminals (Kitty protocol)
- [ ] Large articles may cause UI slowdown
- [ ] Video player windows accumulate on some WMs

### Compatibility Issues
- [ ] Terminal compatibility matrix needed
- [ ] Image viewer fallback chain not exhaustive
- [ ] Video player detection could be more robust
- [ ] mpv YouTube playback requires yt-dlp update

## Documentation Needs

### User Documentation
- [x] Quick Start Guide
- [x] NIP-55 Setup Guide
- [x] Image Viewing Guide
- [x] Video Player Guide
- [x] Tags and Categories Guide
- [x] Setting Up Tags and Categories
- [ ] Troubleshooting Guide (comprehensive)
- [ ] FAQ
- [ ] Keyboard Shortcuts Reference
- [ ] Configuration Guide

### Developer Documentation
- [ ] Architecture Overview
- [ ] Database Schema
- [ ] API Reference
- [ ] Contributing Guide
- [ ] Testing Guide
- [ ] Release Process

## Recently Completed ✅

### Session 1-4 (NIP-55 & Core Features)
- [x] NIP-55 Pleb_Signer integration
- [x] Remote signer (NIP-46) support
- [x] Private key authentication
- [x] Feed syncing from Nostr
- [x] Article fetching and display
- [x] Read status sync to Nostr

### Session 5 (Image & Video Support)
- [x] Image caching system
- [x] External image viewer integration (sxiv, feh, imv, eog)
- [x] Image navigation (arrow keys)
- [x] Video player support (mpv, vlc, mplayer)
- [x] YouTube video detection and playback
- [x] YouTube Shorts support
- [x] Video navigation (Shift+arrow keys)
- [x] Process tracking for external viewers/players
- [x] Tiling window manager support (PID tracking)

### Session 6 (Organization Features)
- [x] Tags view and filtering
- [x] Categories view and filtering
- [x] Uncategorized category
- [x] Multi-feed article aggregation
- [x] Unread count badges on feeds
- [x] Sync status reporting (feeds, tags, categories)

## Priority Ordering

### Immediate (This Week)
1. Validate tags/categories work with NostrFeedz.com
2. Remove debug output for cleaner production experience
3. Write comprehensive troubleshooting guide

### Short Term (This Month)
1. Keyboard shortcuts help screen
2. Search within articles
3. Feed management (add/delete/refresh)
4. Better error handling and messages

### Medium Term (This Quarter)
1. Performance optimization
2. Platform testing and packaging
3. Two-way sync with Nostr
4. Comprehensive test suite

### Long Term (Future)
1. Plugin system
2. Advanced features (recommendations, offline mode)
3. Mobile companion app integration
4. Multi-language support

## Success Metrics

Goals for v1.0 release:
- [ ] All high-priority tasks completed
- [ ] Zero critical bugs
- [ ] Documentation complete
- [ ] Tested on 3+ Linux distros
- [ ] 90%+ of features working smoothly
- [ ] User feedback incorporated

## Contributing

Interested in helping? Check out these good first issues:
- Keyboard shortcuts help screen
- OPML import/export
- Theme customization
- Additional image viewer support
- Testing on different platforms

---

**Last Updated:** 2026-01-29
**Version:** 0.1.0-dev
**Status:** Active Development
