package main

import (
	"log"
	"net/http"
)

func play(url string) {
	go func() {
		_, err := http.Get(api + url)
		if err != nil {
			log.Println("Error autoplaying track:", err)
			panic(err)
		}

		// arLst.SetItemText(arLst.GetCurrentItem(), "[yellow]"+artist, "")
		// trackLst.SetItemText(i, "[yellow]"+name, "")
	}()
}

func stop() {
	go func() {
		_, err := http.Get(api + "/Stop")
		if err != nil {
			log.Println("Error stopping playback:", err)
			panic(err)
		}
	}()
}

func next() {
	go func() {
		_, err := http.Get(api + "/Skip")
		if err != nil {
			log.Println("Error switch to next track:", err)
			panic(err)
		}
	}()
}

func previous() {
	go func() {
		_, err := http.Get(api + "/Back")
		if err != nil {
			log.Println("Error switch to previous track:", err)
			panic(err)
		}
	}()
}
