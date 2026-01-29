package nostr

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

const (
	SubscriptionListKind = 30404
	ReadStatusKind       = 30405
	SubscriptionDTag     = "nostr-feedz-subscriptions"
	ReadStatusDTag       = "nostr-feedz-read-status"
)

type SubscriptionList struct {
	RSS         []string                       `json:"rss"`
	Nostr       []string                       `json:"nostr"`
	Tags        map[string][]string            `json:"tags"`
	Categories  map[string]CategoryInfo        `json:"categories"` // URL/npub -> category info
	Deleted     []string                       `json:"deleted"`
	LastUpdated int64                          `json:"lastUpdated"`
}

type CategoryInfo struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

type ReadStatusList struct {
	ItemGuids   []string `json:"itemGuids"`
	LastUpdated int64    `json:"lastUpdated"`
}

// PublishSubscriptions publishes the subscription list to Nostr
func (c *Client) PublishSubscriptions(list *SubscriptionList) error {
	content, err := json.Marshal(list)
	if err != nil {
		return fmt.Errorf("failed to marshal subscriptions: %w", err)
	}

	event := &nostr.Event{
		Kind:      SubscriptionListKind,
		CreatedAt: nostr.Now(),
		Tags: nostr.Tags{
			{"d", SubscriptionDTag},
			{"client", "nostrfeedz-cli"},
		},
		Content: string(content),
	}

	return c.PublishEvent(event)
}

// FetchSubscriptions fetches the subscription list from Nostr
func (c *Client) FetchSubscriptions(pubkey string) (*SubscriptionList, error) {
	ctx := context.Background()
	filter := nostr.Filter{
		Kinds:   []int{SubscriptionListKind},
		Authors: []string{pubkey},
		Tags:    nostr.TagMap{"d": []string{SubscriptionDTag}},
		Limit:   1,
	}

	events, err := c.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, nil
	}

	var list SubscriptionList
	if err := json.Unmarshal([]byte(events[0].Content), &list); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscriptions: %w", err)
	}

	return &list, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PublishReadStatus publishes the read status list to Nostr
func (c *Client) PublishReadStatus(status *ReadStatusList) error {
	content, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal read status: %w", err)
	}

	event := &nostr.Event{
		Kind:      ReadStatusKind,
		CreatedAt: nostr.Now(),
		Tags: nostr.Tags{
			{"d", ReadStatusDTag},
			{"client", "nostrfeedz-cli"},
		},
		Content: string(content),
	}

	return c.PublishEvent(event)
}

// FetchReadStatus fetches the read status list from Nostr
func (c *Client) FetchReadStatus(pubkey string) (*ReadStatusList, error) {
	ctx := context.Background()
	filter := nostr.Filter{
		Kinds:   []int{ReadStatusKind},
		Authors: []string{pubkey},
		Tags:    nostr.TagMap{"d": []string{ReadStatusDTag}},
		Limit:   1,
	}

	events, err := c.QueryEvents(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, nil
	}

	var status ReadStatusList
	if err := json.Unmarshal([]byte(events[0].Content), &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal read status: %w", err)
	}

	return &status, nil
}

// MergeSubscriptions merges local and remote subscription lists
func MergeSubscriptions(local, remote *SubscriptionList) *SubscriptionList {
	if remote == nil {
		return local
	}
	if local == nil {
		return remote
	}

	merged := &SubscriptionList{
		Tags:        make(map[string][]string),
		Categories:  make(map[string]CategoryInfo),
		LastUpdated: max(local.LastUpdated, remote.LastUpdated),
	}

	// Merge RSS feeds
	rssSet := make(map[string]bool)
	for _, url := range local.RSS {
		rssSet[url] = true
	}
	for _, url := range remote.RSS {
		rssSet[url] = true
	}
	for url := range rssSet {
		merged.RSS = append(merged.RSS, url)
	}

	// Merge Nostr feeds
	nostrSet := make(map[string]bool)
	for _, npub := range local.Nostr {
		nostrSet[npub] = true
	}
	for _, npub := range remote.Nostr {
		nostrSet[npub] = true
	}
	for npub := range nostrSet {
		merged.Nostr = append(merged.Nostr, npub)
	}

	// Merge tags
	for key, tags := range local.Tags {
		merged.Tags[key] = tags
	}
	for key, tags := range remote.Tags {
		if existing, ok := merged.Tags[key]; ok {
			tagSet := make(map[string]bool)
			for _, tag := range existing {
				tagSet[tag] = true
			}
			for _, tag := range tags {
				tagSet[tag] = true
			}
			var mergedTags []string
			for tag := range tagSet {
				mergedTags = append(mergedTags, tag)
			}
			merged.Tags[key] = mergedTags
		} else {
			merged.Tags[key] = tags
		}
	}

	// Merge categories (remote wins on conflicts)
	for key, cat := range local.Categories {
		merged.Categories[key] = cat
	}
	for key, cat := range remote.Categories {
		merged.Categories[key] = cat
	}

	// Merge deleted feeds
	deletedSet := make(map[string]bool)
	for _, item := range local.Deleted {
		deletedSet[item] = true
	}
	for _, item := range remote.Deleted {
		deletedSet[item] = true
	}
	for item := range deletedSet {
		merged.Deleted = append(merged.Deleted, item)
	}

	return merged
}

// MergeReadStatus merges local and remote read status lists
func MergeReadStatus(local, remote *ReadStatusList) *ReadStatusList {
	if remote == nil {
		return local
	}
	if local == nil {
		return remote
	}

	merged := &ReadStatusList{
		LastUpdated: max(local.LastUpdated, remote.LastUpdated),
	}

	guidSet := make(map[string]bool)
	for _, guid := range local.ItemGuids {
		guidSet[guid] = true
	}
	for _, guid := range remote.ItemGuids {
		guidSet[guid] = true
	}

	for guid := range guidSet {
		merged.ItemGuids = append(merged.ItemGuids, guid)
	}

	return merged
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
