package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmcdole/gofeed"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/plebone/nostrfeedz-cli/internal/config"
	"github.com/plebone/nostrfeedz-cli/internal/db"
	nostrClient "github.com/plebone/nostrfeedz-cli/internal/nostr"
)

func (m *Model) updateAuth(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.authState {
	case AuthPrompt:
		switch msg.String() {
		case "1":
			m.authState = AuthPlebSigner
			m.authInput = ""
		case "2":
			m.authState = AuthRemoteSigner
			m.authInput = ""
		case "3":
			m.authState = AuthPrivateKey
			m.authInput = ""
		}
		
	case AuthPlebSigner:
		switch msg.String() {
		case "enter":
			m.authState = AuthConnecting
			return m, m.connectPlebSigner()
		case "esc":
			m.authState = AuthPrompt
			m.authInput = ""
		}
		
	case AuthRemoteSigner:
		switch msg.String() {
		case "enter":
			if m.authInput != "" {
				m.authState = AuthConnecting
				return m, m.connectRemoteSigner(m.authInput)
			}
		case "esc":
			m.authState = AuthPrompt
			m.authInput = ""
		case "backspace":
			if len(m.authInput) > 0 {
				m.authInput = m.authInput[:len(m.authInput)-1]
			}
		default:
			if len(msg.String()) == 1 {
				m.authInput += msg.String()
			}
		}
		
	case AuthPrivateKey:
		switch msg.String() {
		case "enter":
			if m.authInput != "" {
				m.authState = AuthConnecting
				return m, m.connectPrivateKey(m.authInput)
			}
		case "esc":
			m.authState = AuthPrompt
			m.authInput = ""
		case "backspace":
			if len(m.authInput) > 0 {
				m.authInput = m.authInput[:len(m.authInput)-1]
			}
		default:
			if len(msg.String()) == 1 {
				m.authInput += msg.String()
			}
		}
		
	case AuthError:
		m.authState = AuthPrompt
		m.authInput = ""
		m.authError = ""
	}
	
	return m, nil
}

func (m *Model) updateFeeds(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		// Toggle between view modes
		m.viewMode = (m.viewMode + 1) % 3
		m.selectedFeedIdx = 0
		m.selectedTagIdx = 0
		m.selectedCategoryIdx = 0
		
		// Load data for new view mode
		switch m.viewMode {
		case ViewModeFeeds:
			return m, m.loadFeeds()
		case ViewModeTags:
			return m, m.loadTags()
		case ViewModeCategories:
			return m, m.loadCategories()
		}
		
	case "up", "k":
		switch m.viewMode {
		case ViewModeFeeds:
			if m.selectedFeedIdx > 0 {
				m.selectedFeedIdx--
			}
		case ViewModeTags:
			if m.selectedTagIdx > 0 {
				m.selectedTagIdx--
			}
		case ViewModeCategories:
			if m.selectedCategoryIdx > 0 {
				m.selectedCategoryIdx--
			}
		}
		
	case "down", "j":
		switch m.viewMode {
		case ViewModeFeeds:
			if m.selectedFeedIdx < len(m.feeds)-1 {
				m.selectedFeedIdx++
			}
		case ViewModeTags:
			if m.selectedTagIdx < len(m.tags)-1 {
				m.selectedTagIdx++
			}
		case ViewModeCategories:
			if m.selectedCategoryIdx < len(m.categories)-1 {
				m.selectedCategoryIdx++
			}
		}
		
	case "enter":
		switch m.viewMode {
		case ViewModeFeeds:
			if m.selectedFeedIdx < len(m.feeds) {
				m.currentFeed = &m.feeds[m.selectedFeedIdx]
				m.currentView = ArticlesView
				// TODO: Load articles for this feed
			}
		case ViewModeTags:
			if m.selectedTagIdx < len(m.tags) {
				m.currentTag = &m.tags[m.selectedTagIdx]
				m.currentView = ArticlesView
				// Load feeds for this tag and show articles
				return m, m.loadFeedsForTag(m.currentTag.ID)
			}
		case ViewModeCategories:
			if m.selectedCategoryIdx < len(m.categories) {
				m.currentCategory = &m.categories[m.selectedCategoryIdx]
				m.currentView = ArticlesView
				// Load feeds for this category and show articles
				return m, m.loadFeedsForCategory(m.currentCategory.ID)
			}
		}
		
	case "s":
		// Manual sync
		m.statusMessage = "Syncing from Nostr..."
		return m, m.syncFromNostr()
	}
	return m, nil
}

