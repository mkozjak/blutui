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
}

func NewBar(a app.Command, l library.Command) *Bar {
	return &Bar{
		app:     a,
		library: l,
	}
}

func (b *Bar) CreateStatusBar(ch <-chan player.Status) (*tview.Table, error) {
	sb := newStatusBar(b.app, b.library)
	sbc, err := sb.createContainer()
	if err != nil {
		return nil, err
	}

	go sb.listen(ch)
	return sbc, nil
}

func (b *Bar) Show(name string) {
}
