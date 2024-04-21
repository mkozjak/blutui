package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var api string = "http://bluesound.local:11000"

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type track struct {
	name   string
	length int
	disc   int
	number int
}

type album struct {
	name   string
	year   int
	genre  string
	tracks []track
}

type artist struct {
	albums []album
}

type model struct {
	table            table.Model
	albumArtists     map[string]artist
	artists          []string
	currentlyPlaying track
	status           string
	cursor           int
	lines            int
}

type browse struct {
	Items []item `xml:"item"`
}

type item struct {
	Text      string `xml:"text,attr"`  // album name
	Text2     string `xml:"text2,attr"` // artist name
	BrowseKey string `xml:"browseKey,attr"`
	Type      string `xml:"type,attr"`
	PlayURL   string `xml:"playURL,attr"`
}

func fetchData() tea.Msg {
	albumArtists := map[string]artist{}
	albumSectionsEndp := api + "/Browse?key=LocalMusic%3AbySection%2F%252FAlbums%253Fservice%253DLocalMusic"

	resp, err := http.Get(albumSectionsEndp)
	if err != nil {
		fmt.Println("Error fetching album section list:", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	var sections browse
	err = xml.Unmarshal(body, &sections)
	if err != nil {
		fmt.Println("Error parsing the sections XML:", err)
		return nil
	}

	// parse album sections (alphabetical order) from xml
	for _, item := range sections.Items {
		resp, err = http.Get(api + "/Browse?key=" + url.QueryEscape(item.BrowseKey))
		if err != nil {
			fmt.Println("Error fetching album section:", err)
			return nil
		}
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return nil
		}

		var albums browse
		err = xml.Unmarshal(body, &albums)
		if err != nil {
			fmt.Println("Error parsing the albums XML:", err)
			return nil
		}

		// iterate albums and fill albumArtists
		for _, al := range albums.Items {
			// TODO: fetch album tracks here
			// /Browse?key=$browseKey (album.BrowseKey)

			ar, ok := albumArtists[al.Text2]
			if ok {
				ar.albums = append(ar.albums, album{
					name: al.Text,
				})

				albumArtists[al.Text2] = ar
			} else {
				albumArtists[al.Text2] = artist{
					albums: []album{{
						name: al.Text,
					}},
				}
			}
		}
	}

	n := SortArtists(albumArtists)

	// Iterate over sorted artist names
	for _, artistName := range n {
		ar := albumArtists[artistName]

		// Sort albums alphabetically
		sort.Slice(ar.albums, func(i, j int) bool {
			// FIXME: should sort by year instead
			return ar.albums[i].name < ar.albums[j].name
		})

		albumArtists[artistName] = ar
	}

	return albumArtists
}

func (m model) RenderAlbums() error {
	rows := m.table.Rows()

	// init view
	// if len(rows) < 1 {
	// 	names := SortArtists(m.albumArtists)

	// 	for _, n := range names {
	// 		rows = append(rows, table.Row{n})
	// 	}
	// } else {
	// 	// reset all artist rows
	// 	for i, r := range rows {
	// 		if len(r) > 1 {
	// 			rows[i] = table.Row{r[0]}
	// 		}
	// 	}
	// }

	// reset all artist rows
	for i, r := range rows {
		if len(r) > 1 {
			rows[i] = table.Row{r[0]}
		}
	}

	artist := m.table.SelectedRow()[0]

	// render current artist's albums
	for i := 0; i < len(m.albumArtists[artist].albums); i++ {
		rows[i] = append(rows[i], m.albumArtists[artist].albums[i].name)
	}

	m.table.SetRows(rows)
	return nil
}

func (m model) Init() tea.Cmd {
	return fetchData
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width - 2)
		m.table.SetHeight(msg.Height - 10)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "j":
			m.table.MoveDown(1)
			m.RenderAlbums()
			log.Println(m.table)
			return m, nil
		case "k":
			m.table.MoveUp(1)
			m.RenderAlbums()
			return m, nil
		}

	// new data from api received
	case map[string]artist:
		// TODO: should use m.RenderAlbums here?
		m.albumArtists = msg
		r := []table.Row{}

		names := SortArtists(m.albumArtists)

		for _, n := range names {
			r = append(r, table.Row{n})
		}

		for _, album := range m.albumArtists[names[0]].albums {
			r[0] = append(r[0], album.name)
		}

		m.table.SetRows(r)
		return m, nil

	default:
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func main() {
	m := model{
		status: "stopped",
	}

	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		log.Println("error getting terminal size:", err)
		os.Exit(1)
	}

	columns := []table.Column{
		{Title: "Artist", Width: w / 3},
		{Title: "Track", Width: int(float32(w) * 0.6)},
	}

	rows := []table.Row{}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(h),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m.table = t

	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("/tmp/debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}

		defer f.Close()
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Printf("A startup error occurred: %v", err)
		os.Exit(1)
	}
}