func (m *Model) updateArticles(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement article navigation
	switch msg.String() {
	case "esc":
		m.currentView = FeedsView
	}
	return m, nil
}

func (m *Model) updateReader(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement reader navigation
	switch msg.String() {
	case "esc":
		m.currentView = ArticlesView
	}
	return m, nil
}

func (m *Model) initNostrClient() tea.Cmd {
	return func() tea.Msg {
		client := nostrClient.NewClient(m.cfg.Nostr.Relays)
		
		// Try Pleb_Signer first
		if m.cfg.Nostr.PlebSigner.Enabled {
			if err := client.SetPlebSigner(); err != nil {
				return authErrorMsg("Failed to connect to Pleb_Signer: " + err.Error())
			}
		} else if m.cfg.Nostr.RemoteSigner.Enabled && m.cfg.Nostr.RemoteSigner.BunkerURL != "" {
			// Try remote signer
			if err := client.SetRemoteSigner(m.cfg.Nostr.RemoteSigner.BunkerURL, 
				m.cfg.Nostr.RemoteSigner.ConnectionToken); err != nil {
				return authErrorMsg("Failed to connect to remote signer: " + err.Error())
			}
		} else if m.cfg.Nostr.NSEC != "" {
			// Fallback to private key
			if err := client.SetPrivateKeySigner(m.cfg.Nostr.NSEC); err != nil {
				return authErrorMsg("Invalid private key: " + err.Error())
			}
		} else {
			return authErrorMsg("No authentication method configured")
		}
		
		// Test connection
		if err := client.TestConnection(); err != nil {
			return authErrorMsg("Failed to connect to Nostr: " + err.Error())
		}
		
		m.nostr = client
		return authSuccessMsg{}
	}
}

func (m *Model) connectPlebSigner() tea.Cmd {
	return func() tea.Msg {
		client := nostrClient.NewClient(m.cfg.Nostr.Relays)
		
		if err := client.SetPlebSigner(); err != nil {
			return authErrorMsg("Failed to connect: " + err.Error())
		}
		
		if err := client.TestConnection(); err != nil {
			return authErrorMsg("Connection test failed: " + err.Error())
		}
		
		// Save configuration
		m.cfg.Nostr.PlebSigner.Enabled = true
		m.cfg.Nostr.NPUB = client.GetPublicKey()
		m.cfg.Nostr.RemoteSigner.Enabled = false
		if err := config.Save(m.cfg); err != nil {
			return authErrorMsg("Failed to save configuration: " + err.Error())
		}
		
		m.nostr = client
		return authSuccessMsg{}
	}
}

func (m *Model) connectRemoteSigner(bunkerURL string) tea.Cmd {
	return func() tea.Msg {
		client := nostrClient.NewClient(m.cfg.Nostr.Relays)
		
		if err := client.SetRemoteSigner(bunkerURL, ""); err != nil {
			return authErrorMsg("Failed to connect: " + err.Error())
		}
		
		if err := client.TestConnection(); err != nil {
			return authErrorMsg("Connection test failed: " + err.Error())
		}
		
		// Save configuration
		m.cfg.Nostr.RemoteSigner.Enabled = true
		m.cfg.Nostr.RemoteSigner.BunkerURL = bunkerURL
		m.cfg.Nostr.NPUB = client.GetPublicKey()
		if err := config.Save(m.cfg); err != nil {
			return authErrorMsg("Failed to save configuration: " + err.Error())
		}
		
		m.nostr = client
		return authSuccessMsg{}
	}
}

func (m *Model) connectPrivateKey(nsec string) tea.Cmd {
	return func() tea.Msg {
		client := nostrClient.NewClient(m.cfg.Nostr.Relays)
		
		if err := client.SetPrivateKeySigner(nsec); err != nil {
			return authErrorMsg("Invalid private key: " + err.Error())
		}
		
		if err := client.TestConnection(); err != nil {
			return authErrorMsg("Connection test failed: " + err.Error())
		}
		
		// Save configuration
		m.cfg.Nostr.NSEC = nsec
		m.cfg.Nostr.NPUB = client.GetPublicKey()
		m.cfg.Nostr.RemoteSigner.Enabled = false
		m.cfg.Nostr.PlebSigner.Enabled = false
		if err := config.Save(m.cfg); err != nil {
			return authErrorMsg("Failed to save configuration: " + err.Error())
		}
		
		m.nostr = client
		return authSuccessMsg{}
	}
}

