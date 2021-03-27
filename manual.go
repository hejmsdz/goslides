package main

import (
	"fmt"

	"github.com/kjk/notionapi"
)

type Manual struct {
	Steps []string `json:"steps"`
	Image string   `json:"image"`
}

const manualPageID = "e831bd736718447e8177ade7435337cb"

func GetManual(authToken string) (Manual, bool) {
	m := Manual{make([]string, 0), ""}
	client := &notionapi.Client{AuthToken: authToken}
	page, err := client.DownloadPage(manualPageID)

	if err != nil {
		return m, false
	}

	page.ForEachBlock(func(block *notionapi.Block) {
		switch block.Type {
		case notionapi.BlockNumberedList:
			m.Steps = append(m.Steps, extractText(block.GetTitle()))
		case notionapi.BlockImage:
			m.Image = fmt.Sprintf("%s?table=block&id=%s", block.ImageURL, block.ID)
		}
	})

	return m, true
}
