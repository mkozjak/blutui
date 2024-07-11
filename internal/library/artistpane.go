package library

import (
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
	internal "github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/tview"
)

// left pane - artists
func (l *Library) createArtistContainer() *tview.List {
	artistPaneStyle := tcell.Style{}
	artistPaneStyle.Background(tcell.ColorDefault)

	p := tview.NewList().
		SetHighlightFullLine(true).
		SetWrapAround(false).
		SetSelectedTextColor(tcell.ColorWhite).
		SetSelectedBackgroundColor(tcell.ColorCornflowerBlue).
		ShowSecondaryText(false).
		SetMainTextStyle(artistPaneStyle)

	p.SetTitle(" [::b]Artist ").
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft).
		SetCustomBorders(internal.CustomBorders).
		// set artists list keymap
		SetInputCapture(l.artistPaneKeyboardHandler).
		SetFocusFunc(func() {
			p.SetSelectedBackgroundColor(tcell.ColorCornflowerBlue)
		}).
		SetBlurFunc(func() {
			l.app.SetPrevFocused("artistpane")
			p.SetSelectedBackgroundColor(tcell.ColorLightGray)
		})

	return p
}

func (l *Library) FilterArtistPane(f []string) {
	for _, a := range l.artists {
		if !slices.Contains(f, a) {
			r := l.artistPane.FindItems(a, "", false, true)
			l.artistPane.RemoveItem(r[0])
		}
	}

	if len(f) > 0 {
		l.artistPaneFiltered = true
	}
}

func (l *Library) DrawArtistPane() {
	// Delete existing records, possibly after clearing the search results
	l.artistPane.Clear()

	for _, artist := range l.artists {
		l.artistPane.AddItem(artist, "", 0, nil)
	}

	l.artistPane.SetChangedFunc(l.scrollCb)
}

func (l *Library) SelectCpArtist() {
	if l.cpArtistIdx < 0 {
		return
	}

	l.artistPane.SetCurrentItem(l.cpArtistIdx)
}

func (l *Library) MarkCpArtist(name string) {
	// clear previously highlighted items
	if l.cpArtistIdx >= 0 {
		n, _ := l.artistPane.GetItemText(l.cpArtistIdx)
		l.artistPane.SetItemText(l.cpArtistIdx, strings.TrimPrefix(n, "[yellow]"), "")
	}

	if name == "" {
		l.cpArtistIdx = -1
		return
	}

	// highlight artist
	// track is highlighted through l.drawAlbum
	idx := l.artistPane.FindItems(name, "", false, true)
	if len(idx) < 1 {
		return
	}

	n, _ := l.artistPane.GetItemText(idx[0])
	l.artistPane.SetItemText(idx[0], "[yellow]"+n, "")
	l.cpArtistIdx = idx[0]
}
