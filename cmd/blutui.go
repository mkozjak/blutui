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
		CpArtistIdx:  -1,
	}

	err := a.FetchData()
	if err != nil {
		panic(err)
	}

	a.CreateArtistPane()
	a.CreateAlbumPane()
	a.CreateStatusBar()

	// library page
	libFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		// left and right pane
		AddItem(tview.NewFlex().
			AddItem(a.ArtistPane, 0, 1, true).
			AddItem(a.AlbumPane, 0, 2, false), 0, 1, true).
		// status bar
		AddItem(a.StatusBar, 1, 1, false)

	libFlex.SetInputCapture(a.KbLibHandler)

	// draw initial album list for the first artist in the list
	a.Application.SetAfterDrawFunc(func(screen tcell.Screen) {
		l := a.DrawCurrentArtist(a.Artists[0], a.AlbumPane)
		a.AlbumPane.SetRows(l...)

		// disable callback
		a.Application.SetAfterDrawFunc(nil)
	})

	// set global keymap
	a.Application.SetInputCapture(a.KbGlobalHandler)

	// set app root screen
	if err := a.Application.SetRoot(libFlex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
