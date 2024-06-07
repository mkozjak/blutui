package library

import "github.com/gdamore/tcell/v2"

func (l *Library) KeyboardHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		if !l.AlbumPane.HasFocus() {
			a.Application.SetFocus(l.AlbumPane)
			l.ArtistPane.SetSelectedBackgroundColor(tcell.ColorLightGray)
		} else {
			l.Application.SetFocus(l.ArtistPane)
			l.ArtistPane.SetSelectedBackgroundColor(tcell.ColorCornflowerBlue)
		}

		return nil
	case tcell.KeyCtrlB:
		return tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone)
	case tcell.KeyCtrlF:
		return tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone)
	case tcell.KeyCtrlD:
		if l.ArtistPane.HasFocus() == true {
			l.ArtistPane.SetCurrentItem(l.ArtistPane.GetCurrentItem() + 20)
			return nil
		}
	case tcell.KeyCtrlU:
		if l.ArtistPane.HasFocus() == true {
			i := l.ArtistPane.GetCurrentItem()
			if i < 20 {
				l.ArtistPane.SetCurrentItem(0)
			} else {
				l.ArtistPane.SetCurrentItem(i - 20)
			}

			return nil
		}
	}

	return event
}
