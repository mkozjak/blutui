package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/blutui/internal/app"
	"github.com/mkozjak/blutui/internal/bar"
	"github.com/mkozjak/blutui/internal/keyboard"
	"github.com/mkozjak/blutui/internal/library"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/tview"
)

func main() {
	// Create main app
	a := app.New()

	// Create Player and start http long-polling Bluesound for updates
	pUpd := make(chan player.Status)
	p := player.New("http://bluesound.local:11000", pUpd)
	a.Player = p

	// Create Library Page
	lib := library.New("http://bluesound.local:11000", a, p)
	libc, err := lib.CreateContainer()
	if err != nil {
		panic(err)
	}

	// Create a bottom Bar container along with its components
	b := bar.New(a, lib, pUpd)

	// Start listening for Player updates
	go p.PollStatus()

	a.Pages = tview.NewPages().
		AddAndSwitchToPage("library", libc, true)

	a.Pages.SetBackgroundColor(tcell.ColorDefault)

	// Draw initial album list for the first artist in the list
	// Disable callback afterwards
	a.Application.SetAfterDrawFunc(func(screen tcell.Screen) {
		lib.DrawInitAlbums()
		a.Application.SetAfterDrawFunc(nil)
	})

	// Configure global keybindings
	gk := keyboard.NewGlobalHandler(a, a.Player, lib, a.Pages, b)
	a.Application.SetInputCapture(gk.Listen)

	// Configure helpscreen keybindings
	// Attach helpscreen to the app
	hk := keyboard.NewHelpHandler(a.Pages)
	h := internal.CreateHelpScreen(hk.Listen)
	a.Pages.AddPage("help", h, false, false)

	// Draw root app window
	// Root consists of pages (library, etc.) and the status/bottom bar
	a.Root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.Pages, 0, 1, true).
		AddItem(b.StatusContainer(), 1, 0, false)

	// Set app root screen
	if err := a.Application.SetRoot(a.Root, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
