package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/plebone/nostrfeedz-cli/internal/cache"
	"github.com/plebone/nostrfeedz-cli/internal/config"
	"github.com/plebone/nostrfeedz-cli/internal/db"
	"github.com/plebone/nostrfeedz-cli/internal/feed"
	"github.com/plebone/nostrfeedz-cli/internal/nostr"
	"github.com/plebone/nostrfeedz-cli/pkg/styles"
)

type View int

const (
	AuthView View = iota
	FeedsView
	ArticlesView
	ReaderView
)

type ViewMode int

const (
	ViewModeFeeds ViewMode = iota
	ViewModeTags
	ViewModeCategories
)

type AuthState int

const (
	AuthPrompt AuthState = iota
	AuthPlebSigner
	AuthRemoteSigner
	AuthPrivateKey
	AuthConnecting
	AuthSuccess
	AuthError
)

type Model struct {
	cfg      *config.Config
	db       *db.DB
	nostr    *nostr.Client
	fetcher  *feed.Fetcher
	renderer *feed.Renderer
	imgCache *cache.ImageCache
	
	currentView View
	viewMode    ViewMode
	authState   AuthState
	
	// Authentication
	authInput       string
	authError       string
	cursorPos       int
	
	// Data
	feeds           []db.Feed
	tags            []db.Tag
	categories      []db.Category
	articles        []db.FeedItem
	unreadCounts    map[string]int // Feed ID to unread count
	currentFeed     *db.Feed
	currentTag      *db.Tag
	currentCategory *db.Category
	currentArticle  *db.FeedItem
	currentMedia    *feed.MediaLinks
	
	// UI State
	width           int
	height          int
	feedListWidth   int
	articleListWidth int
	selectedFeedIdx int
	selectedTagIdx  int
	selectedCategoryIdx int
	selectedArticleIdx int
	selectedImageIdx int
	selectedVideoIdx int
	articleScrollOffset int
	inlineImageData string
	err             error
	statusMessage   string
	loading         bool
	imageViewerPID  int // Track image viewer process
	videoPlayerPID  int // Track video player process
}

func New(cfg *config.Config, database *db.DB) *Model {
	fetcher := feed.NewFetcher(cfg.Nostr.Relays)
	renderer, _ := feed.NewRenderer(80) // Default width, will update on window resize
	
	// Create image cache directory
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".config", "nostrfeedz", "cache", "images")
	imgCache, _ := cache.NewImageCache(cacheDir, database)
	
	return &Model{
		cfg:              cfg,
		db:               database,
		fetcher:          fetcher,
		renderer:         renderer,
		imgCache:         imgCache,
		currentView:      AuthView,
		viewMode:         ViewModeFeeds,
		authState:        AuthPrompt,
		feedListWidth:    cfg.Display.FeedListWidth,
		articleListWidth: cfg.Display.ArticleListWidth,
		feeds:            []db.Feed{},
		tags:             []db.Tag{},
		categories:       []db.Category{},
		articles:         []db.FeedItem{},
	}
}

