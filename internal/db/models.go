package db

import "time"

type Feed struct {
	ID             string
	Type           string // RSS, NOSTR, NOSTR_VIDEO
	URL            string
	NPUB           string
	Title          string
	Description    string
	LastFetchedAt  *time.Time
	CategoryID     string
	CreatedAt      time.Time
}

type FeedItem struct {
	ID          string
	FeedID      string
	GUID        string
	Title       string
	Content     string
	URL         string
	Author      string
	PublishedAt time.Time
	IsRead      bool
	IsFavorite  bool
	Thumbnail   string
	VideoID     string
	CreatedAt   time.Time
}

type Tag struct {
	ID   string
	Name string
}

type Category struct {
	ID        string
	Name      string
	Color     string
	Icon      string
	SortOrder int
}

type Preference struct {
	Key   string
	Value string
}
