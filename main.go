package main

import (
	"fmt"
	"os"

	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/models"
)

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
