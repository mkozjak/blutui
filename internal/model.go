package internal

import (
	"strings"
	"time"

	"github.com/mkozjak/tview"
)

var CustomBorders = &tview.BoxBorders{
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

var NoBorders = &tview.BoxBorders{}

type App struct {
	Application *tview.Application
	Pages       *tview.Pages
	StatusBar   *tview.Table
	HelpScreen  *tview.Modal
	playerState string
}

type Cache struct {
	Data map[string]CacheItem
}

type CacheItem struct {
	Response   []byte
	Expiration time.Time
}

func (a *App) cpHighlightArtist(name string) {
	// clear previously highlighted items
	if a.CpArtistIdx >= 0 {
		n, _ := a.ArtistPane.GetItemText(a.CpArtistIdx)
		a.ArtistPane.SetItemText(a.CpArtistIdx, strings.TrimPrefix(n, "[yellow]"), "")
	}

	if name == "" {
		a.CpArtistIdx = -1
		return
	}

	// highlight artist
	// track is highlighted through a.newAlbumList
	idx := a.ArtistPane.FindItems(name, "", false, true)
	if len(idx) < 1 {
		return
	}

	n, _ := a.ArtistPane.GetItemText(idx[0])
	a.ArtistPane.SetItemText(idx[0], "[yellow]"+n, "")
	a.CpArtistIdx = idx[0]
}
