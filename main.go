package main

import "fmt"

func main() {
	songsDB := SongsDB{}
	songsDB.Initialize()

	// runServer(songsDB, ":8000")

	deck := Deck{Items: []DeckItem{
		{"75564739-ca02-4447-a5db-061765b8930e"},
		{"4b6bd245-4c8b-452f-9064-9450758131e9"},
		{"f300dbae-5949-41db-984f-1853b1fdcc77"},
	}}

	textDeck, ok := deck.ToTextSlides(songsDB)
	if !ok {
		fmt.Println(":(")
	}
	pdf, err := BuildPDF(textDeck)
	if err != nil {
		fmt.Printf(":( %s", err)
		return
	}
	pdf.WritePdf("out.pdf")
}
