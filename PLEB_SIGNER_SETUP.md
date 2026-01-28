# Pleb_Signer Setup Guide

## Error: "No keys configured"

If you get this error when connecting, it means Pleb_Signer doesn't have any Nostr keys set up yet.

## Solution: Add a Key to Pleb_Signer

### Option 1: Generate a New Key

1. Open Pleb_Signer (from system tray or app menu)
2. Click "Add Key" or "Generate New Key"
3. Give it a name (e.g., "My Nostr Key")
4. Click "Generate"
5. **IMPORTANT**: Back up your nsec somewhere safe!

### Option 2: Import Existing Key

1. Open Pleb_Signer
2. Click "Import Key"
3. Paste your `nsec1...` private key
4. Give it a name
5. Click "Import"

## Verify Key is Added

Run the test script:

```bash
./test-plebsigner.sh
```

You should see your public key displayed.

## Test with NostrFeedz CLI

After adding a key:

```bash
# Run the test
go run ./cmd/test-signer/main.go

# If that works, use the main app
./nostrfeedz
# Choose option 1 - Pleb_Signer
```

## Troubleshooting

### "Pleb_Signer is locked"

Unlock it with your password:
1. Click Pleb_Signer in system tray
2. Enter your password
3. Try connecting again

### "Permission denied"

Grant permission to the app:
1. When you first try to sign, Pleb_Signer shows a dialog
2. Click "Allow" or "Always Allow"
3. For testing, enable "Auto-approve" in settings

### "Failed to connect to Pleb_Signer"

Make sure it's running:
```bash
ps aux | grep pleb-signer
```

If not running:
```bash
pleb-signer
```

## Key Management Best Practices

1. **Use a strong password** for Pleb_Signer
2. **Back up your nsec** immediately after generating
3. **Store nsec offline** in a secure password manager
4. **Never share your nsec** with anyone
5. **Use different keys** for testing vs. production

## Multiple Keys

Pleb_Signer supports multiple Nostr identities:

1. Open Pleb_Signer settings
2. Go to "Keys" tab
3. Add multiple keys
4. Switch between them as needed

NostrFeedz CLI will use your default/selected key.

## Security Note

Pleb_Signer encrypts your keys at rest using:
- **ChaCha20-Poly1305** encryption
- **Argon2** key derivation
- Your password as the encryption key

This means:
- ✅ Keys are safe even if someone accesses your computer
- ✅ Apps can't steal your keys
- ⚠️ **If you forget your password, keys are lost forever!**

Always keep a backup of your nsec in a secure location!
