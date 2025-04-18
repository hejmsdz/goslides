package dtos

import "github.com/hejmsdz/goslides/models"

type AuthGoogleRequest struct {
	IDToken string `json:"idToken"`
}

type AuthRefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type AuthMeResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

func NewAuthMeResponse(user *models.User) *AuthMeResponse {
	return &AuthMeResponse{
		ID:          user.UUID.String(),
		Email:       user.Email,
		DisplayName: user.DisplayName,
	}
}

type AuthResponse struct {
	AccessToken  string          `json:"token"`
	RefreshToken string          `json:"refreshToken"`
	User         *AuthMeResponse `json:"user"`
}

func NewAuthResponse(token string, refreshToken string, user *models.User) *AuthResponse {
	return &AuthResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		User:         NewAuthMeResponse(user),
	}
}
