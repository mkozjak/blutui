package app

import (
	"github.com/mkozjak/blutui/internal/library"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/tview"
)

type Command interface {
	Draw() *tview.Application
	Stop()
	SetFocus(p tview.Primitive) *tview.Application
}

type App struct {
	Application *tview.Application
	Pages       *tview.Pages
	Library     *library.Library
	StatusBar   *tview.Table
	HelpScreen  *tview.Modal
	Player      *player.Player
}

func NewApp() *App {
	return &App{
		Application: tview.NewApplication(),
	}
}

func (a *App) AppDraw() {
	a.Application.Draw()
}

func (a *App) AppQuit() {
	a.Application.Stop()
}

func (a *App) Play(url string) {
	go a.Player.Play(url)
}

func (a *App) GetCurrentPage() string {
	n, _ := a.Pages.GetFrontPage()
	return n
}

// func (a *App) HighlightCpArtist(name string) {
// 	a.Library.HighlightCpArtist(name)
// }

func (a *App) SetAppFocus(p tview.Primitive) {
	a.Application.SetFocus(p)
}

// func (a *App) SetCpTrack(name string) {
// 	a.Library.CpTrackName = name
// }
