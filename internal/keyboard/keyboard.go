package keyboard

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal/app"
	"github.com/mkozjak/blutui/internal/bar"
	"github.com/mkozjak/blutui/internal/library"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/tview"
)

type pagesManager interface {
	GetFrontPage() (string, tview.Primitive)
	ShowPage(name string) *tview.Pages
	HidePage(name string) *tview.Pages
}

type GlobalDependencies struct {
	App     app.Command
	Player  player.Controller
	Library library.Command
	Pages   pagesManager
	Bar     *bar.Bar
}

type HelpDependencies struct {
	Pages pagesManager
}

type GlobalHandler struct {
	GlobalDependencies
}

type HelpHandler struct {
	HelpDependencies
}

func NewGlobalHandler(deps GlobalDependencies) *GlobalHandler {
	return &GlobalHandler{
		GlobalDependencies{
			App:     deps.App,
			Player:  deps.Player,
			Library: deps.Library,
			Pages:   deps.Pages,
			Bar:     deps.Bar,
		},
	}
}

func (h *GlobalHandler) Listen(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlQ:
		h.App.Stop()
	}

	if h.Bar.CurrentContainer() != "status" {
		return event
	}

	switch event.Rune() {
	case 'p':
		go h.Player.Playpause()
	case 's':
		go h.Player.Stop()
	case '>':
		go h.Player.Next()
	case '<':
		go h.Player.Previous()
	case '+':
		go h.Player.VolumeHold(true)
	case '-':
		go h.Player.VolumeHold(false)
	case 'm':
		go h.Player.ToggleMute()
	case 'o':
		if h.Player.State() == "play" {
			h.Library.SelectCpArtist()
		}
	case 'r':
		go h.Player.ToggleRepeatMode()
	case 'u':
		go h.Library.UpdateData()
	case 'h':
		p, _ := h.Pages.GetFrontPage()
		if p != "help" {
			h.Pages.ShowPage("help")
			return nil
		}
	case '/':
		p, _ := h.Pages.GetFrontPage()
		if p == "help" || h.Library.IsFiltered() {
			return event
		}

		h.Bar.Show("search")
		h.App.SetFocus(h.Bar.SearchContainer())
		return nil
	case 'q':
		h.App.Stop()
	}

	return event
}

func NewHelpHandler(deps HelpDependencies) *HelpHandler {
	return &HelpHandler{
		HelpDependencies{
			Pages: deps.Pages,
		},
	}
}

func (k *HelpHandler) Listen(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		k.Pages.HidePage("help")
		return nil
	}

	switch event.Rune() {
	case 'h':
		p, _ := k.Pages.GetFrontPage()
		if p == "help" {
			k.Pages.HidePage("help")
			return nil
		}
	}

	return event
}
