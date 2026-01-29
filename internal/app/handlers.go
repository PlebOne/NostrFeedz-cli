package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/blacktop/go-termimg"
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
				m.selectedArticleIdx = 0
				m.loading = true
				m.statusMessage = "Loading articles..."
				// First try to load from database, then fetch if needed
				return m, tea.Batch(
					m.loadArticlesForFeed(m.currentFeed.ID),
					m.fetchArticles(m.currentFeed),
				)
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
	switch msg.String() {
	case "esc":
		m.currentView = FeedsView
		m.articles = []db.FeedItem{} // Clear articles
		m.selectedArticleIdx = 0
		// Reload unread counts when going back to feeds
		return m, m.loadUnreadCounts()
		
	case "up", "k":
		if m.selectedArticleIdx > 0 {
			m.selectedArticleIdx--
		}
		
	case "down", "j":
		if m.selectedArticleIdx < len(m.articles)-1 {
			m.selectedArticleIdx++
		}
		
	case "enter":
		if m.selectedArticleIdx < len(m.articles) {
			m.currentArticle = &m.articles[m.selectedArticleIdx]
			m.currentView = ReaderView
			m.articleScrollOffset = 0
			m.selectedImageIdx = 0 // Reset to first image
			m.selectedVideoIdx = 0 // Reset to first video
			
			// Mark as read
			m.db.MarkItemRead(m.currentArticle.ID, true)
			m.currentArticle.IsRead = true
			
			// Extract media from content and article URL
			m.currentMedia = m.renderer.ExtractMedia(m.currentArticle.Content, m.currentArticle.URL)
			
			// Preload images in background
			if m.currentMedia != nil && len(m.currentMedia.Images) > 0 {
				m.imgCache.PreloadArticleImages(m.currentMedia.Images)
			}
		}
		
	case "r":
		// Refresh - fetch articles again
		if m.currentFeed != nil {
			m.loading = true
			m.statusMessage = "Refreshing..."
			return m, m.fetchArticles(m.currentFeed)
		}
	}
	return m, nil
}

