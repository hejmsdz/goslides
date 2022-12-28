package main

import (
	"fmt"
	"regexp"
	"strings"
)

type Deck struct {
	Date     string     `json:"date"`
	Items    []DeckItem `json:"items"`
	Hints    bool       `json:"hints"`
	Ratio    string     `json:"ratio"`
	FontSize int        `json:"fontSize"`
}

type DeckItem struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Contents []string `json:"contents"`
}

var dateRegexp = regexp.MustCompile(`^20\d\d-[0-1]\d-[0-3]\d$`)

func (d Deck) IsValid() bool {
	if !dateRegexp.MatchString(d.Date) {
		return false
	}

	if len(d.Items) == 0 {
		return false
	}

	return true
}

func (d Deck) GetPageConfig() PageConfig {
	ratio := 4.0 / 3.0
	fontSize := 36

	if d.Ratio == "16:9" {
		ratio = 16.0 / 9.0
	}

	if d.FontSize > 0 {
		fontSize = d.FontSize
	}

	pageWidth := 768.0

	return PageConfig{
		PageWidth:    pageWidth,
		PageHeight:   pageWidth / ratio,
		Margin:       50,
		FontSize:     fontSize,
		HintFontSize: fontSize * 2 / 3,
		LineSpacing:  1.3,
		Font:         "./fonts/source-sans-pro.ttf",
	}
}

func (d Deck) ToTextSlides(songsDB SongsDB, liturgyDB LiturgyDB) ([][]string, bool) {
	songIDs := make([]string, 0)
	hasLiturgy := false
	for _, item := range d.Items {
		if item.ID != "" {
			songIDs = append(songIDs, item.ID)
		}
		if item.Type == PSALM || item.Type == ACCLAMATION {
			hasLiturgy = true
		}
	}
	songsDB.LoadMissingVerses(songIDs)

	var liturgy Liturgy
	liturgyOk := true
	if hasLiturgy {
		liturgy, liturgyOk = liturgyDB.GetDay(d.Date)
	}

	slides := make([][]string, 0)
	for _, item := range d.Items {
		if item.ID != "" {
			lyrics, ok := songsDB.GetLyrics(item.ID, d.Hints)
			if !ok {
				return slides, false
			}
			slides = append(slides, lyrics)
		} else if item.Type == PSALM && liturgyOk {
			alleluiaticSuffix := ", albo: Alleluja"
			isAlleluiatic := strings.HasSuffix(liturgy.Psalm, alleluiaticSuffix)
			if isAlleluiatic {
				plainPsalm := strings.Replace(liturgy.Psalm, alleluiaticSuffix, "", 1)
				slides = append(slides, []string{plainPsalm, "Alleluja"})
			} else {
				slides = append(slides, []string{liturgy.Psalm})
			}
		} else if item.Type == ACCLAMATION && liturgyOk {
			fullAcclamation := fmt.Sprintf("%s\n\n%s\n\n%s",
				liturgy.Acclamation,
				liturgy.AcclamationVerse,
				liturgy.Acclamation)
			slides = append(slides, []string{fullAcclamation})
		} else if item.Contents != nil && len(item.Contents) > 0 {
			slides = append(slides, item.Contents)
		}
	}

	return slides, true
}
