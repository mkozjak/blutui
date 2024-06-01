package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/tview"
)

func main() {
	a := internal.App{
		Application:  tview.NewApplication(),
		AlbumArtists: map[string]internal.Artist{},
	}

	err := a.FetchData()
	if err != nil {
		panic(err)
	}

	artistPane := a.CreateArtistPane()
	albumPane := a.CreateAlbumPane()
	statusBar := a.CreateStatusBar()

	// draw selected artist's right pane (album items) on artist scroll
	artistPane.SetChangedFunc(func(index int, artist string, _ string, shortcut rune) {
		albumPane.Clear()
		l := a.DrawCurrentArtist(artist, albumPane)
		albumPane.SetRows(l...)
	})

	// app
	appFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		// left and right pane
		AddItem(tview.NewFlex().
			AddItem(artistPane, 0, 1, true).
			AddItem(albumPane, 0, 2, false), 0, 1, true).
		// status bar
		AddItem(statusBar, 1, 1, false)

	// draw initial album list for the first artist in the list
	a.Application.SetAfterDrawFunc(func(screen tcell.Screen) {
		l := a.DrawCurrentArtist(a.Artists[0], albumPane)
		albumPane.SetRows(l...)

		// disable callback
		a.Application.SetAfterDrawFunc(nil)
	})

	// set global keymap
	a.Application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlQ:
			a.Application.Stop()
		case tcell.KeyTab:
			if !albumPane.HasFocus() {
				a.Application.SetFocus(albumPane)
				artistPane.SetSelectedBackgroundColor(tcell.ColorLightGray)
			} else {
				a.Application.SetFocus(artistPane)
				artistPane.SetSelectedBackgroundColor(tcell.ColorCornflowerBlue)
			}

			return nil
		case tcell.KeyCtrlB:
			return tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone)
		case tcell.KeyCtrlF:
			return tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone)
		}

		switch event.Rune() {
		case 'p':
			go internal.Playpause()
		case 's':
			go internal.Stop()
		case '>':
			go internal.Next()
		case '<':
			go internal.Previous()
		case '+':
			go internal.VolumeUp()
		case '-':
			go internal.VolumeDown()
		case 'q':
			a.Application.Stop()
		}

		return event
	})

	if err := a.Application.SetRoot(appFlex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