func (m *Model) updateReader(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// ESC always goes back to articles (closes image if open)
		if m.inlineImageData != "" {
			termimg.ClearAll() // Clear terminal images
			m.inlineImageData = ""
		}
		m.currentView = ArticlesView
		m.currentArticle = nil
		m.articleScrollOffset = 0
		return m, tea.ClearScreen
		
	case "q":
		// q just closes image if showing, otherwise does nothing
		if m.inlineImageData != "" {
			termimg.ClearAll() // Clear terminal images
			m.inlineImageData = ""
			m.statusMessage = ""
			return m, tea.ClearScreen
		}
		
	case "up", "k":
		// Don't scroll if showing image
		if m.inlineImageData == "" && m.articleScrollOffset > 0 {
			m.articleScrollOffset--
		}
		
	case "down", "j":
		// Don't scroll if showing image
		if m.inlineImageData == "" {
			m.articleScrollOffset++
		}
		
	case "pageup":
		m.articleScrollOffset -= m.height - 10
		if m.articleScrollOffset < 0 {
			m.articleScrollOffset = 0
		}
		
	case "pagedown", " ":
		m.articleScrollOffset += m.height - 10
		
	case "o":
		// Open in browser
		if m.currentArticle != nil && m.currentArticle.URL != "" {
			// Try to open with xdg-open
			openInBrowser(m.currentArticle.URL)
			m.statusMessage = "Opened in browser"
		}
		
	case "i":
		// Open image in external viewer
		if m.currentMedia == nil {
			m.statusMessage = "No media found in article"
			return m, nil
		}
		if len(m.currentMedia.Images) == 0 {
			m.statusMessage = "No images found in article"
			return m, nil
		}
		
		// Get cached image path or download if needed
		imageURL := m.currentMedia.Images[m.selectedImageIdx]
		m.statusMessage = fmt.Sprintf("Loading image from cache...")
		
		cachePath, err := m.imgCache.GetCached(imageURL)
		if err != nil {
			// Not cached yet, download it
			m.statusMessage = "Downloading image..."
			cachePath, err = m.imgCache.Download(imageURL)
			if err != nil {
				m.statusMessage = fmt.Sprintf("Failed to load image: %v", err)
				return m, nil
			}
		}
		
		m.statusMessage = fmt.Sprintf("Opening image: %s", cachePath)
		m.openImage(cachePath)
		
		if len(m.currentMedia.Images) > 1 {
			m.statusMessage = fmt.Sprintf("Viewing image %d of %d (← → to navigate, q to close)", 
				m.selectedImageIdx+1, len(m.currentMedia.Images))
		} else {
			m.statusMessage = "Viewing image (press 'q' to close)"
		}
		
	case "left", "h":
		// Previous image
		if m.currentMedia != nil && len(m.currentMedia.Images) > 1 {
			if m.selectedImageIdx > 0 {
				m.selectedImageIdx--
			} else {
				m.selectedImageIdx = len(m.currentMedia.Images) - 1 // Wrap around
			}
			
			// Auto-open the new selection
			imageURL := m.currentMedia.Images[m.selectedImageIdx]
			cachePath, err := m.imgCache.GetCached(imageURL)
			if err != nil {
				cachePath, err = m.imgCache.Download(imageURL)
				if err != nil {
					m.statusMessage = fmt.Sprintf("Failed to load image: %v", err)
					return m, nil
				}
			}
			m.openImage(cachePath)
			m.statusMessage = fmt.Sprintf("Viewing image %d of %d (← → to navigate)", 
				m.selectedImageIdx+1, len(m.currentMedia.Images))
		}
		
	case "right", "l":
		// Next image
		if m.currentMedia != nil && len(m.currentMedia.Images) > 1 {
			if m.selectedImageIdx < len(m.currentMedia.Images)-1 {
				m.selectedImageIdx++
			} else {
				m.selectedImageIdx = 0 // Wrap around
			}
			
			// Auto-open the new selection
			imageURL := m.currentMedia.Images[m.selectedImageIdx]
			cachePath, err := m.imgCache.GetCached(imageURL)
			if err != nil {
				cachePath, err = m.imgCache.Download(imageURL)
				if err != nil {
					m.statusMessage = fmt.Sprintf("Failed to load image: %v", err)
					return m, nil
				}
			}
			m.openImage(cachePath)
			m.statusMessage = fmt.Sprintf("Viewing image %d of %d (← → to navigate)", 
				m.selectedImageIdx+1, len(m.currentMedia.Images))
		}
		
	case "I":
		// Force external image viewer
		if m.currentMedia != nil && len(m.currentMedia.Images) > 0 {
			m.openImage(m.currentMedia.Images[0])
			m.statusMessage = "Opened image in external viewer"
		}
		
	case "v":
		// Open video in external player
		if m.currentMedia == nil {
			m.statusMessage = "No media found in article"
			return m, nil
		}
		if len(m.currentMedia.Videos) == 0 {
			m.statusMessage = "No videos found in article"
			return m, nil
		}
		
		videoURL := m.currentMedia.Videos[m.selectedVideoIdx].URL
		m.statusMessage = fmt.Sprintf("Opening video...")
		m.openVideo(videoURL)
		
		if len(m.currentMedia.Videos) > 1 {
			m.statusMessage = fmt.Sprintf("Playing video %d of %d (Shift+← Shift+→ to navigate)", 
				m.selectedVideoIdx+1, len(m.currentMedia.Videos))
		} else {
			m.statusMessage = "Playing video"
		}
		
	case "shift+left", "H":
		// Previous video
		if m.currentMedia != nil && len(m.currentMedia.Videos) > 1 {
			if m.selectedVideoIdx > 0 {
				m.selectedVideoIdx--
			} else {
				m.selectedVideoIdx = len(m.currentMedia.Videos) - 1 // Wrap around
			}
			
			videoURL := m.currentMedia.Videos[m.selectedVideoIdx].URL
			m.openVideo(videoURL)
			m.statusMessage = fmt.Sprintf("Playing video %d of %d (Shift+← Shift+→ to navigate)", 
				m.selectedVideoIdx+1, len(m.currentMedia.Videos))
		}
		
	case "shift+right", "L":
		// Next video
		if m.currentMedia != nil && len(m.currentMedia.Videos) > 1 {
			if m.selectedVideoIdx < len(m.currentMedia.Videos)-1 {
				m.selectedVideoIdx++
			} else {
				m.selectedVideoIdx = 0 // Wrap around
			}
			
			videoURL := m.currentMedia.Videos[m.selectedVideoIdx].URL
			m.openVideo(videoURL)
			m.statusMessage = fmt.Sprintf("Playing video %d of %d (Shift+← Shift+→ to navigate)", 
				m.selectedVideoIdx+1, len(m.currentMedia.Videos))
		}
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
type unreadCountsLoadedMsg map[string]int

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
		
		// Add "Uncategorized" as first category
		uncategorized := db.Category{
			ID:   "uncategorized",
			Name: "Uncategorized",
			Icon: "📋",
			SortOrder: -1,
		}
		allCategories := append([]db.Category{uncategorized}, categories...)
		
		return categoriesLoadedMsg(allCategories)
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
		var feeds []db.Feed
		var err error
		
		if categoryID == "uncategorized" {
			// Get feeds without a category
			feeds, err = m.db.GetUncategorizedFeeds()
		} else {
			feeds, err = m.db.GetFeedsByCategory(categoryID)
		}
		
		if err != nil {
			return errMsg(err)
		}
		return feedsForCategoryLoadedMsg(feeds)
	}
}

