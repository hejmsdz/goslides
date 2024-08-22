package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
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
	nonce := rand.Float64()
	return getURL(c, fmt.Sprintf("public/%s?v=%f", fileName, nonce))
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

	textDeck, ok := deck.ToTextSlides(*srv.songsDB, srv.liturgyDB)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get lyrics"})
		return
	}

	var deckResult DeckResult
	switch deck.Format {
	case "png+zip":
		zip, contents, err := BuildImages(textDeck, deck.GetPageConfig())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		zipName := deck.Date + ".zip"
		SaveTemporaryFile(zip, zipName)
		deckResult = DeckResult{getPublicURL(c, zipName), contents}

	case "txt":
		text := Tugalize(textDeck)

		txtName := deck.Date + ".txt"
		textReader := strings.NewReader(text)
		SaveTemporaryFile(textReader, txtName)
		deckResult = DeckResult{getPublicURL(c, txtName), []ContentSlide{}}

	default:
		pdf, contents, err := BuildPDF(textDeck, deck.GetPageConfig())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		pdfName := deck.Date + ".pdf"
		SaveTemporaryPDF(pdf, pdfName)

		deckResult = DeckResult{getPublicURL(c, pdfName), contents}
	}

	if !deck.Contents {
		deckResult.Contents = nil
	}

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

func (srv Server) putLive(c *gin.Context) {
	name := "session" // c.Param("name")

	var ls LiveSession
	if err := c.ShouldBind(&ls); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prevLs, exists := LiveSessions[name]
	if exists {
		prevLs.ReplaceDeck(ls.Deck, ls.CurrentPage)
	} else {
		LiveSessions[name] = &ls
	}

	c.JSON(http.StatusOK, gin.H{
		"url": getURL(c, "live"),
	})
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

func (srv Server) postLivePage(c *gin.Context) {
	name := c.Param("name")
	ls, ok := LiveSessions[name]

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Live session not found"})
		return
	}

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page number not valid"})
		return
	}

	ls.ChangePage(page)
}

func (srv Server) Run() {
	r := gin.Default()
	r.Use(corsMiddleware)
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
	v2.PUT("/live/:name", srv.putLive)
	v2.GET("/live/:name", srv.getLive)
	v2.POST("/live/:name/page", srv.postLivePage)
	r.Run(srv.addr)
}
