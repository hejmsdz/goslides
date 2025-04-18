package routers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/common"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/services"
	"github.com/pkg/errors"
)

func RegisterLiveRoutes(r gin.IRouter, dic *di.Container) {
	h := NewLiveHandler(dic)

	r.POST("/live", h.Auth.OptionalAuthMiddleware, h.PostLive)
	r.PUT("/live/:key", h.Auth.OptionalAuthMiddleware, h.PutLive)
	r.GET("/live/:key", h.GetLive)
	r.DELETE("/live/:key", h.DeleteLive)
	r.POST("/live/:key/page", h.PostLivePage)
}

type LiveHandler struct {
	Live *services.LiveService
	Auth *services.AuthService
}

func NewLiveHandler(dic *di.Container) *LiveHandler {
	return &LiveHandler{
		Live: dic.Live,
		Auth: dic.Auth,
	}
}

func (h *LiveHandler) PostLive(c *gin.Context) {
	var input dtos.LiveSessionRequest
	user := h.Auth.GetCurrentUser(c)

	if err := c.ShouldBind(&input); err != nil {
		common.ReturnBadRequestError(c, err)
		return
	}

	if err := input.Validate(); err != nil {
		common.ReturnAPIError(c, http.StatusUnprocessableEntity, "validation failed", err)
		return
	}

	key := h.Live.GenerateLiveSessionKey()
	session, err := h.Live.CreateSession(key, input, user)
	if err != nil {
		common.ReturnError(c, err)
		return
	}

	resp := dtos.NewLiveSessionResponse(key, session.Token)
	c.JSON(http.StatusOK, resp)
}

func (h *LiveHandler) PutLive(c *gin.Context) {
	key := c.Param("key")
	token := c.Query("token")
	user := h.Auth.GetCurrentUser(c)

	var input dtos.LiveSessionRequest
	if err := c.ShouldBind(&input); err != nil {
		common.ReturnBadRequestError(c, err)
		return
	}

	if err := input.Validate(); err != nil {
		common.ReturnAPIError(c, http.StatusUnprocessableEntity, "validation failed", err)
		return
	}

	prevSession, exists := h.Live.GetSession(key)
	if exists {
		if !h.Live.ValidateToken(key, token) {
			common.ReturnAPIError(c, http.StatusForbidden, "invalid token", nil)
			return
		}

		h.Live.UpdateSession(key, input, user)
		h.Live.ExtendSessionTime(key)

		resp := dtos.NewLiveSessionResponse(key, prevSession.Token)
		c.JSON(http.StatusOK, resp)
	} else {
		if !h.Live.ValidateLiveSessionKey(key) {
			common.ReturnAPIError(c, http.StatusUnprocessableEntity, "invalid live session key", nil)
			return
		}
		session, err := h.Live.CreateSession(key, input, user)
		if err != nil {
			common.ReturnError(c, err)
			return
		}

		resp := dtos.NewLiveSessionResponse(key, session.Token)
		c.JSON(http.StatusOK, resp)
	}
}

func (h *LiveHandler) GetLive(c *gin.Context) {
	key := c.Param("key")

	ls, ok := h.Live.GetSession(key)
	if !ok {
		common.ReturnAPIError(c, http.StatusNotFound, "live session not found", nil)
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
		common.ReturnAPIError(c, http.StatusNotFound, "live session not found", nil)
		return
	}

	if !h.Live.ValidateToken(key, token) {
		common.ReturnAPIError(c, http.StatusForbidden, "invalid token", nil)
		return
	}

	h.Live.DeleteSession(key)
	c.Status(http.StatusNoContent)
}

func (h *LiveHandler) PostLivePage(c *gin.Context) {
	key := c.Param("key")
	token := c.Query("token")
	_, ok := h.Live.GetSession(key)

	if !ok {
		common.ReturnAPIError(c, http.StatusNotFound, "live session not found", nil)
		return
	}

	if !h.Live.ValidateToken(key, token) {
		common.ReturnAPIError(c, http.StatusUnauthorized, "invalid token", nil)
		return
	}

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		common.ReturnAPIError(c, http.StatusBadRequest, "page number not valid", errors.Wrap(err, "failed to parse page number"))
		return
	}

	h.Live.ChangeSessionPage(key, page)
	h.Live.ExtendSessionTime(key)

	c.Status(http.StatusNoContent)
}
