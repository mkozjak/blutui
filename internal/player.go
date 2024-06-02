package internal

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

var api string = "http://bluesound.local:11000"

func (a *App) Play(url string) {
	_, err := http.Get(api + url)
	if err != nil {
		log.Println("Error autoplaying track:", err)
		a.sbMessages <- Status{State: "ctrlerr"}
	}
}

func (a *App) Playpause() {
	_, err := http.Get(api + "/Pause?toggle=1")
	if err != nil {
		log.Println("Error toggling play/pause:", err)
		a.sbMessages <- Status{State: "ctrlerr"}
	}
}

func (a *App) Stop() {
	_, err := http.Get(api + "/Stop")
	if err != nil {
		log.Println("Error stopping playback:", err)
		a.sbMessages <- Status{State: "ctrlerr"}
	}
}

func (a *App) Next() {
	_, err := http.Get(api + "/Skip")
	if err != nil {
		log.Println("Error switching to next track:", err)
		a.sbMessages <- Status{State: "ctrlerr"}
	}
}

func (a *App) Previous() {
	_, err := http.Get(api + "/Back")
	if err != nil {
		log.Println("Error switching to previous track:", err)
		a.sbMessages <- Status{State: "ctrlerr"}
	}
}

func (a *App) VolumeUp() {
	v, err := a.currentVolume()
	if err != nil {
		log.Println("Error fetching volume state:", err)
		a.sbMessages <- Status{State: "ctrlerr"}
	}

	_, err = http.Get(fmt.Sprintf("%s/Volume?level=%d", api, v+3))
	if err != nil {
		log.Println("Error setting volume up:", err)
		a.sbMessages <- Status{State: "ctrlerr"}
	}
}

func (a *App) VolumeDown() {
	v, err := a.currentVolume()
	if err != nil {
		log.Println("Error fetching volume state:", err)
		a.sbMessages <- Status{State: "ctrlerr"}
	}

	_, err = http.Get(fmt.Sprintf("%s/Volume?level=%d", api, v-3))
	if err != nil {
		log.Println("Error setting volume down:", err)
		a.sbMessages <- Status{State: "ctrlerr"}
	}
}

func (a *App) currentVolume() (int, error) {
	resp, err := http.Get(api + "/Volume")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var volRes volume

	err = xml.Unmarshal(body, &volRes)
	if err != nil {
		return 0, err
	}

	return volRes.Value, nil
}

func (a *App) PollStatus() {
	etag := ""

	for {
		resp, err := http.Get(api + "/Status?timeout=60" + etag)
		if err != nil {
			uerr := url.Error{Err: err}
			var derr *net.DNSError

			if errors.As(err, &derr) {
				Log("dns error:", err)
				s := Status{State: "neterr"}

				a.sbMessages <- s
				continue
			}

			if uerr.Timeout() {
				Log("polling timeout")
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

		a.sbMessages <- s
		etag = "&etag=" + s.ETag
		time.Sleep(5 * time.Second)
	}
}
