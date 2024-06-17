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

type GlobalHandler struct {
	app     app.Command
	player  player.Command
	library library.Command
	pages   pagesManager
	bar     *bar.Bar
}

func NewGlobalHandler(a app.Command, p player.Command, l library.Command, pg pagesManager, b *bar.Bar) *GlobalHandler {
	return &GlobalHandler{
		app:     a,
		player:  p,
		library: l,
		pages:   pg,
		bar:     b,
	}
}

func (h *GlobalHandler) Listen(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlQ:
		h.app.Stop()
	}

	if h.bar.CurrentContainer() != "status" {
		return event
	}

	switch event.Rune() {
	case 'p':
		go h.player.Playpause()
	case 's':
		go h.player.Stop()
	case '>':
		go h.player.Next()
	case '<':
		go h.player.Previous()
	case '+':
		go h.player.VolumeHold(true)
	case '-':
		go h.player.VolumeHold(false)
	case 'm':
		go h.player.ToggleMute()
	case 'o':
		if h.player.State() == "play" {
			h.library.SelectCpArtist()
		}
	case 'u':
		// TODO: show error on bar
		go h.library.FetchData(false)
	case 'h':
		p, _ := h.pages.GetFrontPage()
		if p != "help" {
			h.pages.ShowPage("help")
			return nil
		}
	case '/':
		p, _ := h.pages.GetFrontPage()
		if p == "help" || h.library.IsFiltered() {
			return event
		}

		h.bar.Show("search")
		h.app.SetFocus(h.bar.SearchContainer())
		return nil
	case 'q':
		h.app.Stop()
	}

	return event
}

type HelpHandler struct {
	pages pagesManager
}

func NewHelpHandler(pg pagesManager) *HelpHandler {
	return &HelpHandler{
		pages: pg,
	}
}

func (k *HelpHandler) Listen(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		k.pages.HidePage("help")
		return nil
	}

	switch event.Rune() {
	case 'h':
		p, _ := k.pages.GetFrontPage()
		if p == "help" {
			k.pages.HidePage("help")
			return nil
		}
	}

	return event
}
