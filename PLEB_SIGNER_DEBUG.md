# Pleb_Signer Signature Mismatch - Troubleshooting

## The Issue

When connecting to Pleb_Signer, you may encounter:
```
Connection test failed: failed to sign test event: Signature mismatch
```

## What This Means

The event was signed by Pleb_Signer, but when go-nostr validates the signature, it doesn't match. This usually means the event fields aren't perfectly synchronized between what was sent and what was signed.

## Recent Fix Applied

The code has been updated to synchronize ALL event fields after signing:
- `ID` - Event identifier
- `PubKey` - Public key
- `Sig` - Signature
- `CreatedAt` - Timestamp (may be adjusted by signer)
- `Kind`, `Tags`, `Content` - Ensure exact match

## Test the Fix

Run the dedicated test tool:

```bash
go run ./cmd/test-signer/main.go
```

This will:
1. Connect to Pleb_Signer
2. Check if it's unlocked
3. Get your public key
4. Create and sign a test event
5. Verify the signature is valid

## Expected Output

```
=== Pleb_Signer Connection Debug ===

1. Connecting to Pleb_Signer...
✓ Connected

2. Checking if Pleb_Signer is unlocked...
✓ Unlocked and ready

3. Getting public key...
✓ Public key: abc123...

4. Creating test event...
✓ Event created (kind=1, created_at=1234567890)

5. Signing event with Pleb_Signer...
   (This will show an approval dialog in Pleb_Signer)
✓ Event signed successfully

6. Verifying signed event...
   ID: def456...
   PubKey: abc123...
   Sig: 789xyz...
   CreatedAt: 1234567890

7. Checking signature validity...
✓ Signature is valid!

=== All Tests Passed! ===
```

## If Still Failing

### 1. Check Pleb_Signer Version

```bash
dbus-send --session --dest=com.plebsigner.Signer \
  --type=method_call --print-reply \
  /com/plebsigner/Signer \
  com.plebsigner.Signer1.Version
```

Make sure you're using the latest version of Pleb_Signer.

### 2. Enable Debug Mode

Uncomment the debug line in `internal/nostr/plebsigner.go`:

```go
// Debug: log the raw response
fmt.Printf("DEBUG: Pleb_Signer response: %s\n", result)
```

Then rebuild and run:
```bash
go build -o nostrfeedz ./cmd/nostrfeedz
./nostrfeedz
```

This will show the exact JSON response from Pleb_Signer.

### 3. Check Event Structure

The issue might be in how Pleb_Signer returns the signed event. The expected JSON structure is:

```json
{
  "event_json": "...",
  "event": {
    "id": "...",
    "pubkey": "...",
    "created_at": 1234567890,
    "kind": 1,
    "tags": [],
    "content": "...",
    "sig": "..."
  }
}
```

If Pleb_Signer returns a different structure, the parsing will fail.

### 4. Verify go-nostr Compatibility

The signature verification uses go-nostr's `CheckSignature()` method. Make sure:
- Event ID is calculated correctly (SHA-256 of canonical JSON)
- Signature is in correct format (Schnorr signature)
- Public key matches the signer

### 5. Test with Direct D-Bus

Try signing manually via D-Bus to see the raw response:

```bash
# Create a test event JSON
EVENT='{"kind":1,"content":"test","tags":[],"created_at":'$(date +%s)'}'

# Sign it
dbus-send --session --dest=com.plebsigner.Signer \
  --type=method_call --print-reply \
  /com/plebsigner/Signer \
  com.plebsigner.Signer1.SignEvent \
  string:"$EVENT" string:"" string:"test-app"
```

Compare the response structure with what the code expects.

## Common Causes

1. **Timestamp Mismatch**: Pleb_Signer might adjust `created_at` for clock skew
   - **Fix**: Copy ALL fields from signed event (done in latest version)

2. **Tag Ordering**: Tags might be reordered
   - **Fix**: Use exact tags from signed event

3. **Content Encoding**: Unicode or escape sequences might differ
   - **Fix**: Use exact content from signed event

4. **PubKey Format**: Hex vs npub
   - **Fix**: Ensure using hex format (64 characters)

## Workaround: Use nsec Temporarily

While we debug, you can use private key authentication:

```bash
./nostrfeedz
# Choose option 3 - Private Key
# Enter your nsec
```

This bypasses Pleb_Signer entirely.

## Report the Issue

If the test tool still fails, please report with:

1. **Pleb_Signer version**: From D-Bus call above
2. **Debug output**: Raw JSON response
3. **Event details**: All fields from the test output
4. **go-nostr version**: Check `go.mod`

## What Was Changed

File: `internal/nostr/plebsigner.go`

```go
// Before - Only some fields
event.ID = signedResp.Event.ID
event.PubKey = signedResp.Event.PubKey
event.Sig = signedResp.Event.Sig

// After - ALL fields
event.ID = signedResp.Event.ID
event.PubKey = signedResp.Event.PubKey
event.Sig = signedResp.Event.Sig
event.CreatedAt = signedResp.Event.CreatedAt  // Added
event.Kind = signedResp.Event.Kind            // Added
event.Tags = signedResp.Event.Tags            // Added
event.Content = signedResp.Event.Content      // Added
```

This ensures perfect synchronization between the CLI's event object and what Pleb_Signer signed.

## Next Steps

1. Rebuild: `go build -o nostrfeedz ./cmd/nostrfeedz`
2. Test: `go run ./cmd/test-signer/main.go`
3. If passing: Use NostrFeedz CLI normally
4. If failing: Enable debug mode and report the issue
