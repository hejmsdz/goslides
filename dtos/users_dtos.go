package dtos

import (
	"errors"

	"github.com/hejmsdz/goslides/models"
)

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

type UserUpdateRequest struct {
	DisplayName string `json:"displayName"`
}

func (u *UserUpdateRequest) Validate() error {
	if u.DisplayName == "" {
		return errors.New("display name is required")
	}

	if len(u.DisplayName) > 100 {
		return errors.New("display name must be less than 100 characters")
	}

	return nil
}
