package bar

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal/app"
	"github.com/mkozjak/blutui/internal/library"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/blutui/spinner"
	"github.com/mkozjak/tview"
)

// switcher is the interface that is implemented by [Bar] that enables
// switching between different Bar children, such as status or search bar components.
type switcher interface {
	Show(name string)
}

type appManager interface {
	app.Focuser
	app.StatusbarShower
	app.PageViewer
	app.Drawer
}

type LibManager interface {
	library.ArtistFilter
	library.CPMarkSetter
}

// A Bar represents a bottom bar that holds containers such as [SearchBar] or [StatusBar].
type Bar struct {
	// The following fields hold interfaces that are used for communicating with
	// app, libraries and spinner instances. App is used for focusing-specific tasks,
	// libraries for music data manipulation by search and status bar components and
	// spinner in order to start or stop the loading indicator.
	app     appManager
	libs    map[string]LibManager
	spinner spinner.Container

	status *StatusBar
	// tview-specific widgets that represent types compatible with flex widget or
	// app focusing methods that are used to draw these widgets to the screen.
	statusc *tview.Grid
	searchc *tview.InputField

	// Currently shown container, such as "status" or "search".
	// Exposed via [CurrentContainer].
	currCont string
}

// New returns a new [Bar] given its dependencies app, libraries and spinner instances
// and a read-only channel that delivers player's updates like play, stream, stop etc.
//
// Returned Bar is suitable to be used for getting tview.Primitive that can be sent to
// tview's components for drawing to the screen. It is also used for switching between
// [StatusBar] and [SearchBar].
func New(a appManager, l map[string]LibManager, sp spinner.Container, ch <-chan player.Status) *Bar {
	bar := &Bar{
		app:     a,
		libs:    l,
		spinner: sp,
	}

	CPMarkSetters := make(map[string]library.CPMarkSetter)
	for k, v := range l {
		CPMarkSetters[k] = v
	}

	stb := newStatusBar(a, CPMarkSetters, sp)
	stbc := stb.createContainer()
	go stb.listen(ch)

	artistFilters := make(map[string]library.ArtistFilter)
	for k, v := range l {
		artistFilters[k] = v
	}

	srb := newSearchBar(a, bar, artistFilters)
	srbc := srb.createContainer()

	bar.status = stb
	bar.statusc = stbc
	bar.searchc = srbc
	bar.currCont = "status"

	return bar
}

// CurrentContainer returns the name of a currently shown container.
func (b *Bar) CurrentContainer() string {
	return b.currCont
}

// StatusContainer returns tview.Primitive for the Status Bar.
// It is a pointer to the tview Grid component that implements tview.Primitive.
func (b *Bar) StatusContainer() tview.Primitive {
	return b.statusc
}

// SearchContainer returns tview.Primitive for the Search Bar.
// It is a pointer to the tview InputField component that implements tview.Primitive.
func (b *Bar) SearchContainer() tview.Primitive {
	return b.searchc
}

// Show switches to a Bar component given its name as the input. It handles keyboard
// focus automatically based on [Bar] component type.
func (b *Bar) Show(name string) {
	switch name {
	case "search":
		b.app.ShowBarComponent(b.searchc)
		b.currCont = "search"
	case "status":
		b.app.ShowBarComponent(b.statusc)
		p := b.app.PrevFocused()
		b.app.SetPrevFocused("search")
		b.currCont = "status"
		b.app.SetFocus(p)
	}
}

func (b *Bar) SetPageOnStatus(name string) {
	b.status.currentPage.SetText(name)

	if name == "local" {
		b.status.currentPage.SetTextColor(tcell.ColorWhite).
			SetBackgroundColor(tcell.ColorCornflowerBlue)
	} else if name == "tidal" {
		b.status.currentPage.SetTextColor(tcell.ColorWhite).
			SetBackgroundColor(tcell.ColorGrey)
	}
}
