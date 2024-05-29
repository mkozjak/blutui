package internal

import (
	"encoding/xml"
	"errors"
	"log"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/tview"
)

var ArtistPaneStyle = &tview.BoxBorders{
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

var AlbumPaneStyle = ArtistPaneStyle
var TrListStyle = &tview.BoxBorders{}

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
	currentlyPlaying    track
	status              string
	currentArtistAlbums []*tview.List
}

type Cache struct {
	Data map[string]CacheItem
}

type CacheItem struct {
	Response   []byte
	Expiration time.Time
}

type browse struct {
	Items []item `xml:"item"`
}

type volume struct {
	XMLName xml.Name `xml:"volume"`
	Value   int      `xml:",chardata"`
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
			ar, ok := a.AlbumArtists[arName]

			if ok {
				ar.albums = append(ar.albums, album{
					name:        al.Text,
					tracks:      albumTracks,
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

		// Sort albums alphabetically
		sort.Slice(ar.albums, func(i, j int) bool {
			// FIXME: should sort by year instead
			return ar.albums[i].name < ar.albums[j].name
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

func (a *App) newAlbumList(artist string, album album, c *tview.Grid) *tview.List {
	textStyle := tcell.Style{}
	textStyle.Background(tcell.ColorDefault)
	d := FormatDuration(album.duration)

	trackLst := tview.NewList().
		SetHighlightFullLine(true).
		SetWrapAround(false).
		SetSelectedFocusOnly(true).
		SetSelectedTextColor(tcell.ColorWhite).
		SetSelectedBackgroundColor(tcell.ColorCornflowerBlue).
		ShowSecondaryText(false).
		SetMainTextStyle(textStyle)

	// create a custom list line for album length etc.
	trackLst.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		centerY := y + height/trackLst.GetItemCount()/2

		for cx := x + len(trackLst.GetTitle()) - 3; cx < x+width-len(d)-2; cx++ {
			screen.SetContent(cx, centerY, tview.BoxDrawingsLightHorizontal, nil,
				tcell.StyleDefault.Foreground(tcell.ColorCornflowerBlue))
		}

		// write album length along the horizontal line
		tview.Print(screen, "[::b]"+d, x+1, centerY, width-2, tview.AlignRight, tcell.ColorWhite)

		// space for other content
		return x + 1, centerY + 1, width - 2, height - (centerY + 1 - y)
	})

	trackLst.SetSelectedFunc(func(i int, trackName, _ string, sh rune) {
		_, autoplay, err := a.getTrackURL(trackName, artist, album.name)
		if err != nil {
			panic(err)
		}

		// play track and add subsequent album tracks to queue
		go Play(autoplay)
	})

	// set album tracklist keymap
	trackLst.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j':
			if trackLst.GetCurrentItem()+1 == trackLst.GetItemCount() {
				// reached the end of current album
				// skip to next one if available
				albumIndex, _ := c.GetOffset()

				if albumIndex+1 != len(a.AlbumArtists[artist].albums) {
					// this will redraw the screen
					// TODO: only use SetOffset if the next album cannot fit into the current screen in its entirety
					c.SetOffset(albumIndex+1, 0)
					a.Application.SetFocus(a.currentArtistAlbums[albumIndex+1])
				}
			}

			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			if trackLst.GetCurrentItem() == 0 {
				// reached the beginning of current album
				// skip to previous one if available
				albumIndex, _ := c.GetOffset()

				if albumIndex != 0 {
					// this will redraw the screen
					// TODO: only use SetOffset if the next album cannot fit into the current screen in its entirety
					c.SetOffset(albumIndex-1, 0)
					a.Application.SetFocus(a.currentArtistAlbums[albumIndex-1])
				}
			}

			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}

		return event
	})

	trackLst.
		SetTitle("[::b]" + album.name).
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft).
		SetCustomBorders(TrListStyle)

	for _, t := range album.tracks {
		trackLst.AddItem(t.name, "", 0, nil)
	}

	return trackLst
}

func (a *App) DrawCurrentArtist(artist string, c *tview.Grid) []int {
	l := []int{}
	a.currentArtistAlbums = nil

	for i, album := range a.AlbumArtists[artist].albums {
		albumList := a.newAlbumList(artist, album, c)
		l = append(l, len(album.tracks)+2)

		// automatically focus the first track from the first album
		// since grid is the parent, it will automatically lose focus
		// and give it to the first album
		if i == 0 {
			c.AddItem(albumList, i, 0, 1, 1, 0, 0, true)
		} else {
			c.AddItem(albumList, i, 0, 1, 1, 0, 0, false)
		}

		a.currentArtistAlbums = append(a.currentArtistAlbums, albumList)
	}

	return l
}
