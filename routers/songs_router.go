package routers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
	"github.com/hejmsdz/goslides/services"
)

func RegisterSongRoutes(r gin.IRouter, dic *di.Container) {
	h := NewSongsHandler(dic)

	r.GET("/songs", h.GetSongs)
	r.POST("/songs", h.PostSong)
	r.GET("/songs/:id", h.GetSong)
	r.PATCH("/songs/:id", h.PatchSong)
	r.DELETE("/songs/:id", h.DeleteSong)
	r.GET("/lyrics/:id", h.GetLyrics)
}

type SongsHandler struct {
	Songs *services.SongsService
}

func NewSongsHandler(dic *di.Container) *SongsHandler {
	return &SongsHandler{dic.Songs}
}

func (h *SongsHandler) GetSongs(c *gin.Context) {
	query := c.Query("query")
	songs := h.Songs.FilterSongs(query)

	resp := dtos.NewSongListResponse(songs)
	c.JSON(http.StatusOK, resp)
}

func (h *SongsHandler) PostSong(c *gin.Context) {
	var input dtos.SongRequest

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	song, err := h.Songs.CreateSong(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := dtos.NewSongDetailResponse(*song)
	c.JSON(http.StatusCreated, resp)
}

func (h *SongsHandler) GetSong(c *gin.Context) {
	id := c.Param("id")
	song, err := h.Songs.GetSong(id)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	resp := dtos.NewSongDetailResponse(song)
	c.JSON(http.StatusOK, resp)
}

func (h *SongsHandler) PatchSong(c *gin.Context) {
	id := c.Param("id")
	var input dtos.SongRequest
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	song, err := h.Songs.UpdateSong(id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := dtos.NewSongDetailResponse(*song)
	c.JSON(http.StatusOK, resp)
}

func (h *SongsHandler) DeleteSong(c *gin.Context) {
	id := c.Param("id")

	err := h.Songs.DeleteSong(id)

	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *SongsHandler) GetLyrics(c *gin.Context) {
	id := c.Param("id")
	raw := c.Query("raw") == "1"
	song, err := h.Songs.GetSong(id)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	resp := song.FormatLyrics(models.FormatLyricsOptions{Raw: raw})
	c.JSON(http.StatusOK, resp)
}
