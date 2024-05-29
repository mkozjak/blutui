package main

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/tview"
)

func main() {
	a := internal.App{
		Application:  tview.NewApplication(),
		AlbumArtists: map[string]internal.Artist{},
	}

	err := a.FetchData()
	if err != nil {
		panic(err)
	}

	// left pane - artists
	ArtistPaneStyle := tcell.Style{}
	ArtistPaneStyle.Background(tcell.ColorDefault)

	ArtistPane := tview.NewList().
		SetHighlightFullLine(true).
		SetWrapAround(false).
		SetSelectedTextColor(tcell.ColorWhite).
		SetSelectedBackgroundColor(tcell.ColorCornflowerBlue).
		ShowSecondaryText(false).
		SetMainTextStyle(ArtistPaneStyle)

	ArtistPane.SetTitle(" [::b]Artist ").
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft).
		SetCustomBorders(internal.ArtistPaneStyle).
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

	// right pane - albums
	AlbumPane := tview.NewGrid().
		SetColumns(0)

	AlbumPane.SetTitle(" [::b]Track ").
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft).
		SetCustomBorders(internal.AlbumPaneStyle)

	// bottom bar - status
	statusBar := tview.NewTable().
		SetFixed(1, 3).
		SetSelectable(false, false).
		SetCell(0, 0, tview.NewTableCell("connecting").
			SetExpansion(1).
			SetTextColor(tcell.ColorDefault).
			SetAlign(tview.AlignLeft)).
		SetCell(0, 1, tview.NewTableCell("progress").
			SetExpansion(2).
			SetTextColor(tcell.ColorDefault).
			SetAlign(tview.AlignCenter)).
		SetCell(0, 2, tview.NewTableCell("something").
			SetExpansion(1).
			SetTextColor(tcell.ColorDefault).
			SetAlign(tview.AlignRight))

	statusBar.SetBackgroundColor(tcell.ColorDefault).SetBorder(false).SetBorderPadding(0, 0, 1, 1)

	// channel for receiving player status updates
	statusCh := make(chan internal.Status)

	// start long-polling for updates
	go a.PollStatus(statusCh)

	// start a goroutine for receiving the updates
	go func() {
		for s := range statusCh {
			var cpTitle string
			var cpFormat string
			var cpQuality string

			switch s.State {
			case "play":
				s.State = "playing"
				cpTitle = s.Artist + " - " + s.Track
				cpFormat = s.Format
				cpQuality = s.Quality
			case "stream":
				s.State = "streaming"
				cpTitle = s.Title2
				cpFormat = s.Format
				cpQuality = s.Quality
			case "stop":
				s.State = "stopped"
				cpTitle = ""
				cpFormat = ""
				cpQuality = ""
			case "pause":
				s.State = "paused"

				if s.Artist == "" && s.Track == "" {
					// streaming, set title to Title3 from /Status
					cpTitle = s.Title3
				} else {
					cpTitle = s.Artist + " - " + s.Track
				}

				cpFormat = s.Format
				cpQuality = s.Quality
			}

			statusBar.GetCell(0, 0).SetText("vol: " + strconv.Itoa(s.Volume) + " | " + s.State)
			statusBar.GetCell(0, 1).SetText(cpTitle)
			statusBar.GetCell(0, 2).SetText(cpQuality + " " + cpFormat)
			a.Application.Draw()
		}
	}()

	// app
	appFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		// left and right pane
		AddItem(tview.NewFlex().
			AddItem(ArtistPane, 0, 1, true).
			AddItem(AlbumPane, 0, 2, false), 0, 1, true).
		// status bar
		AddItem(statusBar, 1, 1, false)

	for _, artist := range a.Artists {
		ArtistPane.AddItem(artist, "", 0, nil)
	}

	// draw selected artist's right pane (album items) on artist scroll
	ArtistPane.SetChangedFunc(func(index int, artist string, _ string, shortcut rune) {
		AlbumPane.Clear()
		l := a.DrawCurrentArtist(artist, AlbumPane)
		AlbumPane.SetRows(l...)
	})

	// draw initial album list for the first artist in the list
	a.Application.SetAfterDrawFunc(func(screen tcell.Screen) {
		l := a.DrawCurrentArtist(a.Artists[0], AlbumPane)
		AlbumPane.SetRows(l...)

		// disable callback
		a.Application.SetAfterDrawFunc(nil)
	})

	// set global keymap
	a.Application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlQ:
			a.Application.Stop()
		case tcell.KeyTab:
			if !AlbumPane.HasFocus() {
				a.Application.SetFocus(AlbumPane)
				ArtistPane.SetSelectedBackgroundColor(tcell.ColorLightGray)
			} else {
				a.Application.SetFocus(ArtistPane)
				ArtistPane.SetSelectedBackgroundColor(tcell.ColorCornflowerBlue)
			}

			return nil
		case tcell.KeyCtrlB:
			return tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone)
		case tcell.KeyCtrlF:
			return tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone)
		}

		switch event.Rune() {
		case 'p':
			go internal.Playpause()
		case 's':
			go internal.Stop()
		case '>':
			go internal.Next()
		case '<':
			go internal.Previous()
		case '+':
			go internal.VolumeUp()
		case '-':
			go internal.VolumeDown()
		case 'q':
			a.Application.Stop()
		}

		return event
	})

	if err := a.Application.SetRoot(appFlex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