func (m *Model) syncFromNostr() tea.Cmd {
	return func() tea.Msg {
		if m.nostr == nil {
			return syncCompleteMsg{0, 0, 0, fmt.Errorf("not connected to Nostr")}
		}

		// Get user's public key
		pubkey := m.nostr.GetPublicKey()
		if pubkey == "" {
			return syncCompleteMsg{0, 0, 0, fmt.Errorf("no public key available")}
		}

		// Fetch subscription list from Nostr
		subs, err := m.nostr.FetchSubscriptions(pubkey)
		if err != nil {
			return syncCompleteMsg{0, 0, 0, fmt.Errorf("failed to fetch subscriptions: %w", err)}
		}

		if subs == nil {
			// No subscriptions found on Nostr yet
			return syncCompleteMsg{0, 0, 0, nil}
		}
		
		// Debug: Log what we received
		fmt.Fprintf(os.Stderr, "DEBUG: Sync received - RSS: %d, Nostr: %d, Tags: %d feeds, Categories: %d feeds\n",
			len(subs.RSS), len(subs.Nostr), len(subs.Tags), len(subs.Categories))

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
		tagsImported := 0
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
				tagsImported++
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
		categoriesImported := 0
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
					} else {
						categoriesImported++
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

		return syncCompleteMsg{feedsAdded, tagsImported, categoriesImported, nil}
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

// Article message types
type articlesLoadedMsg []db.FeedItem
type articlesFetchedMsg struct {
feedID   string
articles []db.FeedItem
err      error
}


type inlineImageMsg struct {
	imageData string
	err       error
}
// fetchArticles fetches articles for a feed (RSS or Nostr)
func (m *Model) fetchArticles(feed *db.Feed) tea.Cmd {
return func() tea.Msg {
var articles []*db.FeedItem
var err error

// Fetch based on feed type
if feed.Type == "rss" {
articles, err = m.fetcher.FetchRSSArticles(feed.URL, feed.ID)
} else if feed.Type == "nostr" {
articles, err = m.fetcher.FetchNostrArticles(feed.NPUB, feed.ID)
} else {
return articlesFetchedMsg{feed.ID, nil, fmt.Errorf("unknown feed type: %s", feed.Type)}
}

if err != nil {
return articlesFetchedMsg{feed.ID, nil, err}
}

// Store articles in database
storedArticles := []db.FeedItem{}
for _, article := range articles {
	if err := m.db.CreateFeedItem(article); err == nil {
		storedArticles = append(storedArticles, *article)
		
		// Preload images for this article in background
		media := m.renderer.ExtractMedia(article.Content, article.URL)
		if media != nil && len(media.Images) > 0 {
			m.imgCache.PreloadArticleImages(media.Images)
		}
	}
}

// Update last fetched timestamp
m.db.UpdateLastFetched(feed.ID)

return articlesFetchedMsg{feed.ID, storedArticles, nil}
}
}

// loadArticlesForFeed loads articles from database for a feed
func (m *Model) loadArticlesForFeed(feedID string) tea.Cmd {
return func() tea.Msg {
articles, err := m.db.GetFeedItemsByFeed(feedID)
if err != nil {
return errMsg(err)
}
return articlesLoadedMsg(articles)
}
}

// loadArticlesForFeeds loads articles for multiple feeds (tags/categories)
func (m *Model) loadArticlesForFeeds(feedIDs []string) tea.Cmd {
return func() tea.Msg {
articles, err := m.db.GetFeedItemsByFeeds(feedIDs)
if err != nil {
return errMsg(err)
}
return articlesLoadedMsg(articles)
}
}

// Helper functions for opening external applications
func openInBrowser(url string) {
exec.Command("xdg-open", url).Start()
}

func (m *Model) openImage(url string) {
// Kill previous image viewer if still running
if m.imageViewerPID > 0 {
// Try to kill the previous viewer
if process, err := os.FindProcess(m.imageViewerPID); err == nil {
process.Kill()
}
m.imageViewerPID = 0
}

viewers := []struct {
cmd  string
args []string
}{
{"sxiv", []string{"-b", "-g", "800x600", url}},
{"feh", []string{"--scale-down", "--auto-zoom", "--borderless", "--geometry", "800x600", url}},
{"imv-wayland", []string{url}},
{"imv-x11", []string{url}},
{"eog", []string{url}},
{"eom", []string{url}},
{"xdg-open", []string{url}},
}

for _, viewer := range viewers {
cmd := exec.Command(viewer.cmd, viewer.args...)
if err := cmd.Start(); err == nil {
m.imageViewerPID = cmd.Process.Pid
return
}
}
}

func (m *Model) openVideo(url string) {
// Kill previous video player if still running
if m.videoPlayerPID > 0 {
if process, err := os.FindProcess(m.videoPlayerPID); err == nil {
process.Kill()
}
m.videoPlayerPID = 0
}

players := []struct {
cmd  string
args []string
}{
{"mpv", []string{"--geometry=800x600", url}},
{"vlc", []string{"--width=800", "--height=600", url}},
{"mplayer", []string{url}},
{"xdg-open", []string{url}},
}

for _, player := range players {
cmd := exec.Command(player.cmd, player.args...)
if err := cmd.Start(); err == nil {
m.videoPlayerPID = cmd.Process.Pid
return
}
}
}

// showInlineImage fetches and displays an image inline in the terminal
func (m *Model) showInlineImage(imageURL string) tea.Cmd {
return func() tea.Msg {
// Check cache first
var imageData string
var err error

if m.imgCache.IsCached(imageURL) {
// Use cached version
cachePath, cacheErr := m.imgCache.GetCached(imageURL)
if cacheErr == nil {
imageData, err = m.renderer.RenderImageInlineFromFile(cachePath, 80, 24)
} else {
err = cacheErr
}
} else {
// Download and cache
cachePath, dlErr := m.imgCache.Download(imageURL)
if dlErr == nil {
imageData, err = m.renderer.RenderImageInlineFromFile(cachePath, 80, 24)
} else {
err = dlErr
}
}

if err != nil {
return inlineImageMsg{"", err}
}

return inlineImageMsg{imageData, nil}
}
}

func (m *Model) loadUnreadCounts() tea.Cmd {
return func() tea.Msg {
counts, err := m.db.GetUnreadCounts()
if err != nil {
return errMsg(err)
}
return unreadCountsLoadedMsg(counts)
}
}
