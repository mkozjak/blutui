package library

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	internal "github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/tview"
)

// left pane - artists
func (l *Library) drawArtistPane() *tview.List {
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
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'j':
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			}

			return event
		})

	for _, artist := range l.Artists {
		p.AddItem(artist, "", 0, nil)
	}

	p.SetChangedFunc(l.scrollCb)

	return p
}

func (l *Library) cpHighlightArtist(name string) {
	// clear previously highlighted items
	if l.CpArtistIdx >= 0 {
		n, _ := l.ArtistPane.GetItemText(l.CpArtistIdx)
		l.ArtistPane.SetItemText(l.CpArtistIdx, strings.TrimPrefix(n, "[yellow]"), "")
	}

	if name == "" {
		l.CpArtistIdx = -1
		return
	}

	// highlight artist
	// track is highlighted through a.newAlbumList
	idx := l.ArtistPane.FindItems(name, "", false, true)
	if len(idx) < 1 {
		return
	}

	n, _ := l.ArtistPane.GetItemText(idx[0])
	l.ArtistPane.SetItemText(idx[0], "[yellow]"+n, "")
	l.CpArtistIdx = idx[0]
}
