package library

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestAlbumBoundaryScrollsOnlyWhenNeeded(t *testing.T) {
	l := New("", "local", testApp{tview.NewApplication()}, nil, testSpinner{})
	container := l.CreateContainer()
	albums := []album{{name: "First", tracks: make([]track, 8)}, {name: "Second", tracks: make([]track, 4)}}
	for i := range albums {
		for row := range albums[i].tracks {
			albums[i].tracks[row].name = fmt.Sprintf("%c%d", 'A'+i, row+1)
		}
	}

	l.ApplyData(FetchDone{artists: []string{"Artist"}, albumArtists: map[string]artist{"Artist": {albums: albums}}})
	l.app.SetFocus(l.currentArtistAlbums[1])
	screen := tcell.NewSimulationScreen("")

	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()

	screen.SetSize(80, 10)
	container.SetRect(0, 0, 80, 10)
	visibleRows := func() []string {
		screen.Clear()
		container.Draw(screen)
		x, y, _, height := l.albumPane.GetInnerRect()
		rows := make([]string, height)
		for i := range rows {
			for col := x + 1; col < x+7; col++ {
				ch, _, _, _ := screen.GetContent(col, y+i)
				rows[i] += string(ch)
			}
			rows[i] = strings.TrimSpace(rows[i])
		}

		return rows
	}
	setFocus := func(p tview.Primitive) { l.app.SetFocus(p) }
	press := func(key rune) {
		container.InputHandler()(tcell.NewEventKey(tcell.KeyRune, key, tcell.ModNone), setFocus)
	}

	before := visibleRows()
	press('k')

	if after := visibleRows(); !slices.Equal(after, before) {
		t.Fatalf("crossing to a visible track changed the viewport from %q to %q", before, after)
	}

	want := []string{"A4", "A5", "A6", "A7", "A8", "", "Second", "B1"}
	if !slices.Equal(before, want) {
		t.Fatalf("visible rows = %q, want %q", before, want)
	}

	x, y, _, _ := l.albumPane.GetInnerRect()
	_, _, style, _ := screen.GetContent(x+1, y+4)
	if _, background, _ := style.Decompose(); background != tcell.ColorCornflowerBlue {
		t.Fatal("previous album's last track is missing the selection highlight")
	}

	for range 5 {
		press('k')
		container.Draw(screen)
	}

	want = []string{"A3", "A4", "A5", "A6", "A7", "A8", "", "Second"}
	if got := visibleRows(); !slices.Equal(got, want) {
		t.Fatalf("scrolling up one line shows %q, want %q", got, want)
	}

	mouse := container.MouseHandler()
	mouse(tview.MouseScrollDown, tcell.NewEventMouse(x+1, y, tcell.WheelDown, tcell.ModNone), setFocus)
	if got := visibleRows(); !slices.Equal(got, before) {
		t.Fatalf("mouse scroll shows %q, want %q", got, before)
	}

	mouse(tview.MouseLeftDown, tcell.NewEventMouse(x+1, y, tcell.Button1, tcell.ModNone), setFocus)
	if got := visibleRows(); !slices.Equal(got, before) {
		t.Fatalf("clicking a visible track changed the viewport to %q", got)
	}

	if row, _ := l.currentArtistAlbums[0].GetSelection(); row != 3 {
		t.Fatalf("clicked track row = %d, want 3", row)
	}

	for range 3 {
		press('k')
		container.Draw(screen)
	}

	want = []string{"First", "A1", "A2", "A3", "A4", "A5", "A6", "A7"}
	if got := visibleRows(); !slices.Equal(got, want) {
		t.Fatalf("scrolling to the first track shows %q, want %q", got, want)
	}

	screen.SetSize(80, 3)
	container.SetRect(0, 0, 80, 3)
	if got := visibleRows(); !slices.Equal(got, []string{"A1"}) {
		t.Fatalf("one-line viewport shows %q, want the selected track", got)
	}
}

