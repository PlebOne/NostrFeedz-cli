package feed

import (
	"context"
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/plebone/nostrfeedz-cli/internal/db"
)

// Fetcher handles fetching articles from RSS and Nostr feeds
type Fetcher struct {
	nostrPool  *nostr.SimplePool
	nostrRelays []string
}

// NewFetcher creates a new feed fetcher
func NewFetcher(nostrRelays []string) *Fetcher {
	return &Fetcher{
		nostrPool:   nostr.NewSimplePool(context.Background()),
		nostrRelays: nostrRelays,
	}
}

// FetchRSSArticles fetches articles from an RSS feed
func (f *Fetcher) FetchRSSArticles(feedURL string, feedID string) ([]*db.FeedItem, error) {
	parser := gofeed.NewParser()
	rssFeed, err := parser.ParseURL(feedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	var articles []*db.FeedItem
	for _, item := range rssFeed.Items {
		// Use item.GUID or link as unique identifier
		guid := item.GUID
		if guid == "" {
			guid = item.Link
		}

		// Parse published date
		var publishedAt time.Time
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			publishedAt = *item.UpdatedParsed
		} else {
			publishedAt = time.Now()
		}

		// Get content (prefer Content over Description)
		content := item.Description
		if item.Content != "" {
			content = item.Content
		}

		// Get author
		author := ""
		if item.Author != nil {
			author = item.Author.Name
		}

		article := &db.FeedItem{
			ID:          fmt.Sprintf("item_%d", time.Now().UnixNano()),
			FeedID:      feedID,
			GUID:        guid,
			Title:       item.Title,
			Content:     content,
			URL:         item.Link,
			Author:      author,
			PublishedAt: publishedAt,
			IsRead:      false,
			IsFavorite:  false,
			CreatedAt:   time.Now(),
		}

		// Handle media enclosures (images, videos)
		if len(item.Enclosures) > 0 {
			// For now, just store the first enclosure
			enc := item.Enclosures[0]
			if enc.Type != "" {
				if enc.Type[:5] == "image" {
					article.Thumbnail = enc.URL
				} else if enc.Type[:5] == "video" {
					article.Thumbnail = enc.URL
					// Try to extract video ID from URL
					article.VideoID = extractVideoID(enc.URL)
				}
			}
		}

		articles = append(articles, article)
	}

	return articles, nil
}

// FetchNostrArticles fetches NIP-23 long-form articles from a Nostr user
func (f *Fetcher) FetchNostrArticles(npub string, feedID string) ([]*db.FeedItem, error) {
	// Convert npub to hex pubkey
	prefix, decoded, err := nip19.Decode(npub)
	if err != nil {
		return nil, fmt.Errorf("failed to decode npub: %w", err)
	}
	
	if prefix != "npub" {
		return nil, fmt.Errorf("invalid npub format")
	}
	
	pubkey := decoded.(string)

	// Query for kind 30023 (NIP-23 long-form content)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := nostr.Filter{
		Kinds:   []int{30023}, // Long-form content
		Authors: []string{pubkey},
		Limit:   50,
	}

	var articles []*db.FeedItem
	for ev := range f.nostrPool.SubManyEose(ctx, f.nostrRelays, nostr.Filters{filter}) {
		if ev.Event == nil {
			continue
		}

		event := ev.Event

		// Extract metadata from tags
		title := ""
		image := ""
		publishedAt := time.Unix(int64(event.CreatedAt), 0)

		for _, tag := range event.Tags {
			if len(tag) < 2 {
				continue
			}
			switch tag[0] {
			case "title":
				title = tag[1]
			case "summary":
			case "image":
				image = tag[1]
			case "published_at":
				if ts, err := time.Parse(time.RFC3339, tag[1]); err == nil {
					publishedAt = ts
				}
			}
		}

		// Use event ID as GUID
		guid := event.ID

		// Get author name from event (we'd need to fetch kind 0 for proper name)
		author := event.PubKey[:8] + "..." // Short pubkey for now
		
		// Encode event ID as note1...
		noteID, _ := nip19.EncodeNote(event.ID)

		article := &db.FeedItem{
			ID:          fmt.Sprintf("item_%d", time.Now().UnixNano()),
			FeedID:      feedID,
			GUID:        guid,
			Title:       title,
			Content:     event.Content,
			URL:         fmt.Sprintf("nostr:%s", noteID),
			Author:      author,
			PublishedAt: publishedAt,
			IsRead:      false,
			IsFavorite:  false,
			Thumbnail:   image,
			CreatedAt:   time.Now(),
		}

		articles = append(articles, article)
	}

	return articles, nil
}

// extractVideoID tries to extract a video ID from common video platforms
func extractVideoID(url string) string {
	// YouTube patterns
	patterns := []string{
		"youtube.com/watch?v=",
		"youtu.be/",
		"youtube.com/embed/",
	}

	for _, pattern := range patterns {
		if idx := len(url); idx > 0 {
			// Simple extraction - could be improved
			return url[len(pattern):]
		}
	}

	return ""
}
