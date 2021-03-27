package main

type Bootstrap struct {
	CurrentVersion string `json:"currentVersion"`
	AppDownloadURL string `json:"appDownloadUrl"`
}

var bootstrap = Bootstrap{"1.0.0", ""}