func TestAlbumNavigation(t *testing.T) {
	for _, tc := range []struct {
		name          string
		key           tcell.Key
		char          rune
		album         int
		row           int
		rememberedRow int
		presses       int
		height        int
		wantAlbum     int
		wantRow       int
	}{
		{name: "k from first track goes to previous album's last track", key: tcell.KeyRune, char: 'k', album: 2, presses: 1, wantAlbum: 1, wantRow: 1},
		{name: "j from last track goes to next album's first track", key: tcell.KeyRune, char: 'j', album: 1, row: 1, rememberedRow: 4, presses: 1, wantAlbum: 2},
		{name: "k stops at first track of first album", key: tcell.KeyRune, char: 'k', presses: 1},
		{name: "j stops at last track of last album", key: tcell.KeyRune, char: 'j', album: 3, row: 2, presses: 1, wantAlbum: 3, wantRow: 2},
		{name: "left bracket selects the previous album's first track", key: tcell.KeyRune, char: '[', album: 2, row: 6, rememberedRow: 1, presses: 1, wantAlbum: 1},
		{name: "right bracket selects the next album's first track", key: tcell.KeyRune, char: ']', album: 1, rememberedRow: 6, presses: 1, wantAlbum: 2},
		{name: "left bracket keeps the selection in the first album", key: tcell.KeyRune, char: '[', row: 2, presses: 1, wantRow: 2},
		{name: "right bracket keeps the selection in the last album", key: tcell.KeyRune, char: ']', album: 3, row: 1, presses: 1, wantAlbum: 3, wantRow: 1},
		{name: "within album", key: tcell.KeyCtrlF, album: 2, row: 1, presses: 1, wantAlbum: 2, wantRow: 7},
		{name: "forward across short albums", key: tcell.KeyCtrlF, row: 2, presses: 1, wantAlbum: 2, wantRow: 2},
		{name: "backward across short albums", key: tcell.KeyCtrlB, album: 2, row: 1, presses: 1, wantRow: 1},
		{name: "forward from last track", key: tcell.KeyCtrlF, row: 3, presses: 1, wantAlbum: 2, wantRow: 3},
		{name: "backward from first track", key: tcell.KeyCtrlB, album: 2, presses: 1},
		{name: "page down key", key: tcell.KeyPgDn, row: 2, presses: 1, wantAlbum: 2, wantRow: 2},
		{name: "page up key", key: tcell.KeyPgUp, album: 2, row: 1, presses: 1, wantRow: 1},
		{name: "stop at final track", key: tcell.KeyCtrlF, presses: 10, wantAlbum: 3, wantRow: 2},
		{name: "stop at first track", key: tcell.KeyCtrlB, album: 3, row: 2, presses: 10},
		{name: "half page down within album", key: tcell.KeyCtrlD, album: 2, row: 1, presses: 1, wantAlbum: 2, wantRow: 4},
		{name: "half page up within album", key: tcell.KeyCtrlU, album: 2, row: 7, presses: 1, wantAlbum: 2, wantRow: 4},
		{name: "half page down across albums", key: tcell.KeyCtrlD, row: 3, presses: 1, wantAlbum: 2},
		{name: "half page up across albums", key: tcell.KeyCtrlU, album: 2, presses: 1, wantRow: 3},
		{name: "half page stop at final track", key: tcell.KeyCtrlD, presses: 10, wantAlbum: 3, wantRow: 2},
		{name: "half page stop at first track", key: tcell.KeyCtrlU, album: 3, row: 2, presses: 10},
		{name: "half page in short viewport", key: tcell.KeyCtrlD, height: 4, presses: 1, wantRow: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			l := New("", "local", testApp{tview.NewApplication()}, nil, testSpinner{})
			container := l.CreateContainer()
			var albums []album

			for _, count := range []int{4, 2, 12, 3} {
				albums = append(albums, album{name: "Album", tracks: make([]track, count)})
			}

			l.ApplyData(FetchDone{artists: []string{"Artist"}, albumArtists: map[string]artist{"Artist": {albums: albums}}})
			l.currentArtistAlbums[tc.wantAlbum].Select(tc.rememberedRow, 0)
			l.currentArtistAlbums[tc.album].Select(tc.row, 0)
			l.app.SetFocus(l.currentArtistAlbums[tc.album])

			screen := tcell.NewSimulationScreen("")

			if err := screen.Init(); err != nil {
				t.Fatal(err)
			}

			defer screen.Fini()
			height := tc.height

			if height == 0 {
				height = 10
			}

			screen.SetSize(80, height)
			container.SetRect(0, 0, 80, height)
			container.Draw(screen)

			for i := 0; i < tc.presses; i++ {
				container.InputHandler()(tcell.NewEventKey(tc.key, tc.char, tcell.ModNone), func(p tview.Primitive) {
					l.app.SetFocus(p)
				})
				container.Draw(screen)
			}

			if got := l.selectedAlbumIdx(); got != tc.wantAlbum {
				t.Fatalf("selected album = %d, want %d", got, tc.wantAlbum)
			}

			selected := l.currentArtistAlbums[tc.wantAlbum]
			row, _ := selected.GetSelection()

			if row != tc.wantRow {
				t.Fatalf("selected row = %d, want %d", row, tc.wantRow)
			}

			if !selected.HasFocus() {
				t.Fatal("selected album does not have focus")
			}
		})
	}
}

func TestArtistPageKeys(t *testing.T) {
	l := New("", "local", testApp{tview.NewApplication()}, nil, testSpinner{})
	l.CreateContainer()
	l.app.SetFocus(l.artistPane)

	for _, tc := range []struct {
		key  tcell.Key
		want tcell.Key
	}{
		{key: tcell.KeyCtrlF, want: tcell.KeyPgDn},
		{key: tcell.KeyCtrlB, want: tcell.KeyPgUp},
	} {
		event := l.KeyboardHandler(tcell.NewEventKey(tc.key, 0, tcell.ModNone))

		if event == nil || event.Key() != tc.want {
			t.Fatalf("artist page event = %v, want key %v", event, tc.want)
		}
	}
}

