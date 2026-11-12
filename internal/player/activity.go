package player

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mkozjak/blutui/internal"
)

type activityRecord struct {
	LastActive time.Time `json:"last_active"`
	State      string    `json:"state"`
}

type activityTracker struct {
	mu             sync.Mutex
	path           string
	record         activityRecord
	initialUnknown bool
}

// EnableActivityTracking stores observations separately for each player.
func (p *Player) EnableActivityTracking() {
	dir, err := os.UserConfigDir()
	if err != nil {
		internal.Log("Error locating player activity:", err)
		return
	}
	id := sha256.Sum256([]byte(p.API))
	t := &activityTracker{path: filepath.Join(dir, "blutui", fmt.Sprintf("activity-%x.json", id[:8]))}
	data, err := os.ReadFile(t.path)
	if err == nil {
		if err = json.Unmarshal(data, &t.record); err != nil {
			internal.Log("Error reading player activity:", err)
			t.record = activityRecord{}
		}
	} else if !os.IsNotExist(err) {
		internal.Log("Error reading player activity:", err)
	}
	t.initialUnknown = t.record.LastActive.IsZero()
	p.activity = t
}

func (t *activityTracker) observe(state string, now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	active := state == "play" || state == "stream" || state == "connecting"
	wasActive := t.record.State == "play" || t.record.State == "stream" || t.record.State == "connecting"
	idle := state == "pause" || state == "stop"
	if !active && !idle {
		return
	}
	changedTime := active || t.record.LastActive.IsZero() || (idle && wasActive)
	if changedTime {
		t.record.LastActive = now
	}
	if !changedTime && t.record.State == state {
		return
	}
	t.record.State = state
	if err := t.save(); err != nil {
		internal.Log("Error saving player activity:", err)
	}
}

func (t *activityTracker) idleLongEnough(now time.Time) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return !t.record.LastActive.IsZero() && now.Sub(t.record.LastActive) > 15*time.Minute
}

// save is called with the tracker locked.
func (t *activityTracker) save() error {
	if err := os.MkdirAll(filepath.Dir(t.path), 0700); err != nil {
		return err
	}
	data, err := json.Marshal(t.record)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(t.path), ".activity-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), t.path)
}
