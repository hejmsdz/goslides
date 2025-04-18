package dtos

import "github.com/hejmsdz/goslides/models"

type TeamResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewTeamResponse(team *models.Team) TeamResponse {
	return TeamResponse{
		ID:   team.UUID.String(),
		Name: team.Name,
	}
}

func NewTeamListResponse(teams []*models.Team) []TeamResponse {
	resp := make([]TeamResponse, len(teams))
	for i, team := range teams {
		resp[i] = NewTeamResponse(team)
	}
	return resp
}
