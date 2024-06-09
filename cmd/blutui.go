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
	Library     *library.Library
	StatusBar   *tview.Table
	HelpScreen  *tview.Modal
	Player      *player.Player
}

func (a *App) AppDraw() {
	a.Application.Draw()
}

func (a *App) Play(url string) {
	go a.Player.Play(url)
}

func (a *App) GetCurrentPage() string {
	n, _ := a.Pages.GetFrontPage()
	return n
}

func (a *App) HighlightCpArtist(name string) {
	a.Library.HighlightCpArtist(name)
}

func (a *App) SetAppFocus(p tview.Primitive) {
	a.Application.SetFocus(p)
}

func (a *App) SetCpTrack(name string) {
	a.Library.CpTrackName = name
}

func main() {
	// Create main app
	a := &App{
		Application: tview.NewApplication(),
	}

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
	sb := statusbar.NewStatusBar(a)
	sbc, err := sb.CreateContainer()
	if err != nil {
		panic(err)
	}

	go sb.Listen(pUpd)

	libc.AddItem(sbc, 0, 1, false)

	// Create Player and hand the instance over to App and Library
	// Start http long-polling Bluesound for updates
	p := player.NewPlayer("http://bluesound.local:11000", pUpd)
	a.Player = p

	go p.PollStatus()

	// a.CreateHelpScreen()

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

	// Set global keymap
	a.Application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return internal.KeyboardHandler(event, a.Library.SelectCpArtist, a.Application.Stop, p)
	})

	// Set app root screen
	if err := a.Application.SetRoot(a.Pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
