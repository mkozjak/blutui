package internal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/tview"
)

func CreateHelpScreen(listen func(event *tcell.EventKey) *tcell.EventKey) *tview.Modal {
	// TODO: do this the smarter way
	c := tview.NewModal().
		SetText("↵ - start playback\np - play/pause\ns - stop\n> - next song\n" +
			"< - previous song\n+ - volume up\n- - volume down\nm - toggle mute\n" +
			"ctrl+f - page down\nctrl+b - page up\n" +
			"ctrl+d - half page down\nctrl+u - half page up\n" +
			"o - jump to currently playing artist\nu - update library\n" +
			"h - show help screen\nq - quit app").
		SetBackgroundColor(tcell.ColorDefault)

	c.SetInputCapture(listen).
		SetBorder(true).
		SetTitle("[::b]Keybindings").
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetCustomBorders(CustomBorders)

	return c
}
