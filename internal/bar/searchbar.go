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
	switcher switcher
	library  library.Command
}

func newSearchBar(s switcher, l library.Command) *SearchBar {
	return &SearchBar{switcher: s, library: l}
}

func (s *SearchBar) createContainer() *tview.InputField {
	in := tview.NewInputField().
		SetLabel("search: ").
		SetLabelColor(tcell.ColorDefault).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorDefault).
		SetAcceptanceFunc(tview.InputFieldMaxLength(40))

	in.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			a := s.library.Artists()
			query := in.GetText()
			var m []int

			for i := 0; i < len(a); i++ {
				var scores []float64

				for _, token := range strings.Split(a[i], " ") {
					scores = append(scores, internal.JWSimilarity(query, token))
				}

				score := slices.Max(scores)
				if score == 1 {
					// Exact match
					internal.Log("EXACT MATCH!")
					m = []int{i}
					break
				} else if score > 0.75 {
					m = append(m, i)
				}
			}

			internal.Log("results:", m)
		case tcell.KeyEscape:
			in.SetText("")
			s.switcher.Show("status")
		}
	})

	in.SetBackgroundColor(tcell.ColorDefault).
		SetTitleColor(tcell.ColorDefault).
		SetBorderPadding(0, 0, 1, 1)

	return in
}
