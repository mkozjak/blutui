package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

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
		fmt.Println("Error decoding cache file:", err)
	}

	return cache, nil
}

func SaveCache(cache *Cache) error {
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
		fmt.Println("Error encoding cache to file:", err)
	}

	return nil
}

func FetchAndCache(url string, cache *Cache) ([]byte, error) {
	var body []byte

	if item, found := cache.Data[url]; found && item.Expiration.After(time.Now()) {
		// Use cached response
		body = item.Response
	} else {
		resp, err := http.Get(url)
		if err != nil {
			log.Println("Error fetching album section list:", err)
			return nil, err
		}
		defer resp.Body.Close()

		body, err = httputil.DumpResponse(resp, true)
		if err != nil {
			log.Println("Error reading response body:", err)
			return nil, err
		}

		cache.Data[url] = CacheItem{
			Response:   body,
			Expiration: time.Now().Add(7 * 24 * time.Hour), // Set cache expiration to 1 week
		}

		if err = SaveCache(cache); err != nil {
			log.Println("Error saving data to local cache:", err)
			return nil, err
		}
	}

	return body, nil
}

func SortArtists(input map[string]Artist) []string {
	// Iterate over the map keys and sort them alphabetically
	names := make([]string, 0, len(input))

	for n := range input {
		names = append(names, n)
	}

	sort.Slice(names, func(i, j int) bool {
		return strings.ToLower(names[i]) < strings.ToLower(names[j])
	})

	return names
}

func Log(data ...interface{}) error {
	file, err := os.Create("/tmp/debug.log")
	if err != nil {
		return err
	}
	defer file.Close()

	for _, datum := range data {
		_, err = file.WriteString(fmt.Sprintf("%v ", datum))
		if err != nil {
			return err
		}
	}

	return nil
}

func FormatDuration(d int) string {
	m := d / 60
	s := d % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func CapitalizeArtist(s string) string {
	if len(s) == 0 {
		return s
	}

	// Decode the first rune in the string
	firstRune, size := utf8.DecodeRuneInString(s)
	if firstRune == utf8.RuneError {
		return s
	}

	// Capitalize the first rune if it's a letter
	firstRune = unicode.ToUpper(firstRune)

	// Combine the capitalized first rune with the rest of the string
	return string(firstRune) + s[size:]

}

func Caser(s string) string {
	var res string

	for _, c := range []cases.Caser{cases.Title(language.English)} {
		res = c.String(s)
	}

	return res
}

func ExtractAlbumYear(y string) (int, error) {
	t, err := time.Parse("2006", y)
	if err == nil {
		return t.Year(), nil
	}

	t, err = time.Parse("2006-01-02", y)
	if err == nil {
		return t.Year(), nil
	}

	return 0, errors.New("invalid date format")
}
