package library

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/cache"
	"github.com/rivo/tview"
)

type testSpinner struct{}

func (testSpinner) Start() {}
func (testSpinner) Stop()  {}

type testApp struct {
	*tview.Application
}

func (testApp) PrevFocused() tview.Primitive { return nil }
func (testApp) SetPrevFocused(string)        {}
func (testApp) QueueUpdateDraw(f func())     { f() }

func libraryResponses(track string) map[string]string {
	return map[string]string{
		localRootEndpoint:     `<browse><item browseKey="section"/></browse>`,
		"/Browse?key=section": `<browse><item text="Album" text2="Artist" browseKey="album"/></browse>`,
		"/Browse?key=album":   fmt.Sprintf(`<browse><item text="%s" duration="60"/></browse>`, track),
		"/Songs?service=LocalMusic&album=Album&artist=Artist": `<songs><song><date>2020</date></song></songs>`,
	}
}

func receiveFetch(t *testing.T, ch <-chan FetchDone) FetchDone {
	t.Helper()

	select {
	case result, ok := <-ch:
		if !ok {
			t.Fatal("fetch closed without a result")
		}

		return result
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for library data")
		return FetchDone{}
	}
}

func TestFetchDataAlbumYearFallback(t *testing.T) {
	for _, tc := range []struct {
		name string
		date string
		want int
	}{
		{name: "invalid date", date: "1", want: 1969},
		{name: "missing date", want: 1969},
		{name: "zero date", date: "0", want: 1969},
		{name: "valid date takes precedence", date: "1970", want: 1970},
	} {
		t.Run(tc.name, func(t *testing.T) {
			responses := libraryResponses("So Excited")
			responses["/Songs?service=LocalMusic&album=Album&artist=Artist"] = fmt.Sprintf(
				`<songs><song><date>%s</date><fn>/var/mnt/TITAN-music/B.B. King/[1969] Completely Well/B.B. King_Completely Well_01_So Excited.flac</fn></song></songs>`, tc.date)
			l := New("", "local", testApp{tview.NewApplication()}, nil, testSpinner{})
			err := l.fetchData(func(endpoint string) ([]byte, error) {
				body, ok := responses[endpoint]

				if !ok {
					return nil, fmt.Errorf("unexpected endpoint: %s", endpoint)
				}

				return []byte(body), nil
			})
			if err != nil {
				t.Fatal(err)
			}

			if got := l.albumArtists["Artist"].albums[0].year; got != tc.want {
				t.Fatalf("album year = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestFetchDataCacheRefresh(t *testing.T) {
	for _, tc := range []struct {
		name      string
		seed      bool
		expired   bool
		cached    bool
		fail      bool
		partial   bool
		wantStale bool
		wantFetch bool
	}{
		{name: "fresh cache", seed: true, cached: true},
		{name: "stale cache", seed: true, expired: true, cached: true, wantStale: true, wantFetch: true},
		{name: "failed refresh", seed: true, expired: true, cached: true, fail: true, wantStale: true, wantFetch: true},
		{name: "cold start", cached: true, wantFetch: true},
		{name: "partial cache", seed: true, partial: true, cached: true, wantFetch: true},
		{name: "manual refresh", seed: true, wantFetch: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			gate := make(chan struct{})
			var releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(gate) }) }
			var requests atomic.Int32
			responses := libraryResponses("Fresh track")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				<-gate

				if tc.fail {
					http.Error(w, "unavailable", http.StatusServiceUnavailable)
					return
				}

				body, found := responses[r.URL.RequestURI()]

				if !found {
					http.NotFound(w, r)
					return
				}

				fmt.Fprint(w, body)
			}))
			defer func() {
				release()
				server.Close()
			}()

			if tc.seed {
				expiration := time.Now().Add(time.Hour)

				if tc.expired {
					expiration = time.Now().Add(-time.Hour)
				}

				c := cache.Cache{Data: make(map[string]cache.CacheItem)}

				for endpoint, body := range libraryResponses("Cached track") {
					if tc.partial && endpoint == "/Browse?key=album" {
						continue
					}

					c.Data[server.URL+endpoint] = cache.CacheItem{Response: []byte(body), Expiration: expiration}
				}

				data, err := json.Marshal(c)
				if err != nil {
					t.Fatal(err)
				}

				path := filepath.Join(os.Getenv("HOME"), ".config", "blutui", "cache")

				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}

				if err := os.WriteFile(path, data, 0644); err != nil {
					t.Fatal(err)
				}
			}

			l := New(server.URL, "local", testApp{tview.NewApplication()}, nil, testSpinner{})
			l.CreateContainer()
			ch := make(chan FetchDone)
			go l.FetchData(tc.cached, ch)

			if tc.wantStale || !tc.wantFetch {
				result := receiveFetch(t, ch)

				if result.Error != nil {
					t.Fatal(result.Error)
				}

				l.ApplyData(result)

				if got := l.albumArtists["Artist"].albums[0].tracks[0].name; got != "Cached track" {
					t.Fatalf("cached track = %q", got)
				}

				if got := l.currentArtistAlbums[0].GetCell(0, 0).Text; got != "Cached track" {
					t.Fatalf("cached UI track = %q", got)
				}
			}

			if tc.wantFetch {
				release()
				result := receiveFetch(t, ch)

				if (result.Error != nil) != tc.fail {
					t.Fatalf("unexpected refresh error: %v", result.Error)
				}

				l.ApplyData(result)
				want := "Fresh track"

				if tc.fail {
					want = "Cached track"
				}

				if got := l.albumArtists["Artist"].albums[0].tracks[0].name; got != want {
					t.Fatalf("displayed track = %q, want %q", got, want)
				}

				if got := l.currentArtistAlbums[0].GetCell(0, 0).Text; got != want {
					t.Fatalf("UI track = %q, want %q", got, want)
				}

				c, err := cache.LoadCache()
				if err != nil {
					t.Fatal(err)
				}

				if got := string(c.Data[server.URL+"/Browse?key=album"].Response); got != libraryResponses(want)["/Browse?key=album"] {
					t.Fatalf("unexpected persisted tracks: %s", got)
				}
			}

			select {
			case _, ok := <-ch:
				if ok {
					t.Fatal("unexpected extra result")
				}
			case <-time.After(5 * time.Second):
				t.Fatal("fetch did not finish")
			}

			if !tc.wantFetch && requests.Load() != 0 {
				t.Fatal("fresh cache triggered network requests")
			}
		})
	}
}

