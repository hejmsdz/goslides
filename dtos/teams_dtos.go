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

type TeamRequest struct {
	Name string `json:"name"`
}

type TeamMember struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TeamMembersResponse = []TeamMember

func NewTeamMembersResponse(members []*models.User) TeamMembersResponse {
	resp := make(TeamMembersResponse, len(members))
	for i, member := range members {
		resp[i] = TeamMember{
			ID:   member.UUID.String(),
			Name: member.DisplayName,
		}
	}
	return resp
}
