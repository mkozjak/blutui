package internal

import "github.com/mkozjak/tview"

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
