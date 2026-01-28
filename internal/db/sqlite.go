package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func New(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS feeds (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		url TEXT,
		npub TEXT,
		title TEXT NOT NULL,
		description TEXT,
		last_fetched_at INTEGER,
		category_id TEXT,
		created_at INTEGER NOT NULL,
		UNIQUE(type, url)
	);

	-- Create unique index for Nostr feeds only (where npub is not null)
	CREATE UNIQUE INDEX IF NOT EXISTS idx_feeds_type_npub 
		ON feeds(type, npub) WHERE npub IS NOT NULL AND npub != '';

	CREATE TABLE IF NOT EXISTS feed_items (
		id TEXT PRIMARY KEY,
		feed_id TEXT NOT NULL,
		guid TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT,
		url TEXT,
		author TEXT,
		published_at INTEGER NOT NULL,
		is_read INTEGER DEFAULT 0,
		is_favorite INTEGER DEFAULT 0,
		thumbnail TEXT,
		video_id TEXT,
		created_at INTEGER NOT NULL,
		FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
		UNIQUE(feed_id, guid)
	);

	CREATE INDEX IF NOT EXISTS idx_feed_items_feed_id ON feed_items(feed_id);
	CREATE INDEX IF NOT EXISTS idx_feed_items_published_at ON feed_items(published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_feed_items_is_read ON feed_items(is_read);

	CREATE TABLE IF NOT EXISTS tags (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL
	);

	CREATE TABLE IF NOT EXISTS feed_tags (
		feed_id TEXT NOT NULL,
		tag_id TEXT NOT NULL,
		PRIMARY KEY(feed_id, tag_id),
		FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
		FOREIGN KEY(tag_id) REFERENCES tags(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS categories (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		color TEXT,
		icon TEXT,
		sort_order INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS preferences (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// Preferences
func (db *DB) GetPreference(key string) (string, error) {
	var value string
	err := db.conn.QueryRow("SELECT value FROM preferences WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (db *DB) SetPreference(key, value string) error {
	_, err := db.conn.Exec(`
		INSERT INTO preferences (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	return err
}

// Feeds
func (db *DB) CreateFeed(feed *Feed) error {
	_, err := db.conn.Exec(`
		INSERT INTO feeds (id, type, url, npub, title, description, last_fetched_at, category_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, feed.ID, feed.Type, feed.URL, feed.NPUB, feed.Title, feed.Description,
		timeToUnix(feed.LastFetchedAt), feed.CategoryID, feed.CreatedAt.Unix())
	return err
}

func (db *DB) GetFeeds() ([]Feed, error) {
	rows, err := db.conn.Query(`
		SELECT id, type, url, npub, title, description, last_fetched_at, category_id, created_at
		FROM feeds ORDER BY title
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []Feed
	for rows.Next() {
		var feed Feed
		var lastFetched sql.NullInt64
		err := rows.Scan(&feed.ID, &feed.Type, &feed.URL, &feed.NPUB, &feed.Title,
			&feed.Description, &lastFetched, &feed.CategoryID, new(int64))
		if err != nil {
			return nil, err
		}
		feed.LastFetchedAt = unixToTime(lastFetched)
		feeds = append(feeds, feed)
	}
	return feeds, rows.Err()
}

func (db *DB) GetFeedByURL(url string) (*Feed, error) {
	var feed Feed
	var lastFetched sql.NullInt64
	err := db.conn.QueryRow(`
		SELECT id, type, url, npub, title, description, last_fetched_at, category_id, created_at
		FROM feeds WHERE url = ?
	`, url).Scan(&feed.ID, &feed.Type, &feed.URL, &feed.NPUB, &feed.Title,
		&feed.Description, &lastFetched, &feed.CategoryID, new(int64))
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	feed.LastFetchedAt = unixToTime(lastFetched)
	return &feed, nil
}

func (db *DB) DeleteFeed(id string) error {
	_, err := db.conn.Exec("DELETE FROM feeds WHERE id = ?", id)
	return err
}

func (db *DB) UpdateFeed(feed *Feed) error {
	_, err := db.conn.Exec(`
		UPDATE feeds 
		SET title = ?, description = ?, last_fetched_at = ?
		WHERE id = ?
	`, feed.Title, feed.Description, timeToUnix(feed.LastFetchedAt), feed.ID)
	return err
}

// Feed Items
func (db *DB) CreateFeedItem(item *FeedItem) error {
	_, err := db.conn.Exec(`
		INSERT OR IGNORE INTO feed_items 
		(id, feed_id, guid, title, content, url, author, published_at, is_read, is_favorite, thumbnail, video_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, item.ID, item.FeedID, item.GUID, item.Title, item.Content, item.URL, item.Author,
		item.PublishedAt.Unix(), boolToInt(item.IsRead), boolToInt(item.IsFavorite),
		item.Thumbnail, item.VideoID, item.CreatedAt.Unix())
	return err
}

func (db *DB) GetFeedItems(feedID string, limit int) ([]FeedItem, error) {
	query := `
		SELECT id, feed_id, guid, title, content, url, author, published_at, is_read, is_favorite, thumbnail, video_id, created_at
		FROM feed_items
	`
	var args []interface{}
	if feedID != "" {
		query += " WHERE feed_id = ?"
		args = append(args, feedID)
	}
	query += " ORDER BY published_at DESC"
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []FeedItem
	for rows.Next() {
		var item FeedItem
		var publishedAt, createdAt int64
		var isRead, isFavorite int
		err := rows.Scan(&item.ID, &item.FeedID, &item.GUID, &item.Title, &item.Content,
			&item.URL, &item.Author, &publishedAt, &isRead, &isFavorite,
			&item.Thumbnail, &item.VideoID, &createdAt)
		if err != nil {
			return nil, err
		}
		item.PublishedAt = time.Unix(publishedAt, 0)
		item.CreatedAt = time.Unix(createdAt, 0)
		item.IsRead = isRead == 1
		item.IsFavorite = isFavorite == 1
		items = append(items, item)
	}
	return items, rows.Err()
}

func (db *DB) MarkItemRead(itemID string, isRead bool) error {
	_, err := db.conn.Exec("UPDATE feed_items SET is_read = ? WHERE id = ?", boolToInt(isRead), itemID)
	return err
}

func (db *DB) MarkItemReadByGUID(guid string) error {
	_, err := db.conn.Exec("UPDATE feed_items SET is_read = 1 WHERE guid = ?", guid)
	return err
}

func (db *DB) ToggleFavorite(itemID string) error {
	_, err := db.conn.Exec("UPDATE feed_items SET is_favorite = NOT is_favorite WHERE id = ?", itemID)
	return err
}

// Helper functions
func timeToUnix(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.Unix()
}

func unixToTime(n sql.NullInt64) *time.Time {
	if !n.Valid {
		return nil
	}
	t := time.Unix(n.Int64, 0)
	return &t
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Tags
func (db *DB) CreateTag(tag *Tag) error {
_, err := db.conn.Exec(`
INSERT OR IGNORE INTO tags (id, name)
VALUES (?, ?)
`, tag.ID, tag.Name)
return err
}

func (db *DB) GetTags() ([]Tag, error) {
rows, err := db.conn.Query("SELECT id, name FROM tags ORDER BY name")
if err != nil {
return nil, err
}
defer rows.Close()

var tags []Tag
for rows.Next() {
var tag Tag
if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
return nil, err
}
tags = append(tags, tag)
}
return tags, rows.Err()
}

func (db *DB) AddFeedTag(feedID, tagID string) error {
_, err := db.conn.Exec(`
INSERT OR IGNORE INTO feed_tags (feed_id, tag_id)
VALUES (?, ?)
`, feedID, tagID)
return err
}

func (db *DB) GetFeedTags(feedID string) ([]Tag, error) {
rows, err := db.conn.Query(`
SELECT t.id, t.name
FROM tags t
JOIN feed_tags ft ON t.id = ft.tag_id
WHERE ft.feed_id = ?
ORDER BY t.name
`, feedID)
if err != nil {
return nil, err
}
defer rows.Close()

var tags []Tag
for rows.Next() {
var tag Tag
if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
return nil, err
}
tags = append(tags, tag)
}
return tags, rows.Err()
}

// Category methods
func (db *DB) CreateCategory(category *Category) error {
_, err := db.conn.Exec(`
INSERT INTO categories (id, name, color, icon, sort_order)
VALUES (?, ?, ?, ?, ?)
`, category.ID, category.Name, category.Color, category.Icon, category.SortOrder)
return err
}

func (db *DB) GetCategories() ([]Category, error) {
rows, err := db.conn.Query(`
SELECT id, name, COALESCE(color, ''), COALESCE(icon, ''), sort_order
FROM categories
ORDER BY sort_order, name
`)
if err != nil {
return nil, err
}
defer rows.Close()

var categories []Category
for rows.Next() {
var cat Category
if err := rows.Scan(&cat.ID, &cat.Name, &cat.Color, &cat.Icon, &cat.SortOrder); err != nil {
return nil, err
}
categories = append(categories, cat)
}
return categories, rows.Err()
}

func (db *DB) GetCategoryByName(name string) (*Category, error) {
var cat Category
err := db.conn.QueryRow(`
SELECT id, name, COALESCE(color, ''), COALESCE(icon, ''), sort_order
FROM categories
WHERE name = ?
`, name).Scan(&cat.ID, &cat.Name, &cat.Color, &cat.Icon, &cat.SortOrder)

if err != nil {
return nil, err
}
return &cat, nil
}

func (db *DB) GetFeedsByCategory(categoryID string) ([]Feed, error) {
rows, err := db.conn.Query(`
SELECT id, type, COALESCE(url, ''), COALESCE(npub, ''), title, 
       COALESCE(description, ''), COALESCE(category_id, ''), created_at
FROM feeds
WHERE category_id = ?
ORDER BY title
`, categoryID)
if err != nil {
return nil, err
}
defer rows.Close()

var feeds []Feed
for rows.Next() {
var feed Feed
var createdAt int64
if err := rows.Scan(&feed.ID, &feed.Type, &feed.URL, &feed.NPUB, &feed.Title,
&feed.Description, &feed.CategoryID, &createdAt); err != nil {
return nil, err
}
feed.CreatedAt = time.Unix(createdAt, 0)
feeds = append(feeds, feed)
}
return feeds, rows.Err()
}

func (db *DB) GetFeedsByTag(tagID string) ([]Feed, error) {
rows, err := db.conn.Query(`
SELECT f.id, f.type, COALESCE(f.url, ''), COALESCE(f.npub, ''), f.title,
       COALESCE(f.description, ''), COALESCE(f.category_id, ''), f.created_at
FROM feeds f
JOIN feed_tags ft ON f.id = ft.feed_id
WHERE ft.tag_id = ?
ORDER BY f.title
`, tagID)
if err != nil {
return nil, err
}
defer rows.Close()

var feeds []Feed
for rows.Next() {
var feed Feed
var createdAt int64
if err := rows.Scan(&feed.ID, &feed.Type, &feed.URL, &feed.NPUB, &feed.Title,
&feed.Description, &feed.CategoryID, &createdAt); err != nil {
return nil, err
}
feed.CreatedAt = time.Unix(createdAt, 0)
feeds = append(feeds, feed)
}
return feeds, rows.Err()
}
