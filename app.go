package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/tview"
)

var api string = "http://bluesound.local:11000"

var arListStyle = &tview.BoxBorders{
	// \u0020 - whitespace
	HorizontalFocus:  rune('\u2500'),
	Horizontal:       rune('\u2500'),
	VerticalFocus:    rune('\u2502'),
	Vertical:         rune('\u2502'),
	TopRightFocus:    rune('\u2510'),
	TopRight:         rune('\u2510'),
	TopLeftFocus:     rune('\u250C'),
	TopLeft:          rune('\u250C'),
	BottomRightFocus: rune('\u2518'),
	BottomRight:      rune('\u2518'),
	BottomLeftFocus:  rune('\u2514'),
	BottomLeft:       rune('\u2514'),
}

var alFlexStyle = arListStyle
var trListStyle = &tview.BoxBorders{}

func (a *app) drawCurrentArtist(artist string, c *tview.Flex) {
	trackLstStyle := tcell.Style{}
	trackLstStyle.Background(tcell.ColorDefault)

	for _, album := range a.albumArtists[artist].albums {
		trackLst := tview.NewList().
			SetHighlightFullLine(true).
			SetWrapAround(false).
			SetSelectedFocusOnly(true).
			SetSelectedTextColor(tcell.ColorWhite).
			SetSelectedBackgroundColor(tcell.ColorCornflowerBlue).
			ShowSecondaryText(false).
			SetMainTextStyle(trackLstStyle)

		trackLst.SetSelectedFunc(func(i int, name, _ string, sh rune) {
			_, autoplay, err := a.getTrackURL(name, artist, album.name)
			if err != nil {
				panic(err)
			}

			// play track and add subsequent album tracks to queue
			go play(autoplay)
		})

		// set album tracklist keymap
		trackLst.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'j':
				if trackLst.GetCurrentItem()+1 == trackLst.GetItemCount() {
					if a.currentAlbumIndex+1 == a.currentAlbumCount {
						// do nothing, return default
						return nil
					} else {
						a.application.SetFocus(c.GetItem(a.currentAlbumIndex + 1))
						a.currentAlbumIndex = a.currentAlbumIndex + 1
						return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
					}
				}

				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				// FIXME
				if trackLst.GetCurrentItem() == 0 {
					if a.currentAlbumIndex == 0 {
						// do nothing, i'm already on 1st album
						return nil
					} else {
						a.application.SetFocus(c.GetItem(a.currentAlbumIndex - 1))
						a.currentAlbumIndex = a.currentAlbumIndex - 1
						return nil
					}
				}

				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			}

			return event
		})

		trackLst.SetTitle("[::b]" + album.name).
			SetBorder(true).
			SetBorderColor(tcell.ColorCornflowerBlue).
			SetBackgroundColor(tcell.ColorDefault).
			SetTitleAlign(tview.AlignLeft).
			SetCustomBorders(trListStyle)

		for _, t := range album.tracks {
			trackLst.AddItem(t.name, "", 0, nil)
		}

		// FIXME: how do I scroll this?
		// c.AddItem(trackLst, trackLst.GetItemCount()+2, 1, true)
		c.AddItem(trackLst, 0, 1, true)
	}
}

func main() {
	a := app{
		application:       tview.NewApplication(),
		albumArtists:      map[string]artist{},
		currentAlbumIndex: 0,
		currentAlbumCount: 0,
	}

	err := a.fetchData()
	if err != nil {
		panic(err)
	}

	// left pane - artists
	arLstStyle := tcell.Style{}
	arLstStyle.Background(tcell.ColorDefault)

	arLst := tview.NewList().
		SetHighlightFullLine(true).
		SetWrapAround(false).
		SetSelectedTextColor(tcell.ColorWhite).
		SetSelectedBackgroundColor(tcell.ColorCornflowerBlue).
		ShowSecondaryText(false).
		SetMainTextStyle(arLstStyle)

	arLst.SetTitle(" [::b]Artist ").
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft).
		SetCustomBorders(arListStyle).
		// set artists list keymap
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'j':
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			}

			return event
		})

	alFlex := tview.NewFlex().
		SetDirection(tview.FlexRow)

	alFlex.SetTitle(" [::b]Track ").
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft).
		SetCustomBorders(alFlexStyle)

	appFlex := tview.NewFlex().
		AddItem(arLst, 0, 1, true).
		AddItem(alFlex, 0, 2, false)

	for _, artist := range a.artists {
		arLst.AddItem(artist, "", 0, nil)
	}

	// draw selected artist's right pane (album items) on artist scroll
	arLst.SetChangedFunc(func(index int, artist string, _ string, shortcut rune) {
		alFlex.Clear()
		a.currentAlbumCount = len(a.albumArtists[artist].albums)
		a.drawCurrentArtist(artist, alFlex)
	})

	// draw initial album list for the first artist in the list
	a.application.SetAfterDrawFunc(func(screen tcell.Screen) {
		a.drawCurrentArtist(a.artists[0], alFlex)

		// disable callback
		a.application.SetAfterDrawFunc(nil)
		return
	})

	// set global keymap
	a.application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlQ:
			a.application.Stop()

		case tcell.KeyTab:
			artistView := appFlex.GetItem(0)
			albumView := appFlex.GetItem(1)

			if !albumView.HasFocus() {
				a.application.SetFocus(alFlex)
				arLst.SetSelectedBackgroundColor(tcell.ColorLightGray)
			} else {
				a.application.SetFocus(artistView)
				arLst.SetSelectedBackgroundColor(tcell.ColorCornflowerBlue)
			}

			a.currentAlbumIndex = 0
			return nil
		}

		switch event.Rune() {
		case 'v':
			go stop()
		case 'b':
			go next()
		case 'z':
			go previous()
		case 'q':
			a.application.Stop()
		}

		return event
	})

	if err := a.application.SetRoot(appFlex, true).SetFocus(appFlex).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
