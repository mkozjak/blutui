package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var api string = "http://bluesound.local:11000"

type track struct {
	name    string
	length  int
	disc    int
	number  int
	playUrl string
}

type album struct {
	name    string
	year    int
	genre   string
	tracks  []track
	playUrl string
}

type artist struct {
	albums []album
}

type model struct {
	albumArtists     map[string]artist
	artists          []string
	currentlyPlaying track
	status           string
	cursor           int
	lines            int
}

type browse struct {
	Items []item `xml:"item"`
}

type item struct {
	Text      string `xml:"text,attr"`  // album name; track name
	Text2     string `xml:"text2,attr"` // artist name
	BrowseKey string `xml:"browseKey,attr"`
	Type      string `xml:"type,attr"`
	PlayURL   string `xml:"playURL,attr"`
}

func (m *model) fetchData() error {
	albumSectionsEndp := api + "/Browse?key=LocalMusic%3AbySection%2F%252FAlbums%253Fservice%253DLocalMusic"

	resp, err := http.Get(albumSectionsEndp)
	if err != nil {
		fmt.Println("Error fetching album section list:", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return err
	}

	var sections browse
	err = xml.Unmarshal(body, &sections)
	if err != nil {
		fmt.Println("Error parsing the sections XML:", err)
		return err
	}

	// parse album sections (alphabetical order) from xml
	for _, item := range sections.Items {
		resp, err = http.Get(api + "/Browse?key=" + url.QueryEscape(item.BrowseKey))
		if err != nil {
			fmt.Println("Error fetching album section:", err)
			return err
		}
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return err
		}

		var albums browse
		err = xml.Unmarshal(body, &albums)
		if err != nil {
			fmt.Println("Error parsing the albums XML:", err)
			return err
		}

		// iterate albums and fill m.albumArtists
		for _, al := range albums.Items {
			// fetch album tracks
			resp, err = http.Get(api + "/Browse?key=" + url.QueryEscape(al.BrowseKey))
			if err != nil {
				fmt.Println("Error fetching album tracks section:", err)
				return err
			}
			defer resp.Body.Close()

			body, err = io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				return err
			}

			var tracks browse
			err = xml.Unmarshal(body, &tracks)
			if err != nil {
				fmt.Println("Error parsing the album tracks XML:", err)
				return err
			}

			var albumTracks []track
			for _, tr := range tracks.Items {
				track := track{
					name:    tr.Text,
					playUrl: tr.PlayURL,
				}

				albumTracks = append(albumTracks, track)
			}

			ar, ok := m.albumArtists[al.Text2]
			if ok {
				ar.albums = append(ar.albums, album{
					name:   al.Text,
					tracks: albumTracks,
				})

				m.albumArtists[al.Text2] = ar
			} else {
				m.albumArtists[al.Text2] = artist{
					albums: []album{{
						name:    al.Text,
						tracks:  albumTracks,
						playUrl: al.PlayURL,
					}},
				}
			}
		}
	}

	m.artists = SortArtists(m.albumArtists)

	// Iterate over sorted artist names
	for _, artistName := range m.artists {
		ar := m.albumArtists[artistName]

		// Sort albums alphabetically
		sort.Slice(ar.albums, func(i, j int) bool {
			// FIXME: should sort by year instead
			return ar.albums[i].name < ar.albums[j].name
		})

		m.albumArtists[artistName] = ar
	}

	return nil
}

func (m *model) getTrackURL(name, artist, album string) (string, error) {
	for _, a := range m.albumArtists[artist].albums {
		if a.name != album {
			continue
		}

		for _, t := range a.tracks {
			if t.name != name {
				continue
			}

			return t.playUrl, nil
		}
	}

	return "", errors.New("no such track")
}

func main() {
	app := tview.NewApplication()
	m := model{albumArtists: map[string]artist{}}

	err := m.fetchData()
	if err != nil {
		panic(err)
	}

	// left pane - artists
	arLst := tview.NewList().
		SetHighlightFullLine(true).
		SetWrapAround(false).
		SetSelectedTextColor(tcell.ColorWhite).
		SetSelectedBackgroundColor(tcell.ColorMediumBlue).
		ShowSecondaryText(false)

	alFlex := tview.NewFlex().
		SetDirection(tview.FlexRow)

	appFlex := tview.NewFlex().
		AddItem(arLst, 0, 1, true).
		AddItem(alFlex, 0, 2, false)

	for _, artist := range m.artists {
		arLst.AddItem(artist, "", 0, nil)
	}

	// draw selected artist's right pane (album items) on artist scroll
	arLst.SetChangedFunc(func(index int, artist string, _ string, shortcut rune) {
		alFlex.Clear()

		for _, album := range m.albumArtists[artist].albums {
			trackLst := tview.NewList().
				SetHighlightFullLine(true).
				SetWrapAround(false).
				SetSelectedFocusOnly(true).
				SetSelectedTextColor(tcell.ColorWhite).
				SetSelectedBackgroundColor(tcell.ColorMediumBlue).
				ShowSecondaryText(false)

			trackLst.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyEnter:
					i := trackLst.GetCurrentItem()
					name, _ := trackLst.GetItemText(i)

					u, err := m.getTrackURL(name, artist, album.name)
					if err != nil {
						panic(err)
					}

					_, err = http.Get(api + u)
					if err != nil {
						fmt.Println("Error playing track:", err)
						panic(err)
					}

				}

				return event
			})

			for _, t := range album.tracks {
				trackLst.AddItem(t.name, "", 0, nil)
			}

			// TODO: add album title somehow
			alFlex.AddItem(trackLst, 0, 2, false)
		}
	})

	// draw initial album list for the first artist in the list
	app.SetAfterDrawFunc(func(screen tcell.Screen) {
		// disable callback
		app.SetAfterDrawFunc(nil)

		return
	})

	// set keymap
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j':
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}

		switch event.Key() {
		case tcell.KeyCtrlQ:
			app.Stop()
		case tcell.KeyTab:
			artistView := appFlex.GetItem(0)
			albumView := appFlex.GetItem(1)

			if !albumView.HasFocus() {
				app.SetFocus(alFlex.GetItem(0))
			} else {
				app.SetFocus(artistView)
			}

			return nil
		}

		return event
	})

	if err := app.SetRoot(appFlex, true).SetFocus(appFlex).Run(); err != nil {
		panic(err)
	}
}
