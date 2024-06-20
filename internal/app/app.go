package app

import (
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/tview"
)

type Command interface {
	Draw() *tview.Application
	CurrentPage() string
	PrevFocused() tview.Primitive
	SetFocus(p tview.Primitive) *tview.Application
	SetPrevFocused(p string)
	ShowComponent(p tview.Primitive)
	Stop()
}

type Drawer interface {
	Draw() *tview.Application
}

type App struct {
	Application *tview.Application
	Root        *tview.Flex
	Library     *tview.Flex
	Pages       *tview.Pages
	StatusBar   *tview.Table
	HelpScreen  *tview.Modal
	Player      *player.Player
	prevFocused string
}

func New() *App {
	return &App{
		Application: tview.NewApplication(),
	}
}

func (a *App) Draw() *tview.Application {
	return a.Application.Draw()
}

func (a *App) CurrentPage() string {
	n, _ := a.Pages.GetFrontPage()
	return n
}

func (a *App) Play(url string) {
	go a.Player.Play(url)
}

func (a *App) PrevFocused() tview.Primitive {
	switch a.prevFocused {
	case "artistpane":
		fallthrough
	case "albumpane":
		return a.Library.GetItem(0)
	}

	return nil
}

func (a *App) SetPrevFocused(p string) {
	a.prevFocused = p
}

func (a *App) SetFocus(p tview.Primitive) *tview.Application {
	return a.Application.SetFocus(p)
}

func (a *App) ShowComponent(c tview.Primitive) {
	bc := a.Root.GetItem(1)
	a.Root.RemoveItem(bc)
	a.Root.AddItem(c, 1, 0, true)
}

func (a *App) Stop() {
	a.Application.Stop()
}
