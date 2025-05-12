package services

import (
	"fmt"
	"strings"

	"github.com/hejmsdz/goslides/core"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
)

type DeckService struct {
	songs   *SongsService
	liturgy *LiturgyService
}

func NewDeckService(songs *SongsService, liturgy *LiturgyService) *DeckService {
	return &DeckService{songs: songs, liturgy: liturgy}
}

func (s *DeckService) GetPageConfig(d dtos.DeckRequest) core.PageConfig {
	ratio := 16.0 / 9.0
	fontSize := 52

	if d.Ratio == "4:3" {
		ratio = 4.0 / 3.0
	}

	if d.FontSize > 0 {
		fontSize = d.FontSize
	}

	pageHeight := 432.0
	pageWidth := pageHeight * ratio

	return core.PageConfig{
		PageWidth:     pageWidth,
		PageHeight:    pageHeight,
		Margin:        8,
		FontSize:      fontSize,
		HintFontSize:  fontSize * 2 / 3,
		LineSpacing:   1.3,
		Font:          "./fonts/source-sans-pro.ttf",
		VerticalAlign: d.VerticalAlign,
	}
}

const PSALM = "PSALM"
const ACCLAMATION = "ACCLAMATION"

func (s *DeckService) BuildTextSlides(d dtos.DeckRequest, user *models.User) ([][]string, bool) {
	hasLiturgy := false
	for _, item := range d.Items {
		if item.Type == PSALM || item.Type == ACCLAMATION {
			hasLiturgy = true
		}
	}

	var liturgy dtos.LiturgyItems
	liturgyOk := true
	if hasLiturgy {
		liturgy, liturgyOk = s.liturgy.GetDay(d.Date)
	}

	slides := make([][]string, 0)
	for _, item := range d.Items {
		if item.ID != "" {
			song, err := s.songs.GetSong(item.ID, user)
			if err != nil {
				return slides, false
			}
			lyrics := song.FormatLyrics(models.FormatLyricsOptions{Order: item.Order, Hints: d.Hints})
			slides = append(slides, lyrics)
		} else if item.Type == PSALM && liturgyOk {
			alleluiaticSuffix := ", albo: Alleluja"
			isAlleluiatic := strings.HasSuffix(liturgy.Psalm, alleluiaticSuffix)
			if isAlleluiatic {
				plainPsalm := strings.Replace(liturgy.Psalm, alleluiaticSuffix, "", 1)
				slides = append(slides, []string{plainPsalm, "Alleluja"})
			} else {
				slides = append(slides, []string{liturgy.Psalm})
			}
		} else if item.Type == ACCLAMATION && liturgyOk {
			fullAcclamation := fmt.Sprintf("%s\n\n%s\n\n%s",
				liturgy.Acclamation,
				liturgy.AcclamationVerse,
				liturgy.Acclamation)
			slides = append(slides, []string{fullAcclamation})
		} else if len(item.Contents) > 0 {
			slides = append(slides, item.Contents)
		}
	}

	return slides, true
}
