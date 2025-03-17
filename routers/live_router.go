package routers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/services"
)

func RegisterLiveRoutes(r gin.IRouter, dic *di.Container) {
	h := NewLiveHandler(dic)

	r.POST("/live", h.PostLive)
	r.PUT("/live/:name", h.PutLive)
	r.GET("/live/:name", h.GetLive)
	r.DELETE("/live/:name", h.DeleteLive)
	r.POST("/live/:name/page", h.PostLivePage)
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

	id := h.Live.GenerateLiveSessionId()
	session := h.Live.CreateSession(id, input)

	resp := dtos.NewLiveSessionResponse(c, id, session.Token)
	c.JSON(http.StatusOK, resp)
}

func (h *LiveHandler) PutLive(c *gin.Context) {
	name := c.Param("name")
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

	prevSession, exists := h.Live.GetSession(name)
	if exists {
		if !h.Live.ValidateToken(name, token) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		h.Live.UpdateSession(name, input)
		h.Live.ExtendSessionTime(name)

		resp := dtos.NewLiveSessionResponse(c, name, prevSession.Token)
		c.JSON(http.StatusOK, resp)
	} else {
		id := h.Live.GenerateLiveSessionId()
		session := h.Live.CreateSession(id, input)

		resp := dtos.NewLiveSessionResponse(c, name, session.Token)
		c.JSON(http.StatusOK, resp)
	}
}

func (h *LiveHandler) GetLive(c *gin.Context) {
	name := c.Param("name")

	ls, ok := h.Live.GetSession(name)

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

func (h *LiveHandler) DeleteLive(c *gin.Context) {
	name := c.Param("name")
	token := c.Query("token")

	_, ok := h.Live.GetSession(name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Live session not found"})
		return
	}

	if !h.Live.ValidateToken(name, token) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	h.Live.DeleteSession(name)
	c.Writer.WriteHeader(http.StatusNoContent)
}

func (h *LiveHandler) PostLivePage(c *gin.Context) {
	name := c.Param("name")
	token := c.Query("token")
	_, ok := h.Live.GetSession(name)

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Live session not found"})
		return
	}

	if !h.Live.ValidateToken(name, token) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page number not valid"})
		return
	}

	h.Live.ChangeSessionPage(name, page)
	h.Live.ExtendSessionTime(name)

	c.Writer.WriteHeader(http.StatusNoContent)
}
