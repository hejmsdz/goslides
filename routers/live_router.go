package routers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/services"
)

func RegisterLiveRoutes(r gin.IRouter, dic *di.Container) {
	h := NewLiveHandler(dic)

	r.POST("/live", h.PostLive)
	r.PUT("/live/:key", h.PutLive)
	r.GET("/live/:key", h.GetLive)
	r.DELETE("/live/:key", h.DeleteLive)
	r.POST("/live/:key/page", h.PostLivePage)
}

type LiveHandler struct {
	Live *services.LiveService
}

func NewLiveHandler(dic *di.Container) *LiveHandler {
	return &LiveHandler{
		Live: dic.Live,
	}
}

func (h *LiveHandler) PostLive(c *gin.Context) {
	var input dtos.LiveSessionRequest
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	key := h.Live.GenerateLiveSessionKey()
	session, err := h.Live.CreateSession(key, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := dtos.NewLiveSessionResponse(c, key, session.Token)
	c.JSON(http.StatusOK, resp)
}

func (h *LiveHandler) PutLive(c *gin.Context) {
	key := c.Param("key")
	token := c.Query("token")

	var input dtos.LiveSessionRequest
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	prevSession, exists := h.Live.GetSession(key)
	if exists {
		if !h.Live.ValidateToken(key, token) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		h.Live.UpdateSession(key, input)
		h.Live.ExtendSessionTime(key)

		resp := dtos.NewLiveSessionResponse(c, key, prevSession.Token)
		c.JSON(http.StatusOK, resp)
	} else {
		if !h.Live.ValidateLiveSessionKey(key) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid live session key"})
			return

		}
		session, err := h.Live.CreateSession(key, input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := dtos.NewLiveSessionResponse(c, key, session.Token)
		c.JSON(http.StatusOK, resp)
	}
}

func (h *LiveHandler) GetLive(c *gin.Context) {
	key := c.Param("key")

	ls, ok := h.Live.GetSession(key)

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
		fmt.Fprintf(w, "retry: 5000\n\n")
		c.SSEvent("start", ls)

		return false
	})

	keepAliveTicker := time.NewTicker(15 * time.Second)
	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-stream:
			if !ok {
				return false
			}

			c.SSEvent(event.Type, event.Data)
			return true
		case <-keepAliveTicker.C:
			c.SSEvent("keepAlive", "")
			return true
		case <-c.Request.Context().Done():
			keepAliveTicker.Stop()
			ls.RemoveMember(stream)
			return false
		}
	})
}

func (h *LiveHandler) DeleteLive(c *gin.Context) {
	key := c.Param("key")
	token := c.Query("token")

	_, ok := h.Live.GetSession(key)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Live session not found"})
		return
	}

	if !h.Live.ValidateToken(key, token) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	h.Live.DeleteSession(key)
	c.Writer.WriteHeader(http.StatusNoContent)
}

func (h *LiveHandler) PostLivePage(c *gin.Context) {
	key := c.Param("key")
	token := c.Query("token")
	_, ok := h.Live.GetSession(key)

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Live session not found"})
		return
	}

	if !h.Live.ValidateToken(key, token) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page number not valid"})
		return
	}

	h.Live.ChangeSessionPage(key, page)
	h.Live.ExtendSessionTime(key)

	c.Writer.WriteHeader(http.StatusNoContent)
}
