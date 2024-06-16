package bar

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/tview"
)

type SearchBar struct {
	switcher switcher
}

func newSearchBar(s switcher) *SearchBar {
	return &SearchBar{switcher: s}
}

func (s *SearchBar) createContainer() *tview.InputField {
	i := tview.NewInputField().
		SetLabel("search: ").
		SetLabelColor(tcell.ColorDefault).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetAcceptanceFunc(tview.InputFieldMaxLength(50))

	i.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			internal.Log(i.GetText())
		case tcell.KeyEscape:
			i.SetText("")
			s.switcher.Show("status")
		}
	})

	i.SetBackgroundColor(tcell.ColorDefault).SetTitleColor(tcell.ColorDefault).SetBorderPadding(0, 0, 1, 1)

	return i
}
