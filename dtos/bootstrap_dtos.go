package dtos

type BootstrapResponse struct {
	CurrentVersion string  `json:"currentVersion"`
	AppDownloadURL string  `json:"appDownloadUrl"`
	SongEditURL    *string `json:"songEditUrl"`
	ContactURL     *string `json:"contactUrl"`
	SupportURL     *string `json:"supportUrl"`
}
