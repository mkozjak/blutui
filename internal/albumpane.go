package internal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/tview"
)

// right pane - albums
func (a *App) CreateAlbumPane() {
	a.AlbumPane = tview.NewGrid().
		SetColumns(0)

	a.AlbumPane.SetTitle(" [::b]Track ").
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft).
		SetCustomBorders(CustomBorders)
}

func (a *App) newAlbumList(artist string, album album, c *tview.Grid) *tview.List {
	textStyle := tcell.Style{}
	textStyle.Background(tcell.ColorDefault)
	d := FormatDuration(album.duration)

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
		_, autoplay, err := a.getTrackURL(trackName, artist, album.name)
		if err != nil {
			panic(err)
		}

		// play track and add subsequent album tracks to queue
		go a.Play(autoplay)
	})

	// set album tracklist keymap
	trackLst.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j':
			if trackLst.GetCurrentItem()+1 == trackLst.GetItemCount() {
				// reached the end of current album
				// skip to next one if available
				albumIndex, _ := c.GetOffset()

				if albumIndex+1 != len(a.AlbumArtists[artist].albums) {
					// this will redraw the screen
					// TODO: only use SetOffset if the next album cannot fit into the current screen in its entirety
					c.SetOffset(albumIndex+1, 0)
					a.Application.SetFocus(a.currentArtistAlbums[albumIndex+1])
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
					a.Application.SetFocus(a.currentArtistAlbums[albumIndex-1])
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
		SetCustomBorders(noBorders)

	for _, t := range album.tracks {
		if a.cpTrackName != "" && a.cpTrackName == cleanTrackName(t.name) {
			trackLst.AddItem("[yellow]"+t.name, "", 0, nil)
		} else {
			trackLst.AddItem(t.name, "", 0, nil)
		}
	}

	return trackLst
}
