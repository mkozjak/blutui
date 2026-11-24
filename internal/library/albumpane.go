package library

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	internal "github.com/mkozjak/blutui/internal"
	"github.com/rivo/tview"
)

type albumTable struct {
	*tview.Table
	clippedTop int
}

// right pane - albums
func (l *Library) createAlbumContainer() *albumViewport {
	p := &albumViewport{Box: tview.NewBox(), library: l}

	p.SetTitle(" [::b]Track ").
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft)

	return p
}

func (l *Library) drawAlbum(artist string, album album) *albumTable {
	durms := internal.FormatDuration(album.duration)

	// Create a new album and set it as not selectable by default
	// so that it's first track doesn't get highlighted.
	c := &albumTable{Table: tview.NewTable().SetSelectable(false, false)}

	c.SetTitle(fmt.Sprintf("[::b]%s (%d)", internal.EscapeStyleTag(album.name), album.year)).
		SetBackgroundColor(tcell.ColorDefault).
		SetFocusFunc(func() {
			// Set this current table as selectable so its selected rows get highlighted
			c.SetSelectable(true, false)
			l.lastFocusedAlbum = c
			l.albumPane.manualScroll = false
		}).
		SetBlurFunc(func() {
			// Set this current table as not selectable so it loses the highlighting
			c.SetSelectable(false, false)
			l.app.SetPrevFocused("albumpane")
		})

	c.SetSelectionChangedFunc(func(_, _ int) {
		l.albumPane.manualScroll = false
	})
	c.SetCell(0, 0, tview.NewTableCell(album.name).SetTransparency(true))

	// Create a custom list line for album length etc.
	c.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		width = max(0, width-2)
		if c.clippedTop == 0 && height > 0 {
			c.drawHeader(screen, x+1, y, width, durms)
			y++
			height--
		}

		trackOffset := max(0, c.clippedTop-1)
		c.SetOffset(trackOffset, 0)
		// Space for other content
		return x + 1, y, width, min(height, c.GetRowCount()-trackOffset)
	})

	// Set album tracklist keymap
	c.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j':
			currRow, _ := c.GetSelection()

			// Reached the end of current album, so skip to next one if available.
			if currRow+1 == c.GetRowCount() {
				albumIndex := l.selectedAlbumIdx()

				if albumIndex >= 0 && albumIndex+1 < len(l.currentArtistAlbums) {
					next := l.currentArtistAlbums[albumIndex+1]
					next.Select(0, 0)
					l.app.SetFocus(next)
					return nil
				}
			}

			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)

		case 'k':
			currRow, _ := c.GetSelection()

			if currRow == 0 {
				albumIndex := l.selectedAlbumIdx()

				if albumIndex > 0 {
					previous := l.currentArtistAlbums[albumIndex-1]
					previous.Select(previous.GetRowCount()-1, 0)
					l.app.SetFocus(previous)
					return nil
				}
			}

			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)

		case 'x':
			currRow, _ := c.GetSelection()
			trackName := c.GetCell(currRow, 0).Text

			u, _, err := l.trackURL(trackName, artist, album.name)
			if err != nil {
				panic(err)
			}

			// play currently selected track only
			go l.player.Play(u)
			return nil
		}

		return event
	})

	c.SetSelectedFunc(func(row, col int) {
		_, autoplay, err := l.trackURL(album.tracks[row].name, artist, album.name)
		if err != nil {
			panic(err)
		}

		// play track and add subsequent album tracks to queue
		go l.player.Play(autoplay)
	})

	// print album tracks
	for i, t := range album.tracks {
		track := tview.NewTableCell("").
			SetTextColor(tcell.ColorDefault).
			SetSelectedStyle(tcell.Style{}.
				Background(tcell.ColorCornflowerBlue).
				Foreground(tcell.ColorWhite)).
			SetAlign(tview.AlignLeft).
			SetExpansion(1).
			SetTransparency(true).
			SetSelectable(true)

		if l.CpTrackName != "" &&
			l.CpTrackName == internal.CleanTrackName(t.name) &&
			l.CpAlbumName == album.name {
			track.SetText("[yellow]" + t.name)
		} else {
			track.SetText(t.name)
		}

		dur := tview.NewTableCell(internal.FormatDuration(t.duration)).
			SetTextColor(tcell.ColorDefault).
			SetSelectedStyle(tcell.Style{}.
				Background(tcell.ColorCornflowerBlue).
				Foreground(tcell.ColorWhite)).
			SetAlign(tview.AlignRight).
			SetExpansion(1).
			SetTransparency(true).
			SetSelectable(true)

		c.SetCell(i, 0, track)
		c.SetCell(i, 1, dur)
	}

	return c
}

