package main

import (
	"fmt"
	"os"
)

var NOTION_TOKEN = os.Getenv("NOTION_TOKEN")
var NOTION_DB = os.Getenv("NOTION_DB")
var NOTION_MANUAL = os.Getenv("NOTION_MANUAL")

func main() {
	manual, _ := GetManual(NOTION_TOKEN, NOTION_MANUAL)

	songsDB := SQLSongsDB{}
	songsDB.Initialize("prod.db")

	liturgyDB := LiturgyDB{}
	liturgyDB.Initialize()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	Server{
		songsDB:   songsDB,
		liturgyDB: liturgyDB,
		manual:    manual,
		addr:      fmt.Sprintf(":%s", port),
	}.Run()
}

/*
func main() {
	n := NotionSongsDB{authToken: NOTION_TOKEN, databaseId: NOTION_DB}
	n.Initialize()

	songsDB := SqlSongsDB{}
	songsDB.Initialize()

	songsDB.Import(n)
}
*/