func (m *Model) loadFeeds() tea.Cmd {
	return func() tea.Msg {
		feeds, err := m.db.GetFeeds()
		if err != nil {
			return errMsg(err)
		}
		return feedsLoadedMsg(feeds)
	}
}

type tagsLoadedMsg []db.Tag
type categoriesLoadedMsg []db.Category
type feedsForTagLoadedMsg []db.Feed
type feedsForCategoryLoadedMsg []db.Feed

func (m *Model) loadTags() tea.Cmd {
	return func() tea.Msg {
		tags, err := m.db.GetTags()
		if err != nil {
			return errMsg(err)
		}
		return tagsLoadedMsg(tags)
	}
}

func (m *Model) loadCategories() tea.Cmd {
	return func() tea.Msg {
		categories, err := m.db.GetCategories()
		if err != nil {
			return errMsg(err)
		}
		return categoriesLoadedMsg(categories)
	}
}

func (m *Model) loadFeedsForTag(tagID string) tea.Cmd {
	return func() tea.Msg {
		feeds, err := m.db.GetFeedsByTag(tagID)
		if err != nil {
			return errMsg(err)
		}
		return feedsForTagLoadedMsg(feeds)
	}
}

func (m *Model) loadFeedsForCategory(categoryID string) tea.Cmd {
	return func() tea.Msg {
		feeds, err := m.db.GetFeedsByCategory(categoryID)
		if err != nil {
			return errMsg(err)
		}
		return feedsForCategoryLoadedMsg(feeds)
	}
}

func (m *Model) syncFromNostr() tea.Cmd {
	return func() tea.Msg {
		if m.nostr == nil {
			return syncCompleteMsg{0, fmt.Errorf("not connected to Nostr")}
		}

		// Get user's public key
		pubkey := m.nostr.GetPublicKey()
		if pubkey == "" {
			return syncCompleteMsg{0, fmt.Errorf("no public key available")}
		}

		// Fetch subscription list from Nostr
		subs, err := m.nostr.FetchSubscriptions(pubkey)
		if err != nil {
			return syncCompleteMsg{0, fmt.Errorf("failed to fetch subscriptions: %w", err)}
		}

		if subs == nil {
			// No subscriptions found on Nostr yet
			return syncCompleteMsg{0, nil}
		}

		feedsAdded := 0

		// Add RSS feeds to local DB
		for _, url := range subs.RSS {
			// Check if feed already exists
			existing, err := m.db.GetFeedByURL(url)
			if err == nil && existing != nil {
				continue // Feed already exists
			}

			// Add new feed
			feed := &db.Feed{
				ID:          fmt.Sprintf("feed_%d", time.Now().UnixNano()),
				Title:       url, // Temporary - will be updated below
				URL:         url,
				Type:        "rss",
				Description: "",
				CategoryID:  "synced",
				CreatedAt:   time.Now(),
			}

			if err := m.db.CreateFeed(feed); err == nil {
				feedsAdded++
				
				// Fetch RSS metadata in background to update title
				go m.updateRSSFeedMetadata(feed)
			}
		}

		// Add Nostr feeds to local DB
		for _, npub := range subs.Nostr {
			// Check if feed already exists
			existing, err := m.db.GetFeedByURL("nostr:" + npub)
			if err == nil && existing != nil {
				continue // Feed already exists
			}

			// Add new feed
			feed := &db.Feed{
				ID:          fmt.Sprintf("feed_%d", time.Now().UnixNano()),
				Title:       npub, // Temporary - will be updated below
				URL:         "nostr:" + npub,
				NPUB:        npub,
				Type:        "nostr",
				Description: "Nostr long-form content",
				CategoryID:  "synced",
				CreatedAt:   time.Now(),
			}

			if err := m.db.CreateFeed(feed); err == nil {
				feedsAdded++
				
				// Fetch Nostr profile metadata in background to update title
				go m.updateNostrFeedMetadata(feed)
			}
		}

		// 3. Import tags from Nostr
		// Tags structure: map[feedURL][]tagNames
		if len(subs.Tags) > 0 {
			// First, collect all unique tag names
			uniqueTags := make(map[string]bool)
			for _, tagNames := range subs.Tags {
				for _, tagName := range tagNames {
					uniqueTags[tagName] = true
				}
			}
			
			// Create all tags
			for tagName := range uniqueTags {
				tag := &db.Tag{
					ID:   fmt.Sprintf("tag_%s", tagName),
					Name: tagName,
				}
				// Try to create tag (ignore if exists)
				m.db.CreateTag(tag)
			}
			
			// Now associate feeds with their tags
			for feedURL, tagNames := range subs.Tags {
				feed, err := m.db.GetFeedByURL(feedURL)
				if err != nil || feed == nil {
					continue // Feed doesn't exist yet or error
				}
				
				// Associate each tag with this feed
				for _, tagName := range tagNames {
					tagID := fmt.Sprintf("tag_%s", tagName)
					m.db.AddFeedTag(feed.ID, tagID)
				}
			}
		}

		// 4. Import categories from Nostr
		if len(subs.Categories) > 0 {
			for feedURL, catInfo := range subs.Categories {
				feed, err := m.db.GetFeedByURL(feedURL)
				if err == nil && feed != nil {
					// Get or create category
					category, err := m.db.GetCategoryByName(catInfo.Name)
					if err != nil || category == nil {
						// Create new category with full info from Nostr
						category = &db.Category{
							ID:    fmt.Sprintf("cat_%s", catInfo.Name),
							Name:  catInfo.Name,
							Color: catInfo.Color,
							Icon:  catInfo.Icon,
						}
						if err := m.db.CreateCategory(category); err != nil {
							fmt.Printf("Warning: failed to create category %s: %v\n", catInfo.Name, err)
							continue
						}
					}
					// Update feed's category
					feed.CategoryID = category.ID
					if err := m.db.UpdateFeed(feed); err != nil {
						fmt.Printf("Warning: failed to update feed category: %v\n", err)
					}
				}
			}
		}

		// 5. Fetch read status from Nostr (kind 30405)
		readStatus, err := m.nostr.FetchReadStatus(pubkey)
		if err != nil {
			// Don't fail the whole sync if read status fails
			// Just log and continue
			fmt.Printf("Warning: failed to fetch read status: %v\n", err)
		} else if readStatus != nil && len(readStatus.ItemGuids) > 0 {
			// Mark items as read in local DB
			for _, guid := range readStatus.ItemGuids {
				// Try to find and mark the item as read
				if err := m.db.MarkItemReadByGUID(guid); err != nil {
					// Item might not exist locally yet, that's okay
					continue
				}
			}
		}

		return syncCompleteMsg{feedsAdded, nil}
	}
}

