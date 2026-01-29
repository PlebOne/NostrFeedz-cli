package feed

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown"
	"github.com/blacktop/go-termimg"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Renderer handles rendering article content
type Renderer struct {
	glamourRenderer *glamour.TermRenderer
	htmlConverter   *md.Converter
	width           int
}

// NewRenderer creates a new content renderer
func NewRenderer(width int) (*Renderer, error) {
	// Create glamour renderer for markdown
	gr, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-4), // Leave margin
	)
	if err != nil {
		return nil, err
	}

	// Create HTML to markdown converter
	conv := md.NewConverter("", true, nil)

	return &Renderer{
		glamourRenderer: gr,
		htmlConverter:   conv,
		width:           width,
	}, nil
}

// RenderContent renders article content (HTML or Markdown) to terminal
func (r *Renderer) RenderContent(content string, isHTML bool) (string, error) {
	// Convert HTML to markdown if needed
	if isHTML {
		markdown, err := r.htmlConverter.ConvertString(content)
		if err != nil {
			return "", fmt.Errorf("failed to convert HTML: %w", err)
		}
		content = markdown
	}

	// Render markdown with glamour
	rendered, err := r.glamourRenderer.Render(content)
	if err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}

	return rendered, nil
}

// RenderContentWithInlineImages renders article content with images embedded inline
func (r *Renderer) RenderContentWithInlineImages(content string, isHTML bool) (string, error) {
	// Convert HTML to markdown if needed
	if isHTML {
		markdown, err := r.htmlConverter.ConvertString(content)
		if err != nil {
			return "", fmt.Errorf("failed to convert HTML: %w", err)
		}
		content = markdown
	}

	// Only embed images if terminal supports it
	if SupportsInlineImages() {
		// Extract image URLs
		imgRegex := regexp.MustCompile(`!\[([^\]]*)\]\((https?://[^\)]+)\)`)
		matches := imgRegex.FindAllStringSubmatch(content, -1)
		
		// Replace each image with inline terminal image
		for _, match := range matches {
			if len(match) >= 3 {
				imageURL := match[2]
				altText := match[1]
				
				// Try to render inline (constrained size for inline display)
				inlineImg, err := r.RenderImageInline(imageURL, r.width-4, 20)
				if err == nil && inlineImg != "" {
					// Replace the markdown image with the terminal image
					replacement := fmt.Sprintf("\n%s\n", inlineImg)
					if altText != "" {
						replacement += fmt.Sprintf("  %s\n", altText)
					}
					content = strings.Replace(content, match[0], replacement, 1)
				}
				// If error, leave the markdown as-is (glamour will handle it)
			}
		}
	}

	// Render markdown with glamour
	rendered, err := r.glamourRenderer.Render(content)
	if err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}

	return rendered, nil
}

