package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/kjk/notionapi"
)

const collectionID = "f56d81ad-0432-4868-b96f-4b9fcee690fa"
const collectionViewID = "db8fc575-ef3b-4392-9a2f-3a93ef50e64d"
const propertyNameNumber = "Numer"
const propertyNameTags = "Kategorie"

type Song struct {
	Id     string `json:"id"`
	Title  string `json:"title"`
	Number string `json:"number"`
	Tags   string `json:"-"`
	Slug   string `json:"slug"`

	numberChapter   int
	numberItem      int
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

var slugReplacer = strings.NewReplacer("ą", "a", "ć", "c", "ę", "e", "ł", "l", "ń", "n", "ó", "o", "ś", "s", "ź", "z", "ż", "z")
var nonAlpha, _ = regexp.Compile("[^a-zA-Z0-9\\. ]+")

func slugify(text string) string {
	text = strings.ToLower(text)
	text = slugReplacer.Replace(text)
	text = nonAlpha.ReplaceAllString(text, "")

	return text
}

func getColumnKeys(recordMap *notionapi.RecordMap) (propertyKeyNumber, propertyKeyTags string) {
	collectionDetails := recordMap.Collections[collectionID].Collection
	for key, column := range collectionDetails.Schema {
		switch column.Name {
		case propertyNameNumber:
			propertyKeyNumber = key
		case propertyNameTags:
			propertyKeyTags = key
		}
	}
	return
}

func (sdb *SongsDB) Initialize() error {
	sdb.client = &notionapi.Client{}
	sdb.Songs = make(map[string]Song, 0)
	sdb.LyricsBlocks = make(map[string]string, 0)

	result, err := sdb.client.QueryCollection(collectionID, collectionViewID, nil, nil)
	if err != nil {
		return err
	}

	propertyKeyNumber, propertyKeyTags := getColumnKeys(result.RecordMap)
	blockMap := result.RecordMap.Blocks
	pageIDs := result.Result.BlockIDS

	for _, pageID := range pageIDs {
		page, ok := blockMap[pageID]
		if !ok {
			continue
		}
		pageBlock := page.Block

		title := extractText(pageBlock.GetTitle())
		number := extractText(pageBlock.GetProperty(propertyKeyNumber))
		tags := extractText(pageBlock.GetProperty(propertyKeyTags))

		numberSplit := strings.Split(number, ".")
		numberChapter, numberItem := -1, -1
		if len(numberSplit) == 2 {
			numberChapter, _ = strconv.Atoi(numberSplit[0])
			numberItem, _ = strconv.Atoi(numberSplit[1])
		}

		sdb.Songs[pageID] = Song{
			Id:     pageID,
			Title:  title,
			Number: number,
			Tags:   tags,
			Slug:   fmt.Sprintf("%s|%s|%s", slugify(title), number, slugify(tags)),

			numberChapter:   numberChapter,
			numberItem:      numberItem,
			contentBlockIDs: pageBlock.ContentIDs,
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

var numberQueryRegexp, _ = regexp.Compile("^\\d")

func (sdb SongsDB) FilterSongs(query string) (results []Song) {
	results = make([]Song, 0)

	if len(query) < 3 {
		return
	}

	querySlug := slugify(query)

	for _, song := range sdb.Songs {
		if strings.Contains(song.Slug, querySlug) {
			results = append(results, song)
		}
	}

	isNumberQuery := numberQueryRegexp.Match([]byte(query))
	sort.Slice(results, func(i, j int) bool {
		if isNumberQuery {
			iChapter, jChapter := results[i].numberChapter, results[j].numberChapter
			iItem, jItem := results[i].numberItem, results[j].numberItem

			if iChapter == jChapter {
				return iItem < jItem
			}
			return iChapter < jChapter
		}
		return results[i].Title < results[j].Title
	})

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
