package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/services"
)

func RegisterLiturgyRoutes(r gin.IRouter, dic *di.Container) {
	h := NewLiturgyHandler(dic)

	r.GET("/liturgy/:date", h.GetLiturgy)
}

type LiturgyHandler struct {
	Liturgy *services.LiturgyService
}

func NewLiturgyHandler(dic *di.Container) *LiturgyHandler {
	return &LiturgyHandler{
		Liturgy: dic.Liturgy,
	}
}

func (h *LiturgyHandler) GetLiturgy(c *gin.Context) {
	date := c.Param("date")

	liturgy, ok := h.Liturgy.GetDay(date)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Liturgy error"})
		return
	}

	c.JSON(http.StatusOK, liturgy)
}
