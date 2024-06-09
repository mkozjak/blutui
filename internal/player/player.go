package player

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

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
}

type Player struct {
	API               string
	Updates           chan<- Status
	status            Status
	volumeHoldCount   int
	volumeHoldBlocker bool
	volumeHoldTicker  *time.Ticker
	volumeHoldMutex   sync.Mutex
}

func NewPlayer(api string, s chan<- Status) *Player {
	return &Player{
		API:     api,
		Updates: s,
	}
}

func (p *Player) GetState() string {
	return p.status.State
}

func (p *Player) Play(url string) {
	_, err := http.Get(p.API + url)
	if err != nil {
		log.Println("Error autoplaying track:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}
}

func (p *Player) Playpause() {
	_, err := http.Get(p.API + "/Pause?toggle=1")
	if err != nil {
		log.Println("Error toggling play/pause:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}
}

func (p *Player) Stop() {
	_, err := http.Get(p.API + "/Stop")
	if err != nil {
		log.Println("Error stopping playback:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}
}

func (p *Player) Next() {
	_, err := http.Get(p.API + "/Skip")
	if err != nil {
		log.Println("Error switching to next track:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}
}

func (p *Player) Previous() {
	_, err := http.Get(p.API + "/Back")
	if err != nil {
		log.Println("Error switching to previous track:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}
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
		log.Println("Error fetching volume state:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}

	_, err = http.Get(fmt.Sprintf("%s/Volume?level=%d", p.API, v+step))
	if err != nil {
		log.Println("Error setting volume up:", err)
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
		log.Println("Error fetching volume state:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}

	_, err = http.Get(fmt.Sprintf("%s/Volume?level=%d", p.API, v-step))
	if err != nil {
		log.Println("Error setting volume down:", err)
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
			return
		}
	}
}

func (p *Player) ToggleMute() {
	_, m, err := p.currentVolume()
	if err != nil {
		log.Println("Error getting mute state:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}

	if m == false {
		_, err = http.Get(p.API + "/Volume?mute=1")
	} else {
		_, err = http.Get(p.API + "/Volume?mute=0")
	}
	if err != nil {
		log.Println("Error toggling mute state:", err)
		p.Updates <- Status{State: "ctrlerr"}
	}
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

			if errors.As(err, &derr) {
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
