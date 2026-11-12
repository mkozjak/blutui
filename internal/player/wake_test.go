package player

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWakeOnStartup(t *testing.T) {
	for _, tc := range []struct {
		name, initial, after string
		wantPlay, wantStop   bool
	}{
		{"paused", "pause", `<state>stream</state><streamUrl>RadioParadise:/0:4</streamUrl>`, true, true},
		{"stopped", "stop", `<state>stream</state><streamUrl>RadioParadise:/0:4</streamUrl>`, true, true},
		{"playing", "play", "", false, false},
		{"streaming", "stream", "", false, false},
		{"connecting", "connecting", "", false, false},
		{"changed source", "pause", `<state>play</state><streamUrl>other</streamUrl>`, true, false},
		{"user paused", "pause", `<state>pause</state><streamUrl>RadioParadise:/0:4</streamUrl>`, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var statuses int
			var played, stopped bool
			var muteCommands []string
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/Status":
					statuses++
					if statuses <= 2 {
						fmt.Fprintf(w, "<status><state>%s</state><mute>0</mute></status>", tc.initial)
					} else {
						fmt.Fprintf(w, "<status>%s</status>", tc.after)
					}
				case "/Play":
					played = r.URL.Query().Get("url") == wakeStation
				case "/Stop":
					stopped = true
				case "/Volume":
					muteCommands = append(muteCommands, r.URL.Query().Get("mute"))
				default:
					t.Errorf("unexpected request: %s", r.URL)
				}
			}))
			defer s.Close()
			p := New(s.URL, nil, nil)
			p.activity = &activityTracker{path: filepath.Join(t.TempDir(), "activity"), record: activityRecord{State: tc.initial, LastActive: time.Now().Add(-16 * time.Minute)}}
			if err := p.wakeOnStartup(s.Client(), 0); err != nil {
				t.Fatal(err)
			}
			if played != tc.wantPlay || stopped != tc.wantStop {
				t.Fatalf("played=%v stopped=%v; want %v, %v", played, stopped, tc.wantPlay, tc.wantStop)
			}
			if tc.wantPlay && strings.Join(muteCommands, ",") != "1,0" {
				t.Fatalf("mute commands = %v", muteCommands)
			}
		})
	}
}

func TestWakeIdleThresholdAndMute(t *testing.T) {
	for _, tc := range []struct {
		name     string
		last     time.Time
		mute     string
		wantPlay bool
		wantMute string
	}{
		{"unknown", time.Time{}, "0", true, "1,0"},
		{"recent", time.Now().Add(-14 * time.Minute), "0", false, ""},
		{"old muted", time.Now().Add(-16 * time.Minute), "1", true, ""},
		{"old unmuted", time.Now().Add(-16 * time.Minute), "0", true, "1,0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			playing := false
			var mutes []string
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/Status":
					if playing {
						fmt.Fprint(w, `<status><state>stream</state><streamUrl>RadioParadise:/0:4</streamUrl></status>`)
					} else {
						fmt.Fprintf(w, "<status><state>stop</state><mute>%s</mute></status>", tc.mute)
					}
				case "/Play":
					playing = true
				case "/Volume":
					mutes = append(mutes, r.URL.Query().Get("mute"))
				case "/Stop":
				default:
					t.Errorf("unexpected request: %s", r.URL)
				}
			}))
			defer s.Close()
			p := New(s.URL, nil, nil)
			p.activity = &activityTracker{path: filepath.Join(t.TempDir(), "activity"), record: activityRecord{State: "stop", LastActive: tc.last}, initialUnknown: tc.last.IsZero()}
			if err := p.wakeOnStartup(s.Client(), 0); err != nil {
				t.Fatal(err)
			}
			if playing != tc.wantPlay || strings.Join(mutes, ",") != tc.wantMute {
				t.Fatalf("played=%v mutes=%v", playing, mutes)
			}
		})
	}
}

func TestWakeDoesNotPlayOnInvalidStatus(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Status" {
			t.Errorf("unexpected request: %s", r.URL)
		}
		fmt.Fprint(w, "<status/>")
	}))
	defer s.Close()
	err := New(s.URL, nil, nil).wakeOnStartup(s.Client(), 0)
	if err == nil || !strings.Contains(err.Error(), "no playback state") {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}

func TestWakeRestoresMuteWhenPlaybackFails(t *testing.T) {
	var mutes []string
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Status":
			fmt.Fprint(w, `<status><state>stop</state><mute>0</mute></status>`)
		case "/Volume":
			mutes = append(mutes, r.URL.Query().Get("mute"))
		case "/Play":
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
		}
	}))
	defer s.Close()
	p := New(s.URL, nil, nil)
	p.activity = &activityTracker{path: filepath.Join(t.TempDir(), "activity"), record: activityRecord{State: "stop", LastActive: time.Now().Add(-time.Hour)}}
	if err := p.wakeOnStartup(s.Client(), 0); err == nil {
		t.Fatal("expected playback error")
	}
	if strings.Join(mutes, ",") != "1,0" {
		t.Fatalf("mute commands = %v", mutes)
	}
}
