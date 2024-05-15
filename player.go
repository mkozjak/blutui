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
	"github.com/mkozjak/tview"
)

var api string = "http://bluesound.local:11000"

type track struct {
	name        string
	length      int
	disc        int
	number      int
	playUrl     string
	autoplayUrl string
}

type album struct {
	name        string
	year        int
	genre       string
	tracks      []track
	playUrl     string
	autoplayUrl string
}

type artist struct {
	albums []album
}

type model struct {
	albumArtists      map[string]artist
	artists           []string
	currentlyPlaying  track
	status            string
	currentAlbumIndex int
	currentAlbumCount int
}

type browse struct {
	Items []item `xml:"item"`
}

type item struct {
	Text        string `xml:"text,attr"`  // album name; track name
	Text2       string `xml:"text2,attr"` // artist name
	BrowseKey   string `xml:"browseKey,attr"`
	Type        string `xml:"type,attr"`
	PlayURL     string `xml:"playURL,attr"`
	AutoplayURL string `xml:"autoplayURL,attr"`
}

var arListStyle = &tview.BoxBorders{
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

var alFlexStyle = arListStyle

var trListStyle = &tview.BoxBorders{}

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
					name:        tr.Text,
					playUrl:     tr.PlayURL,
					autoplayUrl: tr.AutoplayURL,
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
						name:        al.Text,
						tracks:      albumTracks,
						playUrl:     al.PlayURL,
						autoplayUrl: al.AutoplayURL,
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

func (m *model) getTrackURL(name, artist, album string) (string, string, error) {
	for _, a := range m.albumArtists[artist].albums {
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

func main() {
	app := tview.NewApplication()
	m := model{
		albumArtists:      map[string]artist{},
		currentAlbumIndex: 0,
		currentAlbumCount: 0,
	}

	err := m.fetchData()
	if err != nil {
		panic(err)
	}

	// left pane - artists
	arLstStyle := tcell.Style{}
	arLstStyle.Background(tcell.ColorDefault)
	trackLstStyle := arLstStyle

	arLst := tview.NewList().
		SetHighlightFullLine(true).
		SetWrapAround(false).
		SetSelectedTextColor(tcell.ColorWhite).
		SetSelectedBackgroundColor(tcell.ColorCornflowerBlue).
		ShowSecondaryText(false).
		SetMainTextStyle(arLstStyle)

	arLst.SetTitle(" [::b]Artist ").
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft).
		SetCustomBorders(arListStyle).
		// set artists list keymap
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'j':
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			}

			return event
		})

	alFlex := tview.NewFlex().
		SetDirection(tview.FlexRow)

	alFlex.SetTitle(" [::b]Track ").
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft).
		SetCustomBorders(alFlexStyle)

	appFlex := tview.NewFlex().
		AddItem(arLst, 0, 1, true).
		AddItem(alFlex, 0, 2, false)

	for _, artist := range m.artists {
		arLst.AddItem(artist, "", 0, nil)
	}

	// draw selected artist's right pane (album items) on artist scroll
	arLst.SetChangedFunc(func(index int, artist string, _ string, shortcut rune) {
		alFlex.Clear()
		m.currentAlbumCount = len(m.albumArtists[artist].albums)

		for _, album := range m.albumArtists[artist].albums {
			trackLst := tview.NewList().
				SetHighlightFullLine(true).
				SetWrapAround(false).
				SetSelectedFocusOnly(true).
				SetSelectedTextColor(tcell.ColorWhite).
				SetSelectedBackgroundColor(tcell.ColorCornflowerBlue).
				ShowSecondaryText(false).
				SetMainTextStyle(trackLstStyle)

			trackLst.SetSelectedFunc(func(i int, name, _ string, sh rune) {
				_, autoplay, err := m.getTrackURL(name, artist, album.name)
				if err != nil {
					panic(err)
				}

				// play track and add subsequent album tracks to queue
				go func() {
					_, err = http.Get(api + autoplay)
					if err != nil {
						fmt.Println("Error autoplaying track:", err)
						panic(err)
					}

					// arLst.SetItemText(arLst.GetCurrentItem(), "[yellow]"+artist, "")
					// trackLst.SetItemText(i, "[yellow]"+name, "")
				}()
			})

			// set album tracklist keymap
			trackLst.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Rune() {
				case 'j':
					if trackLst.GetCurrentItem()+1 == trackLst.GetItemCount() {
						if m.currentAlbumIndex+1 == m.currentAlbumCount {
							// do nothing, return default
							return nil
						} else {
							app.SetFocus(alFlex.GetItem(m.currentAlbumIndex + 1))
							m.currentAlbumIndex = m.currentAlbumIndex + 1
							return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
						}
					}

					return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
				case 'k':
					// FIXME
					if trackLst.GetCurrentItem() == 0 {
						if m.currentAlbumIndex == 0 {
							// do nothing, i'm already on 1st album
							return nil
						} else {
							app.SetFocus(alFlex.GetItem(m.currentAlbumIndex - 1))
							m.currentAlbumIndex = m.currentAlbumIndex - 1
							return nil
						}
					}

					return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
				}

				return event
			})

			trackLst.SetTitle("[::b]" + album.name).
				SetBorder(true).
				SetBorderColor(tcell.ColorCornflowerBlue).
				SetBackgroundColor(tcell.ColorDefault).
				SetTitleAlign(tview.AlignLeft).
				SetCustomBorders(trListStyle)

			for _, t := range album.tracks {
				trackLst.AddItem(t.name, "", 0, nil)
			}

			alFlex.AddItem(trackLst, trackLst.GetItemCount()+2, 1, true)
		}
	})

	// draw initial album list for the first artist in the list
	app.SetAfterDrawFunc(func(screen tcell.Screen) {
		// disable callback
		app.SetAfterDrawFunc(nil)

		return
	})

	// set global keymap
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlQ:
			app.Stop()
		case tcell.KeyTab:
			artistView := appFlex.GetItem(0)
			albumView := appFlex.GetItem(1)

			if !albumView.HasFocus() {
				app.SetFocus(alFlex)
				arLst.SetSelectedBackgroundColor(tcell.ColorLightGray)
			} else {
				app.SetFocus(artistView)
				arLst.SetSelectedBackgroundColor(tcell.ColorCornflowerBlue)
			}

			m.currentAlbumIndex = 0
			return nil
		}

		return event
	})

	if err := app.SetRoot(appFlex, true).SetFocus(appFlex).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
