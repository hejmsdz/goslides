package dtos

import "github.com/hejmsdz/goslides/models"

type UserMeResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

func NewUserMeResponse(user *models.User) *UserMeResponse {
	return &UserMeResponse{
		ID:          user.UUID.String(),
		Email:       user.Email,
		DisplayName: user.DisplayName,
	}
}
