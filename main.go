package main

import (
	"fmt"
	"os"

	"github.com/hejmsdz/goslides/database"
	"github.com/hejmsdz/goslides/di"
)

func main() {
	db := database.InitializeDB(os.Getenv("DATABASE"))
	container := di.NewContainer(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	server := NewServer(container)
	server.Run(fmt.Sprintf(":%s", port))
}
