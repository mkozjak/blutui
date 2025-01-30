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
		"show local library":                  "1",
		"show tidal library":                  "2",
		"start playback":                      "↵",
		"play selected song only":             "x",
		"play/pause":                          "p",
		"stop":                                "s",
		"next song":                           ">",
		"previous song":                       "<",
		"volume up":                           "+",
		"volume down":                         "-",
		"toggle mute":                         "m",
		"toggle repeat mode (none, all, one)": "r",
		"page down":                           "ctrl+f",
		"page up":                             "ctrl+b",
		"half page down":                      "ctrl+d",
		"half page up":                        "ctrl+u",
		"jump to currently playing artist":    "o",
		"search artists":                      "/",
		"update library":                      "u",
		"show this screen":                    "h",
		"quit app":                            "q",
	}

	order := []string{
		"show local library", "show tidal library", "start playback", "play selected song only", "play/pause", "stop",
		"next song", "previous song", "volume up", "volume down", "toggle mute", "toggle repeat mode (none, all, one)",
		"page down", "page up", "half page down", "half page up", "jump to currently playing artist",
		"search artists", "update library", "show this screen", "quit app",
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
