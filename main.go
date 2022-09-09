package main

import (
	"fmt"
	"os"
)

var NOTION_TOKEN = os.Getenv("NOTION_TOKEN")
var NOTION_DB = os.Getenv("NOTION_DB")

func main() {
	manual, _ := GetManual(NOTION_TOKEN)

	songsDB := SongsDB{}
	songsDB.Initialize(NOTION_TOKEN, NOTION_DB)

	liturgyDB := LiturgyDB{}
	liturgyDB.Initialize()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	runServer(&songsDB, liturgyDB, manual, fmt.Sprintf(":%s", port))
}
