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
	if tracker.idleLongEnough(now.Add(15*time.Minute)) || !tracker.idleLongEnough(now.Add(17*time.Minute)) {
		t.Fatal("idle threshold did not use the pause time")
	}
}
