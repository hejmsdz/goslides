package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/common"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/services"
)

func RegisterTeamRoutes(r gin.IRouter, dic *di.Container) {
	h := NewTeamsHandler(dic)
	auth := dic.Auth.AuthMiddleware

	r.GET("/teams", auth, h.GetTeams)
	r.GET("/teams/:uuid", auth, h.GetTeam)
	r.POST("/teams", auth, h.PostTeam)
	r.POST("/teams/:uuid/invite", auth, h.PostTeamInvite)
	r.POST("/teams/join", auth, h.PostTeamJoin)
	r.POST("/teams/:uuid/leave", auth, h.PostTeamLeave)
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

func (h *TeamsHandler) GetTeam(c *gin.Context) {
	user := h.Auth.GetCurrentUser(c)
	uuid := c.Param("uuid")

	team, err := h.Teams.GetUserTeam(user, uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get team"})
		return
	}

	members, err := h.Teams.GetTeamMembers(team)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get team members"})
		return
	}

	resp := dtos.NewTeamDetailsResponse(team, members)
	c.JSON(http.StatusOK, resp)
}

func (h *TeamsHandler) PostTeam(c *gin.Context) {
	user := h.Auth.GetCurrentUser(c)

	input := &dtos.TeamRequest{}
	if err := c.ShouldBindJSON(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	team, err := h.Teams.CreateTeam(user, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create team"})
		return
	}

	resp := dtos.NewTeamResponse(team)
	c.JSON(http.StatusOK, resp)
}

func (h *TeamsHandler) PostTeamInvite(c *gin.Context) {
	user := h.Auth.GetCurrentUser(c)
	uuid := c.Param("uuid")

	invitation, err := h.Teams.CreateInvitation(user, uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create invitation"})
		return
	}

	c.JSON(http.StatusOK, dtos.NewTeamInvitationResponse(invitation))
}

func (h *TeamsHandler) PostTeamJoin(c *gin.Context) {
	user := h.Auth.GetCurrentUser(c)

	input := &dtos.TeamJoinRequest{}
	if err := c.ShouldBindJSON(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	team, err := h.Teams.JoinTeam(user, input.Token)
	if err != nil {
		common.ReturnError(c, err)
		return
	}

	resp := dtos.NewTeamResponse(team)
	c.JSON(http.StatusOK, resp)
}

func (h *TeamsHandler) PostTeamLeave(c *gin.Context) {
	user := h.Auth.GetCurrentUser(c)
	uuid := c.Param("uuid")

	err := h.Teams.LeaveTeam(user, uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to leave team"})
		return
	}

	c.Status(http.StatusNoContent)
}
