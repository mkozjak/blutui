package library

import (
	"encoding/xml"
	"errors"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"

	internal "github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/blutui/internal/app"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/tview"
)

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

type artist struct {
	albums []album
}

type Command interface {
	Artists() []string
	SelectCpArtist()
	HighlightCpArtist(name string)
	SetCpTrackName(name string)
}

type Library struct {
	container           *tview.Flex
	app                 app.Command
	player              player.Command
	API                 string
	artistPane          *tview.List
	albumPane           *tview.Grid
	albumArtists        map[string]artist
	artists             []string
	currentArtistAlbums []*tview.List
	cpArtistIdx         int
	CpTrackName         string
}

func New(api string, a app.Command, p player.Command) *Library {
	return &Library{
		app:          a,
		player:       p,
		API:          api,
		albumArtists: map[string]artist{},
		cpArtistIdx:  -1,
	}
}

func (l *Library) Artists() []string {
	return l.artists
}

func (l *Library) CreateContainer() (*tview.Flex, error) {
	err := l.fetchData()
	if err != nil {
		return nil, err
	}

	l.artistPane = l.drawArtistPane()
	l.albumPane = l.drawAlbumPane()

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		// left and right pane
		AddItem(tview.NewFlex().
			AddItem(l.artistPane, 0, 1, true).
			AddItem(l.albumPane, 0, 2, false), 0, 1, true)

	flex.SetInputCapture(l.KeyboardHandler)

	return flex, nil
}

func (l *Library) fetchData() error {
	cache, err := internal.LoadCache()
	if err != nil {
		log.Println("Error loading local cache:", err)
		return err
	}

	body, err := internal.FetchAndCache(l.API+"/Browse?key=LocalMusic%3AbySection%2F%252FAlbums%253Fservice%253DLocalMusic", cache)
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
		body, err = internal.FetchAndCache(l.API+"/Browse?key="+url.QueryEscape(item.BrowseKey), cache)
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
			body, err = internal.FetchAndCache(l.API+"/Browse?key="+url.QueryEscape(al.BrowseKey), cache)
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

			arName := internal.Caser(al.Text2)

			// fetch album date from /Songs
			body, err = internal.FetchAndCache(
				strings.ReplaceAll(l.API+"/Songs?service=LocalMusic&album="+al.Text+"&artist="+arName, " ", "+"),
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
						year, err = internal.ExtractAlbumYear(s.Album[0].Song[0].Date)
						if err != nil {
							log.Println("Error extracting album's year:", err)
						}
					}
				}
			}

			ar, ok := l.albumArtists[arName]

			if ok {
				ar.albums = append(ar.albums, album{
					name:        al.Text,
					tracks:      albumTracks,
					year:        year,
					playUrl:     al.PlayURL,
					autoplayUrl: al.AutoplayURL,
					duration:    duration,
				})

				l.albumArtists[arName] = ar
			} else {
				l.albumArtists[arName] = artist{
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

	l.artists = sortArtists(l.albumArtists)

	// Iterate over sorted artist names
	for _, artistName := range l.artists {
		ar := l.albumArtists[artistName]

		// Sort albums by year
		sort.Slice(ar.albums, func(i, j int) bool {
			return ar.albums[i].year < ar.albums[j].year
		})

		l.albumArtists[artistName] = ar
	}

	return nil
}

func (l *Library) RefreshData() {
}

func (l *Library) trackURL(name, artist, album string) (string, string, error) {
	for _, a := range l.albumArtists[artist].albums {
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

func (l *Library) SetCpTrackName(name string) {
	l.CpTrackName = name
}

func sortArtists(input map[string]artist) []string {
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