func (c *albumTable) drawHeader(screen tcell.Screen, x, y, width int, duration string) {
	titleWidth := width
	if width > len(duration)+1 {
		titleWidth -= len(duration) + 1
		// Write album length along the horizontal line
		tview.Print(screen, "[::b]"+duration, x, y, width, tview.AlignRight, tcell.ColorWhite)
	}

	title := c.GetTitle()
	printed, cells := tview.Print(screen, title, x, y, titleWidth, tview.AlignLeft, tview.Styles.TitleColor)
	if printed < len(title) && titleWidth > 0 {
		tview.Print(screen, string(tview.SemigraphicsHorizontalEllipsis), x+titleWidth-1, y, 1, tview.AlignLeft, tview.Styles.TitleColor)
		return
	}

	style := tcell.StyleDefault.Foreground(tcell.ColorCornflowerBlue)
	for col := x + cells + 1; col < x+titleWidth; col++ {
		screen.SetContent(col, y, tview.BoxDrawingsLightHorizontal, nil, style)
	}
}

func (l *Library) DrawInitAlbums() {
	l.drawArtistAlbums(l.artists[0])
}

func (l *Library) MarkCpTrack(track, artist, album string) {
	if l.artistPane.GetItemCount() == 0 {
		return
	}

	shownArtist, _ := l.artistPane.GetItemText(l.artistPane.GetCurrentItem())
	shownArtist = strings.TrimPrefix(shownArtist, "[yellow]")

	for _, al := range l.currentArtistAlbums {
		alName := internal.CleanAlbumName(al.GetTitle())

		for i := 0; i < al.GetRowCount(); i++ {
			c := al.GetCell(i, 0)
			origTrack := strings.TrimPrefix(c.Text, "[yellow]")

			if track != "" && shownArtist == artist && alName == album && internal.CleanTrackName(origTrack) == track {
				c.SetText("[yellow]" + origTrack)
			} else {
				c.SetText(origTrack)
			}
		}
	}
}

func (l *Library) clearAlbumPane() {
	l.albumPane.rowOffset = 0
	l.albumPane.manualScroll = false
	l.currentArtistAlbums = nil
	l.lastFocusedAlbum = nil
}

func (l *Library) drawArtistAlbums(artist string) {
	l.clearAlbumPane()

	// remove style from the string
	cArtist := strings.TrimPrefix(artist, "[yellow]")

	for _, album := range l.albumArtists[cArtist].albums {
		l.currentArtistAlbums = append(l.currentArtistAlbums, l.drawAlbum(cArtist, album))
	}
}

// selectedAlbumIdx returns an index of the currently focused album
// based on the fact whether rows of an album are selectable.
// In this application, only the currently focused album is marked as selectable.
// If none of the albums are selectable, the method returns -1.
func (l *Library) selectedAlbumIdx() int {
	for i, t := range l.currentArtistAlbums {
		r, _ := t.GetSelectable()
		if r == true {
			return i
		}
	}

	return -1
}

// draw selected artist's right pane (album items) on artist scroll
func (l *Library) scrollCb(index int, artist string, _ string, shortcut rune) {
	l.drawArtistAlbums(artist)
}
