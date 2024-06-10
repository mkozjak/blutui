package app

import (
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/tview"
)

type Command interface {
	Draw() *tview.Application
	GetCurrentPage() string
	SetFocus(p tview.Primitive) *tview.Application
	Stop()
}

type App struct {
	Application *tview.Application
	Pages       *tview.Pages
	StatusBar   *tview.Table
	HelpScreen  *tview.Modal
	Player      *player.Player
}

func NewApp() *App {
	return &App{
		Application: tview.NewApplication(),
	}
}

func (a *App) Draw() *tview.Application {
	return a.Application.Draw()
}

func (a *App) GetCurrentPage() string {
	n, _ := a.Pages.GetFrontPage()
	return n
}

func (a *App) Play(url string) {
	go a.Player.Play(url)
}

func (a *App) SetFocus(p tview.Primitive) *tview.Application {
	return a.Application.SetFocus(p)
}

func (a *App) Stop() {
	a.Application.Stop()
}
