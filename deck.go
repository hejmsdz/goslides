package main

import (
	"fmt"
)

type Deck struct {
	Date  string     `json:"date"`
	Items []DeckItem `json:"items"`
}

type DeckItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
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
			lyrics, ok := songsDB.GetLyrics(item.ID)
			if !ok {
				return slides, false
			}
			slides = append(slides, lyrics)
		} else if item.Type == PSALM && liturgyOk {
			slides = append(slides, []string{liturgy.Psalm})
		} else if item.Type == ACCLAMATION && liturgyOk {
			fullAcclamation := fmt.Sprintf("%s\n\n%s\n\n%s",
				liturgy.Acclamation,
				liturgy.AcclamationVerse,
				liturgy.Acclamation)
			slides = append(slides, []string{fullAcclamation})
		}
	}

	return slides, true
}
