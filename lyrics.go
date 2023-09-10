package main

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jomei/notionapi"
	"github.com/rainycape/unidecode"
)

const propertyNameTitle = "Tytuł"
const propertyNameNumber = "Numer"
const propertyNameTags = "Kategorie"
const ordinaryTag = "części stałe"

var nonAlpha = regexp.MustCompile(`[^a-zA-Z0-9\. ]+`)
var verseName = regexp.MustCompile(`^\[(\w+)\]\s+`)
var verseRef = regexp.MustCompile(`^%(\w+)$`)
var numberQueryRegexp = regexp.MustCompile(`^\d`)

const subtitleSeparator = " / "
const commentSymbol = "//"
const lineBreakSymbol = " * "

const hintStartTag = "<hint>"
const hintEndTag = "</hint>"

type Song struct {
	Id         string `json:"id"`
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle,omitempty"`
	Number     string `json:"number"`
	Tags       string `json:"-"`
	Slug       string `json:"slug"`
	IsOrdinary bool   `json:"isOrdinary,omitempty"`

	numberChapter int
	numberItem    int
	updatedAt     time.Time
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

func slugify(text string) string {
	text = strings.ToLower(text)
	text = unidecode.Unidecode(text)
	text = nonAlpha.ReplaceAllString(text, "")
	text = strings.Trim(text, " ")

	return text
}

func splitTitle(fullTitle string) (string, string) {
	titleSplit := strings.SplitN(fullTitle, subtitleSeparator, 2)

	if len(titleSplit) == 2 {
		return titleSplit[0], titleSplit[1]
	}

	return fullTitle, ""
}

func joinTags(tags notionapi.MultiSelectProperty) string {
	joinedTags := ""

	for _, tag := range tags.MultiSelect {
		joinedTags += tag.Name + " "
	}

	return joinedTags
}

func parseNumber(number string) (int, int) {
	numberChapter, numberItem := -1, -1
	numberSplit := strings.Split(number, ".")

	if len(numberSplit) == 2 {
		numberChapter, _ = strconv.Atoi(numberSplit[0])
		numberItem, _ = strconv.Atoi(numberSplit[1])
	}

	return numberChapter, numberItem
}

func (sdb *SongsDB) Initialize(authToken string, databaseId string) error {
	sdb.client = notionapi.NewClient(notionapi.Token(authToken))
	sdb.Songs = make(map[string]Song, 0)
	sdb.LyricsBlocks = make(map[string][]string, 0)

	query := notionapi.DatabaseQueryRequest{
		Sorts:       make([]notionapi.SortObject, 0),
		StartCursor: notionapi.Cursor(""),
		PageSize:    100,
	}

	for {
		result, err := sdb.client.Database.Query(
			context.Background(),
			notionapi.DatabaseID(databaseId),
			&query,
		)
		if err != nil {
			return err
		}

		for _, song := range result.Results {
			sdb.SaveSong(song)
		}

		if result.HasMore {
			query.StartCursor = result.NextCursor
		} else {
			return nil
		}
	}
}

func (sdb SongsDB) SaveSong(song notionapi.Page) {
	pageID := song.ID.String()
	fullTitle := extractText(song.Properties[propertyNameTitle].(*notionapi.TitleProperty).Title)
	number := extractText(song.Properties[propertyNameNumber].(*notionapi.RichTextProperty).RichText)

	title, subtitle := splitTitle(fullTitle)
	tags := joinTags(*song.Properties[propertyNameTags].(*notionapi.MultiSelectProperty))
	numberChapter, numberItem := parseNumber(number)

	sdb.Songs[pageID] = Song{
		Id:         pageID,
		Title:      title,
		Subtitle:   subtitle,
		Number:     number,
		Tags:       tags,
		Slug:       fmt.Sprintf("%s|%s|%s", slugify(fullTitle), number, slugify(tags)),
		IsOrdinary: strings.Contains(tags, ordinaryTag),

		numberChapter: numberChapter,
		numberItem:    numberItem,
		updatedAt:     song.LastEditedTime,
	}
}

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

	isNumberQuery := numberQueryRegexp.MatchString(query)
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
	completed := make(chan string, len(songIDs))
	for _, songID := range songIDs {
		go (func(songID string) {
			defer (func() {
				completed <- songID
			})()

			upToDateSong, err := sdb.client.Page.Get(context.Background(), notionapi.PageID(songID))
			if err != nil {
				return
			}
			isUpdated := upToDateSong.LastEditedTime.After(sdb.Songs[songID].updatedAt)
			_, hasLyrics := sdb.LyricsBlocks[songID]

			if !isUpdated && hasLyrics {
				return
			}

			if isUpdated {
				sdb.SaveSong(*upToDateSong)
			}

			response, _ := sdb.client.Block.GetChildren(
				context.Background(),
				notionapi.BlockID(songID),
				&pagination,
			)

			if response == nil {
				return
			}

			lyrics := make([]string, 0)

			for _, block := range response.Results {
				if block.GetType() != "paragraph" {
					continue
				}

				lyricsBlock := block.(*notionapi.ParagraphBlock)
				text := extractText(lyricsBlock.Paragraph.RichText)
				if text != "" {
					lyrics = append(lyrics, text)
				}
			}
			sdb.LyricsBlocks[songID] = lyrics
		})(songID)
	}

	for range songIDs {
		<-completed
	}

	return nil
}

type GetLyricsOptions struct {
	Hints bool
	Raw   bool
	Order []int
}

func (sdb SongsDB) GetLyrics(songID string, options GetLyricsOptions) ([]string, bool) {
	hasAllVerses := true

	verses, ok := sdb.LyricsBlocks[songID]
	if !ok {
		return nil, false
	}

	lyrics := make([]string, 0)
	namedVerses := make(map[string]string)
	song := sdb.Songs[songID]
	if options.Hints {
		hint := song.Number
		if utfTitle := []rune(song.Title); hint == "" && len(utfTitle) >= 2 {
			hint = string(utfTitle[0:2])
		}
		lyrics = append(lyrics, hintStartTag+hint+hintEndTag)
	}

	var order []int
	if options.Order == nil {
		order = make([]int, len(verses))
		for i := range verses {
			order[i] = i
		}
	} else {
		order = options.Order
	}

	for _, index := range order {
		if index >= len(verses) {
			continue
		}
		verse := verses[index]

		if !options.Raw {
			if strings.HasPrefix(verse, commentSymbol) {
				if options.Order == nil {
					// if order is not given, skip commented out verses
					continue
				}

				// but if order is given, the verse is included deliberately and comment symbol should be ignored
				verse = strings.TrimPrefix(verse, commentSymbol)
				verse = strings.TrimLeft(verse, " ")
			}

			if match := verseRef.FindStringSubmatch(verse); match != nil {
				name := match[1]
				if namedVerse, ok := namedVerses[name]; ok {
					verse = namedVerse
				}
			} else {
				verse = strings.ReplaceAll(verse, lineBreakSymbol, "\n")
				if match := verseName.FindStringSubmatch(verse); match != nil {
					name := match[1]
					verse = verse[len(match[0]):]
					namedVerses[name] = verse
				}
			}
		}

		lyrics = append(lyrics, verse)
	}

	return lyrics, hasAllVerses
}
