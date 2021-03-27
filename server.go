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

func allowCors(w *http.ResponseWriter, methods string) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	(*w).Header().Set("Access-Control-Allow-Methods", methods)
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

func getLiturgy(w http.ResponseWriter, req *http.Request, liturgyDB LiturgyDB) {
	date := req.URL.Query().Get("date")
	liturgy, ok := liturgyDB.GetDay(date)
	if !ok {
		http.Error(w, "liturgy error", http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(liturgy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func getManual(w http.ResponseWriter, req *http.Request, manual Manual) {
	resp, err := json.Marshal(manual)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func postDeck(w http.ResponseWriter, req *http.Request, songsDB SongsDB, liturgyDB LiturgyDB) {
	var deck Deck
	err := json.NewDecoder(req.Body).Decode(&deck)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !deck.IsValid() {
		http.Error(w, "Invalid input", http.StatusUnprocessableEntity)
		return
	}

	textDeck, ok := deck.ToTextSlides(songsDB, liturgyDB)
	if !ok {
		http.Error(w, "Failed to get lyrics", http.StatusInternalServerError)
		return
	}
	pdf, err := BuildPDF(textDeck)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pdfName := deck.Date + ".pdf"
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

func getBootstrap(w http.ResponseWriter, req *http.Request) {
	resp, err := json.Marshal(bootstrap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func postReload(w http.ResponseWriter, req *http.Request, songsDB *SongsDB) {
	err := songsDB.Initialize(NOTION_TOKEN)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func runServer(songsDB *SongsDB, liturgyDB LiturgyDB, manual Manual, addr string) {
	http.HandleFunc("/v2/songs", func(w http.ResponseWriter, req *http.Request) {
		allowCors(&w, "OPTIONS, GET")
		getSongs(w, req, *songsDB)
	})
	http.HandleFunc("/v2/liturgy", func(w http.ResponseWriter, req *http.Request) {
		allowCors(&w, "OPTIONS, GET")
		getLiturgy(w, req, liturgyDB)
	})
	http.HandleFunc("/v2/manual", func(w http.ResponseWriter, req *http.Request) {
		allowCors(&w, "OPTIONS, GET")
		getManual(w, req, manual)
	})
	http.HandleFunc("/v2/deck", func(w http.ResponseWriter, req *http.Request) {
		allowCors(&w, "OPTIONS, POST")
		if (*req).Method == "OPTIONS" {
			return
		}
		postDeck(w, req, *songsDB, liturgyDB)
	})
	http.HandleFunc("/v2/bootstrap", func(w http.ResponseWriter, req *http.Request) {
		allowCors(&w, "OPTIONS, GET")
		getBootstrap(w, req)
	})
	http.HandleFunc("/v2/reload", func(w http.ResponseWriter, req *http.Request) {
		allowCors(&w, "OPTIONS, POST")
		postReload(w, req, songsDB)
	})
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	log.Printf("starting server on %s", addr)
	http.ListenAndServe(addr, nil)
}