func (m *Model) Init() tea.Cmd {
	// Set default dimensions in case WindowSizeMsg hasn't arrived yet
	m.width = 80
	m.height = 24
	
	// Check if already authenticated
	if m.cfg.Nostr.NPUB != "" {
		return m.initNostrClient()
	}
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.currentView == AuthView && m.authState != AuthPrompt {
				// Allow quitting during auth
				return m, tea.Quit
			}
			if m.currentView != AuthView {
				return m, tea.Quit
			}
		}
		
		// Handle auth view input
		if m.currentView == AuthView {
			return m.updateAuth(msg)
		}
		
		// Handle other views
		switch m.currentView {
		case FeedsView:
			return m.updateFeeds(msg)
		case ArticlesView:
			return m.updateArticles(msg)
		case ReaderView:
			return m.updateReader(msg)
		}
		
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Recreate renderer with new width
		if renderer, err := feed.NewRenderer(msg.Width); err == nil {
			m.renderer = renderer
		}
		return m, nil
		
	case authSuccessMsg:
		m.authState = AuthSuccess
		m.currentView = FeedsView
		m.statusMessage = "Successfully authenticated! Syncing from Nostr..."
		return m, tea.Batch(m.loadFeeds(), m.loadTags(), m.loadCategories(), m.syncFromNostr())
		
	case authErrorMsg:
		m.authState = AuthError
		m.authError = string(msg)
		
	case feedsLoadedMsg:
		m.feeds = msg
		// Load unread counts after loading feeds
		return m, m.loadUnreadCounts()
		
	case tagsLoadedMsg:
		m.tags = msg
		
	case categoriesLoadedMsg:
		m.categories = msg
		
	case unreadCountsLoadedMsg:
		m.unreadCounts = msg
		
	case feedsForTagLoadedMsg:
		m.feeds = msg
		m.selectedFeedIdx = 0
		// Load articles from all feeds with this tag
		if len(msg) > 0 {
			feedIDs := make([]string, len(msg))
			for i, feed := range msg {
				feedIDs[i] = feed.ID
			}
			m.loading = true
			m.statusMessage = "Loading articles..."
			return m, m.loadArticlesForFeeds(feedIDs)
		}
		
	case feedsForCategoryLoadedMsg:
		m.feeds = msg
		m.selectedFeedIdx = 0
		// Load articles from all feeds in this category
		if len(msg) > 0 {
			feedIDs := make([]string, len(msg))
			for i, feed := range msg {
				feedIDs[i] = feed.ID
			}
			m.loading = true
			m.statusMessage = "Loading articles..."
			return m, m.loadArticlesForFeeds(feedIDs)
		}
		
	case articlesLoadedMsg:
		m.articles = msg
		m.loading = false
		if len(msg) > 0 {
			m.statusMessage = fmt.Sprintf("Loaded %d articles", len(msg))
		}
		// Reload unread counts after loading articles
		return m, m.loadUnreadCounts()
		
	case articlesFetchedMsg:
		if msg.err != nil {
			m.statusMessage = fmt.Sprintf("Error fetching articles: %s", msg.err)
			m.loading = false
		} else {
			// Reload articles from database (includes newly fetched)
			if m.currentFeed != nil && m.currentFeed.ID == msg.feedID {
				return m, m.loadArticlesForFeed(msg.feedID)
			}
		}
		
	case syncCompleteMsg:
		if msg.error != nil {
			m.statusMessage = fmt.Sprintf("Sync failed: %s", msg.error)
		} else {
			statusParts := []string{}
			if msg.feedsAdded > 0 {
				statusParts = append(statusParts, fmt.Sprintf("%d feeds", msg.feedsAdded))
			}
			if msg.tagsImported > 0 {
				statusParts = append(statusParts, fmt.Sprintf("%d tags", msg.tagsImported))
			}
			if msg.categoriesImported > 0 {
				statusParts = append(statusParts, fmt.Sprintf("%d categories", msg.categoriesImported))
			}
			
			if len(statusParts) > 0 {
				m.statusMessage = fmt.Sprintf("Synced! Added: %s", strings.Join(statusParts, ", "))
			} else {
				m.statusMessage = "Synced from Nostr! (No new data)"
			}
			// Reload all data after sync
			return m, tea.Batch(m.loadFeeds(), m.loadTags(), m.loadCategories())
		}
		
	case inlineImageMsg:
		if msg.err != nil {
			m.statusMessage = fmt.Sprintf("Failed to display image: %s", msg.err)
		} else {
			// Store the inline image data to display
			m.inlineImageData = msg.imageData
			m.statusMessage = "Image displayed inline (ESC to close)"
		}
		
	case errMsg:
		m.err = msg
	}
	
	return m, nil
}

func (m *Model) View() string {
	switch m.currentView {
	case AuthView:
		return m.renderAuth()
	case FeedsView:
		return m.renderFeeds()
	case ArticlesView:
		return m.renderArticles()
	case ReaderView:
		return m.renderReader()
	}
	return ""
}

