package bar

import (
	"github.com/mkozjak/blutui/internal/app"
	"github.com/mkozjak/blutui/internal/library"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/tview"
)

type switcher interface {
	Show(name string)
}

type Bar struct {
	app      app.Command
	library  library.Command
	status   *tview.Grid
	search   *tview.InputField
	currCont string
}

func New(a app.Command, l library.Command, ch <-chan player.Status) *Bar {
	bar := &Bar{
		app:     a,
		library: l,
	}

	stb := newStatusBar(a, l)
	stbc := stb.createContainer()
	go stb.listen(ch)

	srb := newSearchBar(bar, l)
	srbc := srb.createContainer()

	bar.status = stbc
	bar.search = srbc
	bar.currCont = "status"

	return bar
}

func (b *Bar) CurrentContainer() string {
	return b.currCont
}

func (b *Bar) StatusContainer() tview.Primitive {
	return b.status
}

func (b *Bar) SearchContainer() tview.Primitive {
	return b.search
}

func (b *Bar) Show(name string) {
	switch name {
	case "search":
		b.app.ShowComponent(b.search)
		b.currCont = "search"
	case "status":
		b.app.ShowComponent(b.status)
		p := b.app.PrevFocused()
		b.app.SetPrevFocused("search")
		b.currCont = "status"
		b.app.SetFocus(p)
	}
}
