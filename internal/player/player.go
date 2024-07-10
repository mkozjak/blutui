package player

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/blutui/spinner"
)

type repeat struct {
	Mode string `xml:"repeat,attr"`
}

type volume struct {
	XMLName xml.Name `xml:"volume"`
	Value   int      `xml:",chardata"`
	Muted   string   `xml:"mute,attr"`
}

type Status struct {
	ETag     string `xml:"etag,attr"`
	Volume   int    `xml:"volume"`
	Album    string `xml:"album"`
	Artist   string `xml:"artist"`
	Track    string `xml:"name"`
	Title2   string `xml:"title2"`
	Title3   string `xml:"title3"`
	Format   string `xml:"streamFormat"`
	Quality  string `xml:"quality"`
	TrackLen int    `xml:"totlen"`
	Service  string `xml:"service"`
	Secs     int    `xml:"secs"`
	State    string `xml:"state"`
	Repeat   int    `xml:"repeat"`
}

type Controller interface {
	Play(url string)
	Playpause()
	Stop()
	Next()
	Previous()
	VolumeHold(bool)
	ToggleMute()
	ToggleRepeatMode()
	State() string
}

type Player struct {
	API               string
	Updates           chan<- Status
	spinner           spinner.Command
	status            Status
	volumeHoldCount   int
	volumeHoldBlocker bool
	volumeHoldTicker  *time.Ticker
	volumeHoldMutex   sync.Mutex
}

func New(api string, sp spinner.Command, s chan<- Status) *Player {
	return &Player{
		API:     api,
		Updates: s,
		spinner: sp,
	}
}

func (p *Player) State() string {
	return p.status.State
}

func (p *Player) Play(url string) {
	go p.spinner.Start()
	_, err := http.Get(p.API + url)
	if err != nil {
		internal.Log("Error autoplaying track:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}
	p.spinner.Stop()
}

func (p *Player) Playpause() {
	go p.spinner.Start()
	_, err := http.Get(p.API + "/Pause?toggle=1")
	if err != nil {
		internal.Log("Error toggling play/pause:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}
	p.spinner.Stop()
}

func (p *Player) Stop() {
	go p.spinner.Start()
	_, err := http.Get(p.API + "/Stop")
	if err != nil {
		internal.Log("Error stopping playback:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}
	p.spinner.Stop()
}

func (p *Player) Next() {
	go p.spinner.Start()
	_, err := http.Get(p.API + "/Skip")
	if err != nil {
		internal.Log("Error switching to next track:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}
	p.spinner.Stop()
}

func (p *Player) Previous() {
	go p.spinner.Start()
	_, err := http.Get(p.API + "/Back")
	if err != nil {
		internal.Log("Error switching to previous track:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}
	p.spinner.Stop()
}

func (p *Player) volumeUp(bigstep bool) {
	var step int
	if bigstep == true {
		step = 10
	} else {
		step = 3
	}

	v, _, err := p.currentVolume()
	if err != nil {
		internal.Log("Error fetching volume state:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}

	_, err = http.Get(fmt.Sprintf("%s/Volume?level=%d", p.API, v+step))
	if err != nil {
		internal.Log("Error setting volume up:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}
}

func (p *Player) volumeDown(bigstep bool) {
	var step int
	if bigstep == true {
		step = 10
	} else {
		step = 3
	}

	v, _, err := p.currentVolume()
	if err != nil {
		internal.Log("Error fetching volume state:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}

	_, err = http.Get(fmt.Sprintf("%s/Volume?level=%d", p.API, v-step))
	if err != nil {
		internal.Log("Error setting volume down:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}
}

func (p *Player) VolumeHold(up bool) {
	if p.volumeHoldBlocker == true {
		return
	}

	p.volumeHoldCount = p.volumeHoldCount + 1

	if p.volumeHoldTicker != nil {
		return
	}

	p.volumeHoldTicker = time.NewTicker(time.Second)
	done := make(chan bool)

	go func() {
		time.Sleep(500 * time.Millisecond)
		done <- true
	}()

	go p.spinner.Start()

	for {
		select {
		case <-done:
			p.volumeHoldTicker.Stop()
			p.volumeHoldTicker = nil

			close(done)

			if p.volumeHoldCount < 5 {
				if up == true {
					go p.volumeUp(false)
				} else {
					go p.volumeDown(false)
				}
			} else {
				if up == true {
					go p.volumeUp(true)
				} else {
					go p.volumeDown(true)
				}

				p.volumeHoldBlocker = true
				p.volumeHoldMutex.Lock()
				time.Sleep(5 * time.Second)
				p.volumeHoldBlocker = false
				p.volumeHoldMutex.Unlock()
			}

			p.volumeHoldCount = 0
			p.spinner.Stop()
			return
		}
	}
}

func (p *Player) ToggleMute() {
	go p.spinner.Start()
	_, m, err := p.currentVolume()
	if err != nil {
		internal.Log("Error getting mute state:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}

	if m == false {
		_, err = http.Get(p.API + "/Volume?mute=1")
	} else {
		_, err = http.Get(p.API + "/Volume?mute=0")
	}
	if err != nil {
		internal.Log("Error toggling mute state:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}
	go p.spinner.Stop()
}

// ToggleRepeatMode cycles between repeat modes in ascending order
// based on player's current repeat mode. Mode is either 0, 1 or 2.
// 0 means repeat play queue, 1 means repeat a track, and 2 means repeat off.
func (p *Player) ToggleRepeatMode() {
	go p.spinner.Start()
	r, err := p.currentRepeatMode()
	if err != nil {
		internal.Log("Error getting current repeat mode:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}

	switch r % 3 {
	case 0:
		_, err = http.Get(p.API + "/Repeat?state=1")
	case 1:
		_, err = http.Get(p.API + "/Repeat?state=2")
	case 2:
		_, err = http.Get(p.API + "/Repeat?state=0")
	}
	if err != nil {
		internal.Log("Error toggling repeat mode:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}
	go p.spinner.Stop()
}

func (p *Player) currentRepeatMode() (int, error) {
	resp, err := http.Get(p.API + "/Repeat")
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}

	var repeatRes repeat

	err = xml.Unmarshal(body, &repeatRes)
	if err != nil {
		return -1, err
	}

	m, err := strconv.Atoi(repeatRes.Mode)
	if err != nil {
		return -1, err
	}

	return m, nil
}

func (p *Player) currentVolume() (int, bool, error) {
	resp, err := http.Get(p.API + "/Volume")
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, false, err
	}

	var volRes volume

	err = xml.Unmarshal(body, &volRes)
	if err != nil {
		return 0, false, err
	}

	m, err := strconv.ParseBool(volRes.Muted)
	if err != nil {
		return 0, false, err
	}

	return volRes.Value, m, nil
}

func (p *Player) PollStatus() {
	etag := ""

	for {
		resp, err := http.Get(p.API + "/Status?timeout=60" + etag)
		if err != nil {
			uerr := url.Error{Err: err}
			var derr *net.DNSError

			if errors.As(err, &derr) || errors.Is(err, syscall.ECONNREFUSED) {
				s := Status{State: "neterr"}

				p.status = s
				p.Updates <- s
				continue
			}

			if uerr.Timeout() {
				continue
			}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		var s Status
		err = xml.Unmarshal(body, &s)
		if err != nil {
			continue
		}

		p.status = s
		p.Updates <- s
		etag = "&etag=" + s.ETag
		time.Sleep(5 * time.Second)
	}
}