func (m *Model) renderAuth() string {
	var s strings.Builder
	
	// Add some spacing from top
	s.WriteString("\n\n")
	
	// Title
	title := styles.TitleStyle.Render("🚀 Nostr-Feedz CLI")
	s.WriteString(centerText(title, m.width))
	s.WriteString("\n\n\n")
	
	switch m.authState {
	case AuthPrompt:
		s.WriteString(centerText("Welcome! Please choose authentication method:", m.width))
		s.WriteString("\n\n")
		s.WriteString(centerText(styles.KeyStyle.Render("1")+" - Pleb_Signer (D-Bus) "+styles.SuccessStyle.Render("[Recommended]"), m.width))
		s.WriteString("\n")
		s.WriteString(centerText(styles.KeyStyle.Render("2")+" - Remote Signer (NIP-46)", m.width))
		s.WriteString("\n")
		s.WriteString(centerText(styles.KeyStyle.Render("3")+" - Private Key (nsec)", m.width))
		s.WriteString("\n\n")
		s.WriteString(centerText(styles.MutedStyle.Render("Press 1, 2, or 3 to continue"), m.width))
		
	case AuthPlebSigner:
		s.WriteString(centerText(styles.HeaderStyle.Render("Pleb_Signer (D-Bus)"), m.width))
		s.WriteString("\n\n")
		s.WriteString(centerText("Connecting to Pleb_Signer...", m.width))
		s.WriteString("\n")
		s.WriteString(centerText(styles.MutedStyle.Render("Make sure Pleb_Signer is running and unlocked"), m.width))
		s.WriteString("\n\n")
		s.WriteString(centerText(styles.MutedStyle.Render("Press Enter to connect • Esc to go back"), m.width))
		
	case AuthRemoteSigner:
		s.WriteString(centerText(styles.HeaderStyle.Render("Remote Signer (NIP-46)"), m.width))
		s.WriteString("\n\n")
		s.WriteString(centerText("Enter your bunker URL:", m.width))
		s.WriteString("\n")
		s.WriteString(centerText(styles.MutedStyle.Render("Format: bunker://<pubkey>?relay=<relay-url>"), m.width))
		s.WriteString("\n\n")
		
		inputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.AccentColor).
			Padding(0, 1).
			Width(60).
			Render(m.authInput + "▊")
		s.WriteString(centerText(inputBox, m.width))
		s.WriteString("\n\n")
		s.WriteString(centerText(styles.MutedStyle.Render("Press Enter to connect • Esc to go back"), m.width))
		
	case AuthPrivateKey:
		s.WriteString(centerText(styles.HeaderStyle.Render("Private Key Authentication"), m.width))
		s.WriteString("\n\n")
		s.WriteString(centerText("Enter your private key (nsec):", m.width))
		s.WriteString("\n")
		s.WriteString(centerText(styles.ErrorStyle.Render("⚠ Warning: Your key will be stored locally"), m.width))
		s.WriteString("\n\n")
		
		inputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.AccentColor).
			Padding(0, 1).
			Width(60).
			Render(strings.Repeat("*", len(m.authInput)) + "▊")
		s.WriteString(centerText(inputBox, m.width))
		s.WriteString("\n\n")
		s.WriteString(centerText(styles.MutedStyle.Render("Press Enter to login • Esc to go back"), m.width))
		
	case AuthConnecting:
		s.WriteString("\n\n\n\n")
		s.WriteString(centerText("Connecting to Nostr...", m.width))
		s.WriteString("\n\n")
		s.WriteString(centerText(styles.MutedStyle.Render("Please wait..."), m.width))
		
	case AuthError:
		s.WriteString(centerText(styles.RenderError("Authentication Failed"), m.width))
		s.WriteString("\n\n")
		s.WriteString(centerText(m.authError, m.width))
		s.WriteString("\n\n")
		s.WriteString(centerText(styles.MutedStyle.Render("Press any key to try again"), m.width))
	}
	
	return s.String()
}

// centerText centers a string within the given width
func centerText(text string, width int) string {
	if width <= 0 {
		width = 80
	}
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	padding := (width - textWidth) / 2
	return strings.Repeat(" ", padding) + text
}

