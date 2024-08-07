package library

import (
	"github.com/gdamore/tcell/v2"
)

func (l *Library) KeyboardHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		if l.artistPane.HasFocus() {
			// Set first artist's album as selectable and make it focused
			l.currentArtistAlbums[0].SetSelectable(true, false)
			l.app.SetFocus(l.currentArtistAlbums[0])
		} else {
			// Reset offset/scroll and throw focus to artist pane
			l.albumPane.SetOffset(0, 0)
			l.app.SetFocus(l.artistPane)
		}

		return nil
	case tcell.KeyCtrlB:
		return tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone)
	case tcell.KeyCtrlF:
		return tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone)
	case tcell.KeyCtrlD:
		if l.artistPane.HasFocus() == true {
			l.artistPane.SetCurrentItem(l.artistPane.GetCurrentItem() + 20)
			return nil
		}
	case tcell.KeyCtrlU:
		if l.artistPane.HasFocus() == true {
			i := l.artistPane.GetCurrentItem()
			if i < 20 {
				l.artistPane.SetCurrentItem(0)
			} else {
				l.artistPane.SetCurrentItem(i - 20)
			}

			return nil
		}
	}

	switch event.Rune() {
	case 'g':
		if l.artistPane.HasFocus() {
			l.artistPane.SetCurrentItem(0)
		}

		return nil
	case 'G':
		if l.artistPane.HasFocus() {
			l.artistPane.SetCurrentItem(-1)
		}

		return nil
	}

	return event
}

func (l *Library) artistPaneKeyboardHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		if l.artistPaneFiltered {
			l.DrawArtistPane()
			l.artistPaneFiltered = false
			return nil
		}
	}

	switch event.Rune() {
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	}

	return event
}
