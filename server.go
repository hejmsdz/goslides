package main

import (
	"encoding/json"
	"log"
	"net/http"
)

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

func runServer(songsDB SongsDB, addr string) {
	http.HandleFunc("/v2/songs", func(w http.ResponseWriter, req *http.Request) {
		getSongs(w, req, songsDB)
	})
	log.Printf("starting server on %s", addr)
	http.ListenAndServe(addr, nil)
}
