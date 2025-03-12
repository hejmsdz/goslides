package main

import (
	"fmt"
	"os"

	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/models"
)

var NOTION_TOKEN = os.Getenv("NOTION_TOKEN")
var NOTION_DB = os.Getenv("NOTION_DB")
var NOTION_MANUAL = os.Getenv("NOTION_MANUAL")

func main() {
	db := InitializeDB(os.Getenv("DATABASE"), []interface{}{models.Song{}, models.User{}})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	container := di.NewContainer(db)

	Server{
		container: container,
		addr:      fmt.Sprintf(":%s", port),
	}.Run()
}

/*
func main() {
	n := services.NotionSongsDB{}
	n.Initialize()

	db := InitializeDB("prod.db", []interface{}{models.Song{}})
	container := di.NewContainer(db)
	container.Songs.Import(n)
}
*/
