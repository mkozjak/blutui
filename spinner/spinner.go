// Created by Navid Yaghoobi and adopted from
// https://github.com/navidys/tvxwidgets/blob/b555c093da2ad329f1c79eb0f0631a1b9c616efe/spinner.go
// License text: https://github.com/navidys/tvxwidgets/blob/b555c093da2ad329f1c79eb0f0631a1b9c616efe/LICENSE
package spinner

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/blutui/internal/app"
	"github.com/mkozjak/tview"
)

type Command interface {
	GetContainer() tview.Primitive
	Start()
	Stop()
}


// Spinner represents a spinner widget.
type Spinner struct {
	*tview.Box
	counter      int
	currentStyle SpinnerStyle
	styles       map[SpinnerStyle][]rune
	active       bool
	drawer       app.Drawer
	stop         chan bool
}

type SpinnerStyle int

const (
	SpinnerDotsCircling SpinnerStyle = iota
	SpinnerDotsUpDown
	SpinnerBounce
	SpinnerLine
	SpinnerCircleQuarters
	SpinnerSquareCorners
	SpinnerCircleHalves
	SpinnerCorners
	SpinnerArrows
	SpinnerHamburger
	SpinnerStack
	SpinnerGrowHorizontal
	SpinnerGrowVertical
	SpinnerStar
	SpinnerBoxBounce
	spinnerCustom // non-public constant to indicate that a custom style has been set by the user.
)

// NewSpinner returns a new spinner widget.
func New(app app.Drawer) *Spinner {
	return &Spinner{
		Box:          tview.NewBox(),
		currentStyle: SpinnerDotsCircling,
		styles: map[SpinnerStyle][]rune{
			SpinnerDotsCircling:   []rune(`⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`),
			SpinnerDotsUpDown:     []rune(`⠋⠙⠚⠞⠖⠦⠴⠲⠳⠓`),
			SpinnerBounce:         []rune(`⠄⠆⠇⠋⠙⠸⠰⠠⠰⠸⠙⠋⠇⠆`),
			SpinnerLine:           []rune(`|/-\`),
			SpinnerCircleQuarters: []rune(`◴◷◶◵`),
			SpinnerSquareCorners:  []rune(`◰◳◲◱`),
			SpinnerCircleHalves:   []rune(`◐◓◑◒`),
			SpinnerCorners:        []rune(`⌜⌝⌟⌞`),
			SpinnerArrows:         []rune(`⇑⇗⇒⇘⇓⇙⇐⇖`),
			SpinnerHamburger:      []rune(`☰☱☳☷☶☴`),
			SpinnerStack:          []rune(`䷀䷪䷡䷊䷒䷗䷁䷖䷓䷋䷠䷫`),
			SpinnerGrowHorizontal: []rune(`▉▊▋▌▍▎▏▎▍▌▋▊▉`),
			SpinnerGrowVertical:   []rune(`▁▃▄▅▆▇▆▅▄▃`),
			SpinnerStar:           []rune(`✶✸✹✺✹✷`),
			SpinnerBoxBounce:      []rune(`▌▀▐▄`),
		},
		active: false,
		drawer: app,
		stop:   make(chan bool),
	}
}

func (s *Spinner) GetContainer() tview.Primitive {
	return s
}

// Draw draws this primitive onto the screen.
func (s *Spinner) Draw(screen tcell.Screen) {
	internal.Log("check:", s.active)

	if s.active {
		s.Box.DrawForSubclass(screen, s)
		x, y, width, _ := s.Box.GetInnerRect()
		tview.Print(screen, s.getCurrentFrame(), x, y, width, tview.AlignLeft, tcell.ColorDefault)
	} else {
		s.Box.DrawForSubclass(screen, tview.NewTextView())
		x, y, width, _ := s.Box.GetInnerRect()
		tview.Print(screen, "✓", x, y, width, tview.AlignLeft, tcell.ColorDefault)
	}
}

func (s *Spinner) Start() {
	s.active = true
	tick := time.NewTicker(100 * time.Millisecond)

	for {
		select {
		case <-tick.C:
			s.counter++
			s.drawer.Draw()
		case <-s.stop:
			s.counter = 0
			s.active = false
			s.drawer.Draw()
			return
		}
	}
}

func (s *Spinner) Stop() {
	s.stop <- true
}

// SetStyle sets the spinner style.
func (s *Spinner) SetStyle(style SpinnerStyle) *Spinner {
	s.currentStyle = style

	return s
}

func (s *Spinner) getCurrentFrame() string {
	frames := s.styles[s.currentStyle]
	if len(frames) == 0 {
		return ""
	}

	return string(frames[s.counter%len(frames)])
}

// SetCustomStyle sets a list of runes as custom frames to show as the spinner.
func (s *Spinner) SetCustomStyle(frames []rune) *Spinner {
	s.styles[spinnerCustom] = frames
	s.currentStyle = spinnerCustom

	return s
}
