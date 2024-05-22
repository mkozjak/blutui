package main

import (
	"log"
	"net/http"
)

func play(url string) {
	_, err := http.Get(api + url)
	if err != nil {
		log.Println("Error autoplaying track:", err)
		panic(err)
	}

	// arLst.SetItemText(arLst.GetCurrentItem(), "[yellow]"+artist, "")
	// trackLst.SetItemText(i, "[yellow]"+name, "")
}

func playpause() {
	_, err := http.Get(api + "/Pause?toggle=1")
	if err != nil {
		log.Println("Error toggling play/pause:", err)
		panic(err)
	}
}

func stop() {
	_, err := http.Get(api + "/Stop")
	if err != nil {
		log.Println("Error stopping playback:", err)
		panic(err)
	}
}

func next() {
	_, err := http.Get(api + "/Skip")
	if err != nil {
		log.Println("Error switch to next track:", err)
		panic(err)
	}
}

func previous() {
	_, err := http.Get(api + "/Back")
	if err != nil {
		log.Println("Error switch to previous track:", err)
		panic(err)
	}
}
