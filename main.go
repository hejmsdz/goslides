package main

import (
	"fmt"
	"os"
)

func main() {
	manual, _ := GetManual(os.Getenv("NOTION_TOKEN"))

	songsDB := SongsDB{}
	songsDB.Initialize(os.Getenv("NOTION_TOKEN"))

	liturgyDB := LiturgyDB{}
	liturgyDB.Initialize()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	runServer(songsDB, liturgyDB, manual, fmt.Sprintf(":%s", port))
}
