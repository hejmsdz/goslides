package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type DeckResult struct {
	URL      string         `json:"url"`
	Contents []ContentSlide `json:"contents"`
}

func getURL(c *gin.Context, path string) string {
	scheme := "https"
	return fmt.Sprintf("%s://%s/%s", scheme, c.Request.Host, path)
}

func getPublicURL(c *gin.Context, fileName string) string {
	return getURL(c, fmt.Sprintf("public/%s", fileName))
}

func getRandomString(length int) string {
	buffer := make([]byte, length)
	rand.Read(buffer)
	return fmt.Sprintf("%x", buffer)
}

func corsMiddleware(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}

	c.Next()
}

type Server struct {
	songsDB   SongsDB
	liturgyDB LiturgyDB
	manual    Manual
	addr      string
}

func (srv Server) getBootstrap(c *gin.Context) {
	CheckCurrentVersion()
	c.JSON(http.StatusOK, bootstrap)
}

func (srv Server) getSongs(c *gin.Context) {
	query := c.Query("query")
	songs := srv.songsDB.FilterSongs(query)
	c.JSON(http.StatusOK, songs)
}

func (srv Server) postSong(c *gin.Context) {
	var input SongInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !input.IsValid() {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid data"})
		return
	}

	id, err := srv.songsDB.CreateSong(input)
	if err == nil {
		c.JSON(http.StatusCreated, gin.H{"id": id})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (srv Server) getSong(c *gin.Context) {
	id := c.Param("id")
	srv.songsDB.LoadMissingVerses([]string{id})
	song, ok := srv.songsDB.GetSong(id)

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song ID not found"})
		return
	}

	c.JSON(http.StatusOK, song)
}

func (srv Server) patchSong(c *gin.Context) {
	id := c.Param("id")
	var input SongInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !input.IsValid() {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid data"})
		return
	}

	err := srv.songsDB.UpdateSong(id, input)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (srv Server) deleteSong(c *gin.Context) {
	id := c.Param("id")

	err := srv.songsDB.DeleteSong(id)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (srv Server) getLyrics(c *gin.Context) {
	id := c.Param("id")
	raw := c.Query("raw") == "1"
	srv.songsDB.LoadMissingVerses([]string{id})
	lyrics, _ := srv.songsDB.GetLyrics(id, GetLyricsOptions{Hints: false, Raw: raw})

	if lyrics == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song ID not found"})
		return
	}

	c.JSON(http.StatusOK, lyrics)
}

func (srv Server) getLiturgy(c *gin.Context) {
	date := c.Param("date")
	liturgy, ok := srv.liturgyDB.GetDay(date)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Liturgy error"})
		return
	}

	c.JSON(http.StatusOK, liturgy)
}

func (srv Server) getManual(c *gin.Context) {
	c.JSON(http.StatusOK, srv.manual)
}

func (srv Server) postDeck(c *gin.Context) {
	var deck Deck
	if err := c.ShouldBind(&deck); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !deck.IsValid() {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid deck data"})
		return
	}

	textDeck, ok := deck.ToTextSlides(srv.songsDB, srv.liturgyDB)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get lyrics"})
		return
	}

	var deckResult DeckResult
	uid := getRandomString(6)
	switch deck.Format {
	case "png+zip":
		zip, contents, err := BuildImages(textDeck, deck.GetPageConfig())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		zipName := uid + ".zip"
		SaveTemporaryFile(zip, zipName)
		deckResult = DeckResult{getPublicURL(c, zipName), contents}

	case "txt":
		text := Tugalize(textDeck)

		txtName := uid + ".txt"
		textReader := strings.NewReader(text)
		SaveTemporaryFile(textReader, txtName)
		deckResult = DeckResult{getPublicURL(c, txtName), []ContentSlide{}}

	default:
		pdf, contents, err := BuildPDF(textDeck, deck.GetPageConfig())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		pdfName := uid + ".pdf"
		SaveTemporaryPDF(pdf, pdfName)

		deckResult = DeckResult{getPublicURL(c, pdfName), contents}
	}

	if !deck.Contents {
		deckResult.Contents = nil
	}

	c.JSON(http.StatusOK, deckResult)
}

func (srv Server) postReload(c *gin.Context) {
	err := srv.songsDB.Reload()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Writer.WriteHeader(http.StatusNoContent)
}

