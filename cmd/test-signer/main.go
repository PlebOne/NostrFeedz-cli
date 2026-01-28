package main

import (
"fmt"
"os"

"github.com/nbd-wtf/go-nostr"
nostrclient "github.com/plebone/nostrfeedz-cli/internal/nostr"
)

func main() {
fmt.Println("=== Pleb_Signer Connection Debug ===\n")

// Create Pleb_Signer client
fmt.Println("1. Connecting to Pleb_Signer...")
signer, err := nostrclient.NewPlebSignerClient("nostrfeedz-cli-debug")
if err != nil {
fmt.Printf("❌ Failed to connect: %v\n", err)
os.Exit(1)
}
defer signer.Close()
fmt.Println("✓ Connected")

// Check if ready
fmt.Println("\n2. Checking if Pleb_Signer is unlocked...")
ready, err := signer.IsReady()
if err != nil {
fmt.Printf("❌ Failed to check status: %v\n", err)
os.Exit(1)
}
if !ready {
fmt.Println("❌ Pleb_Signer is locked. Please unlock it first.")
os.Exit(1)
}
fmt.Println("✓ Unlocked and ready")

// Get public key
fmt.Println("\n3. Getting public key...")
pubkey, err := signer.GetPublicKey("")
if err != nil {
fmt.Printf("❌ Failed to get public key: %v\n", err)
os.Exit(1)
}
fmt.Printf("✓ Public key: %s\n", pubkey)

// Create a test event
fmt.Println("\n4. Creating test event...")
event := &nostr.Event{
Kind:      1,
Content:   "Test event from NostrFeedz CLI",
CreatedAt: nostr.Now(),
Tags:      nostr.Tags{},
}
fmt.Printf("✓ Event created (kind=%d, created_at=%d)\n", event.Kind, event.CreatedAt)

// Sign the event
fmt.Println("\n5. Signing event with Pleb_Signer...")
fmt.Println("   (This will show an approval dialog in Pleb_Signer)")
err = signer.SignEvent(event, "")
if err != nil {
fmt.Printf("❌ Failed to sign event: %v\n", err)
fmt.Println("\n   Troubleshooting:")
fmt.Println("   - Make sure you clicked 'Approve' in Pleb_Signer")
fmt.Println("   - Check if 'nostrfeedz-cli-debug' has permission")
fmt.Println("   - Try enabling auto-approve for testing")
os.Exit(1)
}
fmt.Println("✓ Event signed successfully")

// Verify the signed event
fmt.Println("\n6. Verifying signed event...")
fmt.Printf("   ID: %s\n", event.ID)
fmt.Printf("   PubKey: %s\n", event.PubKey)
if len(event.Sig) > 32 {
fmt.Printf("   Sig: %s...\n", event.Sig[:32])
} else {
fmt.Printf("   Sig: %s (len=%d)\n", event.Sig, len(event.Sig))
}
fmt.Printf("   CreatedAt: %d\n", event.CreatedAt)

// Check if fields are populated
if event.ID == "" || event.PubKey == "" || event.Sig == "" {
fmt.Println("\n❌ Event not fully signed!")
fmt.Printf("   ID empty: %v\n", event.ID == "")
fmt.Printf("   PubKey empty: %v\n", event.PubKey == "")
fmt.Printf("   Sig empty: %v\n", event.Sig == "")
fmt.Println("\n   This means Pleb_Signer returned an incomplete response.")
fmt.Println("   Enable debug mode in internal/nostr/plebsigner.go to see the raw JSON.")
os.Exit(1)
}

// Check signature validity
fmt.Println("\n7. Checking signature validity...")
ok, err := event.CheckSignature()
if err != nil {
fmt.Printf("❌ Signature check failed: %v\n", err)
fmt.Println("\n   This might indicate:")
fmt.Println("   - Event ID calculation mismatch")
fmt.Println("   - Signature format issue")
fmt.Println("   - Public key mismatch")
os.Exit(1)
}
if !ok {
fmt.Println("❌ Signature is invalid!")
fmt.Println("\n   Event details:")
fmt.Printf("   Kind: %d\n", event.Kind)
fmt.Printf("   Content: %s\n", event.Content)
fmt.Printf("   Tags: %v\n", event.Tags)
fmt.Printf("   Created: %d\n", event.CreatedAt)
fmt.Printf("   PubKey: %s\n", event.PubKey)
fmt.Printf("   ID: %s\n", event.ID)
fmt.Printf("   Sig: %s\n", event.Sig)
os.Exit(1)
}
fmt.Println("✓ Signature is valid!")

fmt.Println("\n=== All Tests Passed! ===")
fmt.Println("\nPleb_Signer integration is working correctly.")
fmt.Println("You can now use option 1 in NostrFeedz CLI.")
}
