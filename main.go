package main

import (
	"fmt"
	"os"
)

func main() {
	songsDB := SongsDB{}
	songsDB.Initialize()

	port := os.Getenv("PORT")
  if port == "" {
    port = "8000"
  }
  runServer(songsDB, fmt.Sprintf(":%s", port))
}
