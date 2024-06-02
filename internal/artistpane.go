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
		SetCustomBorders(ArtistPaneStyle).
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

// draw selected artist's right pane (album items) on artist scroll
func (a *App) scrollCb() {
	a.ArtistPane.SetChangedFunc(func(index int, artist string, _ string, shortcut rune) {
		a.AlbumPane.Clear()
		l := a.DrawCurrentArtist(artist, a.AlbumPane)
		a.AlbumPane.SetRows(l...)
	})
}
