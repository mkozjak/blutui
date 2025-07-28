package library

import (
	"testing"

	"github.com/mkozjak/tview"
)

func TestMarkCpTrack_UnmarksPreviousTracksFromDifferentAlbums(t *testing.T) {
	// Create a minimal library instance for testing
	lib := &Library{
		cpArtistIdx: 0,
	}

	// Set up test data - artist with two albums
	lib.artists = []string{"Kid Rock"}

	// Create album tables to simulate the UI (including album header rows)
	album1 := tview.NewTable()
	album1.SetTitle("[::b]The History Of Rock (2017)")
	album1.SetCell(0, 0, tview.NewTableCell("The History Of Rock"))  // Header row
	album1.SetCell(1, 0, tview.NewTableCell("So Hott"))              // Track row
	album1.SetCell(2, 0, tview.NewTableCell("Born Free"))            // Track row

	album2 := tview.NewTable()
	album2.SetTitle("[::b]Cocky (2001)")
	album2.SetCell(0, 0, tview.NewTableCell("Cocky"))                    // Header row
	album2.SetCell(1, 0, tview.NewTableCell("Forever"))                  // Track row
	album2.SetCell(2, 0, tview.NewTableCell("Lonely Road of Faith"))     // Track row

	lib.currentArtistAlbums = []*tview.Table{album1, album2}

	// First, mark a track in "The History Of Rock"
	lib.MarkCpTrack("So Hott", "Kid Rock", "The History Of Rock")

	// Verify the track is marked
	cell := album1.GetCell(1, 0)  // Row 1, not 0
	if cell.Text != "[yellow]So Hott" {
		t.Errorf("Expected track to be marked, got: %s", cell.Text)
	}

	// Now mark a track in "Cocky" 
	lib.MarkCpTrack("Forever", "Kid Rock", "Cocky")

	// Verify the previous track is unmarked
	cell = album1.GetCell(1, 0)  // Row 1, not 0
	if cell.Text != "So Hott" {
		t.Errorf("Expected previous track to be unmarked, got: %s", cell.Text)
	}

	// Verify the new track is marked
	cell = album2.GetCell(1, 0)  // Row 1, not 0
	if cell.Text != "[yellow]Forever" {
		t.Errorf("Expected new track to be marked, got: %s", cell.Text)
	}
}