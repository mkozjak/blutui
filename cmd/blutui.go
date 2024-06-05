package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/tview"
)

func main() {
	a := internal.App{
		Application:  tview.NewApplication(),
		Pages:        tview.NewPages(),
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

	// help
	// TODO: do this the smarter way
	help := tview.NewModal().
		SetText("enter - start playback\np - play/pause\ns - stop\n> - next song\n" +
			"< - previous song\n+ - volume up\n- - volume down\nm - toggle mute\n" +
			"o - jump to currently playing artist\nu - update library\n" +
			"h - show help screen\nq - quit app").
		SetBackgroundColor(tcell.ColorDefault)

	help.SetInputCapture(a.KbHelpHandler).
		SetBorder(true).
		SetTitle("[::b]Keybindings").
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetCustomBorders(internal.CustomBorders)

	// app
	a.Pages.AddAndSwitchToPage("library", libFlex, true).
		AddPage("help", help, false, false).
		SetBackgroundColor(tcell.ColorDefault)

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
	if err := a.Application.SetRoot(a.Pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