func TestPaneFocusKeys(t *testing.T) {
	l := New("", "local", testApp{tview.NewApplication()}, nil, testSpinner{})
	container := l.CreateContainer()
	l.ApplyData(FetchDone{artists: []string{"Artist"}, albumArtists: map[string]artist{
		"Artist": {albums: []album{{name: "Album", tracks: []track{{name: "Track"}, {name: "Track 2"}}}}},
	}})
	l.app.SetFocus(l.artistPane)
	screen := tcell.NewSimulationScreen("")

	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}

	defer screen.Fini()
	screen.SetSize(80, 24)
	container.SetRect(0, 0, 80, 24)
	container.Draw(screen)
	press := func(key rune) {
		container.InputHandler()(tcell.NewEventKey(tcell.KeyRune, key, tcell.ModNone), func(p tview.Primitive) {
			l.app.SetFocus(p)
		})
		container.Draw(screen)
	}
	press('l')

	if !l.currentArtistAlbums[0].HasFocus() {
		t.Fatal("l did not focus the album pane")
	}

	l.currentArtistAlbums[0].Select(1, 0)
	press('l')
	row, _ := l.currentArtistAlbums[0].GetSelection()

	if row != 1 {
		t.Fatal("l changed selection while already in the album pane")
	}

	l.albumPane.rowOffset = 1
	press('h')
	press('h')

	if !l.artistPane.HasFocus() {
		t.Fatal("h did not focus the artist pane")
	}

	if l.albumPane.rowOffset != 0 {
		t.Fatal("h did not reset the album scroll offset")
	}

	l.ApplyData(FetchDone{artists: []string{"Artist"}, albumArtists: map[string]artist{"Artist": {}}})
	press('l')

	if !l.artistPane.HasFocus() {
		t.Fatal("l moved focus for an artist without albums")
	}
}

func TestJumpToCurrentlyPlayingTrack(t *testing.T) {
	for _, startInAlbum := range []bool{false, true} {
		t.Run(map[bool]string{false: "artist pane", true: "album pane"}[startInAlbum], func(t *testing.T) {
			l := New("", "local", testApp{tview.NewApplication()}, nil, testSpinner{})
			container := l.CreateContainer()
			l.ApplyData(FetchDone{artists: []string{"A", "B"}, albumArtists: map[string]artist{
				"A": {albums: []album{{name: "Other", tracks: []track{{name: "Other"}}}}},
				"B": {albums: []album{
					{name: "First", year: 2020, tracks: make([]track, 8)},
					{name: "Second", year: 2021, tracks: []track{{name: "1. Track"}, {name: "2. Playing"}}},
				}},
			}})
			l.SetCpTrackName("Playing")
			l.SetCpAlbumName("Second")
			l.MarkCpArtist("B")

			if startInAlbum {
				l.app.SetFocus(l.currentArtistAlbums[0])
			} else {
				l.app.SetFocus(l.artistPane)
			}

			screen := tcell.NewSimulationScreen("")
			if err := screen.Init(); err != nil {
				t.Fatal(err)
			}
			defer screen.Fini()

			screen.SetSize(80, 10)
			container.SetRect(0, 0, 80, 10)
			container.Draw(screen)

			albumText := func() string {
				screen.Clear()
				container.Draw(screen)
				x, y, width, height := l.albumPane.GetInnerRect()
				var view strings.Builder
				for row := y; row < y+height; row++ {
					for col := x; col < x+width; col++ {
						ch, _, _, _ := screen.GetContent(col, row)
						view.WriteRune(ch)
					}
					view.WriteByte('\n')
				}
				return view.String()
			}

			l.SelectCpArtist()
			playingView := albumText()

			if l.artistPane.GetCurrentItem() != 1 || !l.currentArtistAlbums[1].HasFocus() {
				t.Fatal("did not focus the playing artist and album")
			}

			if row, _ := l.currentArtistAlbums[1].GetSelection(); row != 1 {
				t.Fatalf("selected track row = %d, want 1", row)
			}

			if !strings.Contains(playingView, "2. Playing") {
				t.Fatalf("playing track is outside the viewport:\n%s", playingView)
			}

			tab := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
			l.KeyboardHandler(tab)

			if !l.artistPane.HasFocus() {
				t.Fatal("Tab did not focus the artist pane")
			}

			if got := albumText(); got != playingView {
				t.Fatalf("Tab changed the track viewport:\n%s\nwant:\n%s", got, playingView)
			}

			l.KeyboardHandler(tab)

			if !l.currentArtistAlbums[1].HasFocus() {
				t.Fatal("Tab did not return focus to the playing album")
			}

			if got := albumText(); got != playingView {
				t.Fatalf("returning focus changed the track viewport:\n%s\nwant:\n%s", got, playingView)
			}
		})
	}
}
