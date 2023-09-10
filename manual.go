package main

import (
	"context"

	"github.com/jomei/notionapi"
)

type Manual struct {
	Steps []string `json:"steps"`
	Image string   `json:"image"`
}

func GetManual(authToken string, manualPageID string) (Manual, bool) {
	m := Manual{make([]string, 0), ""}
	client := notionapi.NewClient(notionapi.Token(authToken))
	pagination := notionapi.Pagination{
		StartCursor: "",
		PageSize:    100,
	}

	blocks, err := client.Block.GetChildren(
		context.Background(),
		notionapi.BlockID(manualPageID),
		&pagination,
	)

	if err != nil {
		return m, false
	}

	for _, block := range blocks.Results {
		switch block.GetType() {
		case notionapi.BlockTypeNumberedListItem:
			itemBlock := block.(*notionapi.NumberedListItemBlock)
			m.Steps = append(m.Steps, extractText(itemBlock.NumberedListItem.RichText))
		case notionapi.BlockTypeImage:
			imageBlock := block.(*notionapi.ImageBlock)
			m.Image = imageBlock.Image.File.URL
		}
	}

	return m, true
}
