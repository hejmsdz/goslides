package main

import (
	"context"

	"github.com/google/go-github/v41/github"
)

type Bootstrap struct {
	CurrentVersion string `json:"currentVersion"`
	AppDownloadURL string `json:"appDownloadUrl"`
}

var bootstrap *Bootstrap

func CheckCurrentVersion() {
	if bootstrap != nil {
		return
	}

	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), "hejmsdz", "slidesui")

	if err != nil {
		return
	}

	currentVersion := *release.TagName
	if currentVersion[0] == 'v' {
		currentVersion = currentVersion[1:]
	}

	bootstrap = &Bootstrap{currentVersion, *release.HTMLURL}
}
