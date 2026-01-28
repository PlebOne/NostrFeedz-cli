# NIP-55 Support via Pleb_Signer

## Overview

NostrFeedz CLI supports [NIP-55](https://github.com/nostr-protocol/nips/blob/master/55.md) (Nostr Signer API) through integration with [Pleb_Signer](https://github.com/PlebOne/Pleb_Signer), a Linux desktop signer that implements the NIP-55 protocol via D-Bus.

## What is NIP-55?

NIP-55 is a protocol for applications to request Nostr event signing from external signer applications. Originally designed for Android (implemented by Amber), Pleb_Signer brings this functionality to Linux desktop environments using D-Bus instead of Android intents.

## Features

- **Secure Key Storage**: Keys are encrypted at rest and never exposed to the CLI
- **User Approval**: Each signature requires explicit user approval (or auto-approve for trusted apps)
- **Multiple Keys**: Support for multiple Nostr identities
- **Full NIP Support**: NIP-04 and NIP-44 encryption/decryption

## Setup

### 1. Install Pleb_Signer

```bash
# Install dependencies (Debian/Ubuntu)
sudo apt install libdbus-1-dev libssl-dev pkg-config

# Clone and build
git clone https://github.com/PlebOne/Pleb_Signer.git
cd Pleb_Signer
cargo build --release

# Install
sudo cp target/release/pleb-signer /usr/local/bin/
```

### 2. Start Pleb_Signer

```bash
# Start the signer
pleb-signer

# Or start minimized to system tray
pleb-signer --minimized
```

### 3. Set Up Your Key

1. Launch Pleb_Signer
2. Create a password
3. Generate a new key or import existing nsec

### 4. Run NostrFeedz CLI

```bash
./nostrfeedz
```

Choose option **1 - Pleb_Signer (D-Bus)** when prompted.

## How It Works

### D-Bus Communication

NostrFeedz CLI communicates with Pleb_Signer over D-Bus:

```
NostrFeedz CLI → D-Bus → Pleb_Signer → User Approval → Signature
```

**D-Bus Service:**
- Service: `com.plebsigner.Signer`
- Object Path: `/com/plebsigner/Signer`
- Interface: `com.plebsigner.Signer1`

### Signing Flow

1. **NostrFeedz creates an event** (e.g., publishing a subscription list)
2. **Sends event to Pleb_Signer** via D-Bus
3. **Pleb_Signer shows approval dialog** to user
4. **User approves** (or auto-approved if trusted)
5. **Pleb_Signer signs** with encrypted key
6. **Returns signed event** to NostrFeedz
7. **NostrFeedz publishes** to relays

### Supported Operations

| Operation | Description |
|-----------|-------------|
| `GetPublicKey` | Get user's public key |
| `SignEvent` | Sign a Nostr event |
| `Nip04Encrypt` | Encrypt with NIP-04 |
| `Nip04Decrypt` | Decrypt with NIP-04 |
| `Nip44Encrypt` | Encrypt with NIP-44 |
| `Nip44Decrypt` | Decrypt with NIP-44 |

## Configuration

After first successful connection, config is saved to `~/.config/nostrfeedz/config.yaml`:

```yaml
nostr:
  npub: "npub1..."  # Auto-filled
  pleb_signer:
    enabled: true
    key_id: ""      # Optional: specific key to use
```

## Permissions

You may need to authorize NostrFeedz CLI in Pleb_Signer:

1. Open Pleb_Signer settings
2. Go to "Permissions"
3. Find "nostrfeedz-cli"
4. Grant necessary permissions:
   - Sign events
   - NIP-04 operations (for DMs)
   - NIP-44 operations (for improved encryption)

## Auto-Approve

For convenience, you can enable auto-approve for NostrFeedz:

1. Open Pleb_Signer
2. Settings → Permissions → nostrfeedz-cli
3. Enable "Auto-approve"
4. Set rate limits if desired

⚠️ **Security Note**: Only enable auto-approve if you trust the application. All events will be signed without confirmation.

## Troubleshooting

### "Failed to connect to Pleb_Signer"

**Possible causes:**
- Pleb_Signer not running
- D-Bus session bus not available
- Permission issues

**Solutions:**
```bash
# Check if Pleb_Signer is running
ps aux | grep pleb-signer

# Check D-Bus
dbus-send --session --dest=com.plebsigner.Signer \
  --type=method_call --print-reply \
  /com/plebsigner/Signer \
  com.plebsigner.Signer1.IsReady

# Start Pleb_Signer
pleb-signer
```

### "Pleb_Signer is locked"

Pleb_Signer requires unlocking after system start.

**Solution:**
1. Open Pleb_Signer from system tray or app menu
2. Enter your password
3. Try connecting again

### "Permission denied"

NostrFeedz CLI needs permission to use Pleb_Signer.

**Solution:**
1. The first signature request will trigger a permission dialog
2. Click "Allow" or "Always Allow"
3. Operation will proceed

## Advantages Over nsec

| Feature | Pleb_Signer | nsec |
|---------|-------------|------|
| **Key Security** | Encrypted, never exposed | Stored in config file |
| **User Control** | Explicit approval per action | Automatic signing |
| **Multiple Keys** | Yes | One per config |
| **Audit Trail** | Yes, in Pleb_Signer | No |
| **Revocation** | Can revoke app access | Must change key |

## NIP-55 Compatibility

While NIP-55 was designed for Android (using intents), Pleb_Signer adapts it for Linux:

| NIP-55 (Android) | Pleb_Signer (Linux) |
|------------------|---------------------|
| Android Intents | D-Bus Method Calls |
| Content Resolver | D-Bus Properties/Methods |
| Package Manager | D-Bus Service Names |

The core protocol and operations remain the same, ensuring compatibility with NIP-55's design principles.

## Development

### Testing D-Bus Connection

```bash
# Check version
dbus-send --session --dest=com.plebsigner.Signer \
  --type=method_call --print-reply \
  /com/plebsigner/Signer \
  com.plebsigner.Signer1.Version

# Get public key
dbus-send --session --dest=com.plebsigner.Signer \
  --type=method_call --print-reply \
  /com/plebsigner/Signer \
  com.plebsigner.Signer1.GetPublicKey string:""
```

### Code Example

See `internal/nostr/plebsigner.go` for the full D-Bus client implementation.

## Related NIPs

- **NIP-55**: Android Signer Application
- **NIP-46**: Nostr Connect (remote signing protocol)
- **NIP-04**: Encrypted Direct Messages
- **NIP-44**: Encrypted Payloads (Versioned)

## Resources

- [Pleb_Signer GitHub](https://github.com/PlebOne/Pleb_Signer)
- [NIP-55 Specification](https://github.com/nostr-protocol/nips/blob/master/55.md)
- [Amber (Android)](https://github.com/greenart7c3/Amber)
- [D-Bus Specification](https://dbus.freedesktop.org/doc/dbus-specification.html)
