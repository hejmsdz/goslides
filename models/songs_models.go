package models

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/hejmsdz/goslides/common"
	"github.com/hejmsdz/goslides/core"
	"gorm.io/gorm"
)

type Song struct {
	gorm.Model
	UUID     uuid.UUID `gorm:"uniqueIndex"`
	Title    string
	Subtitle sql.NullString
	Slug     string
	Lyrics   string
}

var verseName = regexp.MustCompile(`^\[(\w+)\]\s+`)
var verseRef = regexp.MustCompile(`^%(\w+)$`)

const commentSymbol = "//"
const lineBreakSymbol = " * "

func (s *Song) BeforeSave(tx *gorm.DB) (err error) {
	if s.UUID == uuid.Nil {
		s.UUID = uuid.New()
	}

	s.Slug = fmt.Sprintf("%s|%s", common.Slugify(s.Title), common.Slugify(s.Subtitle.String))
	s.Lyrics = strings.ReplaceAll(s.Lyrics, "\r\n", "\n")

	return nil
}

type FormatLyricsOptions struct {
	Raw   bool
	Hints bool
	Order []int
}

func (s Song) FormatLyrics(options FormatLyricsOptions) []string {
	verses := strings.Split(s.Lyrics, "\n\n")

	lyrics := make([]string, 0)
	namedVerses := make(map[string]string)

	if options.Hints {
		utfTitle := []rune(s.Title)
		if len(utfTitle) >= 2 {
			hint := string(utfTitle[0:3])
			lyrics = append(lyrics, core.HintStartTag+hint+core.HintEndTag)
		}
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
