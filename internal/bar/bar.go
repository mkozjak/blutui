package bar

import (
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

// A Bar represents a bottom bar that holds containers such as [SearchBar] or [StatusBar].
type Bar struct {
	// The following fields hold interfaces that are used for communicating with
	// app, library and spinner instances. App is used for focusing-specific tasks,
	// library for music data manipulation by search and status bar components and
	// spinner in order to start or stop the loading indicator.
	app     app.Command
	library library.Command
	spinner spinner.Command

	// tview-specific widgets that represent types compatible with flex widget or
	// app focusing methods that are used to draw these widgets to the screen.
	status *tview.Grid
	search *tview.InputField

	// Currently shown container, such as "status" or "search".
	// Exposed via [CurrentContainer].
	currCont string
}

// New returns a new [Bar] given its dependencies app, library and spinner instances
// and a read-only channel that delivers player's updates like play, stream, stop etc.
//
// Returned Bar is suitable to be used for getting tview.Primitive that can be sent to
// tview's components for drawing to the screen. It is also used for switching between
// [StatusBar] and [SearchBar].
func New(a app.Command, l library.Command, sp spinner.Command, ch <-chan player.Status) *Bar {
	bar := &Bar{
		app:     a,
		library: l,
		spinner: sp,
	}

	stb := newStatusBar(a, l, sp)
	stbc := stb.createContainer()
	go stb.listen(ch)

	srb := newSearchBar(bar, l)
	srbc := srb.createContainer()

	bar.status = stbc
	bar.search = srbc
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
	return b.status
}

// SearchContainer returns tview.Primitive for the Search Bar.
// It is a pointer to the tview InputField component that implements tview.Primitive.
func (b *Bar) SearchContainer() tview.Primitive {
	return b.search
}

// Show switches to a Bar component given its name as the input. It handles keyboard
// focus automatically based on [Bar] component type.
func (b *Bar) Show(name string) {
	switch name {
	case "search":
		b.app.ShowBarComponent(b.search)
		b.currCont = "search"
	case "status":
		b.app.ShowBarComponent(b.status)
		p := b.app.PrevFocused()
		b.app.SetPrevFocused("search")
		b.currCont = "status"
		b.app.SetFocus(p)
	}
}
