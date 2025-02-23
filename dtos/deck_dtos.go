package dtos

import (
	"errors"
	"regexp"

	"github.com/hejmsdz/goslides/core"
)

type DeckRequest struct {
	Date          string     `json:"date"`
	Items         []DeckItem `json:"items"`
	Hints         bool       `json:"hints"`
	Ratio         string     `json:"ratio"`
	FontSize      int        `json:"fontSize"`
	VerticalAlign string     `json:"verticalAlign"`
	Format        string     `json:"format"`
	Contents      bool       `json:"contents"`
}

type DeckItem struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Contents []string `json:"contents"`
	Order    []int    `json:"order"`
}

type DeckResponse struct {
	URL      string              `json:"url"`
	Contents []core.ContentSlide `json:"contents"`
}

func NewDeckResponse(url string, contents []core.ContentSlide) DeckResponse {
	return DeckResponse{
		URL:      url,
		Contents: contents,
	}
}

var dateRegexp = regexp.MustCompile(`^20\d\d-[0-1]\d-[0-3]\d$`)

func (d DeckRequest) Validate() error {
	if !dateRegexp.MatchString(d.Date) {
		return errors.New("invalid date")
	}

	if len(d.Items) == 0 {
		return errors.New("items are empty")
	}

	if len(d.Items) > 100 {
		return errors.New("too many items")
	}

	for _, item := range d.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (i DeckItem) Validate() error {
	// TODO

	return nil
}

func (d DeckRequest) GetPageConfig() core.PageConfig {
	ratio := 4.0 / 3.0
	fontSize := 36

	if d.Ratio == "16:9" {
		ratio = 16.0 / 9.0
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
