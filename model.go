package main

import (
	"encoding/xml"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/tview"
)

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

type app struct {
	application       *tview.Application
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

func (a *app) fetchData() error {
	albumSectionsEndp := api + "/Browse?key=LocalMusic%3AbySection%2F%252FAlbums%253Fservice%253DLocalMusic"

	resp, err := http.Get(albumSectionsEndp)
	if err != nil {
		log.Println("Error fetching album section list:", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
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
		resp, err = http.Get(api + "/Browse?key=" + url.QueryEscape(item.BrowseKey))
		if err != nil {
			log.Println("Error fetching album section:", err)
			return err
		}
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error reading response body:", err)
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
			// fetch album tracks
			resp, err = http.Get(api + "/Browse?key=" + url.QueryEscape(al.BrowseKey))
			if err != nil {
				log.Println("Error fetching album tracks section:", err)
				return err
			}
			defer resp.Body.Close()

			body, err = io.ReadAll(resp.Body)
			if err != nil {
				log.Println("Error reading response body:", err)
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
				}

				albumTracks = append(albumTracks, track)
			}

			ar, ok := a.albumArtists[al.Text2]
			if ok {
				ar.albums = append(ar.albums, album{
					name:   al.Text,
					tracks: albumTracks,
				})

				a.albumArtists[al.Text2] = ar
			} else {
				a.albumArtists[al.Text2] = artist{
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

	a.artists = SortArtists(a.albumArtists)

	// Iterate over sorted artist names
	for _, artistName := range a.artists {
		ar := a.albumArtists[artistName]

		// Sort albums alphabetically
		sort.Slice(ar.albums, func(i, j int) bool {
			// FIXME: should sort by year instead
			return ar.albums[i].name < ar.albums[j].name
		})

		a.albumArtists[artistName] = ar
	}

	return nil
}

func (a *app) getTrackURL(name, artist, album string) (string, string, error) {
	for _, a := range a.albumArtists[artist].albums {
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

func (a *app) newAlbumList(artist, albumName string, tracks []track) *tview.List {
	textStyle := tcell.Style{}
	textStyle.Background(tcell.ColorDefault)

	trackLst := tview.NewList().
		SetHighlightFullLine(true).
		SetWrapAround(false).
		SetSelectedFocusOnly(true).
		SetSelectedTextColor(tcell.ColorWhite).
		SetSelectedBackgroundColor(tcell.ColorCornflowerBlue).
		ShowSecondaryText(false).
		SetMainTextStyle(textStyle)

	trackLst.SetSelectedFunc(func(i int, trackName, _ string, sh rune) {
		_, autoplay, err := a.getTrackURL(trackName, artist, albumName)
		if err != nil {
			panic(err)
		}

		// play track and add subsequent album tracks to queue
		go play(autoplay)
	})

	// set album tracklist keymap
	trackLst.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j':
			if trackLst.GetCurrentItem()+1 == trackLst.GetItemCount() {
				// skip to the next album if available
				if a.currentAlbumIndex+1 == a.currentAlbumCount {
					// do nothing, return default
					return nil
				} else {
					// a.application.SetFocus(c.GetItem(a.currentAlbumIndex + 1))
					// a.currentAlbumIndex = a.currentAlbumIndex + 1

					return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
				}
			}

			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			// FIXME
			if trackLst.GetCurrentItem() == 0 {
				if a.currentAlbumIndex == 0 {
					// do nothing, i'm already on 1st album
					return nil
				} else {
					// a.application.SetFocus(c.GetItem(a.currentAlbumIndex - 1))
					// a.currentAlbumIndex = a.currentAlbumIndex - 1
					return nil
				}
			}

			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}

		return event
	})

	trackLst.SetTitle("[::b]" + albumName).
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft).
		SetCustomBorders(trListStyle)

	for _, t := range tracks {
		trackLst.AddItem(t.name, "", 0, nil)
	}

	return trackLst
}

func (a *app) drawCurrentArtist(artist string, c *tview.Grid) []int {
	l := []int{}

	for i, album := range a.albumArtists[artist].albums {
		albumList := a.newAlbumList(artist, album.name, album.tracks)
		l = append(l, len(album.tracks)+2)

		// automatically focus the first track from the first album
		// since grid is the parent, it will automatically lose focus
		// and give it to the first album
		if i == 0 {
			c.AddItem(albumList, i, 0, 1, 1, 0, 0, true)
		} else {
			c.AddItem(albumList, i, 0, 1, 1, 0, 0, false)
		}
	}

	return l
}
