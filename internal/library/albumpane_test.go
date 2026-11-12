package library

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestAlbumDrawKeepsTitleWithoutFrame(t *testing.T) {
	l := New("", "local", testApp{}, nil, testSpinner{})
	table := l.drawAlbum("Artist", album{name: "Album", year: 2020, tracks: []track{{name: "Track"}}})
	screen := tcell.NewSimulationScreen("")

	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()

	screen.SetSize(30, 4)
	table.SetRect(0, 0, 30, 4)
	table.Draw(screen)

	for _, point := range [][2]int{{0, 0}, {29, 0}, {0, 1}, {29, 1}, {0, 3}, {29, 3}} {
		ch, _, _, _ := screen.GetContent(point[0], point[1])
		if ch != ' ' {
			t.Errorf("frame at %v = %q, want space", point, ch)
		}
	}

	for i, want := range "Album (2020)" {
		ch, _, _, _ := screen.GetContent(i+1, 0)
		if ch != want {
			t.Fatalf("title character %d = %q, want %q", i, ch, want)
		}
	}
}
