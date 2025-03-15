package dtos

type BootstrapResponse struct {
	CurrentVersion string `json:"currentVersion"`
	AppDownloadURL string `json:"appDownloadUrl"`
}
