package dtos

import (
	"fmt"
	"os"
	"time"

	"github.com/hejmsdz/goslides/models"
)

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

type TeamDetailsResponse struct {
	TeamResponse
	Members []TeamMember `json:"members"`
}

func NewTeamDetailsResponse(team *models.Team, members []*models.User) TeamDetailsResponse {
	membersResp := make([]TeamMember, len(members))
	for i, member := range members {
		membersResp[i] = TeamMember{
			ID:   member.UUID.String(),
			Name: member.DisplayName,
		}
	}

	return TeamDetailsResponse{
		TeamResponse: NewTeamResponse(team),
		Members:      membersResp,
	}
}

type TeamInvitationResponse struct {
	Token     string    `json:"token"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func NewTeamInvitationResponse(invitation *models.Invitation) TeamInvitationResponse {
	return TeamInvitationResponse{
		Token:     invitation.Token,
		URL:       fmt.Sprintf("%s/invitation/%s", os.Getenv("FRONTEND_URL"), invitation.Token),
		ExpiresAt: invitation.ExpiresAt,
	}
}

type TeamJoinRequest struct {
	Token string `json:"token"`
}
