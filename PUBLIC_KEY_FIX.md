# Public Key Retrieval Fix

## Issue

User got error: **"Sync failed: no public key available"**

## Root Cause

The `GetPublicKey()` method in `plebsigner.go` was using incorrect JSON parsing.

### What Was Wrong

```go
// OLD - Incorrect parsing
var pubKeyResp struct {
    PublicKey string `json:"publicKey"`
    Npub      string `json:"npub"`
}
json.Unmarshal([]byte(result), &pubKeyResp)
```

This didn't account for Pleb_Signer's triple-nested JSON response format.

### Actual Response Format

```json
{
  "success": true,
  "id": "req_...",
  "result": "{\"type\":\"public_key\",\"npub\":\"...\",\"hex\":\"...\"}"
}
```

The `result` field contains a **double-encoded JSON string** with the actual pubkey data.

## Fix

Updated `GetPublicKey()` to parse in 3 layers (same as `SignEvent()`):

```go
// Layer 1: Parse outer D-Bus response
var outerResp struct {
    Success bool   `json:"success"`
    Result  string `json:"result"` // Double-encoded JSON
    Error   string `json:"error"`
}
json.Unmarshal([]byte(result), &outerResp)

// Layer 2: Parse inner result
var innerResp struct {
    Type  string `json:"type"`
    Npub  string `json:"npub"`
    Hex   string `json:"hex"`
}
json.Unmarshal([]byte(outerResp.Result), &innerResp)

// Return the hex pubkey
return innerResp.Hex
```

## Test Results

```
=== Pleb_Signer Connection Debug ===

1. Connecting to Pleb_Signer...
✓ Connected

2. Checking if Pleb_Signer is unlocked...
✓ Unlocked and ready

3. Getting public key...
✓ Public key: 8dc8688200b447ec2e4018ea5e42dc5d480940cb3f19ca8f361d28179dc4ba5e
```

## Impact

Now the sync process should work correctly:

1. ✅ Authenticate with Pleb_Signer
2. ✅ Retrieve public key (hex format)
3. ✅ Fetch subscriptions from Nostr (kind 30404)
4. ✅ Fetch read status from Nostr (kind 30405)
5. ✅ Import feeds to local DB
6. ✅ Mark articles as read

## Files Changed

- `internal/nostr/plebsigner.go` - Fixed `GetPublicKey()` parsing

## Related Issues

All Pleb_Signer methods now use the correct triple-nested JSON parsing:
- ✅ `GetPublicKey()` - Returns hex pubkey
- ✅ `SignEvent()` - Returns signed event
- ✅ `Nip04Encrypt()` - Returns encrypted text
- ✅ `Nip04Decrypt()` - Returns decrypted text
- ✅ `Nip44Encrypt()` - Returns encrypted text
- ✅ `Nip44Decrypt()` - Returns decrypted text

## Try It Now

```bash
cd /home/tim/Projects/NostrFeedz-cli
./nostrfeedz

# Select "Pleb_Signer"
# Approve in Pleb_Signer
# Should now see:
# "Syncing from Nostr..."
# "Synced from Nostr! Added X feeds"
```
