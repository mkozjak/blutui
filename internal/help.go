package internal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/tview"
)

// This is a library help screen
// TODO: bind to Library
func CreateHelpScreen(listen func(event *tcell.EventKey) *tcell.EventKey) *tview.Modal {
	var text string
	keybindings := map[string]string{
		"start playback":                   "↵",
		"play/pause":                       "p",
		"stop":                             "s",
		"next song":                        ">",
		"previous song":                    "<",
		"volume up":                        "+",
		"volume down":                      "-",
		"toggle mute":                      "m",
		"page down":                        "ctrl+f",
		"page up":                          "ctrl+b",
		"half page down":                   "ctrl+d",
		"half page up":                     "ctrl+u",
		"jump to currently playing artist": "o",
		"search artists":                   "/",
		"update library":                   "u",
		"show this screen":                 "h",
		"quit app":                         "q",
	}

	order := []string{
		"start playback", "play/pause", "stop", "next song", "previous song",
		"volume up", "volume down", "toggle mute", "page down", "page up",
		"half page down", "half page up", "jump to currently playing artist",
		"search artists", "update library", "show this screen", "quit",
	}

	for _, action := range order {
		text = text + keybindings[action] + " - " + action + "\n"
	}

	c := tview.NewModal().
		SetText(text).
		SetBackgroundColor(tcell.ColorDefault)

	c.SetInputCapture(listen).
		SetBorder(true).
		SetTitle("[::b]Keybindings").
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetCustomBorders(CustomBorders)

	return c
}
