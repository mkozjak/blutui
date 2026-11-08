package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mkozjak/blutui/internal"
)

type Cache struct {
	Data map[string]CacheItem
	path string
}

type CacheItem struct {
	Response   []byte
	Expiration time.Time
}

func LoadCache() (*Cache, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cache := &Cache{
		Data: make(map[string]CacheItem),
		path: filepath.Join(home, ".config", "blutui", "cache"),
	}

	file, err := os.Open(cache.path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		return cache, nil
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(cache)
	if err != nil {
		internal.Log("Error decoding cache file:", err)
		cache.Data = make(map[string]CacheItem)
	}

	if cache.Data == nil {
		cache.Data = make(map[string]CacheItem)
	}

	return cache, nil
}

func saveCache(cache *Cache) error {
	if err := os.MkdirAll(filepath.Dir(cache.path), 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(cache.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(cache)
	if err != nil {
		internal.Log("Error encoding cache to file:", err)
	}

	return err
}

func FetchAndCache(url string, cache *Cache, cached bool) ([]byte, error) {
	var body []byte

	if item, found := cache.Data[url]; cached && found && item.Expiration.After(time.Now()) {
		// Use cached response
		body = item.Response
	} else {
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			internal.Log("Error fetching album section list:", err)
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetching %s: %s", url, resp.Status)
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			internal.Log("Error reading response body:", err)
			return nil, err
		}

		cache.Data[url] = CacheItem{
			Response:   body,
			Expiration: time.Now().Add(7 * 24 * time.Hour), // Set cache expiration to 1 week
		}

		if err = saveCache(cache); err != nil {
			internal.Log("Error saving data to local cache:", err)
		}
	}

	return body, nil
}
