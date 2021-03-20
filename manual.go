package main

import (
	"github.com/kjk/notionapi"
)

type Manual struct {
	Steps []string `json:"steps"`
}

const manualPageID = "e831bd736718447e8177ade7435337cb"

func GetManual(authToken string) (Manual, bool) {
	m := Manual{make([]string, 0)}
	client := &notionapi.Client{AuthToken: authToken}
	page, err := client.DownloadPage(manualPageID)

	if err != nil {
		return m, false
	}

	for _, block := range page.Root().Content {
		m.Steps = append(m.Steps, extractText(block.GetTitle()))
	}

	return m, true
}
