package library

import "github.com/gdamore/tcell/v2"

func (l *Library) KeyboardHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		if !l.albumPane.HasFocus() {
			l.app.SetFocus(l.albumPane)
			l.artistPane.SetSelectedBackgroundColor(tcell.ColorLightGray)
		} else {
			l.app.SetFocus(l.artistPane)
			l.artistPane.SetSelectedBackgroundColor(tcell.ColorCornflowerBlue)
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

	return event
}
