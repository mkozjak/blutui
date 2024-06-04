package internal

import (
	"encoding/xml"
	"errors"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mkozjak/tview"
)

var CustomBorders = &tview.BoxBorders{
	// \u0020 - whitespace
	HorizontalFocus:  rune('\u2500'),
	Horizontal:       rune('\u2500'),
	VerticalFocus:    rune('\u2502'),
	Vertical:         rune('\u2502'),
	TopRightFocus:    rune('\u2510'),
	TopRight:         rune('\u2510'),
	TopLeftFocus:     rune('\u250C'),
	TopLeft:          rune('\u250C'),
	BottomRightFocus: rune('\u2518'),
	BottomRight:      rune('\u2518'),
	BottomLeftFocus:  rune('\u2514'),
	BottomLeft:       rune('\u2514'),
}

var noBorders = &tview.BoxBorders{}

type track struct {
	name        string
	duration    int
	disc        int
	number      int
	playUrl     string
	autoplayUrl string
}

type album struct {
	name        string
	year        int
	duration    int
	genre       string
	tracks      []track
	playUrl     string
	autoplayUrl string
}

type Artist struct {
	albums []album
}

type App struct {
	Application         *tview.Application
	AlbumArtists        map[string]Artist
	Artists             []string
	currentArtistAlbums []*tview.List
	ArtistPane          *tview.List
	AlbumPane           *tview.Grid
	StatusBar           *tview.Table
	sbMessages          chan Status
	CpArtistIdx         int // currently playing artist's index in *tview.List
	cpTrackName         string
}

type Cache struct {
	Data map[string]CacheItem
}

type CacheItem struct {
	Response   []byte
	Expiration time.Time
}

// Used for parsing data from /Browse
type browse struct {
	Items []item `xml:"item"`
}

type volume struct {
	XMLName xml.Name `xml:"volume"`
	Value   int      `xml:",chardata"`
	Muted   string   `xml:"mute,attr"`
}

type item struct {
	Text        string `xml:"text,attr"`  // album name; track name
	Text2       string `xml:"text2,attr"` // artist name
	BrowseKey   string `xml:"browseKey,attr"`
	Type        string `xml:"type,attr"`
	PlayURL     string `xml:"playURL,attr"`
	AutoplayURL string `xml:"autoplayURL,attr"`
	Duration    string `xml:"duration,attr"`
}

// Used for parsing data from /Songs
type songs struct {
	Album []struct {
		Song []struct {
			Date string `xml:"date"`
		} `xml:"song"`
	} `xml:"album"`
}

type Status struct {
	ETag     string `xml:"etag,attr"`
	Volume   int    `xml:"volume"`
	Album    string `xml:"album"`
	Artist   string `xml:"artist"`
	Track    string `xml:"name"`
	Title2   string `xml:"title2"`
	Title3   string `xml:"title3"`
	Format   string `xml:"streamFormat"`
	Quality  string `xml:"quality"`
	TrackLen int    `xml:"totlen"`
	Secs     int    `xml:"secs"`
	State    string `xml:"state"`
}

