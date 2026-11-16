package player

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestActivityPersistsIdleTransition(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	tracker := &activityTracker{path: filepath.Join(t.TempDir(), "activity.json")}
	tracker.observe("play", now)
	tracker.observe("pause", now.Add(time.Minute))
	data, err := os.ReadFile(tracker.path)
	if err != nil {
		t.Fatal(err)
	}
	var record activityRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	if record.State != "pause" || !record.LastActive.Equal(now.Add(time.Minute)) {
		t.Fatalf("unexpected persisted activity: %+v", record)
	}
}

func TestStartupWakeUsesLoadedHistory(t *testing.T) {
	now := time.Now()
	for _, tc := range []struct {
		name string
		last time.Time
		want bool
	}{
		{"unknown", time.Time{}, true},
		{"recent", now.Add(-14 * time.Minute), false},
		{"old", now.Add(-16 * time.Minute), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tracker := &activityTracker{path: filepath.Join(t.TempDir(), "activity"), initialLastActive: tc.last, record: activityRecord{LastActive: tc.last, State: "play"}}
			tracker.observe("stop", now)
			if got := tracker.shouldWakeOnStartup(now); got != tc.want {
				t.Fatalf("shouldWakeOnStartup = %v, want %v", got, tc.want)
			}
		})
	}
}
