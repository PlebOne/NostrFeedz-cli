package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/plebone/nostrfeedz-cli/internal/config"
	"github.com/plebone/nostrfeedz-cli/internal/db"
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
	cfg    *config.Config
	db     *db.DB
	nostr  *nostr.Client
	
	currentView View
	authState   AuthState
	
	// Authentication
	authInput       string
	authError       string
	cursorPos       int
	
	// Data
	feeds           []db.Feed
	articles        []db.FeedItem
	currentFeed     *db.Feed
	currentArticle  *db.FeedItem
	
	// UI State
	width           int
	height          int
	feedListWidth   int
	articleListWidth int
	selectedFeedIdx int
	selectedArticleIdx int
	err             error
	statusMessage   string
}

func New(cfg *config.Config, database *db.DB) *Model {
	return &Model{
		cfg:              cfg,
		db:               database,
		currentView:      AuthView,
		authState:        AuthPrompt,
		feedListWidth:    cfg.Display.FeedListWidth,
		articleListWidth: cfg.Display.ArticleListWidth,
		feeds:            []db.Feed{},
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
		return m, nil
		
	case authSuccessMsg:
		m.authState = AuthSuccess
		m.currentView = FeedsView
		m.statusMessage = "Successfully authenticated! Syncing from Nostr..."
		return m, tea.Batch(m.loadFeeds(), m.syncFromNostr())
		
	case authErrorMsg:
		m.authState = AuthError
		m.authError = string(msg)
		
	case feedsLoadedMsg:
		m.feeds = msg
		
	case syncCompleteMsg:
		if msg.error != nil {
			m.statusMessage = fmt.Sprintf("Sync failed: %s", msg.error)
		} else {
			m.statusMessage = fmt.Sprintf("Synced from Nostr! Added %d feeds", msg.feedsAdded)
			// Reload feeds after sync
			return m, m.loadFeeds()
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
	
	// Title bar
	title := styles.HeaderStyle.Render("📰 Nostr-Feedz")
	s.WriteString(title)
	s.WriteString("\n\n")
	
	if len(m.feeds) == 0 {
		s.WriteString(styles.MutedStyle.Render("No feeds yet. Press 'a' to add a feed."))
	} else {
		s.WriteString(styles.SuccessStyle.Render(fmt.Sprintf("📁 %d feeds", len(m.feeds))))
		s.WriteString("\n\n")
		
		for i, feed := range m.feeds {
			if i == m.selectedFeedIdx {
				s.WriteString(styles.SelectedStyle.Render("▸ " + feed.Title))
			} else {
				s.WriteString(styles.FeedItemStyle.Render("  " + feed.Title))
			}
			s.WriteString("\n")
		}
	}
	
	s.WriteString("\n\n")
	
	// Status bar
	statusBar := styles.StatusBarStyle.Render(
		styles.RenderKeyValue("q", "quit") + " • " +
		styles.RenderKeyValue("a", "add feed") + " • " +
		styles.RenderKeyValue("↑↓", "navigate") + " • " +
		styles.RenderKeyValue("enter", "open"))
	s.WriteString(statusBar)
	
	if m.statusMessage != "" {
		s.WriteString("\n" + styles.SuccessStyle.Render(m.statusMessage))
	}
	
	return s.String()
}

func (m *Model) renderArticles() string {
	return "Articles view - Coming soon!"
}

func (m *Model) renderReader() string {
	return "Reader view - Coming soon!"
}

// Message types
type authSuccessMsg struct{}
type authErrorMsg string
type feedsLoadedMsg []db.Feed
type syncCompleteMsg struct {
	feedsAdded int
	error      error
}
type errMsg error
