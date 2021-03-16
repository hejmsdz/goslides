package main

import (
	"strings"

	"github.com/kjk/notionapi"
)

const collectionID = "f56d81ad-0432-4868-b96f-4b9fcee690fa"
const collectionViewID = "db8fc575-ef3b-4392-9a2f-3a93ef50e64d"
const propertyKeyNumber = "Numer"
const propertyKeyTags = "Kategorie"

type Song struct {
	Id     string `json:"id"`
	Title  string `json:"title"`
	Number string `json:"number"`
	Tags   string `json:"-"`

	contentBlockIDs []string
}

type SongsDB struct {
	client       *notionapi.Client
	Songs        map[string]Song
	LyricsBlocks map[string]string
}

func extractText(property []*notionapi.TextSpan) string {
	if len(property) == 0 || property[0] == nil {
		return ""
	}
	return property[0].Text
}

func (sdb *SongsDB) Initialize() error {
	sdb.client = &notionapi.Client{}
	sdb.Songs = make(map[string]Song, 0)
	sdb.LyricsBlocks = make(map[string]string, 0)

	result, err := sdb.client.QueryCollection(collectionID, collectionViewID, nil, nil)
	if err != nil {
		return err
	}

	blockMap := result.RecordMap.Blocks
	pageIDs := result.Result.BlockIDS

	for _, pageID := range pageIDs {
		page, ok := blockMap[pageID]
		if !ok {
			continue
		}
		pageBlock := page.Block

		sdb.Songs[pageID] = Song{
			pageID,
			extractText(pageBlock.GetTitle()),
			extractText(pageBlock.GetProperty(propertyKeyNumber)),
			extractText(pageBlock.GetProperty(propertyKeyTags)),
			pageBlock.ContentIDs,
		}

		for _, contentID := range pageBlock.ContentIDs {
			content, ok := blockMap[contentID]
			if !ok {
				continue
			}
			contentBlock := content.Block

			sdb.LyricsBlocks[contentID] = extractText(contentBlock.GetTitle())
		}
	}

	return nil
}

func (sdb SongsDB) FilterSongs(query string) (result []Song) {
	queryLower := strings.ToLower(query)

	for _, song := range sdb.Songs {
		titleLower := strings.ToLower(song.Title)
		numberLower := strings.ToLower(song.Number)
		tagsLower := strings.ToLower(song.Tags)
		if strings.Contains(titleLower, queryLower) ||
			strings.Contains(numberLower, queryLower) ||
			strings.Contains(tagsLower, queryLower) {
			result = append(result, song)
		}
	}

	return
}

func (sdb SongsDB) LoadMissingVerses(songIDs []string) error {
	missingBlocks := make([]notionapi.RecordRequest, 0)

	for _, songID := range songIDs {
		song := sdb.Songs[songID]

		for _, blockID := range song.contentBlockIDs {
			if _, ok := sdb.LyricsBlocks[blockID]; !ok {
				missingBlocks = append(missingBlocks, notionapi.RecordRequest{Table: "block", ID: blockID})
			}
		}
	}

	res, err := sdb.client.GetRecordValues(missingBlocks)
	if err != nil {
		return err
	}
	for _, record := range res.Results {
		contentBlock := record.Block
		value := extractText(contentBlock.GetTitle())
		sdb.LyricsBlocks[record.ID] = value
	}

	return nil
}

func (sdb SongsDB) GetLyrics(songID string) ([]string, bool) {
	song := sdb.Songs[songID]
	hasAllVerses := true

	lyrics := make([]string, 0)
	for _, blockID := range song.contentBlockIDs {
		verse, ok := sdb.LyricsBlocks[blockID]
		hasAllVerses = hasAllVerses && ok
		lyrics = append(lyrics, verse)
	}

	return lyrics, hasAllVerses
}
