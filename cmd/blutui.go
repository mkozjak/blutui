package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/blutui/internal/app"
	"github.com/mkozjak/blutui/internal/keyboard"
	"github.com/mkozjak/blutui/internal/library"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/blutui/internal/statusbar"
	"github.com/mkozjak/tview"
)

func main() {
	// Create main app
	a := app.NewApp()

	pUpd := make(chan player.Status)

	// Create Library Page
	a.Library = library.NewLibrary("http://bluesound.local:11000", a)
	libc, err := a.Library.CreateContainer()
	if err != nil {
		panic(err)
	}

	// Create Status Bar container and attach it to Library
	// Hand over the Library instance to Status Bar
	// Start listening for Player updates
	sb := statusbar.NewStatusBar(a, a.Library)
	sbc, err := sb.CreateContainer()
	if err != nil {
		panic(err)
	}

	go sb.Listen(pUpd)

	libc.AddItem(sbc, 1, 1, false)

	// Create Player and hand the instance over to App and Library
	// Start http long-polling Bluesound for updates
	p := player.NewPlayer("http://bluesound.local:11000", pUpd)
	a.Player = p

	go p.PollStatus()

	a.Pages = tview.NewPages().
		AddAndSwitchToPage("library", libc, true)
		// AddPage("help", a.HelpScreen, false, false)

	a.Pages.SetBackgroundColor(tcell.ColorDefault)

	// Draw initial album list for the first artist in the list
	// Disable callback afterwards
	a.Application.SetAfterDrawFunc(func(screen tcell.Screen) {
		a.Library.DrawInitAlbums()
		a.Application.SetAfterDrawFunc(nil)
	})

	// Configure global keybindings
	gk := keyboard.NewGlobalHandler(a.Application, a.Player, a.Library, a.Pages)
	a.Application.SetInputCapture(gk.Listen)

	// Configure helpscreen keybindings
	// Attach helpscreen to the app
	hk := keyboard.NewHelpHandler(a.Pages)
	h := internal.CreateHelpScreen(hk.Listen)
	a.Pages.AddPage("help", h, false, false)

	// Set app root screen
	if err := a.Application.SetRoot(a.Pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
