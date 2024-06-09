package statusbar

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/tview"
)

type Container interface {
	tview.Primitive
}

type ContainerWrapper struct {
	tview.Primitive
	table *tview.Table
}

type IStatusBar interface {
}

type Dependencies struct {
}

type StatusBar struct {
	container *tview.Table
	// Library        *library.Library
	// GetCurrentPage func() string
	// AppDraw        func()
}

func NewStatusBar(d Dependencies) *StatusBar {
	return &StatusBar{}
}

// bottom bar - status
func (sb *StatusBar) CreateContainer() (Container, error) {
	t := tview.NewTable().
		SetFixed(1, 3).
		SetSelectable(false, false).
		SetCell(0, 0, tview.NewTableCell("connecting").
			SetExpansion(1).
			SetTextColor(tcell.ColorDefault).
			SetAlign(tview.AlignLeft)).
		SetCell(0, 1, tview.NewTableCell("welcome to blutui =)").
			SetExpansion(2).
			SetTextColor(tcell.ColorDefault).
			SetAlign(tview.AlignCenter).
			SetMaxWidth(40)).
		SetCell(0, 2, tview.NewTableCell("").
			SetExpansion(1).
			SetTextColor(tcell.ColorDefault).
			SetAlign(tview.AlignRight))

	t.SetBackgroundColor(tcell.ColorDefault).SetBorder(false).SetBorderPadding(0, 0, 1, 1)

	return &ContainerWrapper{table: t}, nil
}

func (sb *StatusBar) Listen(ch <-chan player.Status) {
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
			sb.Library.HighlightCpArtist(s.Artist)
			sb.Library.CpTrackName = s.Track
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
			sb.Library.HighlightCpArtist("")
			sb.Library.CpTrackName = ""
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

		currPage := sb.GetCurrentPage()
		format := ""
		if cpQuality != "" || cpFormat != "" {
			format = " | " + cpQuality + " " + cpFormat
		}

		sb.container.GetCell(0, 0).SetText("vol: " + strconv.Itoa(s.Volume) +
			" | " + s.State + format)
		sb.container.GetCell(0, 1).SetText(cpTitle)
		sb.container.GetCell(0, 2).SetText(currPage)
		sb.AppDraw()
	}
}
