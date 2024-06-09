package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/blutui/internal/library"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/blutui/internal/statusbar"
	"github.com/mkozjak/tview"
)

type App struct {
	Application *tview.Application
	Pages       *tview.Pages
	StatusBar   *tview.Table
	HelpScreen  *tview.Modal
	Player      *player.Player
	playerState string
}

func (a *App) getCurrentPage() string {
	n, _ := a.Pages.GetFrontPage()
	return n
}

func main() {
	// Create main app
	a := App{
		Application: tview.NewApplication(),
	}

	pUpd := make(chan player.Status)

	// Create Library Page
	lib := library.NewLibrary("http://bluesound.local:11000", a.Application.SetFocus)
	libc, err := lib.CreateContainer()
	if err != nil {
		panic(err)
	}

	// Create Status Bar container and attach it to Library
	// Hand over the Library instance to Status Bar
	// Start listening for Player updates
	sb := statusbar.NewStatusBar()
	sbc, err := sb.CreateContainer()
	if err != nil {
		panic(err)
	}

	// TODO: move this into a constructor
	sb.Library = lib
	sb.GetCurrentPage = a.getCurrentPage
	sb.AppDraw = a.Application.Draw()
	sb.Listen(pUpd)
	lib.AddItem(sbc, 1, false)

	// Create Player and hand the instance over to App and Library
	// Start http long-polling Bluesound for updates
	p := player.NewPlayer("http://bluesound.local:11000", pUpd)
	a.Player = p
	lib.Player = p

	go p.PollStatus()

	// a.CreateHelpScreen()

	a.Pages = tview.NewPages().
		AddAndSwitchToPage("library", libc, true)
		// AddPage("help", a.HelpScreen, false, false)

	a.Pages.SetBackgroundColor(tcell.ColorDefault)

	// Draw initial album list for the first artist in the list
	// Disable callback afterwards
	a.Application.SetAfterDrawFunc(func(screen tcell.Screen) {
		lib.DrawInitAlbums()
		a.Application.SetAfterDrawFunc(nil)
	})

	// Set global keymap
	a.Application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return internal.KeyboardHandler(event, lib.SelectCpArtist, a.Application.Stop, p)
	})

	// Set app root screen
	if err := a.Application.SetRoot(a.Pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
