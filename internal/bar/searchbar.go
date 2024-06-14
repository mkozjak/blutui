package bar

import "github.com/mkozjak/tview"

type SearchBar struct{}

func newSearchBar() *SearchBar {
	return &SearchBar{}
}

func (s *SearchBar) createContainer() *tview.InputField {
	return tview.NewInputField().SetLabel("search:")
}
