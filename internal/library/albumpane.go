package library

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	internal "github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/tview"
)

// right pane - albums
func (l *Library) drawAlbumPane() *tview.Grid {
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

func (l *Library) newAlbumList(artist string, album album, c *tview.Grid) *tview.List {
	textStyle := tcell.Style{}
	textStyle.Background(tcell.ColorDefault)
	d := internal.FormatDuration(album.duration)

	trackLst := tview.NewList().
		SetHighlightFullLine(true).
		SetWrapAround(false).
		SetSelectedFocusOnly(true).
		SetSelectedTextColor(tcell.ColorWhite).
		SetSelectedBackgroundColor(tcell.ColorCornflowerBlue).
		ShowSecondaryText(false).
		SetMainTextStyle(textStyle)

	// create a custom list line for album length etc.
	trackLst.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		centerY := y + height/trackLst.GetItemCount()/2

		for cx := x + len(trackLst.GetTitle()) - 3; cx < x+width-len(d)-2; cx++ {
			screen.SetContent(cx, centerY, tview.BoxDrawingsLightHorizontal, nil,
				tcell.StyleDefault.Foreground(tcell.ColorCornflowerBlue))
		}

		// write album length along the horizontal line
		tview.Print(screen, "[::b]"+d, x+1, centerY, width-2, tview.AlignRight, tcell.ColorWhite)

		// space for other content
		return x + 1, centerY + 1, width - 2, height - (centerY + 1 - y)
	})

	trackLst.SetSelectedFunc(func(i int, trackName, _ string, sh rune) {
		_, autoplay, err := l.getTrackURL(trackName, artist, album.name)
		if err != nil {
			panic(err)
		}

		// play track and add subsequent album tracks to queue
		l.player.Play(autoplay)
	})

	// set album tracklist keymap
	trackLst.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j':
			if trackLst.GetCurrentItem()+1 == trackLst.GetItemCount() {
				// reached the end of current album
				// skip to next one if available
				albumIndex, _ := c.GetOffset()

				if albumIndex+1 != len(l.albumArtists[artist].albums) {
					// this will redraw the screen
					// TODO: only use SetOffset if the next album cannot fit into the current screen in its entirety
					c.SetOffset(albumIndex+1, 0)
					l.app.SetFocus(l.currentArtistAlbums[albumIndex+1])
				}
			}

			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			if trackLst.GetCurrentItem() == 0 {
				// reached the beginning of current album
				// skip to previous one if available
				albumIndex, _ := c.GetOffset()

				if albumIndex != 0 {
					// this will redraw the screen
					// TODO: only use SetOffset if the next album cannot fit into the current screen in its entirety
					c.SetOffset(albumIndex-1, 0)
					l.app.SetFocus(l.currentArtistAlbums[albumIndex-1])
				}
			}

			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}

		return event
	})

	trackLst.
		SetTitle("[::b]" + album.name).
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft).
		SetCustomBorders(internal.NoBorders)

	for _, t := range album.tracks {
		if l.CpTrackName != "" && l.CpTrackName == internal.CleanTrackName(t.name) {
			trackLst.AddItem("[yellow]"+t.name, "", 0, nil)
		} else {
			trackLst.AddItem(t.name, "", 0, nil)
		}
	}

	return trackLst
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
		albumList := l.newAlbumList(cArtist, album, c)
		alHeights = append(alHeights, len(album.tracks)+2)

		// automatically focus the first track from the first album
		// since grid is the parent, it will automatically lose focus
		// and give it to the first album
		if i == 0 {
			c.AddItem(albumList, i, 0, 1, 1, 0, 0, true)
		} else {
			c.AddItem(albumList, i, 0, 1, 1, 0, 0, false)
		}

		l.currentArtistAlbums = append(l.currentArtistAlbums, albumList)
	}

	return alHeights
}

// draw selected artist's right pane (album items) on artist scroll
func (l *Library) scrollCb(index int, artist string, _ string, shortcut rune) {
	l.albumPane.Clear()
	alHeights := l.drawArtistAlbums(artist, l.albumPane)
	l.albumPane.SetRows(alHeights...)
}
