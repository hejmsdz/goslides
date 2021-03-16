package main

type Deck struct {
	Items []DeckItem
}

type DeckItem struct {
	ID string
}

func (d Deck) ToTextSlides(songsDB SongsDB) ([][]string, bool) {
	songIDs := make([]string, 0)
	for _, item := range d.Items {
		songIDs = append(songIDs, item.ID)
	}
	songsDB.LoadMissingVerses(songIDs)

	slides := make([][]string, 0)

	for _, songID := range songIDs {
		lyrics, ok := songsDB.GetLyrics(songID)
		if !ok {
			return slides, false
		}
		slides = append(slides, lyrics)
	}

	return slides, true
}
