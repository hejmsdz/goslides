package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
)

type DeckResult struct {
	URL string `json:"url"`
}

func getPublicURL(c *gin.Context, fileName string) string {
	scheme := "https"
	return fmt.Sprintf("%s://%s/public/%s", scheme, c.Request.Host, fileName)
}

type Server struct {
	songsDB   *SongsDB
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

func (srv Server) getLyrics(c *gin.Context) {
	id := c.Param("id")
	srv.songsDB.LoadMissingVerses([]string{id})
	lyrics, _ := srv.songsDB.GetLyrics(id, false)

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

	textDeck, ok := deck.ToTextSlides(*srv.songsDB, srv.liturgyDB)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get lyrics"})
		return
	}
	pdf, err := BuildPDF(textDeck, deck.GetPageConfig())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pdfName := deck.Date + ".pdf"
	SaveTemporaryPDF(pdf, pdfName)

	deckResult := DeckResult{getPublicURL(c, pdfName)}
	c.JSON(http.StatusOK, deckResult)
}

func (srv Server) postReload(c *gin.Context) {
	err := srv.songsDB.Initialize(NOTION_TOKEN, NOTION_DB)
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

func (srv Server) Run() {
	r := gin.Default()
	r.Use(cors.Default())
	r.Static("/public", "./public")
	v2 := r.Group("/v2")
	v2.GET("/bootstrap", srv.getBootstrap)
	v2.GET("/songs", srv.getSongs)
	v2.GET("/lyrics/:id", srv.getLyrics)
	v2.GET("/liturgy/:date", srv.getLiturgy)
	v2.GET("/manual", srv.getManual)
	v2.POST("/deck", srv.postDeck)
	v2.POST("/reload", srv.postReload)
	v2.POST("/update_release", srv.postUpdateRelease)
	r.Run(srv.addr)
}