func (m *Model) renderFeeds() string {
	var s strings.Builder
	
	// Title bar with view mode
	viewModeStr := ""
	switch m.viewMode {
	case ViewModeFeeds:
		viewModeStr = "📰 All Feeds"
	case ViewModeTags:
		viewModeStr = "🏷️  Tags"
	case ViewModeCategories:
		viewModeStr = "📂 Categories"
	}
	
	title := styles.HeaderStyle.Render(viewModeStr)
	s.WriteString(title)
	s.WriteString("\n")
	
	// View mode toggle hint
	s.WriteString(styles.MutedStyle.Render("Press Tab to switch views: Feeds • Tags • Categories"))
	s.WriteString("\n\n")
	
	// Render content based on view mode
	switch m.viewMode {
	case ViewModeFeeds:
		if len(m.feeds) == 0 {
			s.WriteString(styles.MutedStyle.Render("No feeds yet. Press 's' to sync from Nostr."))
		} else {
			s.WriteString(styles.SuccessStyle.Render(fmt.Sprintf("📁 %d feeds", len(m.feeds))))
			s.WriteString("\n\n")
			
			for i, feed := range m.feeds {
				unreadCount := m.unreadCounts[feed.ID]
				displayText := feed.Title
				if unreadCount > 0 {
					displayText = fmt.Sprintf("%s (%d)", feed.Title, unreadCount)
				}
				
				if i == m.selectedFeedIdx {
					s.WriteString(styles.SelectedStyle.Render("▸ " + displayText))
				} else {
					s.WriteString(styles.FeedItemStyle.Render("  " + displayText))
				}
				s.WriteString("\n")
			}
		}
		
	case ViewModeTags:
		if len(m.tags) == 0 {
			s.WriteString(styles.MutedStyle.Render("No tags yet. Tags will appear after syncing from Nostr."))
		} else {
			s.WriteString(styles.SuccessStyle.Render(fmt.Sprintf("🏷️  %d tags", len(m.tags))))
			s.WriteString("\n\n")
			
			for i, tag := range m.tags {
				if i == m.selectedTagIdx {
					s.WriteString(styles.SelectedStyle.Render("▸ " + tag.Name))
				} else {
					s.WriteString(styles.FeedItemStyle.Render("  " + tag.Name))
				}
				s.WriteString("\n")
			}
		}
		
	case ViewModeCategories:
		if len(m.categories) == 0 {
			s.WriteString(styles.MutedStyle.Render("No categories yet. Categories will appear after syncing from Nostr."))
		} else {
			s.WriteString(styles.SuccessStyle.Render(fmt.Sprintf("📂 %d categories", len(m.categories))))
			s.WriteString("\n\n")
			
			for i, cat := range m.categories {
				icon := cat.Icon
				if icon == "" {
					icon = "📁"
				}
				displayName := icon + " " + cat.Name
				
				if i == m.selectedCategoryIdx {
					s.WriteString(styles.SelectedStyle.Render("▸ " + displayName))
				} else {
					s.WriteString(styles.FeedItemStyle.Render("  " + displayName))
				}
				s.WriteString("\n")
			}
		}
	}
	
	s.WriteString("\n\n")
	
	// Status bar
	statusBar := styles.StatusBarStyle.Render(
		styles.RenderKeyValue("q", "quit") + " • " +
		styles.RenderKeyValue("tab", "switch view") + " • " +
		styles.RenderKeyValue("↑↓", "navigate") + " • " +
		styles.RenderKeyValue("enter", "open") + " • " +
		styles.RenderKeyValue("s", "sync"))
	s.WriteString(statusBar)
	
	if m.statusMessage != "" {
		s.WriteString("\n" + styles.SuccessStyle.Render(m.statusMessage))
	}
	
	return s.String()
}

func (m *Model) renderArticles() string {
	var s strings.Builder
	
	// Title with feed name
	feedName := "Articles"
	if m.currentFeed != nil {
		feedName = m.currentFeed.Title
	} else if m.currentTag != nil {
		feedName = "Tag: " + m.currentTag.Name
	} else if m.currentCategory != nil {
		feedName = "Category: " + m.currentCategory.Name
	}
	
	title := styles.HeaderStyle.Render("📖 " + feedName)
	s.WriteString(title)
	s.WriteString("\n\n")
	
	if m.loading {
		s.WriteString(styles.MutedStyle.Render("Loading articles..."))
		return s.String()
	}
	
	if len(m.articles) == 0 {
		s.WriteString(styles.MutedStyle.Render("No articles yet. Press 'r' to refresh."))
	} else {
		s.WriteString(styles.SuccessStyle.Render(fmt.Sprintf("📰 %d articles", len(m.articles))))
		s.WriteString("\n\n")
		
		// Show articles (limit to visible)
		maxVisible := m.height - 10
		start := m.selectedArticleIdx - maxVisible/2
		if start < 0 {
			start = 0
		}
		end := start + maxVisible
		if end > len(m.articles) {
			end = len(m.articles)
		}
		
		for i := start; i < end; i++ {
			article := m.articles[i]
			
			// Format article line
			readIndicator := "  "
			if article.IsRead {
				readIndicator = "✓ "
			}
			
			dateStr := article.PublishedAt.Format("Jan 02")
			title := article.Title
			if len(title) > 60 {
				title = title[:57] + "..."
			}
			
			line := fmt.Sprintf("%s%s - %s", readIndicator, dateStr, title)
			
			if i == m.selectedArticleIdx {
				s.WriteString(styles.SelectedStyle.Render("▸ " + line))
			} else {
				if article.IsRead {
					s.WriteString(styles.MutedStyle.Render("  " + line))
				} else {
					s.WriteString(styles.FeedItemStyle.Render("  " + line))
				}
			}
			s.WriteString("\n")
		}
	}
	
	s.WriteString("\n")
	
	// Status bar
	statusBar := styles.StatusBarStyle.Render(
		styles.RenderKeyValue("esc", "back") + " • " +
		styles.RenderKeyValue("↑↓", "navigate") + " • " +
		styles.RenderKeyValue("enter", "read") + " • " +
		styles.RenderKeyValue("r", "refresh"))
	s.WriteString(statusBar)
	
	if m.statusMessage != "" {
		s.WriteString("\n" + styles.SuccessStyle.Render(m.statusMessage))
	}
	
	return s.String()
}

