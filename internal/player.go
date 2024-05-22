package internal

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var api string = "http://bluesound.local:11000"

func Play(url string) {
	_, err := http.Get(api + url)
	if err != nil {
		log.Println("Error autoplaying track:", err)
		panic(err)
	}

	// arLst.SetItemText(arLst.GetCurrentItem(), "[yellow]"+artist, "")
	// trackLst.SetItemText(i, "[yellow]"+name, "")
}

func Playpause() {
	_, err := http.Get(api + "/Pause?toggle=1")
	if err != nil {
		log.Println("Error toggling play/pause:", err)
		panic(err)
	}
}

func Stop() {
	_, err := http.Get(api + "/Stop")
	if err != nil {
		log.Println("Error stopping playback:", err)
		panic(err)
	}
}

func Next() {
	_, err := http.Get(api + "/Skip")
	if err != nil {
		log.Println("Error switch to next track:", err)
		panic(err)
	}
}

func Previous() {
	_, err := http.Get(api + "/Back")
	if err != nil {
		log.Println("Error switch to previous track:", err)
		panic(err)
	}
}

func VolumeUp() {
	v, err := currentVolume()
	if err != nil {
		log.Println("Error fetching volume state:", err)
		panic(err)
	}

	_, err = http.Get(fmt.Sprintf("%s/Volume?level=%d", api, v+5))
	if err != nil {
		log.Println("Error setting volume up:", err)
		panic(err)
	}
}

func VolumeDown() {
	v, err := currentVolume()
	if err != nil {
		log.Println("Error fetching volume state:", err)
		panic(err)
	}

	_, err = http.Get(fmt.Sprintf("%s/Volume?level=%d", api, v-5))
	if err != nil {
		log.Println("Error setting volume down:", err)
		panic(err)
	}
}

func currentVolume() (int, error) {
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

func (a *App) PollStatus(ch chan<- Status) {
	for {
		var s Status
		s.Volume = 20

		ch <- s
		time.Sleep(5 * time.Second)
	}
}
