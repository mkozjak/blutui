package internal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/tview"
)

// right pane - albums
func (a *App) CreateAlbumPane() *tview.Grid {
	p := tview.NewGrid().
		SetColumns(0)

	p.SetTitle(" [::b]Track ").
		SetBorder(true).
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetTitleAlign(tview.AlignLeft).
		SetCustomBorders(AlbumPaneStyle)

	return p
}
