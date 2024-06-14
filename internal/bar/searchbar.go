package bar

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/tview"
)

type SearchBar struct{}

func newSearchBar() *SearchBar {
	return &SearchBar{}
}

func (s *SearchBar) createContainer() *tview.InputField {
	i := tview.NewInputField().
		SetLabel("search: ").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetLabelColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetAcceptanceFunc(tview.InputFieldMaxLength(50))

	i.SetBackgroundColor(tcell.ColorDefault)

	return i
}