func (m *Model) renderReader() string {
	if m.currentArticle == nil {
		return "No article selected"
	}
	
	var s strings.Builder
	
	// Title
	title := styles.HeaderStyle.Render(m.currentArticle.Title)
	s.WriteString(title)
	s.WriteString("\n")
	
	// Metadata
	meta := fmt.Sprintf("By %s • %s", 
		m.currentArticle.Author,
		m.currentArticle.PublishedAt.Format("January 2, 2006"))
	s.WriteString(styles.MutedStyle.Render(meta))
	s.WriteString("\n\n")
	
	// Render content
	isHTML := strings.Contains(m.currentArticle.Content, "<html") || 
	          strings.Contains(m.currentArticle.Content, "<div")
	
	// If we have inline image data (from 'i' key), show it instead
	if m.inlineImageData != "" {
		s.WriteString(lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00")).
			Render("🖼️  IMAGE VIEWER"))
		s.WriteString("\n")
		s.WriteString(strings.Repeat("─", m.width))
		s.WriteString("\n\n")
		s.WriteString(m.inlineImageData)
		s.WriteString("\n\n")
		s.WriteString(strings.Repeat("─", m.width))
		s.WriteString("\n")
		s.WriteString(styles.StatusBarStyle.Render(
			styles.RenderKeyValue("i", "close image") + " • " +
			styles.RenderKeyValue("ESC", "back to articles") + " • " +
			styles.RenderKeyValue("I", "external viewer")))
		s.WriteString("\n")
		if m.statusMessage != "" {
			s.WriteString("\n")
			s.WriteString(styles.MutedStyle.Render(m.statusMessage))
		}
		return s.String()
	}
	
	// Use simple rendering (no inline images to avoid freezing)
	rendered, err := m.renderer.RenderContent(m.currentArticle.Content, isHTML)
	if err != nil {
		s.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("Error rendering: %v", err)))
		s.WriteString("\n\n")
		s.WriteString(m.currentArticle.Content) // Fallback to raw
	} else {
		// Apply scroll offset
		lines := strings.Split(rendered, "\n")
		visibleLines := m.height - 8 // Leave room for header/footer
		
		start := m.articleScrollOffset
		if start >= len(lines) {
			start = len(lines) - visibleLines
		}
		if start < 0 {
			start = 0
		}
		
		end := start + visibleLines
		if end > len(lines) {
			end = len(lines)
		}
		
		for i := start; i < end; i++ {
			s.WriteString(lines[i])
			s.WriteString("\n")
		}
	}
	
	// Media list
	if m.currentMedia != nil {
		s.WriteString(m.renderer.RenderMediaList(m.currentMedia, m.selectedImageIdx, m.selectedVideoIdx))
	}
	
	s.WriteString("\n")
	
	// Status bar
	statusBarKeys := styles.RenderKeyValue("esc", "back") + " • " +
		styles.RenderKeyValue("↑↓", "scroll") + " • " +
		styles.RenderKeyValue("space", "page down")
	
	if m.currentMedia != nil && len(m.currentMedia.Images) > 1 {
		statusBarKeys += " • " + styles.RenderKeyValue("←→", "image")
	}
	
	statusBarKeys += " • " + styles.RenderKeyValue("i", "view") + " • " +
		styles.RenderKeyValue("o", "browser") + " • " +
		styles.RenderKeyValue("v", "video")
	
	statusBar := styles.StatusBarStyle.Render(statusBarKeys)
	s.WriteString(statusBar)
	
	return s.String()
}

// Message types
type authSuccessMsg struct{}
type authErrorMsg string
type feedsLoadedMsg []db.Feed
type syncCompleteMsg struct {
	feedsAdded        int
	tagsImported      int
	categoriesImported int
	error             error
}
type errMsg error
