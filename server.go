package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type DeckResult struct {
	URL string `json:"url"`
}

func getPublicURL(req *http.Request, fileName string) string {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/public/%s", scheme, req.Host, fileName)
}

func getSongs(w http.ResponseWriter, req *http.Request, songsDB SongsDB) {
	query := req.URL.Query().Get("query")
	resp, err := json.Marshal(songsDB.FilterSongs(query))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func postDeck(w http.ResponseWriter, req *http.Request, songsDB SongsDB) {
	var deck Deck
	err := json.NewDecoder(req.Body).Decode(&deck)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	textDeck, ok := deck.ToTextSlides(songsDB)
	if !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pdf, err := BuildPDF(textDeck)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pdfName := "out.pdf"
	SaveTemporaryPDF(pdf, pdfName)

	deckResult := DeckResult{getPublicURL(req, pdfName)}
	resp, err := json.Marshal(deckResult)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func runServer(songsDB SongsDB, addr string) {
	http.HandleFunc("/v2/songs", func(w http.ResponseWriter, req *http.Request) {
		getSongs(w, req, songsDB)
	})
	http.HandleFunc("/v2/deck", func(w http.ResponseWriter, req *http.Request) {
		postDeck(w, req, songsDB)
	})
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	log.Printf("starting server on %s", addr)
	http.ListenAndServe(addr, nil)
}
