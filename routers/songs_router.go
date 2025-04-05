package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
	"github.com/hejmsdz/goslides/services"
)

func RegisterSongRoutes(r gin.IRouter, dic *di.Container) {
	h := NewSongsHandler(dic)
	auth := dic.Auth.AuthMiddleware
	optionalAuth := dic.Auth.OptionalAuthMiddleware

	r.GET("/songs", optionalAuth, h.GetSongs)
	r.POST("/songs", auth, h.PostSong)
	r.GET("/songs/:id", optionalAuth, h.GetSong)
	r.PATCH("/songs/:id", auth, h.PatchSong)
	r.DELETE("/songs/:id", auth, h.DeleteSong)
	r.GET("/lyrics/:id", optionalAuth, h.GetLyrics)
}

type SongsHandler struct {
	Songs *services.SongsService
	Auth  *services.AuthService
}

func NewSongsHandler(dic *di.Container) *SongsHandler {
	return &SongsHandler{dic.Songs, dic.Auth}
}

func (h *SongsHandler) GetSongs(c *gin.Context) {
	query := c.Query("query")
	user := h.Auth.GetCurrentUser(c)
	songs := h.Songs.FilterSongs(query, user)

	resp := dtos.NewSongListResponse(songs)
	c.JSON(http.StatusOK, resp)
}

func (h *SongsHandler) PostSong(c *gin.Context) {
	var input dtos.SongRequest
	user := h.Auth.GetCurrentUser(c)

	if err := c.ShouldBind(&input); err != nil {
		c.Error(err)
		return
	}

	if err := input.Validate(); err != nil {
		c.Error(err)
		return
	}

	song, err := h.Songs.CreateSong(input, user)
	if err != nil {
		c.Error(err)
		return
	}

	resp := dtos.NewSongDetailResponse(song)
	c.JSON(http.StatusCreated, resp)
}

func (h *SongsHandler) GetSong(c *gin.Context) {
	id := c.Param("id")
	user := h.Auth.GetCurrentUser(c)
	song, err := h.Songs.GetSong(id, user)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	resp := dtos.NewSongDetailResponse(song)
	c.JSON(http.StatusOK, resp)
}

func (h *SongsHandler) PatchSong(c *gin.Context) {
	id := c.Param("id")
	user := h.Auth.GetCurrentUser(c)

	var input dtos.SongRequest
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	song, err := h.Songs.UpdateSong(id, input, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := dtos.NewSongDetailResponse(song)
	c.JSON(http.StatusOK, resp)
}

func (h *SongsHandler) DeleteSong(c *gin.Context) {
	id := c.Param("id")
	user := h.Auth.GetCurrentUser(c)
	err := h.Songs.DeleteSong(id, user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *SongsHandler) GetLyrics(c *gin.Context) {
	id := c.Param("id")
	raw := c.Query("raw") == "1"
	user := h.Auth.GetCurrentUser(c)
	song, err := h.Songs.GetSong(id, user)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	resp := song.FormatLyrics(models.FormatLyricsOptions{Raw: raw})
	c.JSON(http.StatusOK, resp)
}
