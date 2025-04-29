package dtos

type BootstrapResponse struct {
	CurrentVersion string  `json:"currentVersion"`
	AppDownloadURL string  `json:"appDownloadUrl"`
	FrontendURL    string  `json:"frontendUrl"`
	ContactURL     *string `json:"contactUrl"`
	SupportURL     *string `json:"supportUrl"`
}