// ExtractMedia extracts image and video URLs from content and article URL
func (r *Renderer) ExtractMedia(content string, articleURL string) *MediaLinks {
	media := &MediaLinks{
		Images: []string{},
		Videos: []VideoInfo{},
	}

	// First check if the article URL itself is a YouTube video (including Shorts)
	if strings.Contains(articleURL, "youtube.com/watch") || strings.Contains(articleURL, "youtu.be/") || strings.Contains(articleURL, "youtube.com/shorts/") {
		// Extract video ID from URL
		var videoID string
		if strings.Contains(articleURL, "youtube.com/watch?v=") {
			parts := strings.Split(articleURL, "v=")
			if len(parts) > 1 {
				videoID = strings.Split(parts[1], "&")[0]
			}
		} else if strings.Contains(articleURL, "youtu.be/") {
			parts := strings.Split(articleURL, "youtu.be/")
			if len(parts) > 1 {
				videoID = strings.Split(parts[1], "?")[0]
			}
		} else if strings.Contains(articleURL, "youtube.com/shorts/") {
			parts := strings.Split(articleURL, "youtube.com/shorts/")
			if len(parts) > 1 {
				videoID = strings.Split(parts[1], "?")[0]
				// Remove any trailing path
				videoID = strings.Split(videoID, "/")[0]
			}
		}
		
		if videoID != "" {
			media.Videos = append(media.Videos, VideoInfo{
				URL:   fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
				Title: "Video from feed",
			})
		}
	}

	// Extract image URLs
	imgRegex := regexp.MustCompile(`!\[.*?\]\((https?://[^\)]+)\)`)
	imgMatches := imgRegex.FindAllStringSubmatch(content, -1)
	for _, match := range imgMatches {
		if len(match) > 1 {
			media.Images = append(media.Images, match[1])
		}
	}

	// Also find HTML img tags
	htmlImgRegex := regexp.MustCompile(`<img[^>]+src="([^"]+)"`)
	htmlMatches := htmlImgRegex.FindAllStringSubmatch(content, -1)
	for _, match := range htmlMatches {
		if len(match) > 1 {
			media.Images = append(media.Images, match[1])
		}
	}

	// Extract video URLs with metadata (YouTube, etc.)
	// Try to extract YouTube embeds with title
	youtubeEmbedRegex := regexp.MustCompile(`<iframe[^>]*src="https?://(?:www\.)?youtube\.com/embed/([\w-]+)"[^>]*(?:title="([^"]*)")?[^>]*>`)
	youtubeMatches := youtubeEmbedRegex.FindAllStringSubmatch(content, -1)
	for _, match := range youtubeMatches {
		if len(match) > 1 {
			videoID := match[1]
			title := ""
			if len(match) > 2 {
				title = match[2]
			}
			url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
			// Check if we already have this video
			found := false
			for _, v := range media.Videos {
				if v.URL == url {
					found = true
					break
				}
			}
			if !found {
				media.Videos = append(media.Videos, VideoInfo{
					URL:   url,
					Title: title,
				})
			}
		}
	}

	// Look for YouTube links in <a> tags
	youtubeLinkRegex := regexp.MustCompile(`<a[^>]*href="(https?://(?:www\.)?(?:youtube\.com/watch\?v=|youtu\.be/)([\w-]+)[^"]*)"`)
	linkMatches := youtubeLinkRegex.FindAllStringSubmatch(content, -1)
	for _, match := range linkMatches {
		if len(match) > 2 {
			videoID := match[2]
			url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
			// Check if we already have this video
			found := false
			for _, v := range media.Videos {
				if v.URL == url {
					found = true
					break
				}
			}
			if !found {
				media.Videos = append(media.Videos, VideoInfo{
					URL: url,
				})
			}
		}
	}

	// Extract plain YouTube URLs from text
	videoPatterns := []struct {
		pattern string
		urlFunc func(string) string
	}{
		{`https?://(?:www\.)?youtube\.com/watch\?v=([\w-]+)`, func(id string) string { return fmt.Sprintf("https://www.youtube.com/watch?v=%s", id) }},
		{`https?://(?:www\.)?youtube\.com/shorts/([\w-]+)`, func(id string) string { return fmt.Sprintf("https://www.youtube.com/watch?v=%s", id) }},
		{`https?://youtu\.be/([\w-]+)`, func(id string) string { return fmt.Sprintf("https://www.youtube.com/watch?v=%s", id) }},
		{`https?://(?:www\.)?vimeo\.com/(\d+)`, func(id string) string { return fmt.Sprintf("https://vimeo.com/%s", id) }},
		{`(https?://[^\s]+\.mp4)`, func(url string) string { return url }},
		{`(https?://[^\s]+\.webm)`, func(url string) string { return url }},
	}

	for _, p := range videoPatterns {
		regex := regexp.MustCompile(p.pattern)
		matches := regex.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				// Check if we already have this video from embed
				url := p.urlFunc(match[1])
				found := false
				for _, v := range media.Videos {
					if v.URL == url {
						found = true
						break
					}
				}
				if !found {
					media.Videos = append(media.Videos, VideoInfo{
						URL: url,
					})
				}
			}
		}
	}

	return media
}