func (srv Server) postUpdateRelease(c *gin.Context) {
	go func() {
		time.Sleep(60 * time.Second)
		ForceCheckCurrentVersion()
	}()

	c.Writer.WriteHeader(http.StatusNoContent)
}

func createLiveSessionResponse(c *gin.Context, name string, token string) gin.H {
	return gin.H{
		"id":    name,
		"url":   getURL(c, fmt.Sprintf("live#%s", name)),
		"token": token,
	}
}

func (srv Server) postLive(c *gin.Context) {
	var ls LiveSession
	if err := c.ShouldBind(&ls); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := GenerateLiveSessionId()
	ls.Initialize()
	LiveSessions[name] = &ls

	c.JSON(http.StatusOK, createLiveSessionResponse(c, name, ls.token))
}

func (srv Server) putLive(c *gin.Context) {
	name := c.Param("name")
	token := c.Query("token")

	var ls LiveSession
	if err := c.ShouldBind(&ls); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prevLs, exists := LiveSessions[name]
	if exists {
		if prevLs.token != token {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		prevLs.ReplaceDeck(ls.Deck, ls.CurrentPage)
		prevLs.ExtendTime()
		c.JSON(http.StatusOK, createLiveSessionResponse(c, name, prevLs.token))
	} else {
		ls.Initialize()
		LiveSessions[name] = &ls
		c.JSON(http.StatusOK, createLiveSessionResponse(c, name, ls.token))
	}
}

func (srv Server) getLive(c *gin.Context) {
	name := c.Param("name")

	ls, ok := LiveSessions[name]

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Live session not found"})
		return
	}

	headers := c.Writer.Header()
	headers.Set("Content-Type", "text/event-stream")
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Connection", "keep-alive")
	headers.Set("Transfer-Encoding", "chunked")

	stream := ls.AddMember()

	c.Stream(func(w io.Writer) bool {
		c.SSEvent("start", ls)

		return false
	})

	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-stream:
			if !ok {
				return false
			}

			c.SSEvent(event.Type, event.Data)
			return true
		case <-c.Request.Context().Done():
			ls.RemoveMember(stream)
			return false
		}
	})
}

func (srv Server) deleteLive(c *gin.Context) {
	name := c.Param("name")
	token := c.Query("token")
	ls, ok := LiveSessions[name]

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Live session not found"})
		return
	}

	if ls.token != token {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	delete(LiveSessions, name)
	c.Writer.WriteHeader(http.StatusNoContent)
}

func (srv Server) postLivePage(c *gin.Context) {
	name := c.Param("name")
	token := c.Query("token")
	ls, ok := LiveSessions[name]

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Live session not found"})
		return
	}

	if ls.token != token {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page number not valid"})
		return
	}

	ls.ChangePage(page)
	ls.ExtendTime()

	c.Writer.WriteHeader(http.StatusNoContent)
}

func (srv Server) Run() {
	r := gin.Default()
	r.Use(corsMiddleware)
	auth := gin.BasicAuth(gin.Accounts{
		"admin": os.Getenv("ADMIN_PASSWORD"),
	})

	r.Static("/public", "./public")
	v2 := r.Group("/v2")
	v2.GET("/bootstrap", srv.getBootstrap)
	v2.GET("/songs", srv.getSongs)
	v2.POST("/songs", auth, srv.postSong)
	v2.GET("/songs/:id", srv.getSong)
	v2.PATCH("/songs/:id", auth, srv.patchSong)
	v2.DELETE("/songs/:id", auth, srv.deleteSong)
	v2.GET("/lyrics/:id", srv.getLyrics)
	v2.GET("/liturgy/:date", srv.getLiturgy)
	v2.GET("/manual", srv.getManual)
	v2.POST("/deck", srv.postDeck)
	v2.POST("/reload", srv.postReload)
	v2.POST("/update_release", srv.postUpdateRelease)
	v2.POST("/live", srv.postLive)
	v2.PUT("/live/:name", srv.putLive)
	v2.GET("/live/:name", srv.getLive)
	v2.DELETE("/live/:name", srv.deleteLive)
	v2.POST("/live/:name/page", srv.postLivePage)
	r.Run(srv.addr)
}
