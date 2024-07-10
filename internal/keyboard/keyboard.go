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
	d GlobalDependencies
}

type HelpHandler struct {
	d HelpDependencies
}

func NewGlobalHandler(deps GlobalDependencies) *GlobalHandler {
	return &GlobalHandler{
		d: GlobalDependencies{
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
		h.d.App.Stop()
	}

	if h.d.Bar.CurrentContainer() != "status" {
		return event
	}

	switch event.Rune() {
	case 'p':
		go h.d.Player.Playpause()
	case 's':
		go h.d.Player.Stop()
	case '>':
		go h.d.Player.Next()
	case '<':
		go h.d.Player.Previous()
	case '+':
		go h.d.Player.VolumeHold(true)
	case '-':
		go h.d.Player.VolumeHold(false)
	case 'm':
		go h.d.Player.ToggleMute()
	case 'o':
		if h.d.Player.State() == "play" {
			h.d.Library.SelectCpArtist()
		}
	case 'r':
		go h.d.Player.ToggleRepeatMode()
	case 'u':
		go h.d.Library.UpdateData()
	case 'h':
		p, _ := h.d.Pages.GetFrontPage()
		if p != "help" {
			h.d.Pages.ShowPage("help")
			return nil
		}
	case '/':
		p, _ := h.d.Pages.GetFrontPage()
		if p == "help" || h.d.Library.IsFiltered() {
			return event
		}

		h.d.Bar.Show("search")
		h.d.App.SetFocus(h.d.Bar.SearchContainer())
		return nil
	case 'q':
		h.d.App.Stop()
	}

	return event
}

func NewHelpHandler(deps HelpDependencies) *HelpHandler {
	return &HelpHandler{
		d: HelpDependencies{
			Pages: deps.Pages,
		},
	}
}

func (k *HelpHandler) Listen(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		k.d.Pages.HidePage("help")
		return nil
	}

	switch event.Rune() {
	case 'h':
		p, _ := k.d.Pages.GetFrontPage()
		if p == "help" {
			k.d.Pages.HidePage("help")
			return nil
		}
	}

	return event
}
