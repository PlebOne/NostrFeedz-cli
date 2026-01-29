package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/plebone/nostrfeedz-cli/internal/db"
)

const (
	// CacheExpiration is how long to keep non-favorited images (30 days)
	CacheExpiration = 30 * 24 * time.Hour
	
	// MaxCacheSize is the maximum cache size in bytes (500 MB)
	MaxCacheSize = 500 * 1024 * 1024
)

// ImageCache manages cached images
type ImageCache struct {
	cacheDir string
	db       *db.DB
}

// NewImageCache creates a new image cache
func NewImageCache(cacheDir string, database *db.DB) (*ImageCache, error) {
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &ImageCache{
		cacheDir: cacheDir,
		db:       database,
	}, nil
}

// GetCachePath returns the cache file path for a URL
func (c *ImageCache) GetCachePath(imageURL string) string {
	// Hash the URL to create a unique filename
	hash := sha256.Sum256([]byte(imageURL))
	filename := hex.EncodeToString(hash[:]) + filepath.Ext(imageURL)
	return filepath.Join(c.cacheDir, filename)
}

// IsCached checks if an image is already cached
func (c *ImageCache) IsCached(imageURL string) bool {
	cachePath := c.GetCachePath(imageURL)
	_, err := os.Stat(cachePath)
	return err == nil
}

// GetCached returns the cached file path if it exists
func (c *ImageCache) GetCached(imageURL string) (string, error) {
	cachePath := c.GetCachePath(imageURL)
	if _, err := os.Stat(cachePath); err != nil {
		return "", fmt.Errorf("image not cached")
	}
	
	// Update access time
	now := time.Now()
	os.Chtimes(cachePath, now, now)
	
	return cachePath, nil
}

// Download downloads and caches an image
func (c *ImageCache) Download(imageURL string) (string, error) {
	// Check if already cached
	if c.IsCached(imageURL) {
		return c.GetCached(imageURL)
	}

	// Download the image
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Create cache file
	cachePath := c.GetCachePath(imageURL)
	file, err := os.Create(cachePath)
	if err != nil {
		return "", fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	// Copy image data to cache
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(cachePath)
		return "", fmt.Errorf("failed to write cache: %w", err)
	}

	return cachePath, nil
}

// DownloadAsync downloads an image in the background
func (c *ImageCache) DownloadAsync(imageURL string, callback func(string, error)) {
	go func() {
		path, err := c.Download(imageURL)
		if callback != nil {
			callback(path, err)
		}
	}()
}

// CleanupExpired removes expired cached images
func (c *ImageCache) CleanupExpired() error {
	now := time.Now()
	
	return filepath.Walk(c.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		// Check if file is expired
		age := now.Sub(info.ModTime())
		if age > CacheExpiration {
			// Check if any favorited article uses this image
			// (We'll need to track image-to-article mapping)
			// For now, just delete if expired
			os.Remove(path)
		}
		
		return nil
	})
}

// CleanupDeleted removes cached images for deleted articles
func (c *ImageCache) CleanupDeleted(articleID string, imageURLs []string) error {
	for _, url := range imageURLs {
		cachePath := c.GetCachePath(url)
		os.Remove(cachePath)
	}
	return nil
}

// GetCacheSize returns the total size of the cache
func (c *ImageCache) GetCacheSize() (int64, error) {
	var size int64
	
	err := filepath.Walk(c.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	
	return size, err
}

// EnforceSizeLimit removes oldest files if cache exceeds size limit
func (c *ImageCache) EnforceSizeLimit() error {
	size, err := c.GetCacheSize()
	if err != nil {
		return err
	}
	
	if size <= MaxCacheSize {
		return nil
	}
	
	// Collect all files with their access times
	type fileInfo struct {
		path    string
		modTime time.Time
		size    int64
	}
	
	var files []fileInfo
	
	filepath.Walk(c.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		files = append(files, fileInfo{
			path:    path,
			modTime: info.ModTime(),
			size:    info.Size(),
		})
		return nil
	})
	
	// Sort by modification time (oldest first)
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].modTime.After(files[j].modTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
	
	// Remove oldest files until under limit
	for _, f := range files {
		if size <= MaxCacheSize {
			break
		}
		os.Remove(f.path)
		size -= f.size
	}
	
	return nil
}

// PreloadArticleImages downloads all images for an article in the background
func (c *ImageCache) PreloadArticleImages(imageURLs []string) {
	for _, url := range imageURLs {
		if !c.IsCached(url) {
			c.DownloadAsync(url, nil)
		}
	}
}
