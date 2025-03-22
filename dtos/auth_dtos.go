package dtos

import "github.com/hejmsdz/goslides/models"

type AuthGoogleRequest struct {
	IDToken string `json:"idToken"`
}

type AuthRefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type AuthResponse struct {
	AccessToken  string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	Name         string `json:"name"`
}

type AuthMeResponse struct {
	UUID        string `json:"uuid"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

func NewAuthMeResponse(user *models.User) *AuthMeResponse {
	return &AuthMeResponse{
		UUID:        user.UUID.String(),
		Email:       user.Email,
		DisplayName: user.DisplayName,
	}
}
