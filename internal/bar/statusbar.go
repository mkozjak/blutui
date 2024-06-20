package bar

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal/app"
	"github.com/mkozjak/blutui/internal/library"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/blutui/spinner"
	"github.com/mkozjak/tview"
)

type StatusCommand interface {
	SetCurrentPage(name string)
}

// injection target
type StatusBar struct {
	container    *tview.Grid
	app          app.Command
	library      library.Command
	spinner      spinner.Command
	volume       *tview.Table
	playerStatus *tview.TextView
	nowPlaying   *tview.TextView
	currentPage  *tview.TextView
}

func newStatusBar(a app.Command, l library.Command, sp spinner.Command) *StatusBar {
	return &StatusBar{
		app:      a,
		library:  l,
		spinner: sp,
	}
}

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
	sb.currentPage.SetTextColor(tcell.ColorDefault).SetBackgroundColor(tcell.ColorDefault)

	sb.container = tview.NewGrid().
		AddItem(sb.spinner.GetContainer(), 0, 0, 1, 1, 1, 1, false).
		AddItem(sb.volume, 0, 1, 1, 1, 1, 8, false).
		AddItem(sb.playerStatus, 0, 2, 1, 1, 1, 20, false).
		AddItem(sb.nowPlaying, 0, 3, 1, 1, 1, 50, false).
		AddItem(sb.currentPage, 0, 4, 1, 1, 1, 10, false).
		SetColumns(3, 8, 20, 0, 10)

	sb.container.SetBackgroundColor(tcell.ColorDefault).SetBorder(false).SetBorderPadding(0, 0, 1, 1)

	return sb.container
}

func (sb *StatusBar) listen(ch <-chan player.Status) {
	for s := range ch {
		var cpTitle string
		var cpFormat string
		var cpQuality string

		switch s.State {
		case "play":
			s.State = "playing"
			cpTitle = s.Artist + " - " + s.Track
			cpFormat = s.Format
			cpQuality = s.Quality

			// TODO: should probably be done elsewhere
			if sb.app.CurrentPage() == "library" {
				if s.Service == "LocalMusic" {
					sb.library.HighlightCpArtist(s.Artist)
					sb.library.SetCpTrackName(s.Track)
				} else {
					sb.library.HighlightCpArtist("")
					sb.library.SetCpTrackName("")
				}
			}
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
			sb.library.HighlightCpArtist("")
			sb.library.SetCpTrackName("")
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

		currPage := sb.app.CurrentPage()
		format := ""
		if cpQuality != "" || cpFormat != "" {
			format = " | " + cpQuality + " " + cpFormat
		}

		sb.volume.SetCell(0, 0, tview.NewTableCell("vol:").SetTextColor(tcell.ColorDefault))
		sb.volume.SetCell(0, 1, tview.NewTableCell(strconv.Itoa(s.Volume)).SetTextColor(tcell.ColorDefault))
		sb.playerStatus.SetText(s.State + format).SetTextAlign(tview.AlignLeft)
		sb.nowPlaying.SetText(cpTitle).SetTextAlign(tview.AlignCenter)
		sb.currentPage.SetText(currPage).SetTextAlign(tview.AlignRight)

		sb.app.Draw()
	}
}

func (sb *StatusBar) setCurrentPage(name string) {
	sb.currentPage.SetText(name)
}