// RenderMediaList renders a list of media items at the bottom
func (r *Renderer) RenderMediaList(media *MediaLinks, selectedImageIdx int, selectedVideoIdx int) string {
	if len(media.Images) == 0 && len(media.Videos) == 0 {
		return ""
	}

	var s strings.Builder
	
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true)

	s.WriteString("\n\n")
	s.WriteString(style.Render("─── Media ───"))
	s.WriteString("\n")

	if len(media.Images) > 0 {
		var hint string
		if len(media.Images) > 1 {
			hint = fmt.Sprintf("📷 %d images [%d/%d] - ← → to select, 'i' to view", 
				len(media.Images), selectedImageIdx+1, len(media.Images))
		} else {
			hint = "📷 1 image - Press 'i' to view"
		}
		s.WriteString(style.Render(hint))
		s.WriteString("\n")
	}

	if len(media.Videos) > 0 {
		var hint string
		if len(media.Videos) > 1 {
			hint = fmt.Sprintf("🎬 %d videos [%d/%d] - Shift+← Shift+→ to select, 'v' to play", 
				len(media.Videos), selectedVideoIdx+1, len(media.Videos))
		} else {
			hint = "🎬 1 video - Press 'v' to play"
		}
		s.WriteString(style.Render(hint))
		s.WriteString("\n")
		
		// Show current video info if available
		if selectedVideoIdx < len(media.Videos) {
			video := media.Videos[selectedVideoIdx]
			if video.Title != "" {
				s.WriteString(style.Render(fmt.Sprintf("  Title: %s", video.Title)))
				s.WriteString("\n")
			}
			if video.Description != "" {
				s.WriteString(style.Render(fmt.Sprintf("  %s", video.Description)))
				s.WriteString("\n")
			}
			s.WriteString(style.Render(fmt.Sprintf("  URL: %s", video.URL)))
			s.WriteString("\n")
		}
	}

	return s.String()
}

// MediaLinks contains extracted media URLs
type MediaLinks struct {
	Images []string
	Videos []VideoInfo
}

type VideoInfo struct {
	URL         string
	Title       string
	Description string
}

// RenderImageInline attempts to display an image inline in the terminal
// Returns the rendered image string if successful, empty string if terminal doesn't support it
func (r *Renderer) RenderImageInline(imageURL string, maxWidth, maxHeight int) (string, error) {
	// Download the image
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	// Create image from reader
	img, err := termimg.From(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse image: %w", err)
	}

	// Set size constraints
	if maxWidth > 0 {
		img = img.Width(maxWidth)
	}
	if maxHeight > 0 {
		img = img.Height(maxHeight)
	}

	// Render to string (auto-detects best protocol)
	rendered, err := img.Render()
	if err != nil {
		return "", fmt.Errorf("failed to render image: %w", err)
	}

	return rendered, nil
}

// SupportsInlineImages checks if the terminal supports inline images
func SupportsInlineImages() bool {
	// termimg will auto-detect and fallback, but we can check
	// if any of the protocols are available
	return termimg.KittySupported() || termimg.SixelSupported() || termimg.ITerm2Supported()
}

// RenderImageInlineFromFile attempts to display a cached image inline in the terminal
func (r *Renderer) RenderImageInlineFromFile(imagePath string, maxWidth, maxHeight int) (string, error) {
// Open cached image file
img, err := termimg.Open(imagePath)
if err != nil {
return "", fmt.Errorf("failed to open cached image: %w", err)
}

// Set size constraints
if maxWidth > 0 {
img = img.Width(maxWidth)
}
if maxHeight > 0 {
img = img.Height(maxHeight)
}

// Render to string (auto-detects best protocol)
rendered, err := img.Render()
if err != nil {
return "", fmt.Errorf("failed to render image: %w", err)
}

return rendered, nil
}
