package dtos

import "github.com/hejmsdz/goslides/models"

type AuthGoogleRequest struct {
	IDToken string `json:"idToken"`
}

type AuthRefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type AuthResponse struct {
	AccessToken  string          `json:"token"`
	RefreshToken string          `json:"refreshToken"`
	User         *UserMeResponse `json:"user"`
}

func NewAuthResponse(token string, refreshToken string, user *models.User) *AuthResponse {
	return &AuthResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		User:         NewUserMeResponse(user),
	}
}
