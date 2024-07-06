package library

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	internal "github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/tview"
)

// right pane - albums
func (l *Library) createAlbumContainer() *tview.Grid {
	p := tview.NewGrid().
		SetColumns(0)

	p.SetTitle(" [::b]Track ").
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft).
		SetCustomBorders(internal.CustomBorders)

	return p
}

func (l *Library) drawAlbum(artist string, album album, g *tview.Grid) *tview.Table {
	durms := internal.FormatDuration(album.duration)

	// Create a new album and set it as not selectable by default
	// so that it's first track doesn't get highlighted.
	c := tview.NewTable().
		SetSelectable(false, false)

	c.SetTitle(fmt.Sprintf("[::b]%s (%d)", album.name, album.year)).
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft).
		SetCustomBorders(internal.NoBorders).
		SetFocusFunc(func() {
			// Set this current table as selectable so its selected rows get highlighted
			c.SetSelectable(true, false)
		}).
		SetBlurFunc(func() {
			// Set this current table as not selectable so it loses the highlighting
			c.SetSelectable(false, false)
			l.app.SetPrevFocused("albumpane")
		})

	c.SetCell(0, 0, tview.NewTableCell(album.name).SetTransparency(true))

	// Create a custom list line for album length etc.
	c.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		centerY := y + height/c.GetRowCount()/2

		for cx := x + len(c.GetTitle()) - 3; cx < x+width-len(durms)-2; cx++ {
			screen.SetContent(cx, centerY, tview.BoxDrawingsLightHorizontal, nil,
				tcell.StyleDefault.Foreground(tcell.ColorCornflowerBlue))
		}

		// Write album length along the horizontal line
		tview.Print(screen, "[::b]"+durms, x+1, centerY, width-2, tview.AlignRight, tcell.ColorWhite)

		// Space for other content
		return x + 1, centerY + 1, width - 2, height - (centerY + 1 - y)
	})

	// Set album tracklist keymap
	c.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j':
			currRow, _ := c.GetSelection()

			// Reached the end of current album, so skip to next one if available.
			if currRow+1 == c.GetRowCount() {
				albumIndex := l.selectedAlbumIdx()

				if albumIndex+1 != len(l.albumArtists[artist].albums) {
					l.app.SetFocus(l.currentArtistAlbums[albumIndex+1])
				}
			}

			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)

		case 'k':
			currRow, _ := c.GetSelection()

			if currRow == 0 {
				albumIndex := l.selectedAlbumIdx()

				if albumIndex != 0 {
					l.app.SetFocus(l.currentArtistAlbums[albumIndex-1])
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

		if l.CpTrackName != "" && l.CpTrackName == internal.CleanTrackName(t.name) {
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

func (l *Library) DrawInitAlbums() {
	r := l.drawArtistAlbums(l.artists[0], l.albumPane)
	l.albumPane.SetRows(r...)
}

func (l *Library) drawArtistAlbums(artist string, c *tview.Grid) []int {
	alHeights := []int{}
	l.currentArtistAlbums = nil

	// remove style from the string
	cArtist := strings.TrimPrefix(artist, "[yellow]")

	for i, album := range l.albumArtists[cArtist].albums {
		albumTable := l.drawAlbum(cArtist, album, c)
		alHeights = append(alHeights, len(album.tracks)+2)

		// automatically focus the first track from the first album
		// since grid is the parent, it will automatically lose focus
		// and give it to the first album
		if i == 0 {
			c.AddItem(albumTable, i, 0, 1, 1, 0, 0, true)
		} else {
			c.AddItem(albumTable, i, 0, 1, 1, 0, 0, false)
		}

		l.currentArtistAlbums = append(l.currentArtistAlbums, albumTable)
	}

	return alHeights
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
	l.albumPane.Clear()
	alHeights := l.drawArtistAlbums(artist, l.albumPane)
	l.albumPane.SetRows(alHeights...)
}
