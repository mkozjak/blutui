package internal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal/player"
)

func KeyboardHandler(event *tcell.EventKey, gotoCpArtist func(), quit func(), p *player.Player) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlQ:
		quit()
	}

	switch event.Rune() {
	case 'p':
		go p.Playpause()
	case 's':
		go p.Stop()
	case '>':
		go p.Next()
	case '<':
		go p.Previous()
	case '+':
		go p.VolumeHold(true)
	case '-':
		go p.VolumeHold(false)
	case 'm':
		go p.ToggleMute()
	case 'o':
		if p.GetState() == "playing" {
			gotoCpArtist()
		}
	case 'u':
		// go p.RefreshData()
	case 'h':
		// p, _ := a.Pages.GetFrontPage()
		// if p != "help" {
		// 	a.Pages.ShowPage("help")
		// 	return nil
		// }
	case 'q':
		quit()
	}

	return event
}

// func KbHelpHandler(event *tcell.EventKey) *tcell.EventKey {
// 	switch event.Key() {
// 	case tcell.KeyEscape:
// 		a.Pages.HidePage("help")
// 		return nil
// 	}

// 	switch event.Rune() {
// 	case 'h':
// 		p, _ := a.Pages.GetFrontPage()
// 		if p == "help" {
// 			a.Pages.HidePage("help")
// 			return nil
// 		}
// 	}

// 	return event
// }
