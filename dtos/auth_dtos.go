package dtos

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
