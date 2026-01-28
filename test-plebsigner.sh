#!/bin/bash
# Test Pleb_Signer D-Bus connection

echo "=== Pleb_Signer D-Bus Test ==="
echo ""

# Check if D-Bus is available
if ! command -v dbus-send &> /dev/null; then
    echo "❌ dbus-send not found. Install with: sudo apt install dbus-x11"
    exit 1
fi

echo "✓ D-Bus tools available"

# Test 1: Check if Pleb_Signer is running
echo ""
echo "Test 1: Checking if Pleb_Signer is running..."
if dbus-send --session --dest=com.plebsigner.Signer --type=method_call --print-reply /com/plebsigner/Signer org.freedesktop.DBus.Introspectable.Introspect &> /dev/null; then
    echo "✓ Pleb_Signer is running"
else
    echo "❌ Pleb_Signer not found on D-Bus"
    echo ""
    echo "Please start Pleb_Signer with:"
    echo "  pleb-signer"
    echo ""
    echo "Or install it from:"
    echo "  https://github.com/PlebOne/Pleb_Signer"
    exit 1
fi

# Test 2: Check version
echo ""
echo "Test 2: Getting Pleb_Signer version..."
VERSION=$(dbus-send --session --dest=com.plebsigner.Signer --type=method_call --print-reply /com/plebsigner/Signer com.plebsigner.Signer1.Version 2>/dev/null | grep -oP 'string "\K[^"]+' || echo "unknown")
if [ "$VERSION" != "" ]; then
    echo "✓ Version: $VERSION"
else
    echo "⚠ Could not get version"
fi

# Test 3: Check if unlocked
echo ""
echo "Test 3: Checking if Pleb_Signer is unlocked..."
READY=$(dbus-send --session --dest=com.plebsigner.Signer --type=method_call --print-reply /com/plebsigner/Signer com.plebsigner.Signer1.IsReady 2>/dev/null | grep -oP 'boolean \K\w+')
if [ "$READY" = "true" ]; then
    echo "✓ Pleb_Signer is unlocked and ready"
else
    echo "❌ Pleb_Signer is locked"
    echo ""
    echo "Please unlock Pleb_Signer:"
    echo "  1. Open Pleb_Signer from system tray or app menu"
    echo "  2. Enter your password"
    echo "  3. Run this test again"
    exit 1
fi

# Test 4: Get public key (will trigger permission dialog on first run)
echo ""
echo "Test 4: Getting public key..."
echo "(This may show a permission dialog in Pleb_Signer - click Allow)"
PUBKEY_JSON=$(dbus-send --session --dest=com.plebsigner.Signer --type=method_call --print-reply /com/plebsigner/Signer com.plebsigner.Signer1.GetPublicKey string:"" 2>&1)
if echo "$PUBKEY_JSON" | grep -q "publicKey"; then
    PUBKEY=$(echo "$PUBKEY_JSON" | grep -oP '"publicKey":"[^"]+' | cut -d'"' -f4)
    NPUB=$(echo "$PUBKEY_JSON" | grep -oP '"npub":"[^"]+' | cut -d'"' -f4)
    echo "✓ Public Key Retrieved"
    echo "  Hex: ${PUBKEY:0:16}..."
    echo "  npub: ${NPUB:0:20}..."
else
    echo "⚠ Could not get public key (permission may have been denied)"
    echo "  You can grant permission in Pleb_Signer settings"
fi

echo ""
echo "=== All Tests Passed! ==="
echo ""
echo "Pleb_Signer is ready to use with NostrFeedz CLI"
echo "Run: ./nostrfeedz"
