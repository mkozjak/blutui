package cache

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/mkozjak/blutui/internal"
)

type Cache struct {
	Data map[string]CacheItem
}

type CacheItem struct {
	Response   []byte
	Expiration time.Time
}

func LoadCache() (*Cache, error) {
	cache := &Cache{Data: make(map[string]CacheItem)}

	file, err := os.Open("/Users/mkozjak/.config/blutui/cache")
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
	}

	return cache, nil
}

func saveCache(cache *Cache) error {
	file, err := os.OpenFile("/Users/mkozjak/.config/blutui/cache", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			file, err = os.Create("/Users/mkozjak/.config/blutui/cache")
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(cache)
	if err != nil {
		internal.Log("Error encoding cache to file:", err)
	}

	return nil
}

func FetchAndCache(url string, cache *Cache, cached bool) ([]byte, error) {
	var body []byte

	if item, found := cache.Data[url]; cached && found && item.Expiration.After(time.Now()) {
		// Use cached response
		body = item.Response
	} else {
		resp, err := http.Get(url)
		if err != nil {
			internal.Log("Error fetching album section list:", err)
			return nil, err
		}
		defer resp.Body.Close()

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
			return nil, err
		}
	}

	return body, nil
}
