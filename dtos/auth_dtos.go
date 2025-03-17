package dtos

type GoogleAuthRequest struct {
	IDToken string `json:"idToken"`
}

type GoogleAuthResponse struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}
