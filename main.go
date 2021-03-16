package main

import "fmt"

func main() {
	songsDB := SongsDB{}
	songsDB.Initialize()

	songs := songsDB.FilterSongs("rorate")
	songsDB.LoadMissingVerses([]string{songs[0].Id})

	lyrics, ok := songsDB.GetLyrics(songs[0].Id)
	if ok {
		fmt.Printf("all verses present\n")
	} else {
		fmt.Printf("some verses missing\n")
	}
	for i, verse := range lyrics {
		fmt.Printf("%d. %s\n\n", i+1, verse)
	}
}
