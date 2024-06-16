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
		SetInputCapture(l.artistPaneKeyboardHandler).
		SetFocusFunc(func() {
			p.SetSelectedBackgroundColor(tcell.ColorCornflowerBlue)
		}).
		SetBlurFunc(func() {
			l.app.SetPrevFocused("artistpane")
			p.SetSelectedBackgroundColor(tcell.ColorLightGray)
		})

	for _, artist := range l.artists {
		p.AddItem(artist, "", 0, nil)
	}

	p.SetChangedFunc(l.scrollCb)

	return p
}

func (l *Library) SelectCpArtist() {
	l.artistPane.SetCurrentItem(l.cpArtistIdx)
}

func (l *Library) HighlightCpArtist(name string) {
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
	// track is highlighted through a.newAlbumList
	idx := l.artistPane.FindItems(name, "", false, true)
	if len(idx) < 1 {
		return
	}

	n, _ := l.artistPane.GetItemText(idx[0])
	l.artistPane.SetItemText(idx[0], "[yellow]"+n, "")
	l.cpArtistIdx = idx[0]
}
