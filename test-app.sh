#!/bin/bash
# Test script for NostrFeedz CLI

echo "Testing NostrFeedz CLI rendering..."
echo ""
echo "The app should display:"
echo "  - Title: 🚀 Nostr-Feedz CLI"
echo "  - Welcome message"
echo "  - Two options: 1 for Remote Signer, 2 for Private Key"
echo ""
echo "If you see a blank screen:"
echo "  - Try pressing any key (sometimes the initial render is delayed)"
echo "  - Try resizing your terminal window slightly"
echo "  - Check that your terminal is at least 80x24"
echo ""
echo "Current terminal size:"
tput cols 2>/dev/null && echo "Width: $(tput cols) columns" || echo "Width: Unable to detect"
tput lines 2>/dev/null && echo "Height: $(tput lines) lines" || echo "Height: Unable to detect"
echo ""
echo "Starting app in 2 seconds... (Press Ctrl+C to quit once inside)"
sleep 2
./nostrfeedz
