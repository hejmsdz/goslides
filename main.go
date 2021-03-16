package main

func main() {
	songsDB := SongsDB{}
	songsDB.Initialize()

	runServer(songsDB, ":8000")
}
