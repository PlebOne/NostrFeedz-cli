package nostr

import (
	"encoding/json"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/nbd-wtf/go-nostr"
)

const (
	dbusService   = "com.plebsigner.Signer"
	dbusPath      = "/com/plebsigner/Signer"
	dbusInterface = "com.plebsigner.Signer1"
)

// PlebSignerClient wraps D-Bus connection to Pleb_Signer
type PlebSignerClient struct {
	conn  *dbus.Conn
	obj   dbus.BusObject
	appID string
}

// NewPlebSignerClient creates a new client for Pleb_Signer
func NewPlebSignerClient(appID string) (*PlebSignerClient, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session bus: %w", err)
	}

	obj := conn.Object(dbusService, dbusPath)

	return &PlebSignerClient{
		conn:  conn,
		obj:   obj,
		appID: appID,
	}, nil
}

// IsReady checks if the signer is unlocked and ready
func (c *PlebSignerClient) IsReady() (bool, error) {
	var ready bool
	err := c.obj.Call(dbusInterface+".IsReady", 0).Store(&ready)
	if err != nil {
		return false, fmt.Errorf("failed to check if ready: %w", err)
	}
	return ready, nil
}

// GetVersion returns the signer version
func (c *PlebSignerClient) GetVersion() (string, error) {
	var version string
	err := c.obj.Call(dbusInterface+".Version", 0).Store(&version)
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}
	return version, nil
}

// GetPublicKey retrieves the public key from the signer
func (c *PlebSignerClient) GetPublicKey(keyID string) (string, error) {
	// Note: GetPublicKey takes no parameters - returns default key

	var result string
	err := c.obj.Call(dbusInterface+".GetPublicKey", 0).Store(&result)
	if err != nil {
		return "", fmt.Errorf("failed to get public key: %w", err)
	}

	// Parse outer response to check for errors
	var outerResp struct {
		Success bool   `json:"success"`
		Result  string `json:"result"` // Double-encoded JSON
		Error   string `json:"error"`
	}
	if err := json.Unmarshal([]byte(result), &outerResp); err != nil {
		return "", fmt.Errorf("failed to parse outer response: %w", err)
	}

	if !outerResp.Success {
		return "", fmt.Errorf("Pleb_Signer error: %s", outerResp.Error)
	}

	// Parse inner result
	var innerResp struct {
		Type  string `json:"type"`
		Npub  string `json:"npub"`
		Hex   string `json:"hex"`
	}
	if err := json.Unmarshal([]byte(outerResp.Result), &innerResp); err != nil {
		return "", fmt.Errorf("failed to parse inner result: %w", err)
	}

	return innerResp.Hex, nil
}

