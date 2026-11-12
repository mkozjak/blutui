package library

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/tview"
)

func TestAlbumPageNavigation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		key       tcell.Key
		album     int
		row       int
		presses   int
		height    int
		wantAlbum int
		wantRow   int
	}{
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
				container.InputHandler()(tcell.NewEventKey(tc.key, 0, tcell.ModNone), func(p tview.Primitive) {
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

	l.albumPane.SetOffset(1, 0)
	press('h')
	press('h')

	if !l.artistPane.HasFocus() {
		t.Fatal("h did not focus the artist pane")
	}

	if offset, _ := l.albumPane.GetOffset(); offset != 0 {
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
			l.CreateContainer()
			l.ApplyData(FetchDone{artists: []string{"A", "B"}, albumArtists: map[string]artist{
				"A": {albums: []album{{name: "Other", tracks: []track{{name: "Other"}}}}},
				"B": {albums: []album{
					{name: "First", year: 2020, tracks: []track{{name: "1. Track"}}},
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

			l.SelectCpArtist()

			if l.artistPane.GetCurrentItem() != 1 || !l.currentArtistAlbums[1].HasFocus() {
				t.Fatal("did not focus the playing artist and album")
			}

			if row, _ := l.currentArtistAlbums[1].GetSelection(); row != 1 {
				t.Fatalf("selected track row = %d, want 1", row)
			}
		})
	}
}
