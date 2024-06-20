package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/blutui/internal/app"
	"github.com/mkozjak/blutui/internal/bar"
	"github.com/mkozjak/blutui/internal/keyboard"
	"github.com/mkozjak/blutui/internal/library"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/blutui/spinner"
	"github.com/mkozjak/tview"
)

func main() {
	// Create main app
	a := app.New()
	sp := spinner.New(a.Draw)

	// Create Player and start http long-polling Bluesound for updates
	pUpd := make(chan player.Status)
	p := player.New("http://bluesound.local:11000", sp, pUpd)
	a.Player = p

	// Create Library Page
	fc := make(chan library.FetchDone)
	lib := library.New("http://bluesound.local:11000", a, p, sp)
	libc := lib.CreateContainer()

	// Start initial fetching of data
	go lib.FetchData(true, fc)

	go func() {
		for {
			msg := <-fc
			if msg.Error != nil {
				// TODO: should probably use os.Exit(1) here
				panic("failed fetching initial data: " + msg.Error.Error())
			}

			// Draw initial album list for the first artist in the list
			lib.DrawArtistPane()
			lib.DrawInitAlbums()
			return
		}
	}()

	a.Library = libc

	// Create a bottom Bar container along with its components
	b := bar.New(a, lib, sp, pUpd)

	// Start listening for Player updates
	go p.PollStatus()

	a.Pages = tview.NewPages().
		AddAndSwitchToPage("library", a.Library, true)

	a.Pages.SetBackgroundColor(tcell.ColorDefault)

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