// SignEvent signs a Nostr event using Pleb_Signer
func (c *PlebSignerClient) SignEvent(event *nostr.Event, keyID string) error {
	// Note: keyID is ignored as it's not part of the D-Bus signature
	// If you need to specify a key, you'd need to configure it in Pleb_Signer directly
	
	// Marshal event to JSON (without signature fields)
	eventJSON, err := json.Marshal(map[string]interface{}{
		"kind":       event.Kind,
		"content":    event.Content,
		"tags":       event.Tags,
		"created_at": event.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	var result string
	// D-Bus signature is (ss): event_json, app_id
	err = c.obj.Call(dbusInterface+".SignEvent", 0, string(eventJSON), c.appID).Store(&result)
	if err != nil {
		return fmt.Errorf("failed to sign event: %w", err)
	}

	// Debug: log the raw response
	// Don't log the full response - it's working fine
	// fmt.Printf("DEBUG: Pleb_Signer response: %s\n", result)

	// Check for error response
	var errorResp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	json.Unmarshal([]byte(result), &errorResp)
	if !errorResp.Success && errorResp.Error != "" {
		return fmt.Errorf("Pleb_Signer error: %s", errorResp.Error)
	}

	// Parse outer response to get the result field (which is double-encoded JSON)
	var outerResp struct {
		Success bool   `json:"success"`
		Result  string `json:"result"` // This is a JSON string
	}
	if err := json.Unmarshal([]byte(result), &outerResp); err != nil {
		return fmt.Errorf("failed to parse outer response: %w", err)
	}

	// Parse the inner result (contains type, event_json, signature)
	var innerResp struct {
		Type      string `json:"type"`
		EventJSON string `json:"event_json"` // This is ALSO a JSON string
		Signature string `json:"signature"`
	}
	if err := json.Unmarshal([]byte(outerResp.Result), &innerResp); err != nil {
		return fmt.Errorf("failed to parse inner result: %w", err)
	}

	// Parse the actual signed event from event_json
	var signedEvent nostr.Event
	if err := json.Unmarshal([]byte(innerResp.EventJSON), &signedEvent); err != nil {
		return fmt.Errorf("failed to parse signed event: %w", err)
	}

	// Update ALL fields from the signed event to ensure consistency
	event.ID = signedEvent.ID
	event.PubKey = signedEvent.PubKey
	event.Sig = signedEvent.Sig
	event.CreatedAt = signedEvent.CreatedAt
	event.Kind = signedEvent.Kind
	event.Tags = signedEvent.Tags
	event.Content = signedEvent.Content

	return nil
}

// Nip04Encrypt encrypts a message using NIP-04
func (c *PlebSignerClient) Nip04Encrypt(plaintext, recipientPubkey, keyID string) (string, error) {
	// Note: keyID parameter ignored - not in D-Bus signature

	var result string
	err := c.obj.Call(dbusInterface+".Nip04Encrypt", 0, plaintext, recipientPubkey, c.appID).Store(&result)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	var encResp struct {
		Ciphertext string `json:"ciphertext"`
	}
	if err := json.Unmarshal([]byte(result), &encResp); err != nil {
		return "", fmt.Errorf("failed to parse encrypt response: %w", err)
	}

	return encResp.Ciphertext, nil
}

// Nip04Decrypt decrypts a message using NIP-04
func (c *PlebSignerClient) Nip04Decrypt(ciphertext, senderPubkey, keyID string) (string, error) {
	// Note: keyID parameter ignored - not in D-Bus signature

	var result string
	err := c.obj.Call(dbusInterface+".Nip04Decrypt", 0, ciphertext, senderPubkey, c.appID).Store(&result)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	var decResp struct {
		Plaintext string `json:"plaintext"`
	}
	if err := json.Unmarshal([]byte(result), &decResp); err != nil {
		return "", fmt.Errorf("failed to parse decrypt response: %w", err)
	}

	return decResp.Plaintext, nil
}

// Nip44Encrypt encrypts a message using NIP-44
func (c *PlebSignerClient) Nip44Encrypt(plaintext, recipientPubkey, keyID string) (string, error) {
	// Note: keyID parameter ignored - not in D-Bus signature

	var result string
	err := c.obj.Call(dbusInterface+".Nip44Encrypt", 0, plaintext, recipientPubkey, c.appID).Store(&result)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	var encResp struct {
		Ciphertext string `json:"ciphertext"`
	}
	if err := json.Unmarshal([]byte(result), &encResp); err != nil {
		return "", fmt.Errorf("failed to parse encrypt response: %w", err)
	}

	return encResp.Ciphertext, nil
}

// Nip44Decrypt decrypts a message using NIP-44
func (c *PlebSignerClient) Nip44Decrypt(ciphertext, senderPubkey, keyID string) (string, error) {
	// Note: keyID parameter ignored - not in D-Bus signature

	var result string
	err := c.obj.Call(dbusInterface+".Nip44Decrypt", 0, ciphertext, senderPubkey, c.appID).Store(&result)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	var decResp struct {
		Plaintext string `json:"plaintext"`
	}
	if err := json.Unmarshal([]byte(result), &decResp); err != nil {
		return "", fmt.Errorf("failed to parse decrypt response: %w", err)
	}

	return decResp.Plaintext, nil
}

// Close closes the D-Bus connection
func (c *PlebSignerClient) Close() error {
	return c.conn.Close()
}
