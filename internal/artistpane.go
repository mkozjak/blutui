package internal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/tview"
)

// left pane - artists
func (a *App) CreateArtistPane() {
	artistPaneStyle := tcell.Style{}
	artistPaneStyle.Background(tcell.ColorDefault)

	a.ArtistPane = tview.NewList().
		SetHighlightFullLine(true).
		SetWrapAround(false).
		SetSelectedTextColor(tcell.ColorWhite).
		SetSelectedBackgroundColor(tcell.ColorCornflowerBlue).
		ShowSecondaryText(false).
		SetMainTextStyle(artistPaneStyle)

	a.ArtistPane.SetTitle(" [::b]Artist ").
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft).
		SetCustomBorders(CustomBorders).
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

	for _, artist := range a.Artists {
		a.ArtistPane.AddItem(artist, "", 0, nil)
	}

	a.scrollCb()
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

// draw selected artist's right pane (album items) on artist scroll
func (a *App) scrollCb() {
	a.ArtistPane.SetChangedFunc(func(index int, artist string, _ string, shortcut rune) {
		a.AlbumPane.Clear()
		l := a.DrawCurrentArtist(artist, a.AlbumPane)
		a.AlbumPane.SetRows(l...)
	})
}
