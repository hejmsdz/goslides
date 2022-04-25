package main

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jomei/notionapi"
)

const databaseId = "26c6e5f0367243e1870b1ee51d742632"
const propertyNameTitle = "Tytuł"
const propertyNameNumber = "Numer"
const propertyNameTags = "Kategorie"
const ordinaryTag = "części stałe"

type Song struct {
	Id         string `json:"id"`
	Title      string `json:"title"`
	Number     string `json:"number"`
	Tags       string `json:"-"`
	Slug       string `json:"slug"`
	IsOrdinary bool   `json:"isOrdinary,omitempty"`

	numberChapter int
	numberItem    int
}

type SongsDB struct {
	client       *notionapi.Client
	Songs        map[string]Song
	LyricsBlocks map[string][]string
}

func extractText(property []notionapi.RichText) (text string) {
	text = ""
	if len(property) == 0 {
		return
	}

	for _, span := range property {
		text += span.PlainText
	}

	return
}

var slugReplacer = strings.NewReplacer("ą", "a", "ć", "c", "ę", "e", "ł", "l", "ń", "n", "ó", "o", "ś", "s", "ź", "z", "ż", "z")
var nonAlpha = regexp.MustCompile("[^a-zA-Z0-9\\. ]+")

func slugify(text string) string {
	text = strings.ToLower(text)
	text = slugReplacer.Replace(text)
	text = nonAlpha.ReplaceAllString(text, "")

	return text
}

func (sdb *SongsDB) Initialize(authToken string) error {
	sdb.client = notionapi.NewClient(notionapi.Token(authToken))
	sdb.Songs = make(map[string]Song, 0)
	sdb.LyricsBlocks = make(map[string][]string, 0)

	query := notionapi.DatabaseQueryRequest{
		PropertyFilter: nil,
		CompoundFilter: nil,
		Sorts:          make([]notionapi.SortObject, 0),
		StartCursor:    notionapi.Cursor(""),
		PageSize:       100,
	}

	for {
		result, err := sdb.client.Database.Query(context.Background(), databaseId, &query)
		if err != nil {
			return err
		}

		for _, song := range result.Results {
			pageID := song.ID.String()
			title := extractText(song.Properties[propertyNameTitle].(*notionapi.TitleProperty).Title)
			number := extractText(song.Properties[propertyNameNumber].(*notionapi.RichTextProperty).RichText)
			tags := song.Properties[propertyNameTags].(*notionapi.MultiSelectProperty)
			joinedTags := ""

			for _, tag := range tags.MultiSelect {
				joinedTags += tag.Name + " "
			}

			numberSplit := strings.Split(number, ".")
			numberChapter, numberItem := -1, -1
			if len(numberSplit) == 2 {
				numberChapter, _ = strconv.Atoi(numberSplit[0])
				numberItem, _ = strconv.Atoi(numberSplit[1])
			}

			sdb.Songs[pageID] = Song{
				Id:         pageID,
				Title:      title,
				Number:     number,
				Tags:       joinedTags,
				Slug:       fmt.Sprintf("%s|%s|%s", slugify(title), number, slugify(joinedTags)),
				IsOrdinary: strings.Contains(joinedTags, ordinaryTag),

				numberChapter: numberChapter,
				numberItem:    numberItem,
			}
		}

		if result.HasMore {
			query.StartCursor = result.NextCursor
		} else {
			return nil
		}
	}
}

var numberQueryRegexp = regexp.MustCompile("^\\d")

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
	pagination := notionapi.Pagination{
		StartCursor: "",
		PageSize:    100,
	}
	for _, songID := range songIDs {
		if _, ok := sdb.LyricsBlocks[songID]; ok {
			continue
		}
		response, _ := sdb.client.Block.GetChildren(context.Background(), notionapi.BlockID(songID), &pagination)

		if response == nil {
			return nil
		}

		lyrics := make([]string, 0)

		for _, block := range response.Results {
			if block.GetType() != "paragraph" {
				continue
			}
			lyricsBlock := block.(*notionapi.ParagraphBlock)
			lyrics = append(lyrics, extractText(lyricsBlock.Paragraph.Text))
		}
		sdb.LyricsBlocks[songID] = lyrics
	}

	return nil
}

func (sdb SongsDB) GetLyrics(songID string, hints bool) ([]string, bool) {
	hasAllVerses := true

	if _, ok := sdb.LyricsBlocks[songID]; !ok {
		return nil, false
	}

	lyrics := make([]string, 0)
	number := sdb.Songs[songID].Number
	if hints && number != "" {
		lyrics = append(lyrics, "<hint>"+number+"</hint>")
	}
	for _, verse := range sdb.LyricsBlocks[songID] {
		verse := strings.ReplaceAll(verse, " * ", "\n")
		if verse != "" && !strings.HasPrefix(verse, "//") {
			lyrics = append(lyrics, verse)
		}
	}

	return lyrics, hasAllVerses
}
