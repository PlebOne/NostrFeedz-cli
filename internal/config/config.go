package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Nostr   NostrConfig   `mapstructure:"nostr"`
	Sync    SyncConfig    `mapstructure:"sync"`
	Reading ReadingConfig `mapstructure:"reading"`
	Display DisplayConfig `mapstructure:"display"`
	Database DatabaseConfig `mapstructure:"database"`
}

type NostrConfig struct {
	NPUB         string              `mapstructure:"npub"`
	NSEC         string              `mapstructure:"nsec"`
	Relays       []string            `mapstructure:"relays"`
	RemoteSigner RemoteSignerConfig  `mapstructure:"remote_signer"`
	PlebSigner   PlebSignerConfig    `mapstructure:"pleb_signer"`
}

type RemoteSignerConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	BunkerURL       string `mapstructure:"bunker_url"`
	ConnectionToken string `mapstructure:"connection_token"`
}

type PlebSignerConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	KeyID   string `mapstructure:"key_id"` // Optional: specific key to use
}

type SyncConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	AutoSyncInterval string `mapstructure:"auto_sync_interval"`
}

type ReadingConfig struct {
	MarkReadBehavior   string `mapstructure:"mark_read_behavior"`
	OrganizationMode   string `mapstructure:"organization_mode"`
}

type DisplayConfig struct {
	Theme             string `mapstructure:"theme"`
	FeedListWidth     int    `mapstructure:"feed_list_width"`
	ArticleListWidth  int    `mapstructure:"article_list_width"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

var DefaultRelays = []string{
	"wss://relay.damus.io",
	"wss://nos.lol",
	"wss://relay.snort.social",
	"wss://relay.nostr.band",
	"wss://nostr-pub.wellorder.net",
}

func Load() (*Config, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	// Set defaults
	viper.SetDefault("nostr.relays", DefaultRelays)
	viper.SetDefault("nostr.remote_signer.enabled", false)
	viper.SetDefault("nostr.pleb_signer.enabled", false)
	viper.SetDefault("sync.enabled", true)
	viper.SetDefault("sync.auto_sync_interval", "15m")
	viper.SetDefault("reading.mark_read_behavior", "on-open")
	viper.SetDefault("reading.organization_mode", "tags")
	viper.SetDefault("display.theme", "default")
	viper.SetDefault("display.feed_list_width", 30)
	viper.SetDefault("display.article_list_width", 40)
	
	dbPath := filepath.Join(getDataDir(), "feeds.db")
	viper.SetDefault("database.path", dbPath)

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// Config file not found, create default one
		if err := createDefaultConfig(configDir); err != nil {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	viper.Set("nostr", cfg.Nostr)
	viper.Set("sync", cfg.Sync)
	viper.Set("reading", cfg.Reading)
	viper.Set("display", cfg.Display)
	viper.Set("database", cfg.Database)

	configPath := filepath.Join(configDir, "config.yaml")
	return viper.WriteConfigAs(configPath)
}

func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "nostrfeedz"), nil
}

func getDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "nostrfeedz")
}

func createDefaultConfig(configDir string) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	defaultConfig := `# Nostr-Feedz CLI Configuration

# User Identity
nostr:
  npub: ""                      # Your Nostr public key
  nsec: ""                      # Your Nostr private key (optional, use remote signer instead)
  relays:
    - "wss://relay.damus.io"
    - "wss://nos.lol"
    - "wss://relay.snort.social"
    - "wss://relay.nostr.band"
    - "wss://nostr-pub.wellorder.net"
  
  # Remote Signer (NIP-46) - For desktop remote signers
  remote_signer:
    enabled: false              # Set to true to use remote signer
    bunker_url: ""              # e.g., bunker://<pubkey>?relay=wss://relay.nsecbunker.com
    connection_token: ""        # Connection secret token
  
  # Pleb_Signer (NIP-55 via D-Bus) - Recommended for Linux
  pleb_signer:
    enabled: false              # Set to true to use Pleb_Signer
    key_id: ""                  # Optional: specific key ID to use

# Sync Settings
sync:
  enabled: true
  auto_sync_interval: "15m"     # Auto-sync every 15 minutes

# Reading Preferences
reading:
  mark_read_behavior: "on-open" # "on-open" | "after-10s" | "never"
  organization_mode: "tags"     # "tags" | "categories"

# Display
display:
  theme: "default"              # "default" | "dark" | "light"
  feed_list_width: 30
  article_list_width: 40

# Database
database:
  path: "~/.local/share/nostrfeedz/feeds.db"
`

	configPath := filepath.Join(configDir, "config.yaml")
	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}

func GetDatabasePath(cfg *Config) string {
	dbPath := cfg.Database.Path
	if dbPath == "" {
		dbPath = filepath.Join(getDataDir(), "feeds.db")
	}
	// Expand ~ to home directory
	if dbPath[0] == '~' {
		home, _ := os.UserHomeDir()
		dbPath = filepath.Join(home, dbPath[1:])
	}
	return dbPath
}