func TestApplyDataSelectionAndEmptyLibrary(t *testing.T) {
	l := New("", "local", testApp{tview.NewApplication()}, nil, testSpinner{})
	l.CreateContainer()
	data := FetchDone{artists: []string{"A", "B"}, albumArtists: map[string]artist{"A": {}, "B": {}}}
	l.ApplyData(data)
	l.artistPane.SetCurrentItem(1)
	l.ApplyData(data)
	selected, _ := l.artistPane.GetItemText(l.artistPane.GetCurrentItem())

	if selected != "B" {
		t.Fatalf("selection changed to %q", selected)
	}

	l.ApplyData(FetchDone{Error: errors.New("offline")})

	if !reflect.DeepEqual(l.artists, data.artists) {
		t.Fatal("failed refresh changed displayed data")
	}

	l.ApplyData(FetchDone{albumArtists: map[string]artist{}})

	if l.artistPane.GetItemCount() != 0 || len(l.currentArtistAlbums) != 0 {
		t.Fatal("empty library retained old UI data")
	}

	l.app.SetFocus(l.artistPane)
	l.KeyboardHandler(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
}

func TestSelectCurrentlyPlayingArtistDuringStartup(t *testing.T) {
	data := FetchDone{artists: []string{"A", "B"}, albumArtists: map[string]artist{"A": {}, "B": {}}}

	for _, statusFirst := range []bool{true, false} {
		l := New("", "local", testApp{tview.NewApplication()}, nil, testSpinner{})
		l.CreateContainer()
		l.SelectCpArtist()

		if statusFirst {
			l.MarkCpArtist("B")
			l.ApplyData(data)
		} else {
			l.ApplyData(data)
			l.MarkCpArtist("B")
		}

		if got := l.artistPane.GetCurrentItem(); got != 1 {
			t.Fatalf("statusFirst=%v: selected index %d, want 1", statusFirst, got)
		}

		l.MarkCpArtist("Missing")
		if l.cpArtistIdx != -1 {
			t.Fatal("missing artist retained the previous artist index")
		}
	}
}

func TestMarkCurrentlyPlayingTrack(t *testing.T) {
	l := New("", "local", testApp{tview.NewApplication()}, nil, testSpinner{})
	l.CreateContainer()
	l.ApplyData(FetchDone{
		artists: []string{"Artist", "Other"},
		albumArtists: map[string]artist{
			"Artist": {albums: []album{{name: "Album", year: 2020, tracks: []track{{name: "1. First"}, {name: "2. Second"}}}}},
			"Other":  {},
		},
	})
	l.MarkCpArtist("Artist")

	for _, tc := range []struct {
		track  string
		artist string
		first  string
		second string
	}{
		{"First", "Artist", "[yellow]1. First", "2. Second"},
		{"First", "Artist", "[yellow]1. First", "2. Second"},
		{"Second", "Artist", "1. First", "[yellow]2. Second"},
		{"First", "Other", "1. First", "2. Second"},
		{"", "", "1. First", "2. Second"},
	} {
		l.MarkCpTrack(tc.track, tc.artist, "Album")
		table := l.currentArtistAlbums[0]

		if first, second := table.GetCell(0, 0).Text, table.GetCell(1, 0).Text; first != tc.first || second != tc.second {
			t.Errorf("track %q by %q: got %q, %q; want %q, %q", tc.track, tc.artist, first, second, tc.first, tc.second)
		}
	}
}

func TestMarkCurrentlyPlayingTrackBeforeLibraryLoads(t *testing.T) {
	l := New("", "local", testApp{tview.NewApplication()}, nil, testSpinner{})
	l.CreateContainer()
	l.MarkCpTrack("Track", "Artist", "Album")
}
