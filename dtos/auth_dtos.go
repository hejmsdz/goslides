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
	IsNewUser    bool            `json:"isNewUser"`
}

func NewAuthResponse(token string, refreshToken string, user *models.User, isNewUser bool) *AuthResponse {
	return &AuthResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		User:         NewUserMeResponse(user),
		IsNewUser:    isNewUser,
	}
}

type AuthNonceVerifyRequest struct {
	Nonce string `json:"nonce"`
}

type AuthNonceResponse struct {
	Nonce string `json:"nonce"`
}

func NewAuthNonceResponse(nonce string) *AuthNonceResponse {
	return &AuthNonceResponse{
		Nonce: nonce,
	}
}
