package main

import (
	"regexp"
	"strings"
	"time"

	"github.com/rainycape/unidecode"
)

type Song struct {
	Id         string `json:"id"`
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle,omitempty"`
	Number     string `json:"number"`
	Tags       string `json:"-"`
	Slug       string `json:"slug"`
	IsOrdinary bool   `json:"isOrdinary,omitempty"`

	numberChapter int
	numberItem    int
	updatedAt     time.Time
}

type GetLyricsOptions struct {
	Hints bool
	Raw   bool
	Order []int
}

var nonAlpha = regexp.MustCompile(`[^a-zA-Z0-9\. ]+`)
var verseName = regexp.MustCompile(`^\[(\w+)\]\s+`)
var verseRef = regexp.MustCompile(`^%(\w+)$`)

const ordinaryTag = "części stałe"

const subtitleSeparator = " / "
const commentSymbol = "//"
const lineBreakSymbol = " * "

const hintStartTag = "<hint>"
const hintEndTag = "</hint>"

type SongsDB interface {
	Reload() error
	LoadMissingVerses(songIDs []string) error
	GetLyrics(songID string, options GetLyricsOptions) ([]string, bool)
	FilterSongs(query string) []Song
}

func Slugify(text string) string {
	text = strings.ToLower(text)
	text = unidecode.Unidecode(text)
	text = nonAlpha.ReplaceAllString(text, "")
	text = strings.Trim(text, " ")

	return text
}

func FormatLyrics(verses []string, title string, number string, options GetLyricsOptions) []string {
	lyrics := make([]string, 0)
	namedVerses := make(map[string]string)
	if options.Hints {
		hint := number
		if utfTitle := []rune(title); hint == "" && len(utfTitle) >= 2 {
			hint = string(utfTitle[0:2])
		}
		lyrics = append(lyrics, hintStartTag+hint+hintEndTag)
	}

	var order []int
	if options.Order == nil {
		order = make([]int, len(verses))
		for i := range verses {
			order[i] = i
		}
	} else {
		order = options.Order
	}

	for _, index := range order {
		if index >= len(verses) {
			continue
		}
		verse := verses[index]

		if !options.Raw {
			if strings.HasPrefix(verse, commentSymbol) {
				if options.Order == nil {
					// if order is not given, skip commented out verses
					continue
				}

				// but if order is given, the verse is included deliberately and comment symbol should be ignored
				verse = strings.TrimPrefix(verse, commentSymbol)
				verse = strings.TrimLeft(verse, " ")
			}

			if match := verseRef.FindStringSubmatch(verse); match != nil {
				name := match[1]
				if namedVerse, ok := namedVerses[name]; ok {
					verse = namedVerse
				}
			} else {
				verse = strings.ReplaceAll(verse, lineBreakSymbol, "\n")
				if match := verseName.FindStringSubmatch(verse); match != nil {
					name := match[1]
					verse = verse[len(match[0]):]
					namedVerses[name] = verse
				}
			}
		}

		lyrics = append(lyrics, verse)
	}

	return lyrics
}
