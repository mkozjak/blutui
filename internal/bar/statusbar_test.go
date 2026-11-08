package bar

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/tview"
)

type statusApp struct {
	*tview.Application
	page   string
	queued int
}

func (a *statusApp) CurrentPage() string            { return a.page }
func (*statusApp) PrevFocused() tview.Primitive     { return nil }
func (*statusApp) SetPrevFocused(string)            {}
func (*statusApp) ShowBarComponent(tview.Primitive) {}
func (a *statusApp) QueueUpdateDraw(f func())       { a.queued++; f() }

type statusSpinner struct{}

func (statusSpinner) Container() tview.Primitive { return tview.NewBox() }

func TestStatusBarPlaybackTransitions(t *testing.T) {
	a := &statusApp{Application: tview.NewApplication(), page: "help"}
	sb := newStatusBar(a, nil, statusSpinner{})
	sb.createContainer()
	ch := make(chan player.Status, 1)

	for _, tc := range []struct {
		status player.Status
		want   string
	}{
		{player.Status{State: "play", Artist: " Artist ", Track: " Song "}, "Artist - Song"},
		{player.Status{State: "ctrlerr"}, "Artist - Song"},
		{player.Status{State: "stream", Artist: "Other", Track: "Next", Title2: "Station"}, "Other - Next"},
		{player.Status{State: "pause", Title2: "Radio track", Title3: "Station"}, "Radio track"},
		{player.Status{State: "stream", Title3: "Fallback"}, "Fallback"},
		{player.Status{State: "play", Track: "Song only"}, "Song only"},
		{player.Status{State: "pause", Artist: "Artist only"}, "Artist only"},
		{player.Status{State: "stop"}, ""},
	} {
		sb.update(tc.status)

		if got := sb.nowPlaying.GetText(false); got != tc.want {
			t.Errorf("state %s: title = %q, want %q", tc.status.State, got, tc.want)
		}
	}

	ch <- player.Status{State: "play", Artist: "Artist", Track: "Song"}
	close(ch)
	sb.listen(ch)

	if a.queued != 1 {
		t.Fatal("status update bypassed the UI queue")
	}

	screen := tcell.NewSimulationScreen("")

	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(48, 1)
	sb.container.SetRect(0, 0, 48, 1)
	sb.container.Draw(screen)
	_, _, width, _ := sb.nowPlaying.GetRect()

	if width <= 0 {
		t.Fatal("now playing hidden on narrow terminal")
	}
}
