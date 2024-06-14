package bar

import (
	"github.com/mkozjak/blutui/internal/app"
	"github.com/mkozjak/blutui/internal/library"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/tview"
)

type Bar struct {
	app     app.Command
	library library.Command
	status  *tview.Table
	search  *tview.InputField
}

func New(a app.Command, l library.Command, ch <-chan player.Status) *Bar {
	stb := newStatusBar(a, l)
	stbc := stb.createContainer()
	go stb.listen(ch)

	srb := newSearchBar()
	srbc := srb.createContainer()

	return &Bar{
		app:     a,
		library: l,
		status: stbc,
		search: srbc,
	}
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
	case "status":
		b.app.ShowComponent(b.status)
	}
}
