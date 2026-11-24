package library

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type albumViewport struct {
	*tview.Box
	library      *Library
	rowOffset    int
	manualScroll bool
}

func (p *albumViewport) Draw(screen tcell.Screen) {
	p.Box.DrawForSubclass(screen, p)
	x, y, width, height := p.GetInnerRect()
	totalRows := 0

	for _, album := range p.library.currentArtistAlbums {
		if album.HasFocus() && !p.manualScroll && height > 0 {
			row, _ := album.GetSelection()
			selectedLine := totalRows + 1 + max(0, row)
			maxOffset := selectedLine

			if row <= 0 && height > 1 {
				maxOffset = totalRows
			}

			// Move just enough to show the selection and the header above the first track.
			p.rowOffset = min(maxOffset, max(p.rowOffset, selectedLine-height+1))
		}

		totalRows += album.GetRowCount() + 2
	}

	p.rowOffset = max(0, min(p.rowOffset, totalRows-height))
	albumTop := 0

	for _, album := range p.library.currentArtistAlbums {
		albumEnd := albumTop + album.GetRowCount() + 2
		visibleTop := max(albumTop, p.rowOffset)
		visibleEnd := min(albumEnd, p.rowOffset+height)

		if width > 0 && visibleTop < visibleEnd {
			album.clippedTop = visibleTop - albumTop
			album.SetRect(x, y+visibleTop-p.rowOffset, width, visibleEnd-visibleTop)
			album.Draw(screen)
		} else {
			album.SetRect(0, 0, 0, 0)
		}

		albumTop = albumEnd
	}
}

func (p *albumViewport) Focus(delegate func(tview.Primitive)) {
	if len(p.library.currentArtistAlbums) > 0 {
		delegate(p.library.currentArtistAlbums[0])
		return
	}

	p.Box.Focus(delegate)
}

func (p *albumViewport) HasFocus() bool {
	album := p.library.lastFocusedAlbum
	return p.Box.HasFocus() || album != nil && album.HasFocus()
}

func (p *albumViewport) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return p.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		album := p.library.lastFocusedAlbum

		if album != nil && album.HasFocus() {
			album.InputHandler()(event, setFocus)
		}
	})
}

func (p *albumViewport) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return p.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		if !p.InRect(event.Position()) {
			return false, nil
		}

		switch action {
		case tview.MouseScrollUp:
			p.rowOffset--
			p.manualScroll = true
			return true, nil
		case tview.MouseScrollDown:
			p.rowOffset++
			p.manualScroll = true
			return true, nil
		}

		for _, album := range p.library.currentArtistAlbums {
			if consumed, capture := album.MouseHandler()(action, event, setFocus); consumed {
				if action == tview.MouseLeftDown {
					row, _ := album.CellAt(event.Position())

					if row >= 0 {
						album.Select(row, 0)
					}
				}

				return consumed, capture
			}
		}

		return false, nil
	})
}
