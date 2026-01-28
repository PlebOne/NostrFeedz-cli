package nostr

import (
	"context"
	"fmt"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

type Client struct {
	pool         *nostr.SimplePool
	relays       []string
	secretKey    string
	pubkey       string
	plebSigner   *PlebSignerClient
	signerType   string // "nsec", "plebsigner", or empty
}

// NewClient creates a new Nostr client with the given relays
func NewClient(relays []string) *Client {
	ctx := context.Background()
	return &Client{
		pool:   nostr.NewSimplePool(ctx),
		relays: relays,
	}
}

// SetPrivateKeySigner sets up signing using an nsec (private key)
func (c *Client) SetPrivateKeySigner(nsec string) error {
	if nsec == "" {
		return fmt.Errorf("private key is empty")
	}
	
	var hex string
	
	// Try to decode as nsec (bech32) first
	if len(nsec) > 4 && nsec[:4] == "nsec" {
		_, value, err := nip19.Decode(nsec)
		if err != nil {
			return fmt.Errorf("invalid nsec format: %w", err)
		}
		hex = value.(string)
	} else {
		// Assume it's already hex
		hex = nsec
	}
	
	pk, err := nostr.GetPublicKey(hex)
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}
	
	c.secretKey = hex
	c.pubkey = pk
	c.signerType = "nsec"
	return nil
}

// SetPlebSigner sets up signing using Pleb_Signer (D-Bus)
func (c *Client) SetPlebSigner() error {
	signer, err := NewPlebSignerClient("nostrfeedz-cli")
	if err != nil {
		return fmt.Errorf("failed to connect to Pleb_Signer: %w", err)
	}
	
	// Check if signer is ready
	ready, err := signer.IsReady()
	if err != nil {
		signer.Close()
		return fmt.Errorf("failed to check signer status: %w", err)
	}
	
	if !ready {
		signer.Close()
		return fmt.Errorf("Pleb_Signer is locked. Please unlock it first")
	}
	
	// Get public key
	pubkey, err := signer.GetPublicKey("")
	if err != nil {
		signer.Close()
		return fmt.Errorf("failed to get public key: %w", err)
	}
	
	c.plebSigner = signer
	c.pubkey = pubkey
	c.signerType = "plebsigner"
	return nil
}

// SetRemoteSigner sets up signing using NIP-46 remote signer (bunker)
// Note: This is a simplified version. Full NIP-46 support requires more work.
func (c *Client) SetRemoteSigner(bunkerURL, connectionToken string) error {
	if bunkerURL == "" {
		return fmt.Errorf("bunker URL is empty")
	}
	
	// For now, return an error indicating NIP-46 needs more implementation
	// This can be enhanced later with proper NIP-46 support
	return fmt.Errorf("NIP-46 remote signer not fully implemented yet. Please use nsec authentication for now")
}

// GetPublicKey returns the current user's public key (hex format)
func (c *Client) GetPublicKey() string {
	return c.pubkey
}

// SignEvent signs a Nostr event using the configured signer
func (c *Client) SignEvent(event *nostr.Event) error {
	switch c.signerType {
	case "nsec":
		if c.secretKey == "" {
			return fmt.Errorf("no private key configured")
		}
		event.PubKey = c.pubkey
		return event.Sign(c.secretKey)
		
	case "plebsigner":
		if c.plebSigner == nil {
			return fmt.Errorf("Pleb_Signer not connected")
		}
		return c.plebSigner.SignEvent(event, "")
		
	default:
		return fmt.Errorf("no signer configured")
	}
}

// PublishEvent publishes a signed event to all relays
func (c *Client) PublishEvent(event *nostr.Event) error {
	if c.secretKey == "" {
		return fmt.Errorf("no signer configured")
	}
	
	// Sign the event
	if err := c.SignEvent(event); err != nil {
		return err
	}
	
	// Publish to all relays
	ctx := context.Background()
	for _, relayURL := range c.relays {
		relay, err := c.pool.EnsureRelay(relayURL)
		if err != nil {
			fmt.Printf("Failed to connect to relay %s: %v\n", relayURL, err)
			continue
		}
		
		if err := relay.Publish(ctx, *event); err != nil {
			fmt.Printf("Failed to publish to %s: %v\n", relayURL, err)
		}
	}
	
	return nil
}

// QueryEvents queries events from relays based on filters
func (c *Client) QueryEvents(ctx context.Context, filter nostr.Filter) ([]*nostr.Event, error) {
	var events []*nostr.Event
	
	for ev := range c.pool.SubManyEose(ctx, c.relays, nostr.Filters{filter}) {
		if ev.Relay != nil {
			events = append(events, ev.Event)
		}
	}
	
	return events, nil
}

// Close closes connections
func (c *Client) Close() {
	if c.plebSigner != nil {
		c.plebSigner.Close()
	}
	// SimplePool doesn't have a Close method, but we should cleanup connections
	// This is handled automatically by the pool
}

// TestConnection tests if we can connect to relays and authenticate
func (c *Client) TestConnection() error {
	if c.signerType == "" {
		return fmt.Errorf("no signer configured")
	}
	
	// Create a test event
	event := &nostr.Event{
		Kind:      1,
		Content:   "NostrFeedz CLI test",
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{},
	}
	
	// Try to sign it
	if err := c.SignEvent(event); err != nil {
		return fmt.Errorf("failed to sign test event: %w", err)
	}
	
	// Try to connect to at least one relay
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	for _, relayURL := range c.relays {
		relay, err := c.pool.EnsureRelay(relayURL)
		if err != nil {
			continue
		}
		
		// Try to query our own profile (kind 0)
		filter := nostr.Filter{
			Kinds:   []int{0},
			Authors: []string{c.pubkey},
			Limit:   1,
		}
		
		events := make([]*nostr.Event, 0)
		for ev := range c.pool.SubManyEose(ctx, []string{relayURL}, nostr.Filters{filter}) {
			if ev.Relay != nil {
				events = append(events, ev.Event)
			}
		}
		
		// Successfully connected to at least one relay
		_ = relay
		return nil
	}
	
	return fmt.Errorf("failed to connect to any relay")
}