// updateRSSFeedMetadata fetches RSS feed metadata and updates the feed title
func (m *Model) updateRSSFeedMetadata(feed *db.Feed) {
parser := gofeed.NewParser()
rssFeed, err := parser.ParseURL(feed.URL)
if err != nil {
fmt.Printf("Warning: Failed to fetch RSS metadata for %s: %v\n", feed.URL, err)
return
}

// Update feed with actual metadata
feed.Title = rssFeed.Title
if rssFeed.Description != "" {
feed.Description = rssFeed.Description
}

// Save to database
if err := m.db.UpdateFeed(feed); err != nil {
fmt.Printf("Warning: Failed to update feed metadata: %v\n", err)
}
}

// updateNostrFeedMetadata fetches Nostr profile metadata and updates the feed title
func (m *Model) updateNostrFeedMetadata(feed *db.Feed) {
ctx := context.Background()

// Convert npub to hex pubkey
pubkey := feed.NPUB
if len(pubkey) > 4 && pubkey[:4] == "npub" {
_, decoded, err := nip19.Decode(pubkey)
if err == nil {
pubkey = decoded.(string)
}
}

filter := nostr.Filter{
Kinds:   []int{0}, // Profile metadata
Authors: []string{pubkey},
Limit:   1,
}

events, err := m.nostr.QueryEvents(ctx, filter)
if err != nil || len(events) == 0 {
return
}

// Parse profile metadata
var profile struct {
Name    string `json:"name"`
About   string `json:"about"`
Picture string `json:"picture"`
}

if err := json.Unmarshal([]byte(events[0].Content), &profile); err != nil {
return
}

// Update feed with profile info
if profile.Name != "" {
feed.Title = profile.Name
}
if profile.About != "" {
feed.Description = profile.About
}

// Save to database
if err := m.db.UpdateFeed(feed); err != nil {
fmt.Printf("Warning: Failed to update Nostr feed metadata: %v\n", err)
}
}
