package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/blutui/internal/library"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/tview"
)

func main() {
	lib := library.NewLibrary("http://bluesound.local:11000")
	libc, err := lib.CreateContainer()
	if err != nil {
		panic(err)
	}

	// libc.AddItem(statusBar, 1, 1, false)

	pUpd := make(chan player.Status)
	p := player.NewPlayer("http://bluesound.local:11000", pUpd)

	// statusBar := statusbar.

	a := internal.App{
		Application: tview.NewApplication(),
	}

	// a.CreateHelpScreen()

	pages := tview.NewPages().
		AddAndSwitchToPage("library", libc, true).
		AddPage("help", a.HelpScreen, false, false)

	pages.SetBackgroundColor(tcell.ColorDefault)

	// draw initial album list for the first artist in the list
	a.Application.SetAfterDrawFunc(func(screen tcell.Screen) {
		l := lib.DrawArtistAlbums(lib.Artists[0], lib.AlbumPane)
		lib.AlbumPane.SetRows(l...)

		// disable callback
		a.Application.SetAfterDrawFunc(nil)
	})

	// set global keymap
	a.Application.SetInputCapture(a.KeyboardHandler)

	// set app root screen
	if err := a.Application.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
