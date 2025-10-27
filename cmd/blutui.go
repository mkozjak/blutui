package main

import (
	"flag"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/blutui/internal"
	"github.com/mkozjak/blutui/internal/app"
	"github.com/mkozjak/blutui/internal/bar"
	"github.com/mkozjak/blutui/internal/keyboard"
	"github.com/mkozjak/blutui/internal/library"
	"github.com/mkozjak/blutui/internal/player"
	"github.com/mkozjak/blutui/spinner"
	"github.com/mkozjak/tview"
)

var appVersion string
var proto string = "http"
var host string = "bluesound.lan"
var port string = "11000"
var bsUrl string = fmt.Sprintf("%s://%s:%s", proto, host, port)

func main() {
	// Define the version flag
	versionFlag := flag.Bool("version", false, "Display app version")
	flag.Parse()

	if *versionFlag {
		fmt.Println(appVersion)
		return
	}

	// Create main app
	a := app.New()
	sp := spinner.New(a.Draw)

	// Create Player and start http long-polling Bluesound for updates
	pUpd := make(chan player.Status)
	p := player.New(bsUrl, sp, pUpd)
	a.Player = p

	// Create Local Library Page
	lfc := make(chan library.FetchDone)
	lib := library.New(bsUrl, "local", a, p, sp)
	libc := lib.CreateContainer()

	// Start initial fetching of data
	go lib.FetchData(true, lfc)

	go func() {
		for {
			msg := <-lfc
			if msg.Error != nil {
				// TODO: should probably use os.Exit(1) here
				panic("failed fetching initial local data: " + msg.Error.Error())
			}

			// Draw initial album list for the first artist in the list
			lib.DrawArtistPane()
			lib.DrawInitAlbums()
			a.Draw()

			return
		}
	}()

	// Create Tidal Page
	tfc := make(chan library.FetchDone)
	tidal := library.New(bsUrl, "tidal", a, p, sp)
	tidalc := tidal.CreateContainer()

	go tidal.FetchData(true, tfc)

	go func() {
		for {
			msg := <-tfc
			if msg.Error != nil {
				// TODO: should probably use os.Exit(1) here
				panic("failed fetching initial tidal data: " + msg.Error.Error())
			}

			// Draw initial album list for the first artist in the list
			tidal.DrawArtistPane()
			tidal.DrawInitAlbums()

			return
		}
	}()

	a.Libs = map[string]*tview.Flex{
		"local": libc,
		"tidal": tidalc,
	}

	// Create a bottom Bar container along with its components
	b := bar.New(a, map[string]bar.LibManager{"local": lib, "tidal": tidal}, sp, pUpd)

	// Start listening for Player updates
	go p.PollStatus()

	a.Pages = tview.NewPages().
		AddAndSwitchToPage("local", libc, true).
		AddPage("tidal", tidalc, true, false)

	a.Pages.SetBackgroundColor(tcell.ColorDefault)

	a.Pages.SetChangedFunc(func() {
		n, _ := a.Pages.GetFrontPage()
		b.SetPageOnStatus(n)
	})

	// Configure global keybindings
	gk := keyboard.NewGlobalHandler(a, a.Player, lib, a.Pages, b)
	a.Application.SetInputCapture(gk.Listen)

	// Configure helpscreen keybindings
	// Attach helpscreen to the app
	hk := keyboard.NewHelpHandler(a.Pages)
	h := internal.CreateHelpScreen(hk.Listen)
	a.Pages.AddPage("help", h, false, false)

	// Draw root app window
	// Root consists of pages (local and tidal lib, etc.) and the status/bottom bar
	a.Root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.Pages, 0, 1, true).
		AddItem(b.StatusContainer(), 1, 0, false)

	// Set app root screen
	if err := a.Application.SetRoot(a.Root, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
