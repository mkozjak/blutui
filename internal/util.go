package internal

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	mu     sync.Mutex
	logger *os.File
)

func init() {
	var err error
	logger, err = os.OpenFile("/tmp/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
}

func Log(data ...interface{}) error {
	mu.Lock()
	defer mu.Unlock()

	for _, datum := range data {
		_, err := fmt.Fprintf(logger, "%v ", datum)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(logger) // Add newline
	return err
}

func FormatDuration(d int) string {
	m := d / 60
	s := d % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func CapitalizeArtist(s string) string {
	if len(s) == 0 {
		return s
	}

	// Decode the first rune in the string
	firstRune, size := utf8.DecodeRuneInString(s)
	if firstRune == utf8.RuneError {
		return s
	}

	// Capitalize the first rune if it's a letter
	firstRune = unicode.ToUpper(firstRune)

	// Combine the capitalized first rune with the rest of the string
	return string(firstRune) + s[size:]

}

func Caser(s string) string {
	var res string

	for _, c := range []cases.Caser{cases.Title(language.English)} {
		res = c.String(s)
	}

	return res
}

func ExtractAlbumYear(y string) (int, error) {
	t, err := time.Parse("2006", y)
	if err == nil {
		return t.Year(), nil
	}

	t, err = time.Parse("2006-01-02", y)
	if err == nil {
		return t.Year(), nil
	}

	return 0, errors.New("invalid date format")
}

// HackAlbumYear extracts year from ContextMenuKey URL.
// This function is used as a fallback way of getting album's
// release year, because the API sometimes doesn't return it.
// It searches key for substrings such as [2002] or (2002) and
// used the last occurence of it to extract the year.
func HackAlbumYear(key string) (int, error) {
	u, err := url.Parse(key)
	if err != nil {
		return 0, errors.New("failed parsing context url")
	}

	q := u.Query()
	for k, v := range q {
		if k != "filename" {
			continue
		}

		r, _ := regexp.Compile(`\[(\d{4})\]|\((\d{4})\)`)
		matches := r.FindStringSubmatch(v[0])

		if matches != nil && len(matches) > 1 {
			y, err := strconv.Atoi(matches[1])
			if err != nil {
				return 0, err
			}

			return y, nil
		}
	}

	return 0, errors.New("year could not be found")
}

// EscapeStyleTag disables tview style tagging when, for example,
// literal square brackets are needed to be printed.
// In example, album name "Chilombo [clean]" should be printed as is.
// To enable this, a closing square bracket needs to be prepended by an
// opening square bracket. This will result in "Chilombo [clean[]".
func EscapeStyleTag(s string) string {
	return strings.Replace(s, "]", "[]", 1)
}

// CleanAlbumName removes all tview style tags and album year annotations
// that are suffixed to the album name.
func CleanAlbumName(s string) string {
	s = strings.Replace(s, "[::b]", "", 1)
	s = strings.Replace(s, "[]", "]", 1)

	r, _ := regexp.Compile(` \(\d{4}\)$`)
	s = r.ReplaceAllString(s, "")

	return s
}

// CleanTrackName removes prefixes such as track numbers.
func CleanTrackName(n string) string {
	re := regexp.MustCompile(`^\d+\.\s`)
	return re.ReplaceAllString(n, "")
}

// JWSimilarity is the implementation of Jaro-Winkler similarity metric
// It returns 1 if there's a 100% match and 0% if there's no matching characters.
func JWSimilarity(s1, s2 string) float64 {
	if s1 == "" || s2 == "" {
		return 0
	}

	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	// input strings are exactly the same
	if s1 == s2 {
		return 1
	}

	ls1 := float64(len(s1))
	ls2 := float64(len(s2))
	dMax := math.Floor(max(ls1, ls2)/2) - 1

	var matches []string

	// Count the number of matching characters
	// faremviel, farmville
	for i, s1Char := range s1 {
		for j, s2Char := range s2 {
			// If the characters match and are not farther than dMax
			if s1Char == s2Char && math.Abs(float64(i)-float64(j)) <= dMax {
				matches = append(matches, string(s1Char))
				break
			}
		}
	}

	// No matching characters found, Jaro similarity score is 0
	m := float64(len(matches))
	if m == 0 {
		return 0
	}

	// Now find the number of transpositions
	var trsp []string

	for i, mc := range matches {
		for j, s2Char := range s2 {
			if mc != string(s2Char) || i == j {
				continue
			}

			// If the previous matching character is equal, don't count it in
			// TODO: optimize if/else
			l := len(trsp)
			if l == 1 {
				if trsp[0] == mc {
					continue
				}
			} else if l > 1 {
				if trsp[l-1] == mc {
					continue
				}
			}

			trsp = append(trsp, mc)
		}
	}

	t := float64(len(trsp) / 2)

	jaro := roundFloat((m/ls1+m/ls2+(m-t)/m)/3, 5)

	// Distance threshold not met
	if jaro < 0.75 {
		return 0
	}

	// Now calculate Jaro-Winkler similarity
	var l float64 = 0
	var p float64 = 0.1

	for i := 0; i < len(s1) && i < len(s2); i++ {
		if s1[i] == s2[i] {
			l++
		}
	}

	sim := jaro + l*p*(1-jaro)

	return roundFloat(sim, 5)
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
