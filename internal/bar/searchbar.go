package bar

import (
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/blutui/internal/library"
	"github.com/mkozjak/tview"
)

type SearchBar struct {
	switcher  switcher
	library   library.Command
	container *tview.InputField
}

func newSearchBar(s switcher, l library.Command) *SearchBar {
	return &SearchBar{switcher: s, library: l}
}

func (s *SearchBar) createContainer() *tview.InputField {
	s.container = tview.NewInputField().
		SetLabel("search: ").
		SetLabelColor(tcell.ColorDefault).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorDefault).
		SetAcceptanceFunc(tview.InputFieldMaxLength(40)).
		SetDoneFunc(s.done)

	s.container.SetBackgroundColor(tcell.ColorDefault).
		SetTitleColor(tcell.ColorDefault).
		SetBorderPadding(0, 0, 1, 1)

	return s.container
}

func (s *SearchBar) done(key tcell.Key) {
	switch key {
	case tcell.KeyEnter:
		a := s.library.Artists()
		query := s.container.GetText()
		var m []string

		for i := 0; i < len(a); i++ {
			var scores []float64

			for _, token := range strings.Split(a[i], " ") {
				scores = append(scores, internal.JWSimilarity(query, token))
			}

			score := slices.Max(scores)
			if score == 1 {
				// Exact match
				m = []string{a[i]}
				break
			} else if score > 0.75 {
				m = append(m, a[i])
			}
		}

		s.library.FilterArtistPane(m)
		s.container.SetText("")
		s.switcher.Show("status")
	case tcell.KeyEscape:
		s.container.SetText("")
		s.switcher.Show("status")
	}
}
