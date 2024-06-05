package internal

import "github.com/gdamore/tcell/v2"

func (a *App) KbGlobalHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlQ:
		a.Application.Stop()
	case tcell.KeyTab:
		if !a.AlbumPane.HasFocus() {
			a.Application.SetFocus(a.AlbumPane)
			a.ArtistPane.SetSelectedBackgroundColor(tcell.ColorLightGray)
		} else {
			a.Application.SetFocus(a.ArtistPane)
			a.ArtistPane.SetSelectedBackgroundColor(tcell.ColorCornflowerBlue)
		}

		return nil
	case tcell.KeyCtrlB:
		return tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone)
	case tcell.KeyCtrlF:
		return tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone)
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
	case 'q':
		a.Application.Stop()
	}

	return event
}
