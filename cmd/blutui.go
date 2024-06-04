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
		CpArtistIdx: -1,
	}

	err := a.FetchData()
	if err != nil {
		panic(err)
	}

	a.CreateArtistPane()
	a.CreateAlbumPane()
	a.CreateStatusBar()

	// app
	appFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		// left and right pane
		AddItem(tview.NewFlex().
			AddItem(a.ArtistPane, 0, 1, true).
			AddItem(a.AlbumPane, 0, 2, false), 0, 1, true).
		// status bar
		AddItem(a.StatusBar, 1, 1, false)

	// draw initial album list for the first artist in the list
	a.Application.SetAfterDrawFunc(func(screen tcell.Screen) {
		l := a.DrawCurrentArtist(a.Artists[0], a.AlbumPane)
		a.AlbumPane.SetRows(l...)

		// disable callback
		a.Application.SetAfterDrawFunc(nil)
	})

	// set global keymap
	a.Application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlQ:
			a.Application.Stop()
		case tcell.KeyTab:
			if !a.AlbumPane.HasFocus() {
				a.Application.SetFocus(a.AlbumPane)
				a.ArtistPane.SetSelectedBackgroundColor(tcell.ColorLightGray)
			} else {
				a.Application.SetFocus(a.ArtistPane)
				a.ArtistPane.SetSelectedBackgroundColor(tcell.ColorCornflowerBlue)
			}

			return nil
		case tcell.KeyCtrlB:
			return tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone)
		case tcell.KeyCtrlF:
			return tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone)
		}

		switch event.Rune() {
		case 'p':
			go a.Playpause()
		case 's':
			go a.Stop()
		case '>':
			go a.Next()
		case '<':
			go a.Previous()
		case '+':
			go a.VolumeUp()
		case '-':
			go a.VolumeDown()
		case 'm':
			go a.ToggleMute()
		case 'q':
			a.Application.Stop()
		}

		return event
	})

	if err := a.Application.SetRoot(appFlex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
