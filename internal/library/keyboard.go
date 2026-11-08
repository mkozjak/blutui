package library

import (
	"github.com/gdamore/tcell/v2"
)

func (l *Library) KeyboardHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		if l.artistPane.HasFocus() {
			if len(l.currentArtistAlbums) == 0 {
				return nil
			}

			// Set first artist's album as selectable and make it focused
			l.currentArtistAlbums[0].SetSelectable(true, false)
			l.app.SetFocus(l.currentArtistAlbums[0])
		} else {
			// Reset offset/scroll and throw focus to artist pane
			l.albumPane.SetOffset(0, 0)
			l.app.SetFocus(l.artistPane)
		}

		return nil
	case tcell.KeyCtrlB, tcell.KeyPgUp:
		if l.albumPane.HasFocus() {
			l.pageAlbums(-1, false)
			return nil
		}

		return tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone)
	case tcell.KeyCtrlF, tcell.KeyPgDn:
		if l.albumPane.HasFocus() {
			l.pageAlbums(1, false)
			return nil
		}

		return tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone)
	case tcell.KeyCtrlD:
		if l.albumPane.HasFocus() {
			l.pageAlbums(1, true)
			return nil
		}

		if l.artistPane.HasFocus() == true {
			l.artistPane.SetCurrentItem(l.artistPane.GetCurrentItem() + 20)
			return nil
		}
	case tcell.KeyCtrlU:
		if l.albumPane.HasFocus() {
			l.pageAlbums(-1, true)
			return nil
		}

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
	case 'h':
		if l.albumPane.HasFocus() {
			l.albumPane.SetOffset(0, 0)
			l.app.SetFocus(l.artistPane)
		}

		return nil
	case 'l':
		if l.artistPane.HasFocus() && len(l.currentArtistAlbums) > 0 {
			l.currentArtistAlbums[0].SetSelectable(true, false)
			l.app.SetFocus(l.currentArtistAlbums[0])
		}

		return nil
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

func (l *Library) pageAlbums(direction int, half bool) {
	index := l.selectedAlbumIdx()

	if index < 0 {
		return
	}

	_, _, _, height := l.albumPane.GetInnerRect()
	step := max(1, height-2)

	if half {
		step = max(1, step/2)
	}

	row, _ := l.currentArtistAlbums[index].GetSelection()
	row += direction * step

	for row >= l.currentArtistAlbums[index].GetRowCount() && index+1 < len(l.currentArtistAlbums) {
		row -= l.currentArtistAlbums[index].GetRowCount()
		index++
	}

	for row < 0 && index > 0 {
		index--
		row += l.currentArtistAlbums[index].GetRowCount()
	}

	album := l.currentArtistAlbums[index]
	row = max(0, min(row, album.GetRowCount()-1))
	album.Select(row, 0)
	l.app.SetFocus(album)
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
