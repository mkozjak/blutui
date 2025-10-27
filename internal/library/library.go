package library

import (
	"encoding/xml"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/mkozjak/blutui/cache"
	internal "github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/blutui/internal/app"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/blutui/spinner"
	"github.com/mkozjak/tview"
)

var localRootEndpoint string = "/Browse?key=LocalMusic%3AbySection%3AAlbum%2F%252Flibrary%252Fv1%252FAlbums%253Fservice%3DLocalMusic"
var tidalRootEndpoint string = "/Browse?key=%2FAlbums%3Fservice%3DTidal%26amp%3BbrowseIsFavouritesContext%3D1%26amp%3Bcategory%3DFAVOURITES"

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
	Text           string `xml:"text,attr"`  // album name; track name
	Text2          string `xml:"text2,attr"` // artist name
	BrowseKey      string `xml:"browseKey,attr"`
	Type           string `xml:"type,attr"`
	PlayURL        string `xml:"playURL,attr"`
	AutoplayURL    string `xml:"autoplayURL,attr"`
	ContextMenuKey string `xml:"contextMenuKey,attr"`
	Duration       string `xml:"duration,attr"`
	Image          string `xml:"image,attr"`
	Tracks         string `xml:"tracks,attr"`
}

// Used for parsing data from /Songs
type songs struct {
	Song []struct {
		Date string `xml:"date"`
		Fn   string `xml:"fn"`
	} `xml:"song"`
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
	FetchData(cached bool, doneCh chan<- FetchDone)
	UpdateData()
	IsFiltered() bool
	SelectCpArtist()
	ArtistFilter
	CPMarkSetter
}

type ArtistFilter interface {
	Artists() []string
	FilterArtistPane(f []string)
}

type CPMarkSetter interface {
	MarkCpArtist(name string)
	MarkCpTrack(track, artist, album string)
	SetCpAlbumName(name string)
	SetCpTrackName(name string)
}

type FetchDone struct {
	Service string
	Error   error
}

type Library struct {
	container *tview.Flex
	app       app.Focuser
	player    player.Controller
	spinner   spinner.StartStopper
	API       string
	service   string

	// TODO: should move these into a separate ap struct?
	artistPane          *tview.List
	artistPaneFiltered  bool
	albumPane           *tview.Grid
	albumArtists        map[string]artist
	artists             []string
	currentArtistAlbums []*tview.Table
	cpArtistIdx         int
	CpAlbumName         string
	CpTrackName         string
}

func New(api, service string, a app.Focuser, p player.Controller, sp spinner.StartStopper) *Library {
	return &Library{
		app:                a,
		player:             p,
		spinner:            sp,
		API:                api,
		service:            service,
		albumArtists:       map[string]artist{},
		cpArtistIdx:        -1,
		artistPaneFiltered: false,
	}
}

func (l *Library) Artists() []string {
	return l.artists
}

func (l *Library) CreateContainer() *tview.Flex {
	l.artistPane = l.createArtistContainer()
	l.DrawArtistPane()
	l.albumPane = l.createAlbumContainer()

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		// left and right pane
		AddItem(tview.NewFlex().
			AddItem(l.artistPane, 0, 1, true).
			AddItem(l.albumPane, 0, 2, false), 0, 1, true)

	flex.SetInputCapture(l.KeyboardHandler)

	return flex
}

func (l *Library) FetchData(cached bool, doneCh chan<- FetchDone) {
	go l.spinner.Start()

	c, err := cache.LoadCache()
	if err != nil {
		internal.Log("Error loading local cache:", err)
		doneCh <- FetchDone{Error: err}
		return
	}

	var albums browse

	fetchAlbums := func(url string) {
		body, err := cache.FetchAndCache(url, c, cached)
		if err != nil {
			internal.Log("Error fetching album sections:", err)
			doneCh <- FetchDone{Error: err}
			return
		}

		err = xml.Unmarshal(body, &albums)
		if err != nil {
			internal.Log("Error parsing the albums XML:", err, url)
			doneCh <- FetchDone{Error: err}
			return
		}
	}

	l.albumArtists = make(map[string]artist)

	if l.service == "local" {
		body, err := cache.FetchAndCache(l.API+localRootEndpoint, c, cached)
		if err != nil {
			internal.Log("Error fetching/caching data:", err)
			doneCh <- FetchDone{Error: err}
			return
		}

		var sections browse
		err = xml.Unmarshal(body, &sections)
		if err != nil {
			internal.Log("Error parsing the sections XML:", err, string(body))
			doneCh <- FetchDone{Error: err}
			return
		}

		// parse album sections (alphabetical order) from xml
		for _, item := range sections.Items {
			fetchAlbums(l.API + "/Browse?key=" + url.QueryEscape(item.BrowseKey))
		}
	} else if l.service == "tidal" {
		fetchAlbums(l.API + tidalRootEndpoint)
	}

	// iterate albums and fill l.albumArtists
	for _, al := range albums.Items {
		var duration int

		// fetch album tracks
		body, err := cache.FetchAndCache(l.API+"/Browse?key="+url.QueryEscape(al.BrowseKey), c, cached)
		if err != nil {
			internal.Log("Error fetching album tracks:", err)
			doneCh <- FetchDone{Error: err}
			return
		}

		var tracks browse

		err = xml.Unmarshal(body, &tracks)
		if err != nil {
			internal.Log("Error parsing the album tracks XML:", err, "body:", string(body))
			doneCh <- FetchDone{Error: err}
			return
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
		alEsc := url.QueryEscape(al.Text)
		arEsc := url.QueryEscape(al.Text2)

		body, err = cache.FetchAndCache(l.API+"/Songs?service=LocalMusic&album="+alEsc+"&artist="+arEsc, c, cached)
		if err != nil {
			internal.Log("Error fetching album date:", err)
			doneCh <- FetchDone{Error: err}
			return
		}

		var s songs
		var year int

		err = xml.Unmarshal(body, &s)
		if err != nil {
			internal.Log("Error parsing the album songs XML:", err, "body:", string(body))
			doneCh <- FetchDone{Error: err}
			return
		}

		if len(s.Song) > 0 {
			d := s.Song[0].Date

			if d != "" && d != "0" {
				year, err = internal.ExtractAlbumYear(d)
				if err != nil {
					internal.Log("Error extracting album's year:", err)
				}
			} else {
				year, err = internal.HackAlbumYear(s.Song[0].Fn)
				if err != nil {
					year, err = internal.ExtractYearFromPath(s.Song[0].Fn)
					if err != nil {
						internal.Log("Error extracting album's year from path:", err)
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

	l.spinner.Stop()
	doneCh <- FetchDone{Error: nil}
}

func (l *Library) IsFiltered() bool {
	return l.artistPaneFiltered
}

func (l *Library) UpdateData() {
	ch := make(chan FetchDone)
	go l.FetchData(false, ch)

	for {
		msg := <-ch
		if msg.Error != nil {
			// TODO: show error on bar
			panic("failed fetching initial data: " + msg.Error.Error())
		}

		// Refresh artist pane
		l.DrawArtistPane()
		l.app.SetFocus(l.artistPane)
	}
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

func (l *Library) SetCpAlbumName(name string) {
	l.CpAlbumName = name
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
