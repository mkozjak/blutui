package bar

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal/library"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/blutui/spinner"
	"github.com/mkozjak/tview"
)

// A StatusBar is a [Bar] component that provides important player information
// such as network activity (net i/o), current volume, state of playback,
// artist and song names and currently shown page, such as library.
// StatusBar is permanently shown on Bar, meaning, all other components fall back to it
// after they are done with their work.
type StatusBar struct {
	// A tview-specific widget that holds all status bar information spread across
	// the grid view.
	container *tview.Grid

	// The following fields hold interfaces that are used for communicating with
	// app, library and spinner instances.
	app     appManager
	libs    map[string]library.CPMarkSetter
	spinner spinner.Container

	// The following fields are tview-specific widgets responsible for holding player
	// information like currently set volume level, player's current playback state,
	// a currently playing song (if any) and a currently shown app page.
	volume       *tview.Table
	playerStatus *tview.TextView
	nowPlaying   *tview.TextView
	currentPage  *tview.TextView
}

// newStatusBar returns a new [StatusBar] given its dependencies app, library and
// spinner instances.
// StatusBar is then used for the creation of its child containers for volume,
// player status, currently played song and currently shown app page.
func newStatusBar(a appManager, l map[string]library.CPMarkSetter, sp spinner.Container) *StatusBar {
	return &StatusBar{
		app:     a,
		libs:    l,
		spinner: sp,
	}
}

// createContainer creates a [StatusBar] container returning a pointer to
// tview's Grid type, that is directly used by app in order to turn on
// the status bar on [Bar] to show important player status messages.
func (sb *StatusBar) createContainer() *tview.Grid {
	sb.volume = tview.NewTable().
		SetFixed(1, 2).SetSelectable(false, false).
		SetCell(0, 0, tview.NewTableCell("").SetTextColor(tcell.ColorDefault)).
		SetCell(0, 1, tview.NewTableCell("").SetTextColor(tcell.ColorDefault))

	sb.playerStatus = tview.NewTextView()
	sb.playerStatus.SetTextColor(tcell.ColorDefault).SetBackgroundColor(tcell.ColorDefault)

	sb.nowPlaying = tview.NewTextView()
	sb.nowPlaying.SetTextColor(tcell.ColorDefault).SetBackgroundColor(tcell.ColorDefault)

	sb.currentPage = tview.NewTextView()
	sb.currentPage.SetChangedFunc(func() {
		sb.app.Draw()
	})

	sb.currentPage.SetTextColor(tcell.ColorDefault).SetBackgroundColor(tcell.ColorDefault)

	sb.container = tview.NewGrid().
		AddItem(sb.spinner.Container(), 0, 0, 1, 1, 1, 1, false).
		AddItem(sb.volume, 0, 1, 1, 1, 1, 8, false).
		AddItem(sb.playerStatus, 0, 2, 1, 1, 1, 20, false).
		AddItem(sb.nowPlaying, 0, 3, 1, 1, 1, 50, false).
		AddItem(sb.currentPage, 0, 4, 1, 1, 1, 10, false).
		SetColumns(3, 8, 20, 0, 10)

	sb.container.SetBackgroundColor(tcell.ColorDefault).SetBorder(false).SetBorderPadding(0, 0, 1, 1)

	return sb.container
}

// listen starts iterating player updates given its input read-only channel
// that is used to communicate player status via its long-polling API.
// This method takes care about reacting to player updates such as playback
// state, song changes, streaming quality information feeds, network and
// management errors.
func (sb *StatusBar) listen(ch <-chan player.Status) {
	for s := range ch {
		var cpTitle string
		var cpFormat string
		var cpQuality string
		currPage := sb.app.CurrentPage()

		switch s.State {
		case "play":
			s.State = "playing"
			cpTitle = s.Artist + " - " + s.Track
			cpFormat = s.Format
			cpQuality = s.Quality

			sb.libs[currPage].MarkCpArtist(s.Artist)
			sb.libs[currPage].MarkCpTrack(s.Track, s.Artist, s.Album)
			sb.libs[currPage].SetCpTrackName(s.Track)
			sb.libs[currPage].SetCpAlbumName(s.Album)
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
			sb.libs[currPage].MarkCpArtist("")
			sb.libs[currPage].SetCpTrackName("")
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
		case "neterr":
			s.State = "network error"
			cpTitle = ""
			cpFormat = ""
			cpQuality = ""
		case "ctrlerr":
			s.State = "player control error"
			cpTitle = ""
			cpFormat = ""
			cpQuality = ""
		}

		format := ""
		if cpQuality != "" || cpFormat != "" {
			format = cpQuality + " " + cpFormat
		}

		var repeat string
		switch s.Repeat {
		case 0:
			repeat = " ♯ "
		case 1:
			repeat = " ∞ "
		default:
			repeat = " "
		}

		sb.volume.SetCell(0, 0, tview.NewTableCell("vol:").SetTextColor(tcell.ColorDefault))
		sb.volume.SetCell(0, 1, tview.NewTableCell(strconv.Itoa(s.Volume)).SetTextColor(tcell.ColorDefault))
		sb.playerStatus.SetText(s.State + repeat + format).SetTextAlign(tview.AlignLeft)
		sb.nowPlaying.SetText(cpTitle).SetTextAlign(tview.AlignCenter)
		sb.currentPage.SetText(currPage).SetTextAlign(tview.AlignCenter).SetTextColor(tcell.ColorBlack).
			SetBackgroundColor(tcell.ColorCornflowerBlue)

		if currPage == "local" {
			sb.currentPage.SetTextColor(tcell.ColorWhite).
				SetBackgroundColor(tcell.ColorCornflowerBlue)
		} else if currPage == "tidal" {
			sb.currentPage.SetTextColor(tcell.ColorWhite).
				SetBackgroundColor(tcell.ColorGrey)
		}

		sb.app.Draw()
	}
}

// SetCurrentPage updates the label showing currently open application page
// such as Library or Help screen given its input page name.
func (sb *StatusBar) SetCurrentPage(name string) {
	sb.currentPage.SetText(name)
}
