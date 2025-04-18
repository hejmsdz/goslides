package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/services"
)

func RegisterTeamRoutes(r gin.IRouter, dic *di.Container) {
	h := NewTeamsHandler(dic)
	auth := dic.Auth.AuthMiddleware

	r.GET("/teams", auth, h.GetTeams)
}

type TeamsHandler struct {
	Teams *services.TeamsService
	Auth  *services.AuthService
}

func NewTeamsHandler(dic *di.Container) *TeamsHandler {
	return &TeamsHandler{dic.Teams, dic.Auth}
}

func (h *TeamsHandler) GetTeams(c *gin.Context) {
	user := h.Auth.GetCurrentUser(c)
	teams, err := h.Teams.GetUserTeams(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get teams"})
		return
	}

	resp := dtos.NewTeamListResponse(teams)
	c.JSON(http.StatusOK, resp)
}
