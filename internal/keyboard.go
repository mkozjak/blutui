package internal

import "github.com/gdamore/tcell/v2"

func (a *App) KeyboardHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlQ:
		a.Application.Stop()
	}

	switch event.Rune() {
	case 'p':
		go a.Playpause()
	case 's':
		go a.Stop()
	case '>':
		go a.Next()
	case '<':
		go a.Previous()
	case '+':
		go a.VolumeHold(true)
	case '-':
		go a.VolumeHold(false)
	case 'm':
		go a.ToggleMute()
	case 'o':
		if a.playerState == "playing" {
			a.ArtistPane.SetCurrentItem(a.CpArtistIdx)
		}
	case 'u':
		go a.RefreshData()
	case 'h':
		p, _ := a.Pages.GetFrontPage()
		if p != "help" {
			a.Pages.ShowPage("help")
			return nil
		}
	case 'q':
		a.Application.Stop()
	}

	return event
}

func (a *App) KbHelpHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		a.Pages.HidePage("help")
		return nil
	}

	switch event.Rune() {
	case 'h':
		p, _ := a.Pages.GetFrontPage()
		if p == "help" {
			a.Pages.HidePage("help")
			return nil
		}
	}

	return event
}
