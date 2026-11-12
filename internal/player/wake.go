package player

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mkozjak/blutui/internal"
)

const wakeStation = "RadioParadise:/0:4"

type wakeStatus struct {
	State     string `xml:"state"`
	StreamURL string `xml:"streamUrl"`
	Mute      string `xml:"mute"`
}

// WakeOnStartup briefly plays a radio stream when the player is idle.
func (p *Player) WakeOnStartup() {
	if err := p.wakeOnStartup(&http.Client{Timeout: 5 * time.Second}, time.Second); err != nil {
		internal.Log("Error waking player:", err)
	}
}

func (p *Player) wakeOnStartup(client *http.Client, delay time.Duration) (result error) {
	status, err := p.wakeStatus(client)
	if err != nil || (status.State != "pause" && status.State != "stop") {
		return err
	}
	if p.activity == nil {
		return nil
	}
	p.activity.observe(status.State, time.Now())
	if !p.activity.initialUnknown && !p.activity.idleLongEnough(time.Now()) {
		return nil
	}
	if status.Mute != "0" && status.Mute != "1" {
		return fmt.Errorf("wake status has no valid mute state")
	}
	if status.Mute == "0" {
		defer func() {
			if err := p.wakeRequest(client, "/Volume?mute=0"); err != nil {
				result = errors.Join(result, fmt.Errorf("restoring mute state: %w", err))
			}
		}()
		if err := p.wakeRequest(client, "/Volume?mute=1"); err != nil {
			return err
		}
	}
	// Another controller may have started playback while mute was set.
	status, err = p.wakeStatus(client)
	if err != nil || (status.State != "pause" && status.State != "stop") {
		return err
	}
	if err := p.wakeRequest(client, "/Play?"+url.Values{"url": {wakeStation}}.Encode()); err != nil {
		return err
	}
	time.Sleep(delay)
	// The player can report the previous idle state while the stream connects.
	for attempt := 0; attempt < 6; attempt++ {
		status, err = p.wakeStatus(client)
		if err != nil {
			return err
		}
		stream, unescapeErr := url.QueryUnescape(status.StreamURL)
		if unescapeErr == nil && stream == wakeStation && (status.State == "stream" || status.State == "play") {
			return p.wakeRequest(client, "/Stop")
		}
		if stream == wakeStation && status.State == "pause" {
			return nil
		}
		if status.State != "connecting" && status.State != "stop" && status.State != "pause" {
			return nil
		}
		if attempt < 5 {
			time.Sleep(time.Second)
		}
	}
	return nil
}

func (p *Player) wakeStatus(client *http.Client) (wakeStatus, error) {
	var status wakeStatus
	resp, err := client.Get(p.API + "/Status")
	if err != nil {
		return status, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return status, fmt.Errorf("fetching wake status: %s", resp.Status)
	}
	if err := xml.NewDecoder(resp.Body).Decode(&status); err != nil {
		return status, err
	}
	if strings.TrimSpace(status.State) == "" {
		return status, fmt.Errorf("wake status has no playback state")
	}
	return status, nil
}

func (p *Player) wakeRequest(client *http.Client, path string) error {
	resp, err := client.Get(p.API + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wake request %s: %s", path, resp.Status)
	}
	return nil
}