func (a *App) FetchData() error {
	cache, err := LoadCache()
	if err != nil {
		log.Println("Error loading local cache:", err)
		return err
	}

	body, err := FetchAndCache(api+"/Browse?key=LocalMusic%3AbySection%2F%252FAlbums%253Fservice%253DLocalMusic", cache)
	if err != nil {
		log.Println("Error fetching/caching data:", err)
		return err
	}

	var sections browse
	err = xml.Unmarshal(body, &sections)
	if err != nil {
		log.Println("Error parsing the sections XML:", err)
		return err
	}

	// parse album sections (alphabetical order) from xml
	for _, item := range sections.Items {
		body, err = FetchAndCache(api+"/Browse?key="+url.QueryEscape(item.BrowseKey), cache)
		if err != nil {
			log.Println("Error fetching album sections:", err)
			return err
		}

		var albums browse
		err = xml.Unmarshal(body, &albums)
		if err != nil {
			log.Println("Error parsing the albums XML:", err)
			return err
		}

		// iterate albums and fill m.albumArtists
		for _, al := range albums.Items {
			var duration int

			// fetch album tracks
			body, err = FetchAndCache(api+"/Browse?key="+url.QueryEscape(al.BrowseKey), cache)
			if err != nil {
				log.Println("Error fetching album tracks:", err)
				return err
			}

			var tracks browse
			err = xml.Unmarshal(body, &tracks)
			if err != nil {
				log.Println("Error parsing the album tracks XML:", err)
				return err
			}

			var albumTracks []track
			for _, tr := range tracks.Items {
				track := track{
					name:        tr.Text,
					playUrl:     tr.PlayURL,
					autoplayUrl: tr.AutoplayURL,
					duration: func() int {
						l, err := strconv.Atoi(tr.Duration)
						if err != nil {
							return 0
						}

						return l
					}(),
				}

				albumTracks = append(albumTracks, track)
				duration += track.duration
			}

			arName := Caser(al.Text2)

			// fetch album date from /Songs
			body, err = FetchAndCache(
				strings.ReplaceAll(api+"/Songs?service=LocalMusic&album="+al.Text+"&artist="+arName, " ", "+"),
				cache)
			if err != nil {
				log.Println("Error fetching album date:", err)
				return err
			}

			var s songs
			var year int
			err = xml.Unmarshal(body, &s)
			if err != nil {
				log.Println("Error parsing the album songs XML:", err)
				return err
			}

			if len(s.Album) > 0 {
				if len(s.Album[0].Song) > 0 {
					if s.Album[0].Song[0].Date != "" {
						year, err = ExtractAlbumYear(s.Album[0].Song[0].Date)
						if err != nil {
							log.Println("Error extracting album's year:", err)
						}
					}
				}
			}

			ar, ok := a.AlbumArtists[arName]

			if ok {
				ar.albums = append(ar.albums, album{
					name:        al.Text,
					tracks:      albumTracks,
					year:        year,
					playUrl:     al.PlayURL,
					autoplayUrl: al.AutoplayURL,
					duration:    duration,
				})

				a.AlbumArtists[arName] = ar
			} else {
				a.AlbumArtists[arName] = Artist{
					albums: []album{{
						name:        al.Text,
						tracks:      albumTracks,
						year:        year,
						playUrl:     al.PlayURL,
						autoplayUrl: al.AutoplayURL,
						duration:    duration,
					}},
				}
			}
		}
	}

	a.Artists = SortArtists(a.AlbumArtists)

	// Iterate over sorted artist names
	for _, artistName := range a.Artists {
		ar := a.AlbumArtists[artistName]

		// Sort albums by year
		sort.Slice(ar.albums, func(i, j int) bool {
			return ar.albums[i].year < ar.albums[j].year
		})

		a.AlbumArtists[artistName] = ar
	}

	return nil
}

func (a *App) getTrackURL(name, artist, album string) (string, string, error) {
	for _, a := range a.AlbumArtists[artist].albums {
		if a.name != album {
			continue
		}

		for _, t := range a.tracks {
			if t.name != name {
				continue
			}

			return t.playUrl, t.autoplayUrl, nil
		}
	}

	return "", "", errors.New("no such track")
}

func (a *App) cpHighlightArtist(name string) {
	// clear previously highlighted items
	if a.CpArtistIdx >= 0 {
		n, _ := a.ArtistPane.GetItemText(a.CpArtistIdx)
		a.ArtistPane.SetItemText(a.CpArtistIdx, strings.TrimPrefix(n, "[yellow]"), "")
	}

	if name == "" {
		a.CpArtistIdx = -1
		return
	}

	// highlight artist
	// track is highlighted through a.newAlbumList
	idx := a.ArtistPane.FindItems(name, "", false, true)
	if len(idx) < 1 {
		return
	}

	n, _ := a.ArtistPane.GetItemText(idx[0])
	a.ArtistPane.SetItemText(idx[0], "[yellow]"+n, "")
	a.CpArtistIdx = idx[0]
}
